package handlers

import (
	"github.com/Tottitov/todo/internal/models"
	"github.com/Tottitov/todo/views/components"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type TodoHandler struct {
	DB *pgxpool.Pool
}

func (h TodoHandler) List(w http.ResponseWriter, r *http.Request) {
	log.Println("List handler hit:", r.Method, r.URL.Path)
	rows, err := h.DB.Query(r.Context(), "SELECT id, title FROM todos")

	if err != nil {
		log.Println("Query error:", err)
		http.Error(w, "error loading todos", http.StatusInternalServerError)
		log.Println("Query error:", err)
		return
	}

	defer rows.Close()
	var todos []models.Todo
	for rows.Next() {
		var t models.Todo
		if err := rows.Scan(&t.ID, &t.Title); err != nil {
			http.Error(w, "error scanning row", http.StatusInternalServerError)
			log.Println("Scan error:", err)
			return
		}
		todos = append(todos, t)
	}

	err = components.Layout(todos).Render(r.Context(), w)
	if err != nil {
		http.Error(w, "error rendering page", http.StatusInternalServerError)
		log.Println("Render error:", err)
	}
}

func (h TodoHandler) Create(w http.ResponseWriter, r *http.Request) {
	log.Println("Create handler hit:", r.Method, r.URL.Path)

	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}

	title := r.FormValue("title")

	if title == "" {
		http.Error(w, "missing title", http.StatusBadRequest)
		return
	}

	// Check for existing todo with same title (case-insensitive)
	var exists bool
	err := h.DB.QueryRow(r.Context(),
		"SELECT EXISTS (SELECT 1 FROM todos WHERE LOWER(title) = LOWER($1))", title,
	).Scan(&exists)

	if err != nil {
		log.Println("Check error:", err)
		http.Error(w, "could not check for duplicates", http.StatusInternalServerError)
		return
	}

	if exists {
		log.Println("Duplicate todo skipped:", title)
		w.WriteHeader(http.StatusConflict) // 409 Conflict
		return
	}

	// Insert if not a duplicate
	var id int
	err = h.DB.QueryRow(r.Context(),
		"INSERT INTO todos (title) VALUES ($1) RETURNING id", title,
	).Scan(&id)

	if err != nil {
		log.Println("Insert error:", err)
		http.Error(w, "could not insert todo", http.StatusInternalServerError)
		return
	}

	todo := models.Todo{ID: id, Title: title}
	err = components.TodoItem(todo).Render(r.Context(), w)

	if err != nil {
		log.Println("Render error:", err)
	}
}

// Delete removes a todo by its ID and returns an empty response so HTMX
// can replace the <li> with nothing (removing it from the DOM).
func (h TodoHandler) Delete(w http.ResponseWriter, r *http.Request) {
	log.Println("Delete handler hit:", r.Method, r.URL.Path)
	// 1. Pull the numeric ID out of the URL path (e.g. "/todos/42").
	idStr := strings.TrimPrefix(r.URL.Path, "/todos/")
	id, err := strconv.Atoi(idStr)

	if err != nil {
		http.Error(w, "invalid todo ID", http.StatusBadRequest)
		return
	}

	// 2. Run the DELETE query
	if _, err := h.DB.Exec(r.Context(),
		"DELETE FROM todos WHERE id = $1", id,
	); err != nil {
		http.Error(w, "error deleting todo", http.StatusInternalServerError)
		return
	}
	// 3. Tell HTMX "all done" with a 200 OK status so HTMX will properly
	// process the response and remove the element from the DOM.
	w.WriteHeader(http.StatusOK)
}
