package main

import (
	"crypto/rand"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/hn275/shorturl/database"
	"github.com/hn275/shorturl/encode"
	"github.com/hn275/shorturl/router"
)

const (
	urlLimitLength = 0x800
)

var handleAssets = http.StripPrefix("/assets/", http.FileServer(http.Dir("assets")))

func init() {
	if os.Getenv("DEBUG") == "1" {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}
}

func main() {
	// db
	db, err := database.Connect(envOrElse("SQLITE_PATH", "/tmp/database.db"))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// mux
	r := router.New()

	r.HandleFunc("/api/qrcode", handleQr).Methods(http.MethodGet)
	r.HandleFunc("/api/shorturl", handleUrl).Methods(http.MethodPost)

	type h = http.HandlerFunc
	r.Handle("/", h(handleHome))
	r.Handle("/assets/", handleAssets)
	r.Handle("/{urlEncoded}", h(handleParams))

	port := envOrElse("PORT", "8000")
	log.Println("Listening on port:", port)
	log.Fatal(http.ListenAndServe(":"+port, handlers.LoggingHandler(os.Stdout, r)))
}

func handleQr(w http.ResponseWriter, r *http.Request) {}

func handleUrl(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Query().Get("url")
	if len(url) > urlLimitLength {
		writeError(w, http.StatusBadRequest, "url too long")
		return
	}

	// insert to db
	nonce := encode.Nonce{}
	if _, err := io.ReadFull(rand.Reader, nonce[:]); err != nil {
		writeError(w, http.StatusInternalServerError, "")
		slog.Error("failed to make nonce", "err", err)
		return
	}

	nonceEncoded := encode.Encoder.EncodeToString(nonce[:])

	db := database.New()

	id, err := db.InsertURL(url, nonceEncoded)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "")
		slog.Error(
			"failed to insert url to database",
			"url", url, "nonce", nonce, "err", err,
		)
		return
	}

	generatedHash := encode.Encode(id, nonce)

	// response
	w.Header().Add("Cache-Control", "no-cache")
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(generatedHash))
}

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
		slog.Error("failed to write response", "err", err)
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
	vars := mux.Vars(r)
	encodedUrl, ok := vars["urlEncoded"]
	if !ok {
		writeError(w, http.StatusBadRequest, "missing encoded url")
		return
	}

	id, nonce, err := encode.Decode(encodedUrl)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	// get url from db
	nonceEncoded := encode.Encoder.EncodeToString(nonce[:])
	db := database.New()
	url, err := db.GetURL(id, nonceEncoded)
	if err != nil {
		writeError(w, http.StatusNotFound, "url not found")
		slog.Error("failed to query URL", "id", id, "nonce", nonceEncoded, "err", err)
		return
	}
	w.Header().Add("Cache-Control", "no-cache")
	http.Redirect(w, r, url, http.StatusPermanentRedirect)
}

func writeError(w http.ResponseWriter, httpCode int, msg string) {
	w.Header().Add("Content-Type", "text/plain")
	w.WriteHeader(httpCode)

	if httpCode == http.StatusInternalServerError {
		slog.Error("server error", "err", msg)
		return
	}

	if _, err := w.Write([]byte(msg)); err != nil {
		stderr("error writing response: %v", err)
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
