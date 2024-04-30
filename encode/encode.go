package encode

import "math"

type ID = uint32

const (
	table     string = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	tableLen uint8  = uint8(len(table))
)

func Encode(id ID) string {
	i := uint16(math.Log(float64(id)) / math.Log(float64(tableLen)))
	out := make([]byte, i+1)
	l := ID(tableLen)
	for ; i != 0; i-- {
		idx := id % l
		out[i] = table[idx]
		id /= l
	}

	// the last one
	idx := id % l
	out[i] = table[idx]
	return string(out)
}

func Decode(digest string) int64 {
	return 0
}
