package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/hn275/shorturl/database"
	"github.com/hn275/shorturl/encode"
	"github.com/hn275/shorturl/router"
)

const (
	urlLimitLength = 0x800
)

func main() {
	// db
	db, err := database.Connect(envOrElse("SQLITE_PATH", "database.db"))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// mux
	r := router.New()

	type h = http.HandlerFunc
	r.Handle("/assets/", handleAssets)
	r.Handle("/index.html", h(handleHome))
	r.Handle("/generate", h(handleGenerate))
	r.Handle("/", h(handleParams))

	port := envOrElse("PORT", "3000")
	log.Println("Listening on port:", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
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
	if len(url) > urlLimitLength {
		writeError(w, http.StatusBadRequest, "url too long")
	}

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

	w.Header().Add("Cache-Control", "no-cache")
	if err := tmpl.Execute(w, data); err != nil {
		stderr("failed to execute templating:\n%v", err)
	}

}

var handleAssets = http.StripPrefix("/assets/", http.FileServer(http.Dir("assets")))

func handleHome(w http.ResponseWriter, r *http.Request) {
	buf, err := os.ReadFile("public/index.html")
	if err != nil {
		stderr("Failed to parse index.html: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Add("Cache-Control", "no-cache")
	w.Header().Add("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(buf); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
}

func handleParams(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// redirecting
	if r.URL.String() == "/" {
		w.Header().Add("Cache-Control", "no-cache")
		http.Redirect(w, r, "/index.html", http.StatusPermanentRedirect)
		return
	}

	if r.URL.String() == "/favicon.ico" {
		w.Header().Add("Cache-Control", "no-cache")
		http.Redirect(w, r, "/assets/favicon.ico", http.StatusPermanentRedirect)
		return
	}

	// decode url
	id, err := encode.Decode(r.URL.String()[1:])
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	// get url from db
	db := database.New()
	url, err := db.GetURL(id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	w.Header().Add("Cache-Control", "no-cache")
	http.Redirect(w, r, url, http.StatusPermanentRedirect)
}

func writeError(w http.ResponseWriter, httpCode int, msg string) {
	w.Header().Add("Content-Type", "text/plain")
	w.WriteHeader(httpCode)
	if _, err := w.Write([]byte(msg)); err != nil {
		stderr("error writing response: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
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
