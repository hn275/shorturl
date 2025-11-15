package encode

import (
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"log/slog"
	"os"
	"slices"

	"lukechampine.com/blake3"
)

type (
	ID    = string
	Nonce = [NonceSize]byte
)

const (
	NonceSize = 4

	kdfContextString = "ushort 2025-11-03 00:28:17 short url context string"
)

var (
	Encoder = base64.RawURLEncoding

	serverSecret = []byte{}
	binEncoder   = binary.LittleEndian
)

func init() {
	secretHex, ok := os.LookupEnv("SECRET")
	if !ok {
		slog.Error("server SECRET not set")
		os.Exit(1)
	}

	var err error
	serverSecret, err = hex.DecodeString(secretHex)
	if err != nil {
		slog.Error("failed to decode server SECRET", "err", err)
		os.Exit(1)
	}

	if len(serverSecret) < 32 {
		slog.Error("weak server secret")
		os.Exit(1)
	}
}

func Encode(id uint64, nonce Nonce) string {
	buf := make([]byte, 8+NonceSize)
	binEncoder.PutUint64(buf[:8], id)

	copy(buf[8:], nonce[:])

	subKey := make([]byte, 8)
	blake3.DeriveKey(
		subKey,
		kdfContextString,
		slices.Concat(serverSecret, nonce[:]),
	)

	for i := range subKey {
		buf[i] ^= subKey[i]
	}

	return Encoder.EncodeToString(buf)
}

func Decode(encodedID string) (uint64, Nonce, error) {
	raw, err := Encoder.DecodeString(encodedID)
	if err != nil {
		return 0, Nonce{}, fmt.Errorf("failed to decode raw id: %w", err)
	}

	nonce := Nonce{}
	copy(nonce[:], raw[8:])

	subKey := make([]byte, 8)

	blake3.DeriveKey(
		subKey,
		kdfContextString,
		slices.Concat(serverSecret, nonce[:]),
	)

	for i := range subKey {
		raw[i] ^= subKey[i]
	}

	id := binEncoder.Uint64(raw[:8])

	return id, nonce, nil
}
