package models

type Task struct {
	ID      int64  `json:"id,string"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

type Response struct {
	ID    int64  `json:"id,omitempty,string"`
	Error string `json:"error,omitempty"`
}

type TaskListResponse struct {
	Tasks []Task `json:"tasks"`
}

type AuthData struct {
	Password string `json:"password"`
}

type TokenResponse struct {
	Token string `json:"token"`
}
