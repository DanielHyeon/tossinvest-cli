# Function Logic Map: `TestAssessApprovedCandidateReturnsPassWithImmutableProvenance`

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
| B1 | exact AST `if` at source line 402: `if err != nil {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B2 | exact AST `if` at source line 405: `if !got.Valid() \|\| !got.Chase().Passed() {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B3 | exact AST `if` at source line 408: `if got.Key() != input.Candidate.Key \|\| !got.FirstSeenAt().Equal(input.Candidate.FirstSeenAt) {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B4 | exact AST `if` at source line 412: `if got.State() != StateActive \|\| !got.LastSeenAt().Equal(input.Candidate.LastSeenAt) \|\|` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B5 | exact AST `if` at source line 416: `if got.ThresholdVersion() != set.Version() \|\| got.SetDigest() != set.SetDigest() \|\|` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B6 | exact AST `range` at source line 423: `for index := range typ.NumField() {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B7 | exact AST `if` at source line 425: `if field.IsExported() {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| parser/type/guard helpers named in `ast.json` | verifies approved identity, current-life proof, threshold provenance, and private fields | no network, timeout, retry, or fallback; parse/type errors fail closed | CodeGraph + AST |

## State mutations and fallbacks

- Mutations are limited to test-local finding/path/type maps and synthetic fixtures; no production candidate, threshold, order, or account state is changed.

## Safety conclusion

- Safe edit boundary: `TestAssessApprovedCandidateReturnsPassWithImmutableProvenance` verifies approved identity, current-life proof, threshold provenance, and private fields and returns findings or test assertions without granting authority.
- High-risk impact: no — test evidence for a high-risk boundary.
