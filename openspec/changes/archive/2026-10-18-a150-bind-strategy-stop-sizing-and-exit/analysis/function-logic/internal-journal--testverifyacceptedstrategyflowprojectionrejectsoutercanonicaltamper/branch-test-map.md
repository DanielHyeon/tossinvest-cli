# Branch Test Map: `TestVerifyAcceptedStrategyflowProjectionRejectsOuterCanonicalTamper`

Current AST branch inventory. Scenario rows name conditions; they do not claim isolated branch coverage. Focused receipts are in the implementation evidence; full regression remains pending.

| Branch | Source line and condition | Verification status |
|---|---|---|
| B1 | 339: `if err := verifyStrategyRiskBinding(wrongHash, base); err == nil {` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
| B2 | 345: `if err := verifyStrategyRiskBinding(badPreimage, base); err == nil {` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
| B3 | 350: `if err != nil {` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
| B4 | 355: `if err := verifyStrategyRiskBinding(wrongKind, base); err == nil {` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
| B5 | 360: `if err := verifyStrategyRiskBinding(wrongAccount, base); err == nil {` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
| B6 | 364: `for name, tampered := range map[string]string{` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
| B7 | 372: `if tampered == payload {` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
| B8 | 377: `if err := verifyStrategyRiskBinding(decision, lineage); err == nil {` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
