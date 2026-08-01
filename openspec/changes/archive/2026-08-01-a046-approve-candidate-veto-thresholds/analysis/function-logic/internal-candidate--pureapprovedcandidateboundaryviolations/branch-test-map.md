# Branch Test Map: `pureApprovedCandidateBoundaryViolations`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | branchless wrapper delegates the complete package and shared file set to the type-checked guard and returns its findings unchanged | all pure-boundary tests | prior spelling guard | focused GREEN |
| P1 | aliases/named pointer/map/slice/chan, any/error/signature/interface, embedded capabilities, candidate.Source, unsafe, and generic capability fail recursively | `TestApprovedCandidatePureBoundaryRejectsAliasedAndNamedCapabilities`, `TestApprovedCandidatePureBoundaryRejectsNestedAndGenericCapabilities` | prior audit returned no type-contract finding | focused GREEN |
| P2 | method/receiver/init/function literal/go/defer/send fail | `TestApprovedCandidatePureBoundaryRejectsMutationAndExecutionForms` | prior audit returned only shallow parameter findings | focused GREEN |
| P3 | ApprovedCandidate accessors and safe builtin on scalar/value struct/fixed array succeed | `TestApprovedCandidatePureBoundaryAllowsScalarStructAndFixedArray` | new success fixture | focused GREEN |
| P4 | method value, type-asserted call, free helper and package call fail | nested/call table plus `TestApprovedCandidatePureBoundaryRejectsScalarPackageStateAndPackageCall` | method/free/package calls were accepted | focused GREEN |
| P5 | dereference/index/selector assignment and channel send fail | mutation fixture and external-selector fixture | mutation forms were accepted | focused GREEN |
| P6 | scalar package state fails while local scalar state and value-only result stay accepted | package-state test plus both success fixtures | scalar package var produced no finding | focused GREEN |
