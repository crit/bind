package bind

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TestReceiver struct {
	Int           int       `env:"BIND_TEST_VALUE_INT" test:"int"`
	Int8          int8      `env:"BIND_TEST_VALUE_INT" test:"int"`
	Int16         int16     `env:"BIND_TEST_VALUE_INT" test:"int"`
	Int32         int32     `env:"BIND_TEST_VALUE_INT" test:"int"`
	Int64         int64     `env:"BIND_TEST_VALUE_INT" test:"int"`
	Uint          uint      `env:"BIND_TEST_VALUE_INT" test:"int"`
	Uint8         uint8     `env:"BIND_TEST_VALUE_INT" test:"int"`
	Uint16        uint16    `env:"BIND_TEST_VALUE_INT" test:"int"`
	Uint32        uint32    `env:"BIND_TEST_VALUE_INT" test:"int"`
	Uint64        uint64    `env:"BIND_TEST_VALUE_INT" test:"int"`
	Bool          bool      `env:"BIND_TEST_VALUE_BOOL" test:"bool"`
	Float32       float32   `env:"BIND_TEST_VALUE_FLOAT" test:"float"`
	Float64       float64   `env:"BIND_TEST_VALUE_FLOAT" test:"float"`
	String        string    `env:"BIND_TEST_VALUE_STRING" test:"string"`
	Time          time.Time `env:"BIND_TEST_VALUE_TIME" test:"time"`
	SliceInt      []int     `test:"slice_int"`    // TODO: env support
	SliceBool     []bool    `test:"slice_bool"`   // TODO: env support
	SliceFloat    []float64 `test:"slice_float"`  // TODO: env support
	SliceString   []string  `test:"slice_string"` // TODO: env support
	StrangeCasing string    `test:"strange_casing"`
}

var (
	TestCase = TestReceiver{
		Int:           23,
		Int8:          23,
		Int16:         23,
		Int32:         23,
		Int64:         23,
		Uint:          23,
		Uint8:         23,
		Uint16:        23,
		Uint32:        23,
		Uint64:        23,
		Bool:          true,
		Float32:       3.14,
		Float64:       3.14,
		String:        "abc",
		Time:          time.Date(2022, 11, 10, 0, 0, 0, 0, time.UTC),
		SliceInt:      []int{1, 2, 3},
		SliceBool:     []bool{true, true, true},
		SliceFloat:    []float64{1.23, 2.34, 3.45},
		SliceString:   []string{"a", "b", "c"},
		StrangeCasing: "abc",
	}

	TestData = map[string][]string{
		"int":            {"23"},
		"bool":           {"true"},
		"float":          {"3.14"},
		"string":         {"abc"},
		"slice_int":      {"1", "2", "3"},
		"slice_bool":     {"true", "true", "true"},
		"slice_float":    {"1.23", "2.34", "3.45"},
		"slice_string":   {"a", "b", "c"},
		"time":           {TestCase.Time.Format(tagTimeFormat)},
		"StrAnGE_CasINg": {"abc"},
	}
)

func TestParse(t *testing.T) {
	var receiver TestReceiver

	err := parse(&receiver, "test", TestData)
	require.Nil(t, err)
	assert.Equal(t, TestCase, receiver)
}

func TestParse_DefaultValue(t *testing.T) {
	receiver := struct {
		Data int `test:"data" default:"23"`
	}{}

	err := parse(&receiver, "test", map[string][]string{})
	require.Nil(t, err)
	assert.Equal(t, 23, receiver.Data)
}

func TestParseError_NonStruct(t *testing.T) {
	receiver := 1

	err := parse(&receiver, "test", TestData)
	require.ErrorIs(t, err, ErrReceiverUnsupportedType)
}

func TestParseError_Time(t *testing.T) {
	data := map[string][]string{
		"date": {"abc"},
	}

	receiver := struct {
		Date time.Time `test:"date"`
	}{}

	err := parse(&receiver, "test", data)
	require.ErrorIs(t, err, ErrFieldTimeFormat)
}

func TestParseError_FieldType(t *testing.T) {
	data := map[string][]string{
		"data": {"23"},
	}

	receiver := struct {
		Data struct {
			Value int
		} `test:"data"`
	}{}

	err := parse(&receiver, "test", data)
	require.ErrorIs(t, err, ErrFieldUnsupportedType)
}
