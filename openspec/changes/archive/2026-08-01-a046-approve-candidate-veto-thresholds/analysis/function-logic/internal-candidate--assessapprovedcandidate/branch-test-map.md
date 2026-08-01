# Branch Test Map: `AssessApprovedCandidate`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | unapproved/zero threshold set is refused with zero result and typed `invalid_set` error | `TestAssessApprovedCandidateFailsClosed` / `invalid set` | compile RED: typed refusal contract absent | focused GREEN |
| B2 | a valid KR set cannot assess a US candidate | `TestAssessApprovedCandidateFailsClosed` / `wrong market` | compile RED: typed refusal contract absent | focused GREEN |
| B3 | incomplete candidate-life identity cannot mint an approved value | `TestAssessApprovedCandidateFailsClosed` / `invalid candidate life`, `zero first seen` | compile RED: typed refusal and life validation absent | focused GREEN |
| B4 | any dangerous measured veto returns zero plus typed `veto_raised` error | `TestAssessApprovedCandidateFailsClosed` / `dangerous` | compile RED: typed refusal absent; current source returns approval without checking `Passed()` | focused GREEN |
| B5 | any unmeasured veto returns zero plus typed `veto_unmeasured` error | `TestAssessApprovedCandidateFailsClosed` / `unmeasured` | compile RED: typed refusal absent; current source returns approval without checking `Passed()` | focused GREEN |
| B4/B5 mixed | all three raised; all three unmeasured; and raised+unmeasured mixtures return exact zero with deterministic typed kind and ordered codes | `TestAssessApprovedCandidateFailsClosedForAllAndMixedVetoStates` | coverage test added before guard fix; focused suite compile RED at missing guard helpers | focused GREEN |
| B6 | no raised/unmeasured code but `Passed()==false` is a defensive impossible-state refusal | structural audit: `Chase` owns exactly D3 states and `Passed`, `Raised`, `NotMeasured` iterate the same private order | N/A: public inputs cannot construct this inconsistent state | source/AST review required |
| Success | all three measured-clear vetoes return an immutable approved value with exact threshold provenance and deterministic candidate-life ID | `TestAssessApprovedCandidateReturnsPassWithImmutableProvenance` | compile RED: immutable accessors and full provenance absent | focused GREEN |
| B6a | same Key + same instant in another `time.Location` yields the same life ID; changed `FirstSeenAt` yields a different ID | `TestApprovedCandidateLifeIdentityUsesKeyAndFirstSeenAt` | compile RED: `CandidateLifeID` absent | focused GREEN |
