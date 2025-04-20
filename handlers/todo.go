package handlers

import (
	"net/http"
	"strconv"

	"github.com/Tottitov/todo/components"
	"github.com/Tottitov/todo/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TodoHandler struct {
	DB *pgxpool.Pool
}

// List handles displaying all todos with optional filters: all, active, completed
func (h *TodoHandler) List(w http.ResponseWriter, r *http.Request) {
	filter := r.URL.Query().Get("filter") // "active", "completed", or "" (default = all)

	query := "SELECT id, title, completed FROM todos"
	switch filter {
	case "active":
		query += " WHERE completed = false"
	case "completed":
		query += " WHERE completed = true"
	}
	query += " ORDER BY id"

	rows, err := h.DB.Query(r.Context(), query)
	if err != nil {
		http.Error(w, "Failed to fetch todos", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var todos []models.Todo
	for rows.Next() {
		var todo models.Todo
		if err := rows.Scan(&todo.ID, &todo.Title, &todo.Completed); err != nil {
			http.Error(w, "Failed to scan todo", http.StatusInternalServerError)
			return
		}
		todos = append(todos, todo)
	}

	w.Header().Set("Content-Type", "text/html")
	if err := components.TodoList(todos, filter).Render(r.Context(), w); err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
	}
}

// Create handles adding a new todo
func (h *TodoHandler) Create(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	title := r.PostForm.Get("title")
	if title == "" {
		http.Error(w, "Todo title cannot be empty", http.StatusBadRequest)
		return
	}

	var id int
	err := h.DB.QueryRow(r.Context(),
		"INSERT INTO todos (title, completed) VALUES ($1, $2) RETURNING id",
		title, false).Scan(&id)
	if err != nil {
		http.Error(w, "Failed to create todo", http.StatusInternalServerError)
		return
	}

	if r.Header.Get("HX-Request") == "true" {
		rows, err := h.DB.Query(r.Context(), "SELECT id, title, completed FROM todos ORDER BY id")
		if err != nil {
			http.Error(w, "failed to reload todos", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var todos []models.Todo
		for rows.Next() {
			var t models.Todo
			if err := rows.Scan(&t.ID, &t.Title, &t.Completed); err != nil {
				http.Error(w, "error reading todos", http.StatusInternalServerError)
				return
			}
			todos = append(todos, t)
		}

		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusCreated)
		components.TodoList(todos, "").Render(r.Context(), w)
	} else {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

// Update handles updating a todo's title
func (h *TodoHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID format", http.StatusBadRequest)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	title := r.PostForm.Get("title")
	if title == "" {
		http.Error(w, "Todo title cannot be empty", http.StatusBadRequest)
		return
	}

	_, err = h.DB.Exec(r.Context(), "UPDATE todos SET title = $1 WHERE id = $2", title, id)
	if err != nil {
		http.Error(w, "Failed to update todo", http.StatusInternalServerError)
		return
	}

	if r.Header.Get("HX-Request") == "true" {
		todo := models.Todo{ID: id, Title: title}
		err = h.DB.QueryRow(r.Context(), "SELECT completed FROM todos WHERE id = $1", id).Scan(&todo.Completed)
		if err != nil {
			http.Error(w, "Failed to get todo status", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html")
		components.TodoItem(todo).Render(r.Context(), w)
	} else {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

// Delete handles removing a todo
func (h *TodoHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID format", http.StatusBadRequest)
		return
	}

	_, err = h.DB.Exec(r.Context(), "DELETE FROM todos WHERE id = $1", id)
	if err != nil {
		http.Error(w, "Failed to delete todo", http.StatusInternalServerError)
		return
	}

	if r.Header.Get("HX-Request") == "true" {
		rows, err := h.DB.Query(r.Context(), "SELECT id, title, completed FROM todos ORDER BY id")
		if err != nil {
			http.Error(w, "error reloading todos", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var todos []models.Todo
		for rows.Next() {
			var t models.Todo
			if err := rows.Scan(&t.ID, &t.Title, &t.Completed); err != nil {
				http.Error(w, "error reading todos", http.StatusInternalServerError)
				return
			}
			todos = append(todos, t)
		}

		w.Header().Set("Content-Type", "text/html")
		components.TodoList(todos, "").Render(r.Context(), w)
	} else {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

// ToggleComplete handles toggling a todo's completion status and updates the entire list
func (h *TodoHandler) ToggleComplete(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID format", http.StatusBadRequest)
		return
	}

	var completed bool
	err = h.DB.QueryRow(r.Context(), "SELECT completed FROM todos WHERE id = $1", id).Scan(&completed)
	if err != nil {
		http.Error(w, "Todo not found", http.StatusNotFound)
		return
	}

	_, err = h.DB.Exec(r.Context(), "UPDATE todos SET completed = $1 WHERE id = $2", !completed, id)
	if err != nil {
		http.Error(w, "Failed to update todo", http.StatusInternalServerError)
		return
	}

	if r.Header.Get("HX-Request") == "true" {
		rows, err := h.DB.Query(r.Context(), "SELECT id, title, completed FROM todos ORDER BY id")
		if err != nil {
			http.Error(w, "failed to reload todos", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var todos []models.Todo
		for rows.Next() {
			var t models.Todo
			if err := rows.Scan(&t.ID, &t.Title, &t.Completed); err != nil {
				http.Error(w, "error reading todos", http.StatusInternalServerError)
				return
			}
			todos = append(todos, t)
		}

		w.Header().Set("Content-Type", "text/html")
		components.TodoList(todos, "").Render(r.Context(), w)
	} else {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

// DeleteCompleted removes all completed todos
func (h TodoHandler) DeleteCompleted(w http.ResponseWriter, r *http.Request) {
	if _, err := h.DB.Exec(r.Context(), "DELETE FROM todos WHERE completed = true"); err != nil {
		http.Error(w, "error clearing completed todos", http.StatusInternalServerError)
		return
	}

	rows, err := h.DB.Query(r.Context(), "SELECT id, title, completed FROM todos ORDER BY id")
	if err != nil {
		http.Error(w, "error reloading todos", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var todos []models.Todo
	for rows.Next() {
		var t models.Todo
		if err := rows.Scan(&t.ID, &t.Title, &t.Completed); err != nil {
			http.Error(w, "error reading todos", http.StatusInternalServerError)
			return
		}
		todos = append(todos, t)
	}

	w.Header().Set("Content-Type", "text/html")
	components.TodoList(todos, "").Render(r.Context(), w)
}
