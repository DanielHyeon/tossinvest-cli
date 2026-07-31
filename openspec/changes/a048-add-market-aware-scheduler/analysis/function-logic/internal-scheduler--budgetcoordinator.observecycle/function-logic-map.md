# Function Logic Map: `BudgetCoordinator.ObserveCycle`

- Source: `internal/scheduler/budget.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| observation | exact endpoint budget response | official rate-budget adapter | not ingested when cycle binding fails |
| cycle | opaque capability returned before the request started | `BeginObservation` | zero/forged/replayed/cross-key/coordinator/generation returns false |
| cycle record | active one-shot record with completion watermark | endpoint state under mutex | deleted before ingestion so replay cannot gain authority |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | nil coordinator or zero cycle | none | false | zero/nil binding test |
| B2 | endpoint missing or coordinator/key/generation differs | none | false | cross-scope tests |
| B3 | active capability missing or record generation differs | none | false | forge/replay tests |
| success | every binding matches | consumes active cycle, preserves issued memory, ingests response with causal watermark | true | held-response, delta, boundary tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `sha256.Sum256` | verify exact endpoint key binding | pure | cross-key test |
| `observeLocked` | update evidence and reconcile only watermark-covered completions | invalid/conflicting evidence remains fail-closed | chronology/reset tests |
| coordinator mutex | make validation, one-shot consumption and ingestion atomic | defer unlock | race test |

## State mutations and fallbacks

- Invalid capability attempts do not consume the valid original cycle.
- A valid cycle is consumed even if its response evidence is old or invalid; replay never retries reconciliation authority.

## Safety conclusion

- Safe edit boundary: dormant low-priority budget observation path only.
- High-risk impact: yes, because replay/cross-scope acceptance could release reserved capacity.
