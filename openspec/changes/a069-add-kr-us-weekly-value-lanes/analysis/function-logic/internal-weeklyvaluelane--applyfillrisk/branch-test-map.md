# Branch Test Map: `ApplyFillRisk`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | invalid/tampered state fails closed; valid conservative transferred floor | TestRiskStatePostMutationAndCopiedMapTamperFailClosed; TestFillRiskFloorFXDuplicateUnknownAndCorruptStatePreservation | yes | yes |
| B2 | unknown actual risk preserves fill and latches | TestFillRiskFloorFXDuplicateUnknownAndCorruptStatePreservation | existing | existing |
| B3 | duplicate fill idempotent | TestFillRiskFloorFXDuplicateUnknownAndCorruptStatePreservation | existing | existing |
| B4 | combined reservation+risk transition is all-or-nothing | TestApplyPositiveFillAtomicCommitsReservationAndRiskTogether | yes | yes |
| B5 | prior fill exact duplicate | TestFillRiskFloorFXDuplicateUnknownAndCorruptStatePreservation | existing | existing |
| B6 | prior fill conflict latches unknown | TestFillRiskFloorFXDuplicateUnknownAndCorruptStatePreservation | existing | existing |
| B7 | identity/plan mismatch preserves fill and latches | TestFillRiskFloorFXDuplicateUnknownAndCorruptStatePreservation | existing | existing |
| B8 | FX/time mismatch preserves fill and latches | TestFillRiskFloorFXDuplicateUnknownAndCorruptStatePreservation | existing | existing |
| B9 | accounting parse/held mismatch latches | TestFillRiskFloorFXDuplicateUnknownAndCorruptStatePreservation | existing | existing |
| B10 | arithmetic overflow or actual overage latches | TestFillRiskFloorFXDuplicateUnknownAndCorruptStatePreservation | existing | existing |
