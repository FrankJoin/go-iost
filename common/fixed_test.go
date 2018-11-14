package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFixed_ToString(t *testing.T) {
	f := Fixed{-9223372036854775808, 4, nil}
	f.ToString()
	assert.Equal(t, f.Err, errOverflow)
	f = Fixed{100, 0, nil}
	s := f.ToString()
	assert.Equal(t, s, "100")
}

func TestNewFixed(t *testing.T) {
	f, err := NewFixed("-323.49494", 12)
	assert.Equal(t, err, nil)
	assert.Equal(t, f.Value, int64(-323494940000000))
	_, err = NewFixed("-9223372036854775808", 0)
	assert.Equal(t, err, errOverflow)
	_, err = NewFixed("-323.49494", 40)
	assert.Equal(t, err, errOverflow)
	_, err = NewFixed("323.494.94", 10)
	assert.Equal(t, err, errDoubleDot)
}

func TestFixed_Multiply(t *testing.T) {
	f1 := Fixed{-9223372036854775807, 4, nil}
	f2 := Fixed{-9223372036854775807, 4, nil}
	f1.Multiply(&f2)
	assert.Equal(t, f1.Err, errOverflow)
}

func TestFixed_Times(t *testing.T) {
	f1 := Fixed{-9223372036854775807, 4, nil}
	f1.Times(3)
	assert.Equal(t, f1.Err, errOverflow)
}

func TestFixed_Marshal(t *testing.T) {
	f := Fixed{1230, 2, nil}
	assert.Equal(t, "12.30", f.Marshal())
	f2, err := UnmarshalFixed("12.30")
	assert.Equal(t, nil, err)
	assert.Equal(t, 2, f2.Decimal)
	assert.Equal(t, nil, f2.Err)
	assert.Equal(t, int64(1230), f2.Value)
}
