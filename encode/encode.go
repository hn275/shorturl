package encode

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"log/slog"
	"os"
	"slices"
	"strconv"

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
	idStr := strconv.Itoa(int(id))
	buf := slices.Concat(nonce[:], []byte(idStr))

	subKey := make([]byte, len(buf))
	blake3.DeriveKey(subKey, kdfContextString, serverSecret)
	for i := range buf {
		buf[i] ^= subKey[i]
	}

	return Encoder.EncodeToString(buf)
}

func Decode(encodedID string) (uint64, Nonce, error) {
	raw, err := Encoder.DecodeString(encodedID)
	if err != nil {
		return 0, Nonce{}, fmt.Errorf("failed to decode raw id: %w", err)
	}

	xorSecret(raw)

	idStr, nonceRaw := raw[NonceSize:], raw[:NonceSize]

	if len(nonceRaw) != NonceSize {
		return 0, Nonce{}, fmt.Errorf("invalid nonce")
	}

	var nonce Nonce
	copy(nonce[:], nonceRaw)

	id, err := strconv.Atoi(string(idStr))
	if err != nil {
		return 0, Nonce{}, fmt.Errorf("failed to decode id: %w", err)
	}

	return uint64(id), nonce, nil
}

func xorSecret(buf []byte) []byte {
	subKey := make([]byte, len(buf))
	blake3.DeriveKey(subKey, kdfContextString, serverSecret)
	for i := range buf {
		buf[i] ^= subKey[i]
	}
	return buf
}
