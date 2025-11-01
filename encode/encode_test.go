package encode_test

import (
	"crypto/rand"
	"io"
	"testing"

	"github.com/hn275/shorturl/encode"
	"github.com/stretchr/testify/assert"
)

func TestEncodeDecode(t *testing.T) {
	const testCtr = 0x10

	for i := range testCtr {
		id := uint64(i)
		nonce := encode.Nonce{}
		if _, err := io.ReadFull(rand.Reader, nonce[:]); err != nil {
			t.Fatal(err)
		}

		encoded := encode.Encode(id, nonce)
		decodedID, decodedNonce, err := encode.Decode(encoded)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, id, decodedID, "invalid id")
		assert.Equal(t, nonce, decodedNonce, "invalid nonce")
	}
}
