# Function Logic Map: `pureApprovedCandidateBoundaryViolations`

- Source: `internal/candidate/approved_consumer_guard_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `packageRel` | canonical repository-relative Go package | repository AST walk | an unrecognized direct reader uses the original strict value-only checker and fails if capability-bearing |
| `module` | non-empty module import prefix | `go.mod` resolved by the audit | type-check failure is reported as a finding |
| `fset`, `files` | parsed production syntax for one direct-reader package | `auditApprovedCandidateBoundaries` | malformed/unresolved syntax becomes a finding, never an allowance |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | package is the dedicated `internal/strategycandidate` read-only production sanitizer | none; route to the exact sanitizer checker | return its import/call/output findings | paired sanitizer guard fixtures and repository-wide boundary audit |
| B2 | every other direct reader, including `internal/strategy` | none; preserve the existing scalar-only checker | return strict type/call/capability findings | existing approved-boundary guard suite |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `typeCheckProductionCandidateSanitizer` | enforce the new bridge's exact read-only import/call/output contract | any deviation is a finding; no retry or fallback | RED sanitizer guard plus repository audit |
| `typeCheckPureApprovedCandidateBoundary` | preserve the existing value-only `strategy.SealApproved` boundary | any pointer, collection, callback, foreign import or non-accessor call is a finding | existing guard suite + AST |

## State mutations and fallbacks

- No production state is mutated. The function only selects one of two static-analysis policies.
- The sanitizer exception is exact by repository-relative package name and does not broaden the default policy.

## Safety conclusion

- Safe edit boundary: add one exact package branch before the existing default and require a dedicated adversarial guard that rejects execution/journal/broker/config writers, mutating candidate-store calls, non-approved outputs and arbitrary callbacks.
- High-risk impact: yes. This function is the enforcement point preventing an approved candidate from acquiring execution authority before sanitization.
