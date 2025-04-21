package handlers

import (
	"context"
	"net/http"

	"github.com/Tottitov/todo/components"
	"github.com/Tottitov/todo/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TodoHandler struct {
	DB *pgxpool.Pool
}

func (h *TodoHandler) List(w http.ResponseWriter, r *http.Request) {
	filter := r.URL.Query().Get("filter")
	allTodos, err := h.fetchAllTodos(r.Context())
	if err != nil {
		sendError(w, "Failed to fetch todos", http.StatusInternalServerError)
		return
	}

	activeCount := countActive(allTodos)

	displayTodos := filterTodos(allTodos, filter)

	setHTMLHeader(w)
	components.TodoList(displayTodos, filter, activeCount).Render(r.Context(), w)
}

func (h *TodoHandler) Create(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		sendError(w, "Failed to parse form", http.StatusBadRequest)
		return
	}
	title := r.PostForm.Get("title")
	if title == "" {
		sendError(w, "Todo title cannot be empty", http.StatusBadRequest)
		return
	}

	_, err := h.DB.Exec(r.Context(),
		"INSERT INTO todos (title, completed) VALUES ($1, $2)",
		title, false)
	if err != nil {
		sendError(w, "Failed to create todo", http.StatusInternalServerError)
		return
	}

	allTodos, err := h.fetchAllTodos(r.Context())
	if err != nil {
		sendError(w, "Failed to reload todos", http.StatusInternalServerError)
		return
	}
	displayTodos := filterTodos(allTodos, "")
	activeCount := countActive(allTodos)

	setHTMLHeader(w)
	w.WriteHeader(http.StatusCreated)
	components.TodoListContent(displayTodos, "", activeCount).Render(r.Context(), w)
}

func (h *TodoHandler) Edit(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		sendError(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var todo models.Todo
	err = h.DB.QueryRow(r.Context(),
		"SELECT id, title, completed FROM todos WHERE id = $1",
		id,
	).Scan(&todo.ID, &todo.Title, &todo.Completed)
	if err != nil {
		sendError(w, "Todo not found", http.StatusNotFound)
		return
	}

	setHTMLHeader(w)
	components.TodoEdit(todo).Render(r.Context(), w)
}

func (h *TodoHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		sendError(w, "Invalid ID format", http.StatusBadRequest)
		return
	}

	if err := r.ParseForm(); err != nil {
		sendError(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	title := r.PostForm.Get("title")
	if title == "" {
		sendError(w, "Todo title cannot be empty", http.StatusBadRequest)
		return
	}

	_, err = h.DB.Exec(r.Context(), "UPDATE todos SET title = $1 WHERE id = $2", title, id)
	if err != nil {
		sendError(w, "Failed to update todo", http.StatusInternalServerError)
		return
	}

	if r.Header.Get("HX-Request") == "true" {
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
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *TodoHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		sendError(w, "Invalid ID format", http.StatusBadRequest)
		return
	}

	_, err = h.DB.Exec(r.Context(), "DELETE FROM todos WHERE id = $1", id)
	if err != nil {
		sendError(w, "Failed to delete todo", http.StatusInternalServerError)
		return
	}

	allTodos, err := h.fetchAllTodos(r.Context())
	if err != nil {
		sendError(w, "Failed to reload todos", http.StatusInternalServerError)
		return
	}
	displayTodos := filterTodos(allTodos, "")
	activeCount := countActive(allTodos)

	setHTMLHeader(w)
	components.TodoListContent(displayTodos, "", activeCount).Render(r.Context(), w)
}

func (h *TodoHandler) ToggleComplete(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		sendError(w, "Invalid ID format", http.StatusBadRequest)
		return
	}

	var completed bool
	err = h.DB.QueryRow(r.Context(),
		"SELECT completed FROM todos WHERE id = $1", id,
	).Scan(&completed)
	if err != nil {
		sendError(w, "Todo not found", http.StatusNotFound)
		return
	}

	_, err = h.DB.Exec(r.Context(),
		"UPDATE todos SET completed = $1 WHERE id = $2", !completed, id,
	)
	if err != nil {
		sendError(w, "Failed to update todo", http.StatusInternalServerError)
		return
	}

	allTodos, err := h.fetchAllTodos(r.Context())
	if err != nil {
		sendError(w, "Failed to reload todos", http.StatusInternalServerError)
		return
	}
	displayTodos := filterTodos(allTodos, "")
	activeCount := countActive(allTodos)

	setHTMLHeader(w)
	components.TodoListContent(displayTodos, "", activeCount).Render(r.Context(), w)
}

func (h *TodoHandler) DeleteCompleted(w http.ResponseWriter, r *http.Request) {
	_, err := h.DB.Exec(r.Context(), "DELETE FROM todos WHERE completed = true")
	if err != nil {
		sendError(w, "Error clearing completed todos", http.StatusInternalServerError)
		return
	}

	allTodos, err := h.fetchAllTodos(r.Context())
	if err != nil {
		sendError(w, "Failed to reload todos", http.StatusInternalServerError)
		return
	}
	displayTodos := filterTodos(allTodos, "")
	activeCount := countActive(allTodos)

	setHTMLHeader(w)
	components.TodoListContent(displayTodos, "", activeCount).Render(r.Context(), w)
}

func (h *TodoHandler) fetchAllTodos(ctx context.Context) ([]models.Todo, error) {
	query := "SELECT id, title, completed FROM todos ORDER BY id"

	rows, err := h.DB.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

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

func filterTodos(todos []models.Todo, filter string) []models.Todo {
	var filtered []models.Todo
	for _, todo := range todos {
		switch filter {
		case "active":
			if !todo.Completed {
				filtered = append(filtered, todo)
			}
		case "completed":
			if todo.Completed {
				filtered = append(filtered, todo)
			}
		default:
			filtered = append(filtered, todo)
		}
	}
	return filtered
}

func countActive(todos []models.Todo) int {
	count := 0
	for _, todo := range todos {
		if !todo.Completed {
			count++
		}
	}
	return count
}
