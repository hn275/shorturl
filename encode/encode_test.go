package encode_test

import (
	"testing"

	"github.com/hn275/shorturl/encode"
	"github.com/stretchr/testify/assert"
)

func TestEncodeDecode(t *testing.T) {
	// technically the id is uint32, but looping from
	// 0 to 1 << 0xffffffff takes a really long time
	for i := 1; i < 0xffff; i++ {
		id := encode.ID(i)
		enc := encode.Encode(id)
		dec, err := encode.Decode(enc)
		assert.Nil(t, err)
		assert.Equal(t, id, dec)
	}

	p := encode.ID(0xffffffff)
	encodedStr := encode.Encode(p)
	decoded, err := encode.Decode(encodedStr)
	assert.Nil(t, err)
	assert.Equal(t, p, decoded)
}
