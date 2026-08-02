# Function Logic Map: `adoptionCurrency`

- Source: `internal/app/engine/adoption.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| Function contract | `adoptionCurrency(params=1, results=2)` | current Go signature and callers | errors/unknown values propagate without inventing effective state |
| Runtime or persisted state | values supplied by the owning engine/journal/read boundary | current HEAD plus approved a052 spec | unavailable or ambiguous facts remain unknown/deferred |
| Safety boundary | read-only projection, or pre-journal fail-closed validation | a052 design and TossOS safety invariants | no live order, reconciliation resolution, or operating-toggle mutation |

## Branches and early returns

| Branch | AST kind/location | Condition/control path | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|---|
| B1 | `switch` at line 300 | `switch strings.ToLower(strings.TrimSpace(market)) {` | none beyond calls listed below | early return/error path nearby | TestUSAdoptionRefusesWrongOrEmptyQuoteCurrency; TestUSIncludedSymbolFoldsAdoptsAndOpensExitT0 |
| B2 | `case` at line 301 | `case "kr":` | none beyond calls listed below | early return/error path nearby | TestUSAdoptionRefusesWrongOrEmptyQuoteCurrency; TestUSIncludedSymbolFoldsAdoptsAndOpensExitT0 |
| B3 | `case` at line 303 | `case "us":` | none beyond calls listed below | early return/error path nearby | TestUSAdoptionRefusesWrongOrEmptyQuoteCurrency; TestUSIncludedSymbolFoldsAdoptsAndOpensExitT0 |
| B4 | `case` at line 305 | `default:` | none beyond calls listed below | early return/error path nearby | TestUSAdoptionRefusesWrongOrEmptyQuoteCurrency; TestUSIncludedSymbolFoldsAdoptsAndOpensExitT0 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `strings.ToLower` | execute the explicit dependency at line 300 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `strings.TrimSpace` | execute the explicit dependency at line 300 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |

## State mutations and fallbacks

- AST records 0 assignment(s), 3 return statement(s), and 0 goroutine launch(es).
- Fallback is fail-closed: missing, mismatched, unavailable, or ambiguous operational truth is not promoted to managed/effective/actionable state.
- The a052 path adds no reconciliation-resolution command and does not authorize a live trade.

## Safety conclusion

- Safe edit boundary: Only the approved read/projection or fail-closed adoption boundary may change; no order placement, reconciliation resolution, or live toggle mutation is authorized.
- High-risk impact: yes; adoption provenance, reconciliation blocking, or persisted position lifecycle is money-sensitive.
