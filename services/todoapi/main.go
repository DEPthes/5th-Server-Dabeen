package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

type Todo struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	Done      bool      `json:"done"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

var (
	masterDB *sql.DB
	slaveDB  *sql.DB
)

func main() {
	masterDB = mustConnectDB(buildDSN(getEnv("MYSQL_MASTER_HOST", "mysql-master")), "master")
	slaveDB = connectDB(buildDSN(getEnv("MYSQL_SLAVE_HOST", "mysql-slave")), "slave")
	if slaveDB == nil {
		log.Println("slave 연결 실패 → master를 읽기에도 사용합니다")
		slaveDB = masterDB
	}

	r := mux.NewRouter()
	r.Use(loggingMiddleware)
	r.Use(corsMiddleware)

	r.HandleFunc("/health", healthHandler).Methods(http.MethodGet)

	api := r.PathPrefix("/api").Subrouter()
	api.HandleFunc("/todos", listTodos).Methods(http.MethodGet)
	api.HandleFunc("/todos", createTodo).Methods(http.MethodPost)
	api.HandleFunc("/todos/{id:[0-9]+}", getTodo).Methods(http.MethodGet)
	api.HandleFunc("/todos/{id:[0-9]+}", updateTodo).Methods(http.MethodPut)
	api.HandleFunc("/todos/{id:[0-9]+}", deleteTodo).Methods(http.MethodDelete)
	// OPTIONS preflight
	api.Methods(http.MethodOptions).HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	port := getEnv("PORT", "8080")
	log.Printf("todoapi 시작 → :%s  (master=%s, slave=%s)",
		port,
		getEnv("MYSQL_MASTER_HOST", "mysql-master"),
		getEnv("MYSQL_SLAVE_HOST", "mysql-slave"),
	)
	log.Fatal(http.ListenAndServe(":"+port, r))
}

// ── DB 연결 ────────────────────────────────────────────────────

func buildDSN(host string) string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&charset=utf8mb4&loc=Asia%%2FSeoul",
		getEnv("MYSQL_USER", "todo"),
		getEnv("MYSQL_PASSWORD", "todo_pass_123"),
		host,
		getEnv("MYSQL_PORT", "3306"),
		getEnv("MYSQL_DATABASE", "tododb"),
	)
}

func mustConnectDB(dsn, label string) *sql.DB {
	db := connectDB(dsn, label)
	if db == nil {
		log.Fatalf("[%s] DB 연결 실패로 종료합니다", label)
	}
	return db
}

func connectDB(dsn, label string) *sql.DB {
	for i := 1; i <= 30; i++ {
		db, err := sql.Open("mysql", dsn)
		if err == nil {
			db.SetMaxOpenConns(25)
			db.SetMaxIdleConns(5)
			db.SetConnMaxLifetime(5 * time.Minute)
			if pingErr := db.Ping(); pingErr == nil {
				log.Printf("[%s] DB 연결 성공", label)
				return db
			}
		}
		log.Printf("[%s] DB 준비 대기 (%d/30)...", label, i)
		time.Sleep(3 * time.Second)
	}
	return nil
}

// ── Middleware ─────────────────────────────────────────────────

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s  %v", r.Method, r.URL.Path, time.Since(start))
	})
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		next.ServeHTTP(w, r)
	})
}

// ── Helpers ────────────────────────────────────────────────────

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v) //nolint:errcheck
}

func errJSON(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// ── Handlers ───────────────────────────────────────────────────

func healthHandler(w http.ResponseWriter, r *http.Request) {
	masterOK := masterDB.Ping() == nil
	slaveOK := slaveDB.Ping() == nil
	status := http.StatusOK
	if !masterOK {
		status = http.StatusServiceUnavailable
	}
	writeJSON(w, status, map[string]any{
		"status": "ok",
		"master": masterOK,
		"slave":  slaveOK,
	})
}

func listTodos(w http.ResponseWriter, r *http.Request) {
	rows, err := slaveDB.QueryContext(r.Context(),
		`SELECT id, title, done, created_at, updated_at
		 FROM todos ORDER BY created_at DESC`)
	if err != nil {
		errJSON(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	todos := make([]Todo, 0)
	for rows.Next() {
		var t Todo
		if err := rows.Scan(&t.ID, &t.Title, &t.Done, &t.CreatedAt, &t.UpdatedAt); err != nil {
			errJSON(w, http.StatusInternalServerError, err.Error())
			return
		}
		todos = append(todos, t)
	}
	writeJSON(w, http.StatusOK, todos)
}

func createTodo(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Title string `json:"title"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Title == "" {
		errJSON(w, http.StatusBadRequest, "title은 필수입니다")
		return
	}

	result, err := masterDB.ExecContext(r.Context(),
		"INSERT INTO todos (title) VALUES (?)", req.Title)
	if err != nil {
		errJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	id, _ := result.LastInsertId()
	todo, err := fetchTodo(r, masterDB, id)
	if err != nil {
		errJSON(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, todo)
}

func getTodo(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	todo, err := fetchTodo(r, slaveDB, id)
	if err == sql.ErrNoRows {
		errJSON(w, http.StatusNotFound, "todo를 찾을 수 없습니다")
		return
	}
	if err != nil {
		errJSON(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, todo)
}

func updateTodo(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)

	current, err := fetchTodo(r, masterDB, id)
	if err == sql.ErrNoRows {
		errJSON(w, http.StatusNotFound, "todo를 찾을 수 없습니다")
		return
	}
	if err != nil {
		errJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	var req struct {
		Title *string `json:"title"`
		Done  *bool   `json:"done"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errJSON(w, http.StatusBadRequest, "잘못된 요청 형식입니다")
		return
	}
	if req.Title != nil {
		current.Title = *req.Title
	}
	if req.Done != nil {
		current.Done = *req.Done
	}

	if _, err := masterDB.ExecContext(r.Context(),
		"UPDATE todos SET title = ?, done = ? WHERE id = ?",
		current.Title, current.Done, id,
	); err != nil {
		errJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	updated, _ := fetchTodo(r, masterDB, id)
	writeJSON(w, http.StatusOK, updated)
}

func deleteTodo(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	result, err := masterDB.ExecContext(r.Context(),
		"DELETE FROM todos WHERE id = ?", id)
	if err != nil {
		errJSON(w, http.StatusInternalServerError, err.Error())
		return
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		errJSON(w, http.StatusNotFound, "todo를 찾을 수 없습니다")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func fetchTodo(r *http.Request, db *sql.DB, id int64) (*Todo, error) {
	var t Todo
	err := db.QueryRowContext(r.Context(),
		`SELECT id, title, done, created_at, updated_at
		 FROM todos WHERE id = ?`, id,
	).Scan(&t.ID, &t.Title, &t.Done, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &t, nil
}
