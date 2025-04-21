package main

import (
	"context"
	"github.com/Tottitov/todo/handlers"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"net/http"
	"os"
)

func main() {
	dbUrl := os.Getenv("DATABASE_URL")
	dbPool, err := pgxpool.New(context.Background(), dbUrl)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer dbPool.Close()

	todoHandler := &handlers.TodoHandler{DB: dbPool}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /", todoHandler.List)
	mux.HandleFunc("POST /todos", todoHandler.Create)
	mux.HandleFunc("GET /todos/{id}/edit", todoHandler.Edit)
	mux.HandleFunc("PATCH /todos/{id}", todoHandler.Update)
	mux.HandleFunc("DELETE /todos/{id}", todoHandler.Delete)
	mux.HandleFunc("POST /todos/{id}/toggle", todoHandler.ToggleComplete)
	mux.HandleFunc("POST /todos/completed", func(w http.ResponseWriter, r *http.Request) {
		if r.FormValue("_method") == "DELETE" {
			todoHandler.DeleteCompleted(w, r)
			return
		}
		http.NotFound(w, r)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Server running at http://localhost:%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}
