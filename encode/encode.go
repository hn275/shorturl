package encode

import (
	"fmt"
	"math"
	"strings"
)

type ID = uint32

const (
	table string = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	maxID uint32 = 0xffffffff
)

var (
	tableLen   uint8 = uint8(len(table))
	encodedLen uint8 = uint8(math.Ceil(math.Log(float64(maxID)) / math.Log(float64(tableLen))))
)

func Encode(id ID) string {
	buf := make([]byte, encodedLen)
	base := ID(tableLen)
	i := 0
	for {
		buf[i] = table[id%base]
		i++
		id /= base
		if id == 0 {
			break
		}
	}
	return string(buf[:i])
}

func Decode(encodedID string) (ID, error) {
	var id ID = 0
	for i, char := range encodedID {
        // TODO: optimize this index, math it out from the ascii table
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
