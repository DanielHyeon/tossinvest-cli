# Function Logic Map: `auditApprovedCandidateBoundaries`

- Source: `internal/candidate/approved_consumer_guard_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `root`, `module`, `files` | repository root, declared module, repository-relative `.go` files outside candidate | current worktree via `goFilesUnder` | parse/read error is returned; audit fails closed |
| `approvedCandidateBoundaries` | direct pure readers only, each with non-empty rationale | a047 handoff review | missing/stale/empty permission becomes a finding |
| `approvedCandidateAuthorityBridges` | tainted dependents intentionally joining decision to authority, each with non-empty rationale | explicit Guardian/decision bridge review | authority reach without this separate permission becomes a finding |
| `forbidden` | canonical authority roots | `isolation_test.go` | transitive reach becomes a finding unless bridge is explicitly approved |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | test file | none | skip | repository audit fixture |
| B2 | production file parse fails | none | wrapped error | repository audit fixture |
| B3 | dot import or qualified approved symbol | mark package as direct reader | continue scan | direct-reader positive controls |
| B4 | internal/external import | append only module-internal edge; external ignored | continue scan | dependency positive controls |
| B5 | package imports a direct reader transitively | mark every such dependent tainted, independent of accessor spelling or return type | continue | `TestApprovedCandidateBoundaryRejectsReversePrimitiveLaundering` |
| B6 | direct reader is missing/empty/stale in pure-boundary allowlist | append finding | continue | existing boundary audit controls |
| B7 | tainted package reaches an authority root and lacks a separate bridge approval | append full-path finding | continue | reverse primitive fixture; journal fixture |
| B8 | bridge approval is empty, stale, or package reaches no authority root | append finding | continue | bridge permission controls |
| B9 | no violations | none | empty findings, nil error | `TestApprovedCandidateConsumersStayInsidePureBoundaries` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `parser.ParseFile` | parse imports and approved symbols in production Go | any parse error aborts audit | current AST |
| `candidateImportName`, `namesApprovedCandidateSymbol` | identify direct readers including aliases/dot imports | conservative direct-reader classification | positive controls |
| `transitiveDependency` | compute reverse-taint reach and reconstruct paths | breadth-first, cycle-safe, no retry | H1 RED fixture |
| `transitiveAuthorityDependency` | detect any canonical authority root in dependency closure | breadth-first, cycle-safe, no retry | H1/H2 RED fixtures |

## State mutations and fallbacks

- Local maps are accumulated from parsed source only. No production state, files, network, config, or trading capability is mutated.
- There is no primitive/bool fallback: package-level reverse taint follows imports, so type erasure cannot erase provenance.

## Safety conclusion

- Safe edit boundary: strengthen the test-only repository audit and permission taxonomy; production candidate/order code remains unchanged.
- High-risk impact: yes — this is the fail-closed evidence preventing approved-candidate provenance from reaching execution/authority implicitly.
