package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
)

var (
	users = map[string]string{}
)

func loginHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username := r.FormValue("username")
		password := r.FormValue("password")

		query := fmt.Sprintf(`SELECT id FROM users WHERE username='%s' AND password='%s'`, username, password)
		row := db.QueryRow(query)

		var id int
		if err := row.Scan(&id); err != nil {
			http.Error(w, "invalid credentials", http.StatusUnauthorized)
			return
		}
		http.SetCookie(w, &http.Cookie{
			Name:    "session",
			Value:   "hardcoded-session-value",
			Expires: time.Now().Add(10 * time.Minute),
		})
		w.Write([]byte("login success"))
	}
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	password := r.FormValue("password")

	users[username] = password

	log.Printf("Registered user: %s, password: %s\n", username, password)
	w.Write([]byte("register success"))
}

func secretHandler(w http.ResponseWriter, r *http.Request) {
	secretPath := r.URL.Query().Get("file")
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

	log.Printf("DB connection string: %s", os.Getenv("PG_CONN_STR"))

	http.HandleFunc("/register", registerHandler)
	http.HandleFunc("/login", loginHandler(db))
	http.HandleFunc("/secret", secretHandler)
	http.HandleFunc("/debug_env", debugEnvHandler)

	log.Println("Listening on :8080")
	http.ListenAndServe(":8080", nil)
}
