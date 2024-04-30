package database

import (
	"errors"
	"log"

	"github.com/hn275/shorturl/encode"
	"github.com/jmoiron/sqlx"
	sqlite3 "github.com/mattn/go-sqlite3"
)

type DB struct {
	*sqlx.DB
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
        url TEXT NOT NULL UNIQUE
    );
    `
	if _, err = sqlite3DB.Exec(schemas); err != nil {
		return nil, err
	}

	db = &DB{sqlite3DB}
	return db, nil
}

func (db *DB) InsertURL(url string) (encode.ID, error) {
	result, err := db.Exec("INSERT INTO urls (url) VALUES (?);", url)
	urlExists := false
	if err != nil {
		dbErr := &sqlite3.Error{}
		isSqlite3Err := errors.As(err, dbErr)
		if !isSqlite3Err || dbErr.ExtendedCode != sqlite3.ErrConstraintUnique {
			return 0, err
		}
		urlExists = true
	}

	if !urlExists {
		_id, err := result.LastInsertId()
		return encode.ID(_id), err
	}

	var id encode.ID
	q := "SELECT id FROM urls WHERE url = ?;"
	err = db.Get(&id, q, url)
	return id, err

}

func (db *DB) GetURL(id encode.ID) (string, error) {
    var url string
    return url, db.Get(&url, "SELECT url from urls WHERE id = ?;", id)
}
