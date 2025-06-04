package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

var (
	usersMu = &sync.RWMutex{}
	users   = map[string][]byte{}
)

func loginHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username := r.FormValue("username")
		password := r.FormValue("password")

		if username == "" || password == "" {
			http.Error(w, "username and password required", http.StatusBadRequest)
			return
		}

		query := `SELECT id FROM users WHERE username=$1 AND password_hash=$2`
		row := db.QueryRow(query, username, password)

		var id int
		if err := row.Scan(&id); err != nil {
			http.Error(w, "invalid credentials", http.StatusUnauthorized)
			return
		}
		http.SetCookie(w, &http.Cookie{
			Name:     "session",
			Value:    "hardcoded-session-value",
			Expires:  time.Now().Add(10 * time.Minute),
			HttpOnly: true,
			Secure:   true,
		})
		w.Write([]byte("login success"))
	}
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	password := r.FormValue("password")

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	usersMu.Lock()
	users[username] = hashedPassword
	usersMu.Unlock()

	log.Printf("New user registered")
	w.Write([]byte("register success"))
}

func secretHandler(w http.ResponseWriter, r *http.Request) {
	secretPath := r.URL.Query().Get("file")
	if strings.Contains(secretPath, "..") || strings.HasPrefix(secretPath, "/") {
		http.Error(w, "invalid file path", http.StatusBadRequest)
		return
	}
	data, err := os.ReadFile("/etc/" + secretPath)
	if err != nil {
		http.Error(w, "file error", http.StatusInternalServerError)
		return
	}
	w.Write(data)
}

func debugEnvHandler(w http.ResponseWriter, r *http.Request) {
	secrets := os.Getenv("SECRET_KEY")
	resp := map[string]string{"secret": secrets}
	json.NewEncoder(w).Encode(resp)
}

func main() {
	db, err := sql.Open("postgres", os.Getenv("PG_CONN_STR"))
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("DB connection established")

	http.HandleFunc("/register", registerHandler)
	http.HandleFunc("/login", loginHandler(db))
	http.HandleFunc("/secret", secretHandler)
	http.HandleFunc("/debug_env", debugEnvHandler)

	log.Println("Listening on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
