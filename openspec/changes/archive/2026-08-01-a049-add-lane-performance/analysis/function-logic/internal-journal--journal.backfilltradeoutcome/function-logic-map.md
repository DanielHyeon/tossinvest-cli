# Function Logic Map: `Journal.BackfillTradeOutcome`

- Source: `internal/journal/trade_outcomes.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `positionID` | exact CLOSED position missing an outcome | journal projection | existing row returns `ErrTradeOutcomeExists`; invalid evidence returns `ErrInvalidRequest` |
| `model` | configured validated cost model | `internal/costs` | refuses rather than writing free/guessed cost |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | AST `if` at line 758: `if _, err := j.TradeOutcomeOf(ctx, id); err == nil {` | immutable outcome insert only after exact recomputation; conflicts fail closed | explicit success/error/continue path; no invented fallback | `TestABackfillRecoversTheGapAndRefusesToRewriteIt` |
| B2 | AST `else` at line 760: `} else if !errors.Is(err, ErrTradeOutcomeNotFound) {` | immutable outcome insert only after exact recomputation; conflicts fail closed | explicit success/error/continue path; no invented fallback | `TestABackfillRecoversTheGapAndRefusesToRewriteIt` |
| B3 | AST `if` at line 760: `} else if !errors.Is(err, ErrTradeOutcomeNotFound) {` | immutable outcome insert only after exact recomputation; conflicts fail closed | explicit success/error/continue path; no invented fallback | `TestABackfillRecoversTheGapAndRefusesToRewriteIt` |
| B4 | AST `if` at line 765: `if !ok {` | immutable outcome insert only after exact recomputation; conflicts fail closed | explicit success/error/continue path; no invented fallback | `TestABackfillRecoversTheGapAndRefusesToRewriteIt` |
| B5 | AST `if` at line 769: `if _, err := j.db.ExecContext(ctx, \`` | immutable outcome insert only after exact recomputation; conflicts fail closed | explicit success/error/continue path; no invented fallback | `TestABackfillRecoversTheGapAndRefusesToRewriteIt` |
| B6 | AST `if` at line 779: `if isUniqueViolation(err) {` | immutable outcome insert only after exact recomputation; conflicts fail closed | explicit success/error/continue path; no invented fallback | `TestABackfillRecoversTheGapAndRefusesToRewriteIt` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `TradeOutcomeOf` | enforce no rewrite | read only | current HEAD |
| `computeTradeOutcome` | exact frozen outcome and cost | no fallback | function map + tests |
| `db.ExecContext` | append once | uniqueness conflict is typed | current HEAD |

## State mutations and fallbacks

- Backfill remains append-only. It never updates legacy rows and never recomputes an existing outcome.

## Safety conclusion

- Safe edit boundary: include the same exact cost scalar as the close-time insert.
- High-risk impact: yes — journal mutation, protected by uniqueness and migration tests.
