// Package handlers provides HTTP request handlers for the todo application.
// It manages all CRUD operations for todos and handles HTMX-based partial page updates.
package handlers

import (
	"context"
	"net/http"

	"github.com/Tottitov/todo/components"
	"github.com/Tottitov/todo/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

// TodoHandler encapsulates the dependencies and methods needed to handle todo-related HTTP requests.
// It maintains a connection pool to the PostgreSQL database for persistent storage.
type TodoHandler struct {
	DB *pgxpool.Pool // Connection pool for PostgreSQL database
}

// List handles GET requests to display all todos.
// It supports filtering todos by their completion status using the 'filter' query parameter.
// The handler renders either all todos, active todos, or completed todos based on the filter.
// It also calculates and passes the count of active todos to the template.
func (h *TodoHandler) List(w http.ResponseWriter, r *http.Request) {
	// Extract the filter parameter from URL query string (all, active, completed)
	filter := r.URL.Query().Get("filter")

	// Fetch all todos from the database to calculate counts and apply filters
	allTodos, err := h.fetchAllTodos(r.Context())
	if err != nil {
		sendError(w, "Failed to fetch todos", http.StatusInternalServerError)
		return
	}

	// Calculate the number of active (incomplete) todos for the counter
	activeCount := countActive(allTodos)

	// Apply the filter to show only relevant todos
	displayTodos := filterTodos(allTodos, filter)

	// Set content type to HTML and render the todo list component
	setHTMLHeader(w)
	components.TodoList(displayTodos, filter, activeCount).Render(r.Context(), w)
}

// Create handles POST requests to add a new todo.
// It expects a 'title' field in the form data.
// After creating the todo, it returns an updated todo list component for HTMX to swap.
func (h *TodoHandler) Create(w http.ResponseWriter, r *http.Request) {
	// Parse the form data from the request
	if err := r.ParseForm(); err != nil {
		sendError(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	// Extract and validate the todo title
	title := r.PostForm.Get("title")
	if title == "" {
		sendError(w, "Todo title cannot be empty", http.StatusBadRequest)
		return
	}

	// Insert the new todo into the database (defaults to not completed)
	_, err := h.DB.Exec(r.Context(),
		"INSERT INTO todos (title, completed) VALUES ($1, $2)",
		title, false)
	if err != nil {
		sendError(w, "Failed to create todo", http.StatusInternalServerError)
		return
	}

	// Fetch the updated list of todos to reflect the new addition
	allTodos, err := h.fetchAllTodos(r.Context())
	if err != nil {
		sendError(w, "Failed to reload todos", http.StatusInternalServerError)
		return
	}

	// Apply any active filters and calculate the new active count
	displayTodos := filterTodos(allTodos, "")
	activeCount := countActive(allTodos)

	// Return the updated todo list component with status 201 Created
	setHTMLHeader(w)
	w.WriteHeader(http.StatusCreated)
	components.TodoListContent(displayTodos, "", activeCount).Render(r.Context(), w)
}

// Edit handles GET requests to show the edit form for a specific todo.
// It's triggered by double-clicking a todo item and returns the edit form component.
func (h *TodoHandler) Edit(w http.ResponseWriter, r *http.Request) {
	// Extract and validate the todo ID from the URL
	id, err := parseID(r)
	if err != nil {
		sendError(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	// Fetch the todo from the database
	var todo models.Todo
	err = h.DB.QueryRow(r.Context(),
		"SELECT id, title, completed FROM todos WHERE id = $1",
		id,
	).Scan(&todo.ID, &todo.Title, &todo.Completed)
	if err != nil {
		sendError(w, "Todo not found", http.StatusNotFound)
		return
	}

	// Render the edit form component for the todo
	setHTMLHeader(w)
	components.TodoEdit(todo).Render(r.Context(), w)
}

// Update handles PATCH requests to modify a todo's title.
// It's triggered by the edit form submission or when the input loses focus.
// Returns either the updated todo item component or redirects to the home page.
func (h *TodoHandler) Update(w http.ResponseWriter, r *http.Request) {
	// Extract and validate the todo ID from the URL
	id, err := parseID(r)
	if err != nil {
		sendError(w, "Invalid ID format", http.StatusBadRequest)
		return
	}

	// Parse the form data containing the updated title
	if err := r.ParseForm(); err != nil {
		sendError(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	// Validate the new title
	title := r.PostForm.Get("title")
	if title == "" {
		sendError(w, "Todo title cannot be empty", http.StatusBadRequest)
		return
	}

	// Update the todo's title in the database
	_, err = h.DB.Exec(r.Context(), "UPDATE todos SET title = $1 WHERE id = $2", title, id)
	if err != nil {
		sendError(w, "Failed to update todo", http.StatusInternalServerError)
		return
	}

	// Handle HTMX requests differently from regular form submissions
	if r.Header.Get("HX-Request") == "true" {
		// For HTMX requests, return the updated todo item component
		todo := models.Todo{ID: id, Title: title}
		err = h.DB.QueryRow(r.Context(),
			"SELECT completed FROM todos WHERE id = $1", id,
		).Scan(&todo.Completed)
		if err != nil {
			sendError(w, "Failed to get todo status", http.StatusInternalServerError)
			return
		}
		setHTMLHeader(w)
		components.TodoItem(todo).Render(r.Context(), w)
		return
	}
	// For regular requests, redirect to the home page
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// Delete handles DELETE requests to remove a specific todo.
// After deletion, it returns the updated todo list component.
func (h *TodoHandler) Delete(w http.ResponseWriter, r *http.Request) {
	// Extract and validate the todo ID from the URL
	id, err := parseID(r)
	if err != nil {
		sendError(w, "Invalid ID format", http.StatusBadRequest)
		return
	}

	// Delete the todo from the database
	_, err = h.DB.Exec(r.Context(), "DELETE FROM todos WHERE id = $1", id)
	if err != nil {
		sendError(w, "Failed to delete todo", http.StatusInternalServerError)
		return
	}

	// Fetch the updated list of todos
	allTodos, err := h.fetchAllTodos(r.Context())
	if err != nil {
		sendError(w, "Failed to reload todos", http.StatusInternalServerError)
		return
	}

	// Apply any active filters and calculate the new active count
	displayTodos := filterTodos(allTodos, "")
	activeCount := countActive(allTodos)

	// Return the updated todo list component
	setHTMLHeader(w)
	components.TodoListContent(displayTodos, "", activeCount).Render(r.Context(), w)
}

// ToggleComplete handles POST requests to toggle a todo's completion status.
// After toggling, it returns the updated todo list component.
func (h *TodoHandler) ToggleComplete(w http.ResponseWriter, r *http.Request) {
	// Extract and validate the todo ID from the URL
	id, err := parseID(r)
	if err != nil {
		sendError(w, "Invalid ID format", http.StatusBadRequest)
		return
	}

	// Fetch the current completion status
	var completed bool
	err = h.DB.QueryRow(r.Context(),
		"SELECT completed FROM todos WHERE id = $1", id,
	).Scan(&completed)
	if err != nil {
		sendError(w, "Todo not found", http.StatusNotFound)
		return
	}

	// Toggle the completion status in the database
	_, err = h.DB.Exec(r.Context(),
		"UPDATE todos SET completed = $1 WHERE id = $2", !completed, id,
	)
	if err != nil {
		sendError(w, "Failed to update todo", http.StatusInternalServerError)
		return
	}

	// Fetch the updated list of todos
	allTodos, err := h.fetchAllTodos(r.Context())
	if err != nil {
		sendError(w, "Failed to reload todos", http.StatusInternalServerError)
		return
	}

	// Apply any active filters and calculate the new active count
	displayTodos := filterTodos(allTodos, "")
	activeCount := countActive(allTodos)

	// Return the updated todo list component
	setHTMLHeader(w)
	components.TodoListContent(displayTodos, "", activeCount).Render(r.Context(), w)
}

// DeleteCompleted handles POST requests to remove all completed todos.
// After deletion, it returns the updated todo list component.
func (h *TodoHandler) DeleteCompleted(w http.ResponseWriter, r *http.Request) {
	// Delete all completed todos from the database
	_, err := h.DB.Exec(r.Context(), "DELETE FROM todos WHERE completed = true")
	if err != nil {
		sendError(w, "Error clearing completed todos", http.StatusInternalServerError)
		return
	}

	// Fetch the updated list of todos
	allTodos, err := h.fetchAllTodos(r.Context())
	if err != nil {
		sendError(w, "Failed to reload todos", http.StatusInternalServerError)
		return
	}

	// Apply any active filters and calculate the new active count
	displayTodos := filterTodos(allTodos, "")
	activeCount := countActive(allTodos)

	// Return the updated todo list component
	setHTMLHeader(w)
	components.TodoListContent(displayTodos, "", activeCount).Render(r.Context(), w)
}

// fetchAllTodos is a helper function that retrieves all todos from the database.
// Todos are ordered by their ID to maintain a consistent display order.
func (h *TodoHandler) fetchAllTodos(ctx context.Context) ([]models.Todo, error) {
	// Query all todos ordered by ID
	query := "SELECT id, title, completed FROM todos ORDER BY id"

	rows, err := h.DB.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Scan rows into todo structs
	var todos []models.Todo
	for rows.Next() {
		var t models.Todo
		if err := rows.Scan(&t.ID, &t.Title, &t.Completed); err != nil {
			return nil, err
		}
		todos = append(todos, t)
	}
	return todos, nil
}

// filterTodos is a helper function that filters todos based on their completion status.
// It supports three filter modes: all (empty string), active, and completed.
func filterTodos(todos []models.Todo, filter string) []models.Todo {
	var filtered []models.Todo
	for _, todo := range todos {
		switch filter {
		case "active":
			// Only include incomplete todos
			if !todo.Completed {
				filtered = append(filtered, todo)
			}
		case "completed":
			// Only include completed todos
			if todo.Completed {
				filtered = append(filtered, todo)
			}
		default:
			// Include all todos when no filter is specified
			filtered = append(filtered, todo)
		}
	}
	return filtered
}

// countActive is a helper function that counts the number of incomplete todos.
// This count is displayed in the UI as "X items left".
func countActive(todos []models.Todo) int {
	count := 0
	for _, todo := range todos {
		if !todo.Completed {
			count++
		}
	}
	return count
}
