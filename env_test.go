package bind

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnv(t *testing.T) {
	os.Setenv("BIND_TEST_VALUE_INT", "23")
	defer os.Unsetenv("BIND_TEST_VALUE_INT")

	os.Setenv("BIND_TEST_VALUE_BOOL", "true")
	defer os.Unsetenv("BIND_TEST_VALUE_BOOL")

	os.Setenv("BIND_TEST_VALUE_FLOAT", "3.14")
	defer os.Unsetenv("BIND_TEST_VALUE_FLOAT")

	os.Setenv("BIND_TEST_VALUE_STRING", "abc")
	defer os.Unsetenv("BIND_TEST_VALUE_STRING")

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
		String:  "abc",
	}

	var receiver TestReceiver

	err := Env(&receiver)
	require.Nil(t, err)
	assert.Equal(t, receiver, expected)
}

func TestEnv_NilReceiver(t *testing.T) {
	err := Env(nil)
	require.Nil(t, err) // Env did nothing, no error even if unintentional.
}

func TestEnv_DefaultValue(t *testing.T) {
	receiver := struct {
		Data int `env:"BIND_TEST_DEFAULT_VALUE" default:"23"`
	}{}

	err := Env(&receiver)
	require.Nil(t, err)
	assert.Equal(t, 23, receiver.Data)
}

func TestEnvError_NonStruct(t *testing.T) {
	receiver := 1

	err := Env(&receiver)
	require.ErrorIs(t, err, ErrReceiverUnsupportedType)
}

func TestEnvError_AnonStructField(t *testing.T) {
	t.Skip("unable to create an accurate test case")

	os.Setenv("BIND_TEST_ENV_ANON_STRUCT_FIELD", "23")
	defer os.Unsetenv("BIND_TEST_ENV_ANON_STRUCT_FIELD")

	receiver := struct {
		Data struct {
			Value int
		} `env:"BIND_TEST_ENV_ANON_STRUCT_FIELD"`
	}{}

	err := Env(&receiver)
	require.ErrorIs(t, err, ErrFieldAnonymousStruct)
}

func TestEnvError_Slice(t *testing.T) {
	os.Setenv("BIND_TEST_ENV_SLICE", "a,b,c")
	defer os.Unsetenv("BIND_TEST_ENV_SLICE")

	receiver := struct {
		Data []string `env:"BIND_TEST_ENV_SLICE"`
	}{}

	err := Env(&receiver)
	require.ErrorIs(t, err, ErrFieldSliceType)
}

func TestEnvError_Time(t *testing.T) {
	os.Setenv("BIND_TEST_ENV_TIME", "a,b,c")
	defer os.Unsetenv("BIND_TEST_ENV_TIME")

	receiver := struct {
		Data time.Time `env:"BIND_TEST_ENV_TIME"`
	}{}

	err := Env(&receiver)
	require.ErrorIs(t, err, ErrFieldTimeFormat)
}

func TestEnvError_FieldType(t *testing.T) {
	os.Setenv("BIND_TEST_ENV_FIELD_TYPE", "23")
	defer os.Unsetenv("BIND_TEST_ENV_FIELD_TYPE")

	receiver := struct {
		Data struct {
			Value int
		} `env:"BIND_TEST_ENV_FIELD_TYPE"`
	}{}

	err := Env(&receiver)
	require.ErrorIs(t, err, ErrFieldUnsupportedType)
}
