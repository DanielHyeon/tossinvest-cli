# Function Logic Map: `TestTheSeamsWithVerbExemptionsAreEnumeratedAndArgued`

- Source: `internal/console/orders_static_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| capability map | every `Options` seam with exact methods/exemptions | static safety model | unlisted/widened seam fails |
| update seams | fixed updater plus engine/verify guards, no broker capability | `Options` types | empty/unargued exemption fails |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B8 | enumerate exemption-bearing seams and compare exact names/reasons | none | test failure |
| B9-B10 | ensure no stale expected seam | none | test failure |
| B11-B14 | scan every capability/method against forbidden account verbs | none | test failure |
| B15-B17 | validate fields/signatures and reject unexpected account reachability | none | test failure |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `consoleCapabilities` | closed capability model | compile/static data | AST |
| reflection/AST helpers | inspect method sets and signatures | deterministic | existing static suite |

## State mutations and fallbacks

- `SystemUpdater` permits only `Inspect` and `Install`; lock/check funcs carry no account methods.
- Verb exemptions are exact and justified rather than weakening the global forbidden list.

## Safety conclusion

- Safe edit boundary: enumerate the three new narrow seams and exact verb spellings.
- High-risk impact: yes — widened injected capabilities could bypass order/account boundaries.
