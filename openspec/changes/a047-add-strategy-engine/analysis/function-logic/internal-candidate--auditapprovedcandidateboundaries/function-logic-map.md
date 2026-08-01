# Function Logic Map: `auditApprovedCandidateBoundaries`

- Source: `internal/candidate/approved_consumer_guard_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| function inputs | repository AST/types or sealed candidate test fixtures, as declared in the signature | current source and persisted a047 base | violation/error/test failure; no approval is minted |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | exact AST `range` at source line 188: `for _, rel := range files {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B2 | exact AST `if` at source line 189: `if strings.HasSuffix(rel, "_test.go") {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B3 | exact AST `if` at source line 194: `if err != nil {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B4 | exact AST `if` at source line 202: `if importsCandidate && candidatePkg == "." {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B5 | exact AST `else` at source line 204: `} else if namesApprovedCandidateSymbol(parsed, candidatePkg) {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B6 | exact AST `if` at source line 204: `} else if namesApprovedCandidateSymbol(parsed, candidatePkg) {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B7 | exact AST `range` at source line 207: `for _, spec := range parsed.Imports {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B8 | exact AST `if` at source line 209: `if path == module {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B9 | exact AST `else` at source line 211: `} else if strings.HasPrefix(path, module+"/") {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B10 | exact AST `if` at source line 211: `} else if strings.HasPrefix(path, module+"/") {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B11 | exact AST `range` at source line 216: `for _, record := range parsedFiles {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B12 | exact AST `if` at source line 217: `if namesApprovedCandidateAccessor(record.file) {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B13 | exact AST `range` at source line 222: `for packageRel := range directReaders {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B14 | exact AST `range` at source line 226: `for packageRel := range allPackages {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B15 | exact AST `if` at source line 227: `if directReaders[packageRel] {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B16 | exact AST `if` at source line 235: `if reachesReader {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B17 | exact AST `range` at source line 242: `for packageRel := range directReaders {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B18 | exact AST `if` at source line 244: `if !allowed {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B19 | exact AST `if` at source line 246: `if len(accessorFiles[packageRel]) != 0 {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B20 | exact AST `if` at source line 252: `if allowed && strings.TrimSpace(reason) == "" {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B21 | exact AST `range` at source line 258: `for packageRel := range approvedCandidateBoundaries {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B22 | exact AST `if` at source line 259: `if !directReaders[packageRel] {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B23 | exact AST `range` at source line 263: `for packageRel := range tainted {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B24 | exact AST `if` at source line 266: `if bad {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B25 | exact AST `if` at source line 268: `if len(taintedFrom[packageRel]) != 0 {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| parser/type/guard helpers named in `ast.json` | audits repository packages, sanitizer boundaries, and reachable authority paths | no network, timeout, retry, or fallback; parse/type errors fail closed | CodeGraph + AST |

## State mutations and fallbacks

- Mutations are limited to test-local finding/path/type maps and synthetic fixtures; no production candidate, threshold, order, or account state is changed.

## Safety conclusion

- Safe edit boundary: `auditApprovedCandidateBoundaries` audits repository packages, sanitizer boundaries, and reachable authority paths and returns findings or test assertions without granting authority.
- High-risk impact: yes — static guard logic protects the candidate-to-strategy authority boundary.
