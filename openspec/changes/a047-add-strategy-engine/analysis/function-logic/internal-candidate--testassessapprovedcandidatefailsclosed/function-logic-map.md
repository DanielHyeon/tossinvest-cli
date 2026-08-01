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
| B1 | exact AST `range` at source line 358: `for _, tc := range []struct {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B2 | exact AST `if` at source line 381: `if got != (ApprovedCandidate{}) {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B3 | exact AST `if` at source line 385: `if !errors.As(err, &approvalErr) {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B4 | exact AST `if` at source line 388: `if approvalErr.Kind() != tc.wantKind {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B5 | exact AST `if` at source line 391: `if !reflect.DeepEqual(approvalErr.Vetoes(), tc.wantVetoes) {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| parser/type/guard helpers named in `ast.json` | table-tests invalid set, candidate life, current-life expiry, scope, and veto refusals | no network, timeout, retry, or fallback; parse/type errors fail closed | CodeGraph + AST |

## State mutations and fallbacks

- Mutations are limited to test-local finding/path/type maps and synthetic fixtures; no production candidate, threshold, order, or account state is changed.

## Safety conclusion

- Safe edit boundary: `TestAssessApprovedCandidateFailsClosed` table-tests invalid set, candidate life, current-life expiry, scope, and veto refusals and returns findings or test assertions without granting authority.
- High-risk impact: no — test evidence for a high-risk boundary.
