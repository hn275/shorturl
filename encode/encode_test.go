package encode

import (
	"crypto/rand"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncodeDecode(t *testing.T) {
	const testCtr = 0x1000

	for i := range testCtr {
		id := uint64(i)
		nonce := Nonce{}
		if _, err := io.ReadFull(rand.Reader, nonce[:]); err != nil {
			t.Fatal(err)
		}

		encoded := Encode(id, nonce)
		decodedID, decodedNonce, err := Decode(encoded)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, id, decodedID, "invalid id")
		assert.Equal(t, nonce, decodedNonce, "invalid nonce")
	}
}

func TestXorServerSecret(t *testing.T) {
	buf := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, buf); err != nil {
		t.Fatal(err)
	}

	bufCpy := make([]byte, len(buf))
	copy(bufCpy, buf)

	xorSecret(bufCpy)
	assert.NotEqual(t, buf, bufCpy)

	xorSecret(bufCpy)
	assert.Equal(t, buf, bufCpy)
}
