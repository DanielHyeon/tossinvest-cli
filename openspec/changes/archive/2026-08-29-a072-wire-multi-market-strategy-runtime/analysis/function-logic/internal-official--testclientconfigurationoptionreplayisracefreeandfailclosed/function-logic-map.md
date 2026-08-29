# Function Logic Map: `TestClientConfigurationOptionReplayIsRaceFreeAndFailClosed`

- Source: `internal/official/client_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| retained client and concurrent replay/read goroutines | sealed default-origin client | adversarial race fixture | race detector and assertions fail on any unsynchronized mutation |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | launch 32 replay/read pairs | concurrent option replay and provenance observations | join all goroutines | this race test |
| B2 | base URL changed after concurrent replay | none | `t.Fatalf` | this race test |
| B3 | official authority changed after concurrent replay | none | `t.Fatal` | this race test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| AuthorityOrigin/BaseURL and public Options | exercise read/write synchronization | all goroutines complete; no timeout/retry | `go test -race` RED/GREEN |

## State mutations and fallbacks

- Mutations target only the in-memory test client; sealed configuration must make every replay a no-op.

## Safety conclusion

- Safe edit boundary: deterministic start barrier maximizes the original TOCTOU race window.
- High-risk impact: validates race-free official origin/read boundary.
