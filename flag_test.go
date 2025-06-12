package bind

import (
	"flag"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFlag(t *testing.T) {
	exampleTime := time.Date(2022, 11, 10, 0, 0, 0, 0, time.UTC)

	flag.String("int", "", "string value")
	flag.String("bool", "", "string value")
	flag.String("float", "", "string value")
	flag.String("string", "", "string value")
	flag.String("time", "", "string value")

	flag.Set("int", "23")
	flag.Set("bool", "true")
	flag.Set("float", "3.14")
	flag.Set("string", "abc")
	flag.Set("time", exampleTime.Format(tagTimeFormat))

	expected := TestReceiver{
		Int:     23,
		Int8:    int8(23),
		Int16:   int16(23),
		Int32:   int32(23),
		Int64:   int64(23),
		Uint:    uint(23),
		Uint8:   uint8(23),
		Uint16:  uint16(23),
		Uint32:  uint32(23),
		Uint64:  uint64(23),
		Bool:    true,
		Float32: float32(3.14),
		Float64: 3.14,
		Time:    exampleTime,
		String:  "abc",
	}

	var receiver TestReceiver

	err := Flag(&receiver)
	require.Nil(t, err)
	assert.Equal(t, receiver, expected)
}

func TestFlag_NilReceiver(t *testing.T) {
	err := Flag(nil)
	require.Nil(t, err) // Flag did nothing, no error even if unintentional
}

func TestFlag_DefaultValue(t *testing.T) {
	receiver := struct {
		Data int `flag:"data" default:"23"`
	}{}

	err := Flag(&receiver)
	require.Nil(t, err)
	assert.Equal(t, 23, receiver.Data)
}

func TestFlagError_NonStruct(t *testing.T) {
	receiver := 1

	err := Flag(&receiver)
	require.ErrorIs(t, err, ErrReceiverUnsupportedType)
}

func TestFlagError_Slice(t *testing.T) {
	flag.String("slice", "", "string value")
	flag.Set("slice", "a,b,c")

	receiver := struct {
		Data []string `flag:"slice"`
	}{}

	err := Flag(&receiver)
	require.ErrorIs(t, err, ErrFieldSliceType)
}

func TestFlagError_Time(t *testing.T) {
	flag.String("time_error", "", "string value")
	flag.Set("time_error", "a,b,c")

	receiver := struct {
		Data time.Time `flag:"time_error"`
	}{}

	err := Flag(&receiver)
	require.ErrorIs(t, err, ErrFieldTimeFormat)
}

func TestFlagError_FieldType(t *testing.T) {
	flag.String("data", "", "string value")
	flag.Set("data", "23")

	receiver := struct {
		Data struct {
			Value int
		} `flag:"data"`
	}{}

	err := Flag(&receiver)
	require.ErrorIs(t, err, ErrFieldUnsupportedType)
}
