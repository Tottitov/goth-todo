package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/Tottitov/todo/handlers"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	// DATABASE_URL should be something like "postgres://user:pass@host:port/dbname"
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	dbPool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		log.Fatalf("unable to connect to database: %v", err)
	}
	defer dbPool.Close()

	todoHandler := &handlers.TodoHandler{DB: dbPool}

	r := chi.NewRouter()

	// List & create
	r.Get("/", todoHandler.List)
	r.Post("/todos", todoHandler.Create)

	// Inline‑edit form
	r.Get("/todos/{id}/edit", todoHandler.Edit)

	// Update, delete, toggle complete
	r.Patch("/todos/{id}", todoHandler.Update)
	r.Delete("/todos/{id}", todoHandler.Delete)
	r.Post("/todos/{id}/toggle", todoHandler.ToggleComplete)

	// Bulk delete completed
	r.Post("/todos/completed", func(w http.ResponseWriter, r *http.Request) {
		// expecting a form _method=DELETE
		if r.FormValue("_method") == "DELETE" {
			todoHandler.DeleteCompleted(w, r)
			return
		}
		http.NotFound(w, r)
	})

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Server running at http://localhost:%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
