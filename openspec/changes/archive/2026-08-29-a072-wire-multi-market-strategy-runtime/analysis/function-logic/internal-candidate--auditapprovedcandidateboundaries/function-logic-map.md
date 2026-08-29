# Function Logic Map: `auditApprovedCandidateBoundaries`

- Source: `internal/candidate/approved_consumer_guard_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| root/module/files | repository production Go graph | module root scanner | parse/build constraint errors fail closed |
| build selection | current production build constraints | `go/build.MatchFile` | excluded test seams do not enter production graph |
| raw candidate readers | exact approved-candidate symbols/accessors | AST audit | unallowlisted reader is a finding |
| sanitizer stop set | strategy, strategycandidate, strategyengine, strategyflow | audited boundary registry | downstream opaque consumers are not reverse-tainted |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B5 | enumerate files, skip tests, enforce build constraints | read-only source scan | return parse/build error or skip excluded file | production/test-seam boundary tests |
| B6-B12 | parse source, identify direct candidate readers and module imports | in-memory graph only | record exact package edges | alias/dot-import tests |
| B13-B18 | identify accessor laundering and propagate taint until an audited sanitizer | in-memory graph only | record tainted source path | sanitizer-stop tests |
| B19-B24 | audit direct readers, reasons and pure-boundary violations | append findings | continue complete audit | pure-boundary positive/negative controls |
| B25-B27 | audit tainted packages for authority interfaces/dependencies | append exact path findings | return complete findings | authority-root matrix |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `build.Default.MatchFile` | apply real production build constraints | deterministic filesystem metadata read | test-seam regression |
| `parser.ParseFile` | build syntax/import graph | fail closed on malformed source | parser control test |
| `transitiveDependencyWithStop` | prevent reverse taint past opaque sanitizer | bounded graph traversal | sanitizer stop tests |
| `pureApprovedCandidateBoundaryViolations` | enforce allowed direct-reader shapes | deterministic AST rules | pure boundary test matrix |

## State mutations and fallbacks

- Reads source and builds in-memory maps only; no production mutation.
- Build-tag exclusion narrows the graph to shipped production files and cannot allow a production file through.

## Safety conclusion

- Safe edit boundary: repository static authority audit.
- High-risk impact: yes; false negatives could hide an execution-authority edge.
