package encode

import (
	"encoding/base64"
	"fmt"
	"slices"
	"strconv"
)

type (
	ID    = string
	Nonce = [NonceSize]byte
)

const (
	NonceSize = 8
)

var (
	encoding = base64.RawURLEncoding

	EncodeToString = encoding.EncodeToString
	DecodeString   = encoding.DecodeString
)

func Encode(id uint64, nonce Nonce) string {
	idStr := strconv.Itoa(int(id))
	buf := slices.Concat(nonce[:], []byte(idStr))
	return EncodeToString(buf)
}

func Decode(encodedID string) (uint64, Nonce, error) {
	raw, err := DecodeString(encodedID)
	if err != nil {
		return 0, Nonce{}, fmt.Errorf("failed to decode raw id: %w", err)
	}

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
