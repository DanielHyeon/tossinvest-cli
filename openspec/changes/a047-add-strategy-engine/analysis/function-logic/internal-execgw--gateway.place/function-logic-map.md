# Function Logic Map: `Gateway.Place`

- Source: `internal/execgw/gateway.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
| --- | --- | --- | --- |
| `PlaceRequest.Intent` | normalized official order intent | `orderintent` and persisted Guardian preimage | `submit` rejects mismatch before broker call |
| `PlaceRequest.Decision` | unexpired, unspent journal reference | Guardian/journal | typed gateway refusal |
| gateway dependencies | official trading service, journal, entry/protection/mode gates | startup interlock | constructor/startup refusal |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
| --- | --- | --- | --- | --- |
| B1 | `g.preflight != nil` | installs `CheckPlace` callback on the local plan | continue to common plan construction | symbol/protection preflight tests |
| Scenario | `g.preflight == nil` | leaves the local callback nil | continue to common plan construction | ordinary `Gateway.Place` tests |
| Scenario | plan is complete | delegates once to `g.submit`; all later checks and broker outcomes belong to that callee | `Outcome` or error from `submit` | gateway round-trip and in-doubt tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
| --- | --- | --- | --- |
| `preflight.CheckPlace` | broker/capability fail-closed check | before mutation | CodeGraph + AST |
| `PreviewPlace` | obtain official confirmation token for exact normalized request | no alternate mutation path | CodeGraph + AST |
| `PlaceWireBody` | persist exact replay bytes | serialization failure prevents submit | CodeGraph + AST |
| `g.submit` | verify Guardian, mode, protection, reservation, durability and dispatch | places once; unknown outcome is in-doubt, not automatic retry | CodeGraph + AST |

## State mutations and fallbacks

- Normalizes routing identity but derives exposure class from the actual BUY
  shape rather than caller labels.
- Overwrites caller client-order ID with the decision-derived key.
- Only the official trading service is reachable; no paper/shadow/canary path.

## Safety conclusion

- Safe edit boundary: unchanged for a047. The orchestrator hands it the exact
  intent and Guardian reference only after submit-time manifest revalidation.
- High-risk impact: yes.
