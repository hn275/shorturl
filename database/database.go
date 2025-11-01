package database

import (
	"fmt"
	"log"
	"sync"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	*sqlx.DB
	mtx *sync.Mutex
}

var db *DB

func New() *DB {
	return db
}

func Connect(dbName string) (*DB, error) {
	log.Println("connecting to database:", dbName)
	sqlite3DB, err := sqlx.Connect("sqlite3", dbName)
	if err != nil {
		return nil, err
	}

	schemas := `
    CREATE TABLE IF NOT EXISTS urls (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
		nonce TEXT NOT NULL,
        url TEXT NOT NULL UNIQUE
    );
    `
	if _, err = sqlite3DB.Exec(schemas); err != nil {
		return nil, err
	}

	db = &DB{
		DB:  sqlite3DB,
		mtx: &sync.Mutex{},
	}

	return db, nil
}

func (db *DB) InsertURL(nonce, url string) (uint64, error) {
	db.mtx.Lock()
	defer db.mtx.Unlock()

	result, err := db.Exec("INSERT INTO urls (nonce,url) VALUES (?,?);", nonce, url)
	if err != nil {
		return 0, fmt.Errorf("failed to insert to database: %w", err)
	}

	lastInsertID, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last inserted id: %w", err)
	}

	return uint64(lastInsertID), err
}

func (db *DB) GetURL(id uint64, nonce string) (string, error) {
	db.mtx.Lock()
	defer db.mtx.Unlock()

	var url string
	return url, db.Get(&url, "SELECT url from urls WHERE id = ? AND nonce = ?;", id, nonce)
}
