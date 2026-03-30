# Changelog

All notable changes to this project will be documented in this file.

## [2.0.0] - 2026-03-30

### Breaking Changes

- `RegisterFlags` signature changed:
  - **Before:** `func RegisterFlags(receivers ...any)`
  - **Now:** `func RegisterFlags(receivers ...any) error`
- `Env` and `Flag` now support slice fields via CSV parsing.
  - Previous behavior returned `ErrFieldSliceType` for slice fields.
  - New behavior parses values like `a,b,c` (including quoted CSV values).

### Added

- Configurable time parsing:
  - `SetTimeLayout(layout string) error`
  - `TimeLayout() string`
  - `ResetTimeLayout()`
- Per-field time layout override via struct tag:
  - ``time_layout:"..."``
- FlagSet-based APIs:
  - `RegisterFlagsWithSet(fs *flag.FlagSet, receivers ...any) error`
  - `FlagWithSet(fs *flag.FlagSet, receiver any) error`
- CSV parsing error sentinel:
  - `ErrFieldCSVFormat`
- Invalid time layout error sentinel:
  - `ErrInvalidTimeLayout`
- Benchmarks:
  - `BenchmarkParse`
  - `BenchmarkEnv`
  - `BenchmarkFlagWithSet`

### Changed

- Reflection receiver validation hardened across binder paths (`parse`, `Env`, `Flag`) to avoid panic-prone `.Elem()` assumptions.
- Map binding improved:
  - initializes nil maps before writes
  - skips empty input slices safely
- Pointer field binding improved:
  - allocates nil pointer targets before assignment
- Env semantics clarified:
  - set-but-empty env values are respected
  - defaults apply only when env var is unset
- CI modernized:
  - Go matrix now includes `1.19` and `1.22`
  - GitHub Actions updated (`checkout@v4`, `setup-go@v5`)

### Fixed

- Flag binding now uses parsed/current flag values rather than registration-time defaults.
- Removed process-terminating behavior (`log.Fatal`) from library code.
- Added/expanded tests for:
  - nil/non-pointer/typed-nil receiver behavior
  - anonymous embedded field tag error handling
  - malformed CSV and unparseable slice element values
  - time layout override and invalid/mismatched time formats

### Docs

- Added comprehensive `README.md` with:
  - API usage examples
  - supported tags and types
  - receiver behavior contract
  - troubleshooting
  - benchmark command/results

---

## Upgrade Notes (from pre-2.0)

1. Update calls to `RegisterFlags(...)` to handle returned errors:

```go
if err := bind.RegisterFlags(MyConfig{}); err != nil {
	return err
}
```

2. If you relied on `Env`/`Flag` returning `ErrFieldSliceType` for slices, update behavior expectations:
   - slices now parse from CSV (`a,b,c`).

3. If needed, set a global time layout or per-field `time_layout` tag for non-`2006-01-02` formats.
