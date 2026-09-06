# Branch Test Map: `Project`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | blank version, market mismatch, blank issuer, blank mapping version and a zero evaluation time each refuse | `TestProjectRefusesAnInvalidPolicyOrSnapshotScope` | no test reached this arm; all four existing tests supply a valid scope and `t.Fatal` on error | PASS |
| B2 | the declared fatal facts are walked | `TestProjectKeepsFatalAndLaneEvidenceSeparate` | existing coverage | PASS |
| B3 | a missing fatal fact is recorded whether or not it was required | `TestProjectRecordsMissingFatalEvidenceEvenWhenItIsOptional` | the optional case left no trace in `Fatal`; the field did not exist | PASS |
| B4 | a missing required fatal fact still blocks with its reason | `TestProjectRecordsMissingFatalEvidenceEvenWhenItIsOptional` | existing coverage | PASS |
| B5 | a fatal payload with no explicit blocked state fails closed | `TestProjectRejectsFatalPayloadWithoutExplicitBlockedState` | existing coverage | PASS |
| B6 | a blocking fatal fact outranks a high lane score | `TestProjectKeepsFatalAndLaneEvidenceSeparate` | existing coverage | PASS |
| B7 | required and optional lane requirements are walked together | `TestProjectRequiredMissingAndStaleFailClosedWhileOptionalIsTypedUnavailable` | existing coverage | PASS |
| B8 | a missing optional lane fact is marked unavailable without blocking | `TestProjectRequiredMissingAndStaleFailClosedWhileOptionalIsTypedUnavailable` | existing coverage | PASS |
| B9 | conflict, currency, unit and identity mismatches each make the lane ineligible | `TestProjectRequiredEvidenceConflictsAndScopeMismatchesFailClosed` | existing coverage | PASS |
| B10 | two missing required fatal facts come back ordered by kind, not by declaration order | `TestProjectOrdersFatalReasonsAndLaneRefusalsDeterministically` | no test produced two reasons; flipping the comparator changed nothing | PASS |
| B11 | two lane refusals come back ordered by kind, not by declaration order | `TestProjectOrdersFatalReasonsAndLaneRefusalsDeterministically` | no test produced two refusals; flipping the comparator changed nothing | PASS |
