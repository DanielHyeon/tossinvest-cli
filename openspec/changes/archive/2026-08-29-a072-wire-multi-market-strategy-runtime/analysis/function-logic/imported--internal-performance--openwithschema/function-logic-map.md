# Function Logic Map: `openWithSchema`

- Source: `internal/performance/store.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| function parameters/state | validated journal-derived trades, observations, query/window, and derived-store state | current Go signature and persisted/server-owned data | invalid, missing, or corrupt evidence follows explicit error/not-measured/test-failure paths |
| safety boundary | server-owned identities and fixed contracts only | approved a049 OpenSpec plus current code | never invents lineage/cost and never expands trading authority |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | AST `if` at `internal/performance/store.go:43`: `if strings.TrimSpace(path) == "" \|\| path == "." {` | local analytics values or derived SQLite state only; corrupt/unknown values remain errors or not-measured | condition determines the documented success/error/assertion path | `TestStoreSchemaIsSeparateAppendOnlyAndVersioned`, `TestCollectExactReplayIsIdempotentAcrossRestartAndConcurrency`, `TestCollectDivergentReplayFailsClosed`, `TestCollectPersistsUnknownCostAsSQLNullWithoutInventingZero` |
| B2 | AST `if` at `internal/performance/store.go:47`: `if err := os.MkdirAll(dir, 0o700); err != nil {` | local analytics values or derived SQLite state only; corrupt/unknown values remain errors or not-measured | condition determines the documented success/error/assertion path | `TestStoreSchemaIsSeparateAppendOnlyAndVersioned`, `TestCollectExactReplayIsIdempotentAcrossRestartAndConcurrency`, `TestCollectDivergentReplayFailsClosed`, `TestCollectPersistsUnknownCostAsSQLNullWithoutInventingZero` |
| B3 | AST `if` at `internal/performance/store.go:57`: `if err != nil {` | local analytics values or derived SQLite state only; corrupt/unknown values remain errors or not-measured | condition determines the documented success/error/assertion path | `TestStoreSchemaIsSeparateAppendOnlyAndVersioned`, `TestCollectExactReplayIsIdempotentAcrossRestartAndConcurrency`, `TestCollectDivergentReplayFailsClosed`, `TestCollectPersistsUnknownCostAsSQLNullWithoutInventingZero` |
| B4 | AST `if` at `internal/performance/store.go:63`: `if err := db.Ping(); err != nil {` | local analytics values or derived SQLite state only; corrupt/unknown values remain errors or not-measured | condition determines the documented success/error/assertion path | `TestStoreSchemaIsSeparateAppendOnlyAndVersioned`, `TestCollectExactReplayIsIdempotentAcrossRestartAndConcurrency`, `TestCollectDivergentReplayFailsClosed`, `TestCollectPersistsUnknownCostAsSQLNullWithoutInventingZero` |
| B5 | AST `if` at `internal/performance/store.go:67`: `if err := store.migrate(context.Background(), schema, targetVersion); err != nil {` | local analytics values or derived SQLite state only; corrupt/unknown values remain errors or not-measured | condition determines the documented success/error/assertion path | `TestStoreSchemaIsSeparateAppendOnlyAndVersioned`, `TestCollectExactReplayIsIdempotentAcrossRestartAndConcurrency`, `TestCollectDivergentReplayFailsClosed`, `TestCollectPersistsUnknownCostAsSQLNullWithoutInventingZero` |
| B6 | AST `range` at `internal/performance/store.go:71`: `for _, name := range []string{path, path + "-wal", path + "-shm"} {` | local analytics values or derived SQLite state only; corrupt/unknown values remain errors or not-measured | condition determines the documented success/error/assertion path | `TestStoreSchemaIsSeparateAppendOnlyAndVersioned`, `TestCollectExactReplayIsIdempotentAcrossRestartAndConcurrency`, `TestCollectDivergentReplayFailsClosed`, `TestCollectPersistsUnknownCostAsSQLNullWithoutInventingZero` |
| B7 | AST `if` at `internal/performance/store.go:72`: `if _, err := os.Stat(name); err == nil {` | local analytics values or derived SQLite state only; corrupt/unknown values remain errors or not-measured | condition determines the documented success/error/assertion path | `TestStoreSchemaIsSeparateAppendOnlyAndVersioned`, `TestCollectExactReplayIsIdempotentAcrossRestartAndConcurrency`, `TestCollectDivergentReplayFailsClosed`, `TestCollectPersistsUnknownCostAsSQLNullWithoutInventingZero` |
| B8 | AST `else` at `internal/performance/store.go:77`: `} else if !errors.Is(err, os.ErrNotExist) {` | local analytics values or derived SQLite state only; corrupt/unknown values remain errors or not-measured | condition determines the documented success/error/assertion path | `TestStoreSchemaIsSeparateAppendOnlyAndVersioned`, `TestCollectExactReplayIsIdempotentAcrossRestartAndConcurrency`, `TestCollectDivergentReplayFailsClosed`, `TestCollectPersistsUnknownCostAsSQLNullWithoutInventingZero` |
| B9 | AST `if` at `internal/performance/store.go:73`: `if err := os.Chmod(name, 0o600); err != nil {` | local analytics values or derived SQLite state only; corrupt/unknown values remain errors or not-measured | condition determines the documented success/error/assertion path | `TestStoreSchemaIsSeparateAppendOnlyAndVersioned`, `TestCollectExactReplayIsIdempotentAcrossRestartAndConcurrency`, `TestCollectDivergentReplayFailsClosed`, `TestCollectPersistsUnknownCostAsSQLNullWithoutInventingZero` |
| B10 | AST `if` at `internal/performance/store.go:77`: `} else if !errors.Is(err, os.ErrNotExist) {` | local analytics values or derived SQLite state only; corrupt/unknown values remain errors or not-measured | condition determines the documented success/error/assertion path | `TestStoreSchemaIsSeparateAppendOnlyAndVersioned`, `TestCollectExactReplayIsIdempotentAcrossRestartAndConcurrency`, `TestCollectDivergentReplayFailsClosed`, `TestCollectPersistsUnknownCostAsSQLNullWithoutInventingZero` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `filepath.Clean` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestStoreSchemaIsSeparateAppendOnlyAndVersioned`, `TestCollectExactReplayIsIdempotentAcrossRestartAndConcurrency`, `TestCollectDivergentReplayFailsClosed`, `TestCollectPersistsUnknownCostAsSQLNullWithoutInventingZero` |
| `strings.TrimSpace` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestStoreSchemaIsSeparateAppendOnlyAndVersioned`, `TestCollectExactReplayIsIdempotentAcrossRestartAndConcurrency`, `TestCollectDivergentReplayFailsClosed`, `TestCollectPersistsUnknownCostAsSQLNullWithoutInventingZero` |
| `errors.New` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestStoreSchemaIsSeparateAppendOnlyAndVersioned`, `TestCollectExactReplayIsIdempotentAcrossRestartAndConcurrency`, `TestCollectDivergentReplayFailsClosed`, `TestCollectPersistsUnknownCostAsSQLNullWithoutInventingZero` |
| `filepath.Dir` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestStoreSchemaIsSeparateAppendOnlyAndVersioned`, `TestCollectExactReplayIsIdempotentAcrossRestartAndConcurrency`, `TestCollectDivergentReplayFailsClosed`, `TestCollectPersistsUnknownCostAsSQLNullWithoutInventingZero` |
| `os.MkdirAll` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestStoreSchemaIsSeparateAppendOnlyAndVersioned`, `TestCollectExactReplayIsIdempotentAcrossRestartAndConcurrency`, `TestCollectDivergentReplayFailsClosed`, `TestCollectPersistsUnknownCostAsSQLNullWithoutInventingZero` |
| `fmt.Errorf` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestStoreSchemaIsSeparateAppendOnlyAndVersioned`, `TestCollectExactReplayIsIdempotentAcrossRestartAndConcurrency`, `TestCollectDivergentReplayFailsClosed`, `TestCollectPersistsUnknownCostAsSQLNullWithoutInventingZero` |
| `q.Add` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestStoreSchemaIsSeparateAppendOnlyAndVersioned`, `TestCollectExactReplayIsIdempotentAcrossRestartAndConcurrency`, `TestCollectDivergentReplayFailsClosed`, `TestCollectPersistsUnknownCostAsSQLNullWithoutInventingZero` |
| `q.Set` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestStoreSchemaIsSeparateAppendOnlyAndVersioned`, `TestCollectExactReplayIsIdempotentAcrossRestartAndConcurrency`, `TestCollectDivergentReplayFailsClosed`, `TestCollectPersistsUnknownCostAsSQLNullWithoutInventingZero` |

## State mutations and fallbacks

- local analytics values or derived SQLite state only; corrupt/unknown values remain errors or not-measured.
- There is no hidden broker polling, live-order fallback, or user-entered identifier path in this function.
- Missing, ambiguous, or corrupt evidence is preserved as an error/not-measured state or an explicit test failure.

## Safety conclusion

- Safe edit boundary: `internal/performance/store.go` function `openWithSchema` and its documented derived/test state.
- High-risk impact: analytics only; no order, toggle, broker, or LIVE capability.
