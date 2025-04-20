package models

type Todo struct {
	ID        int
	Title     string
	Completed bool
}

func NewTodo(title string) Todo {
	return Todo{
		ID:        0,
		Title:     title,
		Completed: false,
	}
}
