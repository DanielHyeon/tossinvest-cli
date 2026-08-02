# Function Logic Map: `httpAPIReader.Optimization`

- Source: `cmd/tossctl/httpapi_reader.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| Function contract | `httpAPIReader.Optimization(params=1, results=2)` | current Go signature and callers | errors/unknown values propagate without inventing effective state |
| Runtime or persisted state | values supplied by the owning engine/journal/read boundary | current HEAD plus approved a052 spec | unavailable or ambiguous facts remain unknown/deferred |
| Safety boundary | read-only projection, or pre-journal fail-closed validation | a052 design and TossOS safety invariants | no live order, reconciliation resolution, or operating-toggle mutation |

## Branches and early returns

| Branch | AST kind/location | Condition/control path | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|---|
| B1 | `if` at line 391 | `if r == nil \|\| r.optimization == nil {` | local/projection assignment | early return/error nearby | go test ./cmd/tossctl |
| B2 | `if` at line 395 | `if err != nil {` | local/projection assignment | early return/error nearby | go test ./cmd/tossctl |
| B3 | `if` at line 398 | `if r.adoptionDesired == nil {` | local/projection assignment | early return/error nearby | go test ./cmd/tossctl |
| B4 | `if` at line 402 | `if err != nil {` | local/projection assignment | early return/error nearby | go test ./cmd/tossctl |
| B5 | `if` at line 410 | `if runtime.EffectiveKnown {` | local/projection assignment | early return/error nearby | go test ./cmd/tossctl |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `errors.New` | explicit dependency at line 392 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `r.optimization.Read` | explicit dependency at line 394 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `r.adoptionDesired` | explicit dependency at line 401 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `positionpolicy.NewAdoptionSettings` | explicit dependency at line 405 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `r.readManagementRuntime` | explicit dependency at line 407 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `adoptionSettingsFrom` | explicit dependency at line 408 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |

## State mutations and fallbacks

- AST records 7 assignment(s), 5 return statement(s), and 0 goroutine launch(es).
- Fallback is fail-closed: missing, mismatched, unavailable, or ambiguous operational truth is not promoted to managed/effective/actionable state.
- The a052 path adds no reconciliation-resolution command and does not authorize a live trade.

## Safety conclusion

- Safe edit boundary: Keep the change inside transport/read-model/rendering behavior; effective runtime facts must never be inferred from desired configuration.
- High-risk impact: no direct order side effect; nevertheless unknown data and account identity remain fail-closed/masked.
