package main

import (
	"context"
	"github.com/Tottitov/todo/internal/handlers"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"net/http"
)

func main() {
	dbUrl := "postgres://todo_user:Burp123!@localhost:5432/todo?sslmode=disable"

	// 2) Run all pending migrations from ./migrations
	m, err := migrate.New("file://../../migrations", dbUrl)
	if err != nil {
		log.Fatal("migrate.New:", err)
	}

	// If there’s nothing to do, ErrNoChange is fine
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatal("migrate.Up:", err)
	}

	// Connect to Postgres
	dbpool, err := pgxpool.New(context.Background(), dbUrl)
	if err != nil {
		log.Fatal("Unable to connect to database:", err)
	}
	defer dbpool.Close()

	// Wire up the handlers
	todoHandler := handlers.TodoHandler{DB: dbpool}
	http.HandleFunc("/", todoHandler.List)         // GET
	http.HandleFunc("/todos", todoHandler.Create)  // POST
	http.HandleFunc("/todos/", todoHandler.Delete) // DELETE

	// Start server
	log.Println("Server running at http://localhost:8080")
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}
