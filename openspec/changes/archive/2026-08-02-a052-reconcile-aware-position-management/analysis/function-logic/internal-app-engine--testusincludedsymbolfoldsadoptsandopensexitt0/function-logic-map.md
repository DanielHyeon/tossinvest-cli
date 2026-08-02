# Function Logic Map: `TestUSIncludedSymbolFoldsAdoptsAndOpensExitT0`

- Source: `internal/app/engine/adoption_include_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| Function contract | `TestUSIncludedSymbolFoldsAdoptsAndOpensExitT0(params=1, results=0)` | current Go signature and callers | errors/unknown values propagate without inventing effective state |
| Runtime or persisted state | values supplied by the owning engine/journal/read boundary | current HEAD plus approved a052 spec | unavailable or ambiguous facts remain unknown/deferred |
| Safety boundary | read-only projection, or pre-journal fail-closed validation | a052 design and TossOS safety invariants | no live order, reconciliation resolution, or operating-toggle mutation |

## Branches and early returns

| Branch | AST kind/location | Condition/control path | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|---|
| B1 | `if` at line 70 | `if cycle.Err != nil \|\| cycle.Folded != 1 \|\| cycle.Adopted != 1 \|\| cycle.Unmanaged != 0 {` | local/read-model state only; see AST assignments | continues through the function contract | TestUSIncludedSymbolFoldsAdoptsAndOpensExitT0 |
| B2 | `if` at line 75 | `if !p.Adopted() \|\| !p.ExitEligible() \|\| provenance != positionpolicy.ProvenanceExternalAdoption \|\|` | local/read-model state only; see AST assignments | continues through the function contract | TestUSIncludedSymbolFoldsAdoptsAndOpensExitT0 |
| B3 | `if` at line 80 | `if err != nil \|\| adoption.ObservedPrice != "200" \|\| adoption.SyntheticStop != "190" {` | local/read-model state only; see AST assignments | continues through the function contract | TestUSIncludedSymbolFoldsAdoptsAndOpensExitT0 |
| B4 | `if` at line 84 | `if err != nil \|\| exitState.EntryPrice != "200" \|\| exitState.HighWater != "200" \|\|` | local/read-model state only; see AST assignments | continues through the function contract | TestUSIncludedSymbolFoldsAdoptsAndOpensExitT0 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `newDriverHarness` | execute the explicit dependency at line 64 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `includeOnly` | execute the explicit dependency at line 65 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `h.holdsMarket` | execute the explicit dependency at line 67 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `h.cycle` | execute the explicit dependency at line 69 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `t.Fatalf` | execute the explicit dependency at line 71 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `h.positionMarket` | execute the explicit dependency at line 73 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `positionpolicy.ClassifyProvenance` | execute the explicit dependency at line 74 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `p.Adopted` | execute the explicit dependency at line 75 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `p.ExitEligible` | execute the explicit dependency at line 75 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `h.journal.AdoptionOf` | execute the explicit dependency at line 79 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `t.Context` | execute the explicit dependency at line 79 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `h.journal.ExitState` | execute the explicit dependency at line 83 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |

## State mutations and fallbacks

- AST records 7 assignment(s), 0 return statement(s), and 0 goroutine launch(es).
- Fallback is fail-closed: missing, mismatched, unavailable, or ambiguous operational truth is not promoted to managed/effective/actionable state.
- The a052 path adds no reconciliation-resolution command and does not authorize a live trade.

## Safety conclusion

- Safe edit boundary: Keep the change inside transport/read-model/rendering behavior; effective runtime facts must never be inferred from desired configuration.
- High-risk impact: no direct order side effect; nevertheless unknown data and account identity remain fail-closed/masked.
