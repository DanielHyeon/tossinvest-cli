# Branch Test Map: `mintDecision`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | provenance/opaque proof refusal | `TestParkerRejectsZeroAndForgedProofsWithoutPanic` | yes | yes |
| B2 | candidate/session/age clocks | candidate, session and signal-age boundary tests | yes | yes |
| B3 | derived prices/RR/drift and ordered reasons | synthetic derivation + translated StockOS parity tests | yes | pass |
| B4 | HVN and optional evidence | frozen gate table | yes | yes |
| B5 | canonical identity | golden decision Valid assertion | existing | yes |
| B6 | candidate identity/provenance | candidate boundary tests | yes | yes |
| B7 | clock/session/freshness | session/age tests | yes | yes |
| B8 | positive decimal loop | golden/sealer tests | existing | yes |
| B9 | tangled threshold | frozen gate table | yes | yes |
| B10 | optional evidence loop | gate table | yes | yes |
| B11 | derived arithmetic equality | golden test | yes | yes |
| B12 | live-price fallback binding | missing-price gate row | yes | yes |
| B13 | HVN presence | HVN table | yes | yes |
| B14 | HVN distance | HVN table | yes | yes |
| B15 | reason order/identity | golden reasons and Valid assertion | yes | yes |
| B16 | cutoff differs from frozen close-minus-45m rule | regular/early-close cutoff derivation tests | arbitrary 15:20 accepted | pass |
