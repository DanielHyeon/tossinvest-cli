# Branch Test Map: `namesVerdict`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | non-selector AST nodes do not terminate traversal | repository scan and all positive controls | guard helper compile RED | focused GREEN |
| B2 | aliased `candidate.AssessApprovedCandidate`, checked only through `err == nil`, is still a verdict read | `TestApprovedCandidateGuardRejectsAliasedErrNilOrderConsumer` | current `verdictSymbols` omission confirmed by review; focused suite compile RED before detector implementation | focused GREEN |
| B2a | the same fixture also names an order verb, proving the forbidden intersection | `TestApprovedCandidateGuardRejectsAliasedErrNilOrderConsumer` | focused suite compile RED before detector implementation | focused GREEN |
| B3 | helper-returned `Chase.Passed()` without candidate import remains detected | `TestTheVerdictDetectorSeesAReaderThatNeverImportsThisPackage` | existing GREEN | preserve |
| B4 | an unrelated file without verdict selectors stays unclassified | repository allowlist liveness pass | existing GREEN | preserve |
| PB1 | any package directly naming ApprovedCandidate/AssessApprovedCandidate must be explicitly allowlisted as a pure boundary | `TestApprovedCandidateConsumersStayInsidePureBoundaries` | compile RED: boundary audit absent | focused GREEN; current production reader/allowlist count both zero |
| PB2 | `.Valid()` and provenance accessors in the same package, or in a dependent package whose import path reaches the approved-candidate reader, remain inside the same boundary | `TestApprovedCandidateBoundaryDetectsHelperAccessorLaundering` | compile RED: symbol/accessor detectors absent | focused GREEN |
| PB3 | a direct or transitive path from an approved boundary to an order package is rejected | `TestApprovedCandidateBoundaryDetectsTransitiveOrderDependency` | compile RED: transitive dependency detector absent | focused GREEN |
| PB4 | every package that imports an approved boundary is tainted even after the boundary erases the value to bool/error; authority reach requires a separate bridge | `TestApprovedCandidateBoundaryRejectsReversePrimitiveLaundering` | security re-review H1 RED fixture | focused GREEN |
| PB5 | authority roots cannot drift from candidate isolation, including `journal` and all risk/domain/engine roots | `TestApprovedCandidateBoundaryDetectsAllAuthorityRoots`, `TestApprovedCandidateBoundaryDetectsTransitiveJournalDependency` | security re-review H2 RED fixture | focused GREEN |
