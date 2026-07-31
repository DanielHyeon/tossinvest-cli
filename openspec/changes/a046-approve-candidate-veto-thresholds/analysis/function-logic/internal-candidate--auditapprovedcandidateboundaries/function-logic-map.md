# Function Logic Map: `auditApprovedCandidateBoundaries`

- Source: `internal/candidate/approved_consumer_guard_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `root`, `module`, `files` | repository root, declared module, repository-relative `.go` files outside candidate | current worktree via `goFilesUnder` | parse/read error is returned; audit fails closed |
| `approvedCandidateBoundaries` | direct pure readers only, each with non-empty rationale | a047 handoff review | missing/stale/empty permission becomes a finding |
| pure reader package shape | candidate-only imports, no package state, interface/function capability, mutable API shapes, or calls on injected parameters | a046 return-only evaluator contract | any capability-bearing shape becomes a finding |
| `forbidden` | canonical authority roots | `isolation_test.go` | every transitive reach becomes an unconditional finding; a046 has no bridge exemption |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | test file | none | skip | repository audit fixture |
| B2 | production file parse fails | none | wrapped error | repository audit fixture |
| B3 | dot import or qualified approved symbol | mark package as direct reader | continue scan | direct-reader positive controls |
| B4 | internal/external import | append only module-internal edge; external ignored | continue scan | dependency positive controls |
| B5 | package imports a direct reader transitively | mark every such dependent tainted, independent of accessor spelling or return type | continue | `TestApprovedCandidateBoundaryRejectsReversePrimitiveLaundering` |
| B6 | direct reader is missing/empty/stale in pure-boundary allowlist | append finding | continue | existing boundary audit controls |
| B7 | direct reader violates return-only pure boundary | append each capability/import/API finding | continue | `TestApprovedCandidatePureBoundaryRejectsInjectedAuthority` |
| B8 | tainted package reaches any authority root | append full-path finding unconditionally | continue | bridge privilege-expansion and reverse primitive fixtures |
| B9 | no violations | none | empty findings, nil error | repository audit and value-only fixture |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `parser.ParseFile` | parse imports and approved symbols in production Go | any parse error aborts audit | current AST |
| `candidateImportName`, `namesApprovedCandidateSymbol` | identify direct readers including aliases/dot imports | conservative direct-reader classification | positive controls |
| `transitiveDependency` | compute reverse-taint reach and reconstruct paths | breadth-first, cycle-safe, no retry | H1 RED fixture |
| `pureApprovedCandidateBoundaryViolations` | reject injected or hidden authority in direct reader syntax/imports | conservative AST check; candidate-only imports | H2 RED fixture |
| `transitiveAuthorityDependency` | detect any canonical authority root in dependency closure | breadth-first, cycle-safe, no retry and no exemption | H1 fixture |

## State mutations and fallbacks

- Local maps are accumulated from parsed source only. No production state, network, config, or trading capability is mutated.
- There is no primitive/bool fallback: package-level reverse taint follows imports, so type erasure cannot erase provenance.
- There is no bridge permission in a046. a047 must introduce typed Guardian/decision wiring together with exact equality between its approved root/path set and every reachable authority root/path; a reason-only package map is structurally forbidden by the unconditional test.

## Safety conclusion

- Safe edit boundary: strengthen the test-only repository audit and permission taxonomy; production candidate/order code remains unchanged.
- High-risk impact: yes — this is the fail-closed evidence preventing approved-candidate provenance from reaching execution/authority implicitly.
