# Branch Test Map: `ParkerConservativeLane.Evaluate`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | zero/forged opaque inputs | `TestParkerRejectsZeroAndForgedProofsWithoutPanic` | yes | yes |
| B2 | approval activation/current-life | `TestParkerRequiresApprovalActivatedAndCandidateCurrentlyActive` | yes | yes |
| B3 | evaluated session open/cutoff | `TestParkerSessionUsesInjectedEvaluationTimeAndInclusiveCutoff` | yes | yes |
| B4 | exact frozen gate order/boundaries | `TestParkerFrozenGateBoundariesAndRefusalOrder` | yes | yes |
| B5 | exact 15s and +1ns | `TestParkerSignalAgeUsesNanosecondHalfOpenExpiry` | yes | yes |
| B6 | pass and derived evidence | `TestParkerConservativeLaneGoldenContractFixture` | yes | yes |
| B7 | scope guard | zero proofs | existing | yes |
| B8 | source guard | zero proofs | yes | yes |
| B9 | market bundle guard | zero proofs | yes | yes |
| B10 | candidate clock guard | candidate boundary table | yes | yes |
| B11 | session calendar guard | session boundaries | yes | yes |
| B12 | official bar proof guard | zero proofs | yes | yes |
| B13 | bar/evaluation clock guard | session/age boundaries | yes | yes |
| B14 | state proof guard | zero proofs | yes | yes |
| B15 | position proof guard | zero proofs | yes | yes |
| B16 | bar decimal guard | golden/integrity tests | existing | yes |
| B17 | indicator decimal guard | sealer laundering tests | yes | yes |
| B18 | VWAP gate | frozen gate table | yes | yes |
| B19 | slope gate | frozen gate table | yes | yes |
| B20 | EMA9 gate | frozen gate table | yes | yes |
| B21 | LVN gate | frozen gate table | yes | yes |
| B22 | tangled gate | frozen gate table | yes | yes |
| B23 | optional expansion presence | frozen gate table | yes | yes |
| B24 | expansion threshold | frozen gate table | yes | yes |
| B25 | RR threshold | derived golden evidence | yes | yes |
| B26 | optional HVN presence | frozen gate table | yes | yes |
| B27 | HVN distance gate | frozen gate table | yes | yes |
| B28 | age gate | nanosecond age table | yes | yes |
| B29 | optional live-price fallback | frozen gate table | yes | yes |
| B30 | drift gate | frozen gate table | yes | yes |
| B31 | decision identity/mint | golden Valid assertion | existing | yes |
| B32 | nonpositive observed live price maps to source drift refusal | `TestParkerFrozenGateBoundariesAndRefusalOrder/nonpositive_live_price_is_drift_refusal` | yes | yes |
| B33 | negative live-entry delta is normalized with absolute value | golden/current-price boundary cases | existing | yes |
| B34 | derived drift above 0.20 percent is refused | `TestParkerFrozenGateBoundariesAndRefusalOrder/drift_above_limit` | yes | yes |
| B35 | decision identity construction fails closed | golden decision identity contract | existing | yes |
| B36 | final decision mint validation fails closed | golden Valid assertion plus forged/zero proof refusals | existing | yes |
