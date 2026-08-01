# Function Logic Map: `compareAndAppendTrade`

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
| B1 | AST `if` at `internal/performance/store.go:286`: `if err != nil {` | local analytics values or derived SQLite state only; corrupt/unknown values remain errors or not-measured | condition determines the documented success/error/assertion path | `TestStoreSchemaIsSeparateAppendOnlyAndVersioned`, `TestCollectExactReplayIsIdempotentAcrossRestartAndConcurrency`, `TestCollectDivergentReplayFailsClosed`, `TestCollectPersistsUnknownCostAsSQLNullWithoutInventingZero` |
| B2 | AST `if` at `internal/performance/store.go:289`: `if exists {` | local analytics values or derived SQLite state only; corrupt/unknown values remain errors or not-measured | condition determines the documented success/error/assertion path | `TestStoreSchemaIsSeparateAppendOnlyAndVersioned`, `TestCollectExactReplayIsIdempotentAcrossRestartAndConcurrency`, `TestCollectDivergentReplayFailsClosed`, `TestCollectPersistsUnknownCostAsSQLNullWithoutInventingZero` |
| B3 | AST `if` at `internal/performance/store.go:290`: `if !equal {` | local analytics values or derived SQLite state only; corrupt/unknown values remain errors or not-measured | condition determines the documented success/error/assertion path | `TestStoreSchemaIsSeparateAppendOnlyAndVersioned`, `TestCollectExactReplayIsIdempotentAcrossRestartAndConcurrency`, `TestCollectDivergentReplayFailsClosed`, `TestCollectPersistsUnknownCostAsSQLNullWithoutInventingZero` |
| B4 | AST `if` at `internal/performance/store.go:295`: `if _, err := tx.ExecContext(ctx, insertTradeSQL, wanted...); err != nil {` | local analytics values or derived SQLite state only; corrupt/unknown values remain errors or not-measured | condition determines the documented success/error/assertion path | `TestStoreSchemaIsSeparateAppendOnlyAndVersioned`, `TestCollectExactReplayIsIdempotentAcrossRestartAndConcurrency`, `TestCollectDivergentReplayFailsClosed`, `TestCollectPersistsUnknownCostAsSQLNullWithoutInventingZero` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `tradeArgs` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestStoreSchemaIsSeparateAppendOnlyAndVersioned`, `TestCollectExactReplayIsIdempotentAcrossRestartAndConcurrency`, `TestCollectDivergentReplayFailsClosed`, `TestCollectPersistsUnknownCostAsSQLNullWithoutInventingZero` |
| `immutableRowEqual` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestStoreSchemaIsSeparateAppendOnlyAndVersioned`, `TestCollectExactReplayIsIdempotentAcrossRestartAndConcurrency`, `TestCollectDivergentReplayFailsClosed`, `TestCollectPersistsUnknownCostAsSQLNullWithoutInventingZero` |
| `fmt.Errorf` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestStoreSchemaIsSeparateAppendOnlyAndVersioned`, `TestCollectExactReplayIsIdempotentAcrossRestartAndConcurrency`, `TestCollectDivergentReplayFailsClosed`, `TestCollectPersistsUnknownCostAsSQLNullWithoutInventingZero` |
| `tx.ExecContext` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestStoreSchemaIsSeparateAppendOnlyAndVersioned`, `TestCollectExactReplayIsIdempotentAcrossRestartAndConcurrency`, `TestCollectDivergentReplayFailsClosed`, `TestCollectPersistsUnknownCostAsSQLNullWithoutInventingZero` |

## State mutations and fallbacks

- local analytics values or derived SQLite state only; corrupt/unknown values remain errors or not-measured.
- There is no hidden broker polling, live-order fallback, or user-entered identifier path in this function.
- Missing, ambiguous, or corrupt evidence is preserved as an error/not-measured state or an explicit test failure.

## Safety conclusion

- Safe edit boundary: `internal/performance/store.go` function `compareAndAppendTrade` and its documented derived/test state.
- High-risk impact: analytics only; no order, toggle, broker, or LIVE capability.
