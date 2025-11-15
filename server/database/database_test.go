package database

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"testing"
)

func TestDataBaseOperations(t *testing.T) {
	db, err := Connect(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	const testCtr = 10

	var (
		testUrls   = [testCtr]string{}
		testNonces = [testCtr]string{}
		testIds    = [testCtr]uint64{}
	)

	for i := range testCtr {
		testUrls[i] = fmt.Sprintf("http://haln.dev/%d", i)

		nonce := [8]byte{}
		if _, err := io.ReadFull(rand.Reader, nonce[:]); err != nil {
			t.Fatal(err)
		}
		testNonces[i] = base64.RawStdEncoding.EncodeToString(nonce[:])

		var err error
		testIds[i], err = db.InsertURL(testNonces[i], testUrls[i])
		if err != nil {
			t.Fatal(err)
		}
	}

	for i := range testCtr {
		queryUrl, err := db.GetURL(testIds[i], testNonces[i])
		if err != nil {
			t.Fatal(err)
		}

		if testUrls[i] != queryUrl {
			t.Fatal("invalid url")
		}
	}
}
