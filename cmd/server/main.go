package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/Tottitov/todo/internal/handlers"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	dbUrl := os.Getenv("DATABASE_URL")
	if dbUrl == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	// 2) Connect to Postgres
	dbpool, err := pgxpool.New(context.Background(), dbUrl)
	if err != nil {
		log.Fatal("Unable to connect to database:", err)
	}
	defer dbpool.Close()

	// 3) Set up HTTP handlers
	todoHandler := handlers.TodoHandler{DB: dbpool}
	http.HandleFunc("/", todoHandler.List)
	http.HandleFunc("/todos", todoHandler.Create)
	http.HandleFunc("/todos/", todoHandler.Delete)

	// 4) Start server
	log.Println("Server running on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
