package main

import (
	"errors"
	"fmt"
	"html/template"
	"log"
	"math"
	"net/http"
	"os"

	"github.com/jmoiron/sqlx"
	sqlite3 "github.com/mattn/go-sqlite3"
)

const (
	ENCODE_TABLE     string = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	ENCODE_TABLE_LEN uint8  = uint8(len(ENCODE_TABLE))
)

var (
	encoder = Encoder{}
)

type ID = uint32

type Encoder struct{}

func (e *Encoder) Encode(id ID) string {
	i := uint16(math.Log(float64(id)) / math.Log(float64(ENCODE_TABLE_LEN)))
	out := make([]byte, i+1)
	l := ID(ENCODE_TABLE_LEN)
	for ; i != 0; i-- {
		idx := id % l
		out[i] = ENCODE_TABLE[idx]
		id /= l
	}

	// the last one
	idx := id % l
	out[i] = ENCODE_TABLE[idx]
	return string(out)
}

func (e *Encoder) Decode(digest string) int64 {
	return 0
}

func main() {
	// db
	db, err := databaseConnect(envOrElse("SQLITE_PATH", "database.db"))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	http.Handle("/assets/", handleAssets)
	http.Handle("/index.html", handleHome())
	http.Handle("/generate", handleGenerate(db))
	http.Handle("/", handleParams())

	port := envOrElse("PORT", "3000")
	log.Println("Listening on port:", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
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

		var id ID
		if urlExists {
			q := "SELECT id FROM urls WHERE url = ?;"
			if err := db.Get(&id, q, url); err != nil {
				log.Fatal(err)
			}
		} else {
			var err error
			_id, err := result.LastInsertId()
			if err != nil {
				log.Fatal(err)
			}
			id = ID(_id)
		}

		// id -> string
		data := struct {
			GeneratedHash string
		}{
			GeneratedHash: encoder.Encode(id),
		}

		// response
		tmpl, err := template.ParseFiles("public/generate.html")
		if err != nil {
			stderr(err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if err := tmpl.Execute(w, data); err != nil {
			stderr("failed to execute templating:\n%v", err)
		}
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
	log.Println("connecting to database:", dbName)
	schemas := `
    CREATE TABLE IF NOT EXISTS urls (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        url TEXT NOT NULL UNIQUE
    );
    `
	db.MustExec(schemas)
	log.Println("database migrated")
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

func envOrElse(k string, defaultValue string) string {
	v := os.Getenv(k)
	if v == "" {
		log.Printf("Environment variable `%s` not set, using default `%s`\n", k, defaultValue)
		return defaultValue
	}
	return v
}
