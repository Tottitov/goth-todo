package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/Tottitov/todo/internal/database"
	"github.com/Tottitov/todo/internal/models"
	supa "github.com/nedpals/supabase-go"
)

// TodoHandler handles todo-related HTTP requests
type TodoHandler struct {
	Client *supa.Client
}

// Create handles the creation of a new todo
func (h *TodoHandler) Create(w http.ResponseWriter, r *http.Request) {
	// Parse the incoming todo from the request body
	var newTodo models.Todo
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&newTodo); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Create the todo in the database
	createdTodo, err := h.DB.CreateTodo(newTodo)
	if err != nil {
		http.Error(w, "Failed to create todo", http.StatusInternalServerError)
		return
	}

	// Respond with the created todo
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdTodo)
}

// Delete handles the deletion of a todo
func (h *TodoHandler) Delete(w http.ResponseWriter, r *http.Request) {
	// Extract the todo ID from the URL
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 3 {
		http.Error(w, "Invalid todo ID", http.StatusBadRequest)
		return
	}
	todoID := parts[2]

	// Delete the todo from the database
	err := h.DB.DeleteTodo(todoID)
	if err != nil {
		http.Error(w, "Failed to delete todo", http.StatusInternalServerError)
		return
	}

	// Respond with a success status
	w.WriteHeader(http.StatusNoContent)
}

// List retrieves and returns all todos
func (h *TodoHandler) List(w http.ResponseWriter, r *http.Request) {
	// Fetch todos from the database
	todos, err := h.DB.ListTodos()
	if err != nil {
		http.Error(w, "Failed to fetch todos", http.StatusInternalServerError)
		return
	}

	// Respond with the list of todos
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(todos)
}
