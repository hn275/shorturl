package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/hn275/shorturl/database"
	"github.com/hn275/shorturl/encode"
)

func main() {
	// db
	db, err := database.Connect(envOrElse("SQLITE_PATH", "database.db"))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	type h = http.HandlerFunc
	http.Handle("/assets/", handleAssets)
	http.Handle("/index.html", handleHome())
	http.Handle("/generate", h(handleGenerate))
	http.Handle("/", handleParams())

	port := envOrElse("PORT", "3000")
	log.Println("Listening on port:", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func handleGenerate(w http.ResponseWriter, r *http.Request) {
	// user input validation
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		stderr("parse form error: %v", err)
		writeError(w, http.StatusBadRequest, "failed to parse url encoded form")
		return
	}

	url := r.Form.Get("url")

	// insert to db
	db := database.New()
	id, err := db.InsertURL(url)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		stderr("failed to create a new url: %v", err)
		return
	}

	// id -> string
	data := struct {
		GeneratedHash string
	}{
		GeneratedHash: encode.Encode(id),
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
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		if r.URL.String() == "/" {
			http.Redirect(w, r, "/index.html", http.StatusPermanentRedirect)
			return
		}
		idEncoded := r.URL.String()[1:]
		fmt.Println(idEncoded)
		w.WriteHeader(http.StatusOK)
	})
}

func writeError(w http.ResponseWriter, httpCode int, msg string) {
	if _, err := w.Write([]byte(msg)); err != nil {
		stderr("error writing response: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Add("Content-Type", "text/plain")
	w.WriteHeader(httpCode)
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
