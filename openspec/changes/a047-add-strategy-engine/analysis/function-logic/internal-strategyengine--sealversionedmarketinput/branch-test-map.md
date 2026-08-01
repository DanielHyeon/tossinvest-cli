# Branch Test Map: `SealVersionedMarketInput`

| Branch | Scenario | Test | RED | GREEN |
|---|---|---|---|---|
| B1 | caller provenance/version laundering | `TestVersionedMarketInputRejectsCallerProvenanceLaundering` | existing | existing |
| B2 | regular close derives 14:45 exact cutoff | `TestVersionedMarketInputDerivesFrozenEntryCutoff` | arbitrary 15:20 accepted | pass |
| B3 | early close derives close-minus-45m | `TestVersionedMarketInputDerivesFrozenEntryCutoff` | caller-authored cutoff | pass |
| B4 | session no longer than 45m refuses | `TestVersionedMarketInputRejectsSessionShorterThanFrozenBuffer` | not covered | pass |
| B5 | opening/cutoff inclusive edges | `TestParkerSessionUsesInjectedEvaluationTimeAndInclusiveCutoff` | wrong 15:20 fixture | pass |
| B6 | required/optional decimal validation | frozen gate and laundering tables | existing | existing |
| B7 | LVN decimal invalid | laundering/decimal tests | existing | existing |
| B8 | tangled decimal invalid | laundering/decimal tests | existing | existing |
| B9 | optional current price present | live-price boundary table | existing | existing |
| B10 | current price decimal invalid | nonpositive/invalid live-price rows | existing | existing |
| B11 | optional expansion/HVN iteration | optional gate table | existing | existing |
| B12 | present optional decimal invalid | optional decimal tests | existing | existing |
