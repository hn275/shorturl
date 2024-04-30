package encode

import (
	"fmt"
	"math"
	"strings"
)

type ID = uint32

const (
	table string = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

var (
	encodedLen uint8 = uint8(math.Ceil(math.Log(float64(1<<32-1)) / math.Log(float64(tableLen))))
	tableLen   uint8 = uint8(len(table))
)

func Encode(id ID) string {
	out := make([]byte, encodedLen)
	l := ID(tableLen)
	i := 0
	for ; id >= l; i++ {
		idx := id % l
		out[i] = table[idx]
		id /= l
	}
	// the last one
	idx := id % l
	out[i] = table[idx]
	return string(out[:i+1])
}

func Decode(encodedID string) (ID, error) {
	var id ID = 0
	for i, char := range encodedID {
		index := strings.IndexRune(table, char)
		if index == -1 {
			break
		}
		id += ID(index) * ID(math.Pow(float64(tableLen), float64(i)))
	}

	if id == 0 {
		return id, fmt.Errorf("failed to decode id %s", encodedID)
	}

	return id, nil
}
