package main

import (
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/jmoiron/sqlx"
	sqlite3 "github.com/mattn/go-sqlite3"
)

func main() {
	// db
	db, err := databaseConnect("database.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	http.Handle("/assets/", handleAssets)
	http.Handle("/index.html", handleHome())
	http.Handle("/generate", handleGenerate(db))
	http.Handle("/", handleParams())
	log.Fatal(http.ListenAndServe(":3000", nil))
}

func handleGenerate(db *sqlx.DB) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// user input validation
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		if err := r.ParseForm(); err != nil {
			stderr("parse form error: %v", err)
			writeError(w, "failed to parse url encoded form", http.StatusBadRequest)
			return
		}

		url := r.Form.Get("url")

		// insert to db
		result, err := db.Exec("INSERT INTO urls (url) VALUES (?);", url)
		urlExists := false
		if err != nil {
			dbErr := &sqlite3.Error{}
			isSqlite3Err := errors.As(err, dbErr)
			if !isSqlite3Err || dbErr.ExtendedCode != sqlite3.ErrConstraintUnique {
				stderr("failed to save url {%s} to database: %v", url, err)
				writeError(w, "Failed to generate URL.", http.StatusInternalServerError)
				return
			}
			urlExists = true
		}

		var id int64
		if urlExists {
			q := "SELECT id FROM urls WHERE url = ?;"
			if err := db.Get(&id, q, url); err != nil {
				log.Fatal(err)
			}
		} else {
			var err error
			id, err = result.LastInsertId()
			if err != nil {
				log.Fatal(err)
			}
		}

		// id -> string
		fmt.Println(id)
		data := struct {
			GeneratedURL string `db:"url"`
		}{}

		// response
		tmpl, err := template.ParseFiles("public/generate.html")
		if err != nil {
			stderr(err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		tmpl.Execute(w, data)
	})

}

var handleAssets = http.StripPrefix("/assets/", http.FileServer(http.Dir("assets")))

func handleHome() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t, err := template.ParseFiles("public/index.html")
		if err != nil {
			stderr("Failed to parse index.html: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if err := t.Execute(w, nil); err != nil {
			stderr("failed to write template to response: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	})
}

func handleParams() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.String() == "/" {
			http.Redirect(w, r, "/index.html", http.StatusPermanentRedirect)
			return
		}
		w.WriteHeader(http.StatusOK)
	})
}

func databaseConnect(dbName string) (*sqlx.DB, error) {
	db, err := sqlx.Connect("sqlite3", dbName)
	if err != nil {
		return nil, err
	}
	schemas := `
    CREATE TABLE IF NOT EXISTS urls (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        url TEXT NOT NULL UNIQUE
    );
    `
	if _, err := db.Exec(schemas); err != nil {
		return nil, err
	}
	return db, nil
}

func writeError(w http.ResponseWriter, msg string, httpCode int) {
	w.WriteHeader(httpCode)
	if _, err := w.Write([]byte(msg)); err != nil {
		stderr("error writing response: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Add("Content-Type", "text/plain")

}

func stderr(format string, args ...any) {
	msg := fmt.Sprintf(format+"\n", args...)
	if _, err := os.Stderr.WriteString(msg); err != nil {
		log.Fatal(err)
	}
}
