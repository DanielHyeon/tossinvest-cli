# Function Logic Map: `namesVerdict`

- Source: `internal/candidate/consumer_guard_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `file` | parsed non-nil Go AST outside `internal/candidate` | `goFilesUnder` + `parser.ParseFile` caller | parse failures are fatal in caller; detector must not silently skip AST nodes |
| `pkg` | local import name for `internal/candidate`, or empty when absent | `candidateImportName` resolves aliases | empty disables only qualified package selectors; unqualified helper selectors are still scanned |
| `verdictSymbols` | every exported type/constructor that can mint or carry a chase verdict | a046/a047 handoff surface | omission is fail-open; every added approval constructor needs a positive control |
| `unqualifiedVerdictSelectors` | verdict fields/predicates usable after same-package helper laundering | current AST guard architecture | generic spellings are kept narrow to avoid unrelated type collisions |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | inspected node is not a selector | none | continue traversal | existing repository scan |
| B2 | selector is `pkg.<verdictSymbols member>` | local `found=true` only | caller treats the file as a verdict reader | aliased `AssessApprovedCandidate` + `err==nil` + order positive control |
| B3 | selector name is an unqualified verdict predicate/field | local `found=true` only | catches same-package helper laundering without an import | existing `Chase.Passed` positive control |
| B4 | no selector matches | none | false; file is not granted reader status | repository allowlist liveness check |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `ast.Inspect` | walk every selector in one parsed file | pure in-memory traversal; no timeout/retry/I/O | CodeGraph + AST |
| `candidateImportName` (caller-side) | bind aliases before this function runs | dot/blank imports are not treated as usable candidate identifiers | current source + alias positive control |
| approved-candidate package-boundary guard | close `.Valid()` and other accessor laundering across files and dependent packages, and reject forbidden order dependency closure | separate new leaf helpers; cannot prove arbitrary cross-package primitive/boolean dataflow | security re-review High finding |

## State mutations and fallbacks

- Only the local `found` flag changes; no repository or runtime state is mutated.
- The detector does not compile or type-check fixtures. Its positive controls must therefore exercise exact parsed selector shapes.
- Direct `AssessApprovedCandidate` is detected through `verdictSymbols`; it cannot be inferred from the result or `err == nil` alone.
- Same-package `.Valid()` laundering is covered once any file in that package names the approved type/constructor. Cross-package accessor laundering is covered when the accessor package has an import path to that approved-candidate reader.
- Conversion to an unrelated primitive/boolean that erases the approved-candidate accessor lineage is outside this AST guard’s evidence. The explicit pure boundary plus transitive no-order dependency rule is the handoff point for a047.

## Safety conclusion

- Safe edit boundary: extend detector data and add positive controls/package dependency analysis only; do not add production order authority.
- High-risk impact: yes. This static guard is the reverse dependency control for a future LIVE entry consumer and must fail closed on unknown approved-candidate readers.
