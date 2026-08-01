# Function Logic Map: `TestAssessApprovedCandidateFailsClosed`

- Source: `internal/candidate/thresholdset_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| function inputs | repository AST/types or sealed candidate test fixtures, as declared in the signature | current source and persisted a047 base | violation/error/test failure; no approval is minted |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | follows the complete AST branch list in `ast.json`; table-tests invalid set, candidate life, current-life expiry, scope, and veto refusals | test-local collections only | deterministic finding/result | `TestAssessApprovedCandidateFailsClosed` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| parser/type/guard helpers named in `ast.json` | table-tests invalid set, candidate life, current-life expiry, scope, and veto refusals | no network, timeout, retry, or fallback; parse/type errors fail closed | CodeGraph + AST |

## State mutations and fallbacks

- Mutations are limited to test-local finding/path/type maps and synthetic fixtures; no production candidate, threshold, order, or account state is changed.

## Safety conclusion

- Safe edit boundary: `TestAssessApprovedCandidateFailsClosed` table-tests invalid set, candidate life, current-life expiry, scope, and veto refusals and returns findings or test assertions without granting authority.
- High-risk impact: no — test evidence for a high-risk boundary.
