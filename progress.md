# Progress Plan

## Epic 1: Stabilize and Correct Flag Binding

- [x] Fix `RegisterFlags` so it reads parsed flag values (not registration-time defaults).
- [x] Remove `log.Fatal` from `RegisterFlags`; return errors instead.
- [x] Refactor API to support `*flag.FlagSet` instead of relying only on global `flag.CommandLine`.
- [x] Add tests for:
  - [x] Parsed values are correctly bound.
  - [x] Default values are used when no flag is passed.
  - [x] Duplicate registration behavior.
  - [x] Error paths without process termination.
- [x] Ensure `go test ./...` passes for all flag-related tests.

## Epic 2: Harden Reflection Receiver Validation

- [x] Add shared validation helper(s) for binder entry points (`parse`, `Env`, `Flag`).
- [x] Validate receiver is:
  - [x] Non-nil.
  - [x] Pointer where required.
  - [x] Non-nil pointer value.
  - [x] Struct (or supported map path).
- [x] Replace panic-prone reflection assumptions (`.Elem()` on invalid values).
- [x] Add tests for invalid receiver inputs (non-pointer, typed nil pointer, nil interface).

## Epic 3: Fix Map and Pointer Edge Cases

- [ ] In map binding path, guard against empty value slices before indexing `v[0]`.
- [ ] Initialize map receiver when nil before calling `SetMapIndex`.
- [ ] In pointer field binding, allocate nil pointers before recursive set.
- [ ] Add tests for map and pointer edge cases.

## Epic 4: Improve Environment Variable Semantics

- [ ] Replace `os.Getenv` with `os.LookupEnv` in `Env` binding.
- [ ] Define behavior for empty env var values vs unset env vars.
- [ ] Ensure default-tag fallback behavior is documented and tested.

## Epic 5: Consistent Error and No-Op Behavior

- [ ] Define expected behavior for nil receivers across all binders.
- [ ] Decide and document whether nil receiver should return error or no-op.
- [ ] Normalize returned errors for unsupported receiver types.
- [ ] Add/adjust tests to match the finalized contract.

## Epic 6: CI and Repository Hygiene

- [ ] Update GitHub Actions Go version to match or exceed `go.mod` (`go 1.19`).
- [ ] Add version matrix if multi-version support is desired.
- [ ] Remove debug/build artifacts from repository (e.g. `tmp/__debug_bin*`).
- [ ] Expand `.gitignore` to prevent committing local binaries/artifacts.
- [ ] Confirm CI runs cleanly with `go build ./...` and `go test ./...`.

## Epic 7: Documentation and Developer Experience

- [ ] Add `README.md` with package overview and usage examples.
- [ ] Document supported tags: `query`, `form`, `header`, `env`, `flag`, `default`.
- [ ] Document supported field types and known limitations (e.g. env/flag slice support).
- [ ] Document time format behavior (`2006-01-02`) and customization plans.
- [ ] Add a short troubleshooting section for common binding errors.

## Epic 8: Performance and Maintainability

- [ ] Evaluate caching of struct metadata to reduce reflection overhead.
- [ ] Add benchmarks for parse/env/flag workflows.
- [ ] Refactor common binding logic to reduce duplication and improve readability.
- [ ] Keep error messages consistent and actionable.
