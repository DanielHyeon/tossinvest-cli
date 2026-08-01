# Function Logic Map: `Store.recoverConflict`

- Source: `internal/optimization/store.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| token/actor | non-empty; actor equals persisted actor; token authenticates exact raw candidate payload | capability row plus HMAC | `ErrCapabilityInvalid` |
| candidate metadata | canonical times, 0/1 booleans, known category/origin, non-empty exact changes | persisted candidate and current registry | invalid/tampered value rejected |
| output | attempted changes copy plus latest verified snapshot/registry | immutable candidate/current store | read-only; no capability consumption |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | token or actor blank | none | invalid capability | actor tests |
| B2 | capability hash not found | none | invalid capability | invalid token test |
| B3 | candidate row query fails | none | error | DB error coverage |
| B4 | actor/time/boolean/MAC validation fails | none | invalid capability | tamper/cross-actor matrix |
| B5 | category or origin invalid | none | invalid capability | metadata tamper matrix |
| B6 | changes JSON invalid/empty | none | invalid capability | payload tamper test |
| B7-B8 | iterate attempted changes; field/category/control/timing/safety/option invalid | none | invalid capability | registry/payload tamper matrix |
| B9 | latest snapshot read fails | none | error | snapshot corruption coverage |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| candidate query and strict parsers | load immutable raw payload | one read; any corruption fails closed | tamper tests |
| `verifyCandidatePayloadMAC` | binds token to ID, actor, raw JSON, schedule and booleans | constant-time compare | MAC/replay test |
| registry validation | revalidates attempted controls against current authority | every change checked | configuration drift tests |
| current snapshot read | returns latest recovery context | digest verified | snapshot corruption test |

## State mutations and fallbacks

- Entire path is read-only and copies attempted changes. It does not consume the capability, apply settings, or rewrite candidate/history.
- There is no legacy or partial recovery fallback for invalid MAC/metadata/actor.

## Safety conclusion

- Safe edit boundary: authenticated read-only conflict explanation.
- High-risk impact: yes; must not disclose/normalize tampered or cross-actor candidate intent.
