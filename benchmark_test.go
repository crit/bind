package bind

import (
	"flag"
	"io"
	"os"
	"testing"
	"time"
)

type benchReceiver struct {
	Int     int       `test:"int" env:"BIND_BENCH_INT" flag:"bench_int"`
	Bool    bool      `test:"bool" env:"BIND_BENCH_BOOL" flag:"bench_bool"`
	Float   float64   `test:"float" env:"BIND_BENCH_FLOAT" flag:"bench_float"`
	String  string    `test:"string" env:"BIND_BENCH_STRING" flag:"bench_string"`
	Date    time.Time `test:"date" env:"BIND_BENCH_DATE" flag:"bench_date"`
	Ints    []int     `test:"ints"`
	Strings []string  `test:"strings"`
}

var benchData = map[string][]string{
	"int":     {"23"},
	"bool":    {"true"},
	"float":   {"3.14"},
	"string":  {"abc"},
	"date":    {"2022-11-10"},
	"ints":    {"1", "2", "3"},
	"strings": {"a", "b", "c"},
}

func resetRegisteredFlags() {
	flagsMux.Lock()
	defer flagsMux.Unlock()
	registeredFlags = make(map[string]map[string]struct{})
}

func BenchmarkParse(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var r benchReceiver
		if err := parse(&r, "test", benchData); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkEnv(b *testing.B) {
	_ = os.Setenv("BIND_BENCH_INT", "23")
	_ = os.Setenv("BIND_BENCH_BOOL", "true")
	_ = os.Setenv("BIND_BENCH_FLOAT", "3.14")
	_ = os.Setenv("BIND_BENCH_STRING", "abc")
	_ = os.Setenv("BIND_BENCH_DATE", "2022-11-10")
	b.Cleanup(func() {
		_ = os.Unsetenv("BIND_BENCH_INT")
		_ = os.Unsetenv("BIND_BENCH_BOOL")
		_ = os.Unsetenv("BIND_BENCH_FLOAT")
		_ = os.Unsetenv("BIND_BENCH_STRING")
		_ = os.Unsetenv("BIND_BENCH_DATE")
	})

	for i := 0; i < b.N; i++ {
		var r benchReceiver
		if err := Env(&r); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkFlagWithSet(b *testing.B) {
	resetRegisteredFlags()

	fs := flag.NewFlagSet("bench", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	if err := RegisterFlagsWithSet(fs, benchReceiver{}); err != nil {
		b.Fatal(err)
	}
	if err := fs.Set("bench_int", "23"); err != nil {
		b.Fatal(err)
	}
	if err := fs.Set("bench_bool", "true"); err != nil {
		b.Fatal(err)
	}
	if err := fs.Set("bench_float", "3.14"); err != nil {
		b.Fatal(err)
	}
	if err := fs.Set("bench_string", "abc"); err != nil {
		b.Fatal(err)
	}
	if err := fs.Set("bench_date", "2022-11-10"); err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var r benchReceiver
		if err := FlagWithSet(fs, &r); err != nil {
			b.Fatal(err)
		}
	}
}

