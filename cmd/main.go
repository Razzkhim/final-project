package main

import (
	"final-project/internal/database"
	"final-project/internal/handlers"
	"final-project/internal/middleware"
	"final-project/tests"
	"fmt"
	"github.com/joho/godotenv"
	"log"
	_ "modernc.org/sqlite"
	"net/http"
	"os"
	"strconv"
)

func main() {

	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
	}

	port, err := strconv.Atoi(os.Getenv("TODO_PORT"))
	if err != nil {
		port = tests.Port
	}

	db, err := database.InitDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	http.Handle("/", http.FileServer(http.Dir("web")))
	http.HandleFunc("/api/signin", handlers.SignInHandler)
	http.HandleFunc("/api/nextdate", handlers.NextDateHandler)
	http.HandleFunc("/api/task", middleware.Auth(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handlers.GetTaskHandler(w, r, db)
		case http.MethodPost:
			handlers.AddTaskHandler(w, r, db)
		case http.MethodPut:
			handlers.UpdateTaskHandler(w, r, db)
		case http.MethodDelete:
			handlers.DeleteTaskHandler(w, r, db)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))
	http.HandleFunc("/api/tasks", func(w http.ResponseWriter, r *http.Request) {
		handlers.GetTasksHandler(w, r, db)
	})
	http.HandleFunc("/api/task/done", func(w http.ResponseWriter, r *http.Request) {
		handlers.TaskDoneHandler(w, r, db)
	})

	if err = http.ListenAndServe(":"+strconv.Itoa(port), nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

}
