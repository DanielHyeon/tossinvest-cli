# Branch Test Map: `auditApprovedCandidateBoundaries`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | scan every supplied file | repository audit and temporary fixture audits | H1/H2 fixtures compile RED | focused GREEN |
| B2 | `_test.go` input is excluded from production taint | repository audit | prior GREEN | focused regression GREEN |
| B3 | malformed source fails the audit closed | `TestApprovedCandidateBoundaryParserAndDotImportControls` | Low finding requested a real fixture | focused GREEN |
| B4 | dot-imported candidate package is a direct reader | `TestApprovedCandidateBoundaryParserAndDotImportControls` | Low finding requested a real fixture | focused GREEN |
| B5 | qualified/aliased approved symbol is checked when not dot imported | reverse primitive strategy fixture | H1 fixture compile RED | focused GREEN |
| B6 | qualified/aliased approved symbol makes the package a direct reader | reverse primitive strategy fixture | H1 fixture compile RED | focused GREEN |
| B7 | scan every import in a production file | reverse primitive engine fixture | H1 fixture compile RED | focused GREEN |
| B8 | exact module import is recorded as root-relative `.` | parser/import graph contract | prior GREEN | focused regression GREEN |
| B9 | non-root import follows the module-prefix branch | reverse primitive engine fixture | H1 fixture compile RED | focused GREEN |
| B10 | module-internal import is added to the graph; external import is ignored | reverse primitive engine fixture | H1 fixture compile RED | focused GREEN |
| B11 | inspect every parsed production file for diagnostic accessors | helper/accessor laundering fixture | prior RED/GREEN | focused regression GREEN |
| B12 | accessor spelling records file detail but does not determine taint | helper/accessor laundering fixture | prior RED/GREEN | focused regression GREEN |
| B13 | every direct reader starts tainted | strategy direct-reader fixture | H1 fixture compile RED | focused GREEN |
| B14 | consider every production package as a reverse-taint source | engine fixture | H1 fixture compile RED | focused GREEN |
| B15 | direct reader needs no reverse traversal | strategy direct-reader fixture | H1 fixture compile RED | focused GREEN |
| B16 | transitive import path to a direct reader taints package regardless of return type/accessor | `TestApprovedCandidateBoundaryRejectsReversePrimitiveLaundering` | H1 fixture compile RED | focused GREEN |
| B17 | audit every direct reader's pure-boundary permission | repository/strategy fixture | prior RED/GREEN | focused regression GREEN |
| B18 | missing pure-boundary permission becomes a finding | existing repository boundary controls | prior RED/GREEN | focused regression GREEN |
| B19 | accessor files enrich a missing-boundary finding | helper/accessor laundering fixture | prior RED/GREEN | focused regression GREEN |
| B20 | empty pure-boundary rationale becomes a finding | existing boundary permission contract | prior RED/GREEN | focused regression GREEN |
| B21 | audit every general pure-boundary allowlist entry for staleness | reverse primitive fixture | H1 fixture compile RED | focused GREEN |
| B22 | non-direct entry is stale even if the package is transitively tainted | `TestApprovedCandidateBoundaryRejectsReversePrimitiveLaundering` | H1 fixture compile RED | focused GREEN |
| B23 | audit every tainted package's full authority closure | reverse primitive and journal fixtures | H1/H2 fixtures compile RED | focused GREEN |
| B24 | authority reach unconditionally becomes a full-path finding; a pure-boundary reason cannot exempt it | `TestApprovedCandidateBoundaryRejectsReversePrimitiveLaundering`, `TestApprovedCandidateAuthorityReachCannotBeAllowedByBoundaryReason` | reason-only bridge produced empty findings | focused GREEN |
| B25 | reverse-taint path is attached to authority finding | reverse primitive fixture | H1 fixture compile RED | focused GREEN |

Current repository success (empty findings) remains covered by `TestApprovedCandidateConsumersStayInsidePureBoundaries`.
The unbranched pure-boundary validation call is covered by `TestApprovedCandidatePureBoundaryRejectsInjectedAuthority`
(external import, package/local state, local interface, function parameter/field/variable/literal, pointer field,
and injected method/function calls) plus `TestApprovedCandidatePureBoundaryAllowsValueOnlyEvaluator`.
