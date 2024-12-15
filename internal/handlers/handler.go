package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"final-project/internal/config"
	"final-project/internal/database"
	"final-project/internal/models"
	"final-project/internal/services"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"os"
	"strconv"
	"time"
)

func SignInHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		responseError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var authData models.AuthData
	var buf bytes.Buffer

	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		responseError(w, "failed to read the request body", http.StatusBadRequest)
		return
	}

	if err = json.Unmarshal(buf.Bytes(), &authData); err != nil {
		responseError(w, "failed to deserialize JSON", http.StatusBadRequest)
		return
	}

	storedPassword := os.Getenv("TODO_PASSWORD")
	if storedPassword == "" || authData.Password != storedPassword {
		responseError(w, "invalid password", http.StatusUnauthorized)
		return
	}

	claims := jwt.MapClaims{
		"pass": authData.Password,
		"exp":  jwt.NewNumericDate(time.Now().Add(8 * time.Hour)),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(config.JWTKey)
	if err != nil {
		responseError(w, "failed to generate token", http.StatusInternalServerError)
		return
	}
	response := models.TokenResponse{Token: tokenString}
	writeJSON(w, response, http.StatusOK)
}

func NextDateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	nowStr := r.FormValue("now")
	dateStr := r.FormValue("date")
	repeat := r.FormValue("repeat")

	if nowStr == "" || dateStr == "" || repeat == "" {
		http.Error(w, "missing required parameters: now, date or repeat", http.StatusBadRequest)
		return
	}

	now, err := time.Parse("20060102", nowStr)
	if err != nil {
		http.Error(w, "invalid format for 'now': "+nowStr, http.StatusBadRequest)
		return
	}

	nextDate, err := services.NextDate(now, dateStr, repeat)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(nextDate))
}

func AddTaskHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	var task models.Task
	var buf bytes.Buffer

	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		responseError(w, "failed to read the request body", http.StatusBadRequest)
		return
	}

	if err = json.Unmarshal(buf.Bytes(), &task); err != nil {
		responseError(w, "failed to deserialize JSON", http.StatusBadRequest)
		return
	}

	if err = services.ProcessTask(&task); err != nil {
		responseError(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err = database.AddTask(db, &task); err != nil {
		responseError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := models.Response{ID: task.ID}
	writeJSON(w, response, http.StatusOK)
}

func responseError(w http.ResponseWriter, message string, statusCode int) {
	response := models.Response{Error: message}
	writeJSON(w, response, statusCode)
}

func writeJSON(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(statusCode)

	jsonResponse, err := json.Marshal(data)
	if err != nil {
		http.Error(w, `{"error":"failed to serialize JSON"}`, http.StatusInternalServerError)
		return
	}
	w.Write(jsonResponse)
}

func GetTasksHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	if r.Method != http.MethodGet {
		responseError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	search := r.URL.Query().Get("search")

	tasks, err := database.GetTasksWithSearch(db, search)
	if err != nil {
		responseError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := models.TaskListResponse{Tasks: tasks}
	writeJSON(w, response, http.StatusOK)
}

func GetTaskHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	id := r.URL.Query().Get("id")
	if id == "" {
		responseError(w, "task ID is required", http.StatusBadRequest)
		return
	}

	parsedId, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		responseError(w, "invalid task ID", http.StatusBadRequest)
		return
	}

	task, err := database.GetTask(db, parsedId)
	if err != nil {
		responseError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := models.Task{
		ID:      task.ID,
		Date:    task.Date,
		Title:   task.Title,
		Comment: task.Comment,
		Repeat:  task.Repeat,
	}
	writeJSON(w, response, http.StatusOK)
}

func UpdateTaskHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	var task models.Task
	var buf bytes.Buffer

	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		responseError(w, "failed to read the request body", http.StatusBadRequest)
		return
	}

	if err = json.Unmarshal(buf.Bytes(), &task); err != nil {
		responseError(w, "failed to deserialize JSON", http.StatusBadRequest)
		return
	}

	if err = services.ProcessTask(&task); err != nil {
		responseError(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = database.UpdateTask(db, &task)
	if err != nil {
		responseError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, map[string]interface{}{}, http.StatusOK)
}

func DeleteTaskHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	id := r.URL.Query().Get("id")
	if id == "" {
		responseError(w, "task ID is required", http.StatusBadRequest)
		return
	}

	parsedId, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		responseError(w, "invalid task ID", http.StatusBadRequest)
		return
	}

	err = database.DeleteTask(db, parsedId)
	if err != nil {
		responseError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, map[string]interface{}{}, http.StatusOK)
}

func TaskDoneHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	if r.Method != http.MethodPost {
		responseError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		responseError(w, "task ID is required", http.StatusBadRequest)
		return
	}

	parsedId, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		responseError(w, "invalid task ID", http.StatusBadRequest)
		return
	}

	task, err := database.GetTask(db, parsedId)
	if err != nil {
		responseError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if task.Repeat == "" {
		err = database.DeleteTask(db, parsedId)
		if err != nil {
			responseError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, map[string]interface{}{}, http.StatusOK)
		return
	}

	parsedDate, err := time.Parse("20060102", task.Date)
	if err != nil {
		responseError(w, "invalid date format, expected YYYYMMDD", http.StatusBadRequest)
	}

	nextDate, err := services.NextDate(parsedDate, task.Date, task.Repeat)
	if err != nil {
		responseError(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = database.UpdateTaskDate(db, parsedId, nextDate)
	if err != nil {
		responseError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, map[string]interface{}{}, http.StatusOK)
}
