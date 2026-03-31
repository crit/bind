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
	flag.String("duration", "", "string value")
	flag.String("ptr_duration", "", "string value")

	flag.Set("int", "23")
	flag.Set("bool", "true")
	flag.Set("float", "3.14")
	flag.Set("string", "abc")
	flag.Set("time", exampleTime.Format(tagTimeFormat))
	flag.Set("duration", "1m30s")
	flag.Set("ptr_duration", "250ms")

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
		Float32:     float32(3.14),
		Float64:     3.14,
		Time:        exampleTime,
		String:      "abc",
		Duration:    90 * time.Second,
		PtrDuration: durationPtr(250 * time.Millisecond),
	}

	var receiver TestReceiver
	RegisterFlags(receiver)

	err := Flag(&receiver)
	require.Nil(t, err)
	assert.Equal(t, expected, receiver)
}

func TestFlag_NilReceiver(t *testing.T) {
	err := Flag(nil)
	require.Nil(t, err)
}

func TestFlag_DefaultValue(t *testing.T) {
	receiver := struct {
		Data int `flag:"data" default:"23"`
	}{}

	RegisterFlags(receiver)

	err := Flag(&receiver)
	require.Nil(t, err)
	assert.Equal(t, 23, receiver.Data)
}

func TestFlagError_NonStruct(t *testing.T) {
	receiver := 1
	RegisterFlags(receiver) // unable to register

	err := Flag(&receiver)
	require.ErrorIs(t, err, ErrReceiverUnsupportedType)
}

func TestFlagError_NonPointerReceiver(t *testing.T) {
	receiver := TestReceiver{}

	err := Flag(receiver)
	require.ErrorIs(t, err, ErrReceiverUnsupportedType)
}

func TestFlagError_TypedNilPointer(t *testing.T) {
	var receiver *TestReceiver

	err := Flag(receiver)
	require.ErrorIs(t, err, ErrReceiverUnsupportedType)
}

func TestFlag_SliceCSV(t *testing.T) {
	flag.String("slice", "", "string value")
	flag.Set("slice", "a,b,c")

	receiver := struct {
		Data []string `flag:"slice"`
	}{}
	RegisterFlags(receiver)

	err := Flag(&receiver)
	require.NoError(t, err)
	assert.Equal(t, []string{"a", "b", "c"}, receiver.Data)
}

func TestFlag_SliceCSV_Quoted(t *testing.T) {
	flag.String("slice_quoted", "", "string value")
	flag.Set("slice_quoted", "\"a,b\",c")

	receiver := struct {
		Data []string `flag:"slice_quoted"`
	}{}
	RegisterFlags(receiver)

	err := Flag(&receiver)
	require.NoError(t, err)
	assert.Equal(t, []string{"a,b", "c"}, receiver.Data)
}

func TestFlagError_SliceCSVMalformed(t *testing.T) {
	flag.String("slice_bad", "", "string value")
	flag.Set("slice_bad", "\"a,b")

	receiver := struct {
		Data []string `flag:"slice_bad"`
	}{}
	RegisterFlags(receiver)

	err := Flag(&receiver)
	require.ErrorIs(t, err, ErrFieldCSVFormat)
	assert.ErrorContains(t, err, "Data is an")
}

func TestFlagError_SliceCSV_UnparseableElement(t *testing.T) {
	flag.String("slice_int_bad", "", "string value")
	flag.Set("slice_int_bad", "1,abc")

	receiver := struct {
		Data []int `flag:"slice_int_bad"`
	}{}
	RegisterFlags(receiver)

	err := Flag(&receiver)
	require.Error(t, err)
	assert.ErrorContains(t, err, "Data is an")
	assert.ErrorContains(t, err, "invalid syntax")
}

func TestFlagError_Time(t *testing.T) {
	flag.String("time_error", "", "string value")
	flag.Set("time_error", "a,b,c")

	receiver := struct {
		Data time.Time `flag:"time_error"`
	}{}
	RegisterFlags(receiver)

	err := Flag(&receiver)
	require.ErrorIs(t, err, ErrFieldTimeFormat)
}

func TestFlagError_FieldType(t *testing.T) {
	flag.String("data_error", "", "string value")
	flag.Set("data_error", "23")

	receiver := struct {
		Data struct {
			Value int
		} `flag:"data_error"`
	}{}
	RegisterFlags(receiver)

	err := Flag(&receiver)
	require.ErrorIs(t, err, ErrFieldUnsupportedType)
}

func TestFlagError_Duration(t *testing.T) {
	flag.String("duration_error", "", "string value")
	flag.Set("duration_error", "abc")

	receiver := struct {
		Data time.Duration `flag:"duration_error"`
	}{}
	RegisterFlags(receiver)

	err := Flag(&receiver)
	require.Error(t, err)
	assert.ErrorContains(t, err, "Data is an")
	assert.ErrorContains(t, err, "invalid duration")
}

func TestFlagError_PtrDuration(t *testing.T) {
	flag.String("ptr_duration_error", "", "string value")
	flag.Set("ptr_duration_error", "abc")

	receiver := struct {
		Data *time.Duration `flag:"ptr_duration_error"`
	}{}
	RegisterFlags(receiver)

	err := Flag(&receiver)
	require.Error(t, err)
	assert.ErrorContains(t, err, "Data is an")
	assert.ErrorContains(t, err, "invalid duration")
}
