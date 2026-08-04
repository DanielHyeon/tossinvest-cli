# Function Logic Map: `validateRiskBucketAdmission`

- Source: `internal/journal/risk_bucket.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| admission plan/decision | exact q_final transaction, owner and five ordered buckets | Guardian + `riskbucket.CalculateAdmission` | typed snapshot/owner mismatch |
| policy window | exact policy authority observed/fresh timestamps | opaque policy provenance | reject cross-wired or stale reference |
| snapshot window | exact journal snapshot observed/fresh timestamps | opaque snapshot provenance | reject cross-wired or stale reference |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | transaction identity, creation time or five-bucket cardinality is incomplete | none | snapshot mismatch | invalid admission suites |
| B2 | owner/account/market/symbol/lane/campaign identity is invalid | none | owner conflict | owner conflict suites |
| B3 | iterate the canonical five-dimensional order | none | validate every dimension | paired admission suites |
| B4 | key/version/digest/time/amount provenance differs | none | dimension-scoped snapshot mismatch | tamper + distinct-window tests |
| B5 | identify market and symbol binding dimensions | none | accumulate exact owner binding | market/symbol binding tests |
| B6 | market bucket must equal owner market | none | set market binding result | cross-market tests |
| B7 | symbol bucket must equal owner symbol | none | set symbol binding result | symbol substitution tests |
| B8 | market or symbol binding missing | none | owner bucket mismatch | binding tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `policyWindow` | select explicit policy times with legacy fallback | pure; no retry | AST + distinct-window tests |
| `snapshotWindow` | select explicit snapshot times with legacy fallback | pure; no retry | AST + distinct-window tests |
| `BoundEvidence` | obtain immutable policy/snapshot evidence | pure, fail closed | riskbucket authority tests |

## State mutations and fallbacks

- Pure validation only; no rows are written before this function succeeds.
- Legacy `ObservedAt/FreshUntil` remains accepted only when both explicit windows are absent.

## Safety conclusion

- Safe edit boundary: pre-transaction authority validation.
- High-risk impact: yes; an incorrect time binding can admit stale risk authority.
