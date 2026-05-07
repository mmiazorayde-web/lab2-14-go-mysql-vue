package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
	"github.com/gorilla/mux"
)

type Student struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Age   int    `json:"age"`
	Grade string `json:"grade"`
}

var db *sql.DB

func main() {
	var err error
	db, err = sql.Open("sqlite3", "./students.db")
	if err != nil {
		log.Fatal("❌ SQLite:", err)
	}
	defer db.Close()

	// 🗄️ Создать БД + тестовые данные
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS students (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			age INTEGER,
			grade TEXT
		);
		INSERT OR IGNORE INTO students (name, age, grade) VALUES 
			('Иван Петров', 20, 'A'),
			('Мария Сидорова', 19, 'B'),
			('Алексей Иванов', 21, 'A'),
			('Елена Козлова', 20, 'B');
	`)
	if err != nil {
		log.Fatal("❌ Создание таблицы:", err)
	}

	r := mux.NewRouter()
	r.HandleFunc("/", homePage)
	r.HandleFunc("/api/students", getStudents).Methods("GET")
	r.HandleFunc("/api/students", createStudent).Methods("POST")
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./")))

	fmt.Println("🚀 Сервер запущен: http://localhost:8080")
	fmt.Println("📱 API: http://localhost:8080/api/students")
	log.Fatal(http.ListenAndServe(":8080", r))
}

func homePage(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("index.html")
	t.Execute(w, nil)
}

func getStudents(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT id, name, age, grade FROM students ORDER BY name")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var students []Student
	for rows.Next() {
		var s Student
		rows.Scan(&s.ID, &s.Name, &s.Age, &s.Grade)
		students = append(students, s)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(students)
}

func createStudent(w http.ResponseWriter, r *http.Request) {
	var s Student
	json.NewDecoder(r.Body).Decode(&s)

	result, err := db.Exec("INSERT INTO students (name, age, grade) VALUES (?, ?, ?)",
		s.Name, s.Age, s.Grade)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	id, _ := result.LastInsertId()
	s.ID = int(id)

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(s)
}
