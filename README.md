# bind

Lightweight struct binding for Go.

`bind` maps input values into struct fields using tags:

- HTTP/query-style maps (`query`, `form`, `header`)
- Environment variables (`env`)
- Command-line flags (`flag`)
- Fallback defaults (`default`)

## Install

```bash
go get github.com/crit/bind
```

## Quick start

```go
package main

import (
	"log"
	"net/http"

	"github.com/crit/bind"
)

type Input struct {
	CustomerID int    `query:"customer_id"`
	Name       string `query:"name" default:"guest"`
}

func handler(w http.ResponseWriter, r *http.Request) {
	var in Input
	if err := bind.Query(&in, r.URL.Query()); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("customer_id=%d name=%s", in.CustomerID, in.Name)
}
```

## Supported tags

- `query:"key"` — bind from map key for query params.
- `form:"key"` — bind from map key for form values.
- `header:"key"` — bind from map key for headers.
- `env:"NAME"` — bind from environment variable.
- `flag:"name"` — bind from command-line flag.
- `default:"value"` — fallback value when input key/var is missing.

## APIs

### Query / Form / Header

```go
err := bind.Query(&cfg, map[string][]string{"port": {"8080"}})
err := bind.Form(&cfg, formValues)
err := bind.Header(&cfg, headerValues)
```

### Env

```go
err := bind.Env(&cfg)
```

Env semantics:
- If env var is **set** (even to empty string), that value is used.
- If env var is **unset**, `default` is used (if present).

### Flag

Default command-line set:

```go
if err := bind.RegisterFlags(Config{}); err != nil {
	return err
}
if err := bind.Flag(&cfg); err != nil {
	return err
}
```

Custom `*flag.FlagSet`:

```go
fs := flag.NewFlagSet("app", flag.ContinueOnError)
if err := bind.RegisterFlagsWithSet(fs, Config{}); err != nil {
	return err
}
_ = fs.Parse([]string{"-port=8080"})
if err := bind.FlagWithSet(fs, &cfg); err != nil {
	return err
}
```

## Supported field types

Scalar types:
- `int`, `int8`, `int16`, `int32`, `int64`
- `uint`, `uint8`, `uint16`, `uint32`, `uint64`
- `bool`
- `float32`, `float64`
- `string`
- `time.Time`
- pointers to supported scalar types

Collection support:
- `Query`/`Form`/`Header`: supports slices (`[]T`) of supported scalar element types.
- `Env`: supports slices via CSV input (e.g. `A,B,C`).
- `Flag`: supports slices via CSV input (e.g. `-names=A,B,C`).

## Time format

Default `time.Time` parsing layout:

```text
2006-01-02
```

You can configure time parsing globally:

```go
if err := bind.SetTimeLayout("02-01-2006"); err != nil {
	return err
}
defer bind.ResetTimeLayout()
```

Or override per field using `time_layout`:

```go
type Config struct {
	Birthday time.Time `env:"BIRTHDAY" time_layout:"2006/01/02"`
}
```

If parsing fails, bind returns `ErrFieldTimeFormat`.

## Receiver behavior

- Nil receiver: no-op (returns `nil`).
- Non-pointer, typed nil pointer, or unsupported receiver kind: returns `ErrReceiverUnsupportedType`.

## Known limitations

- No custom delimiter configuration for CSV parsing yet (currently standard CSV semantics).

## Benchmarks

Run:

```bash
go test -run '^$' -bench 'Benchmark(Parse|Env|FlagWithSet)$' -benchmem ./...
```

Example result (Apple M1):

- `BenchmarkParse`: ~700 ns/op, 256 B/op, 6 allocs/op
- `BenchmarkEnv`: ~710 ns/op, 136 B/op, 2 allocs/op
- `BenchmarkFlagWithSet`: ~1420 ns/op, 240 B/op, 8 allocs/op

## Troubleshooting

- **"receiver was not a struct"**
  - Pass a pointer to a struct (or pointer to map for map binding path).
- **"unsupported type"**
  - Field type is not currently bindable (e.g., nested struct value field).
- **"unable to parse csv"**
  - Ensure env/flag slice values are valid CSV (quotes must be balanced).
- **"unable to parse time"**
  - Ensure value matches `YYYY-MM-DD` (`2006-01-02`).
- **"flag not registered with bind.RegisterFlags"**
  - Call `RegisterFlags(...)`/`RegisterFlagsWithSet(...)` before `Flag(...)`.
