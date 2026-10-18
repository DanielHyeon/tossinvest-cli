# Branch Test Map: `TestProductionFirstLegAuthorityLoaderPairedKRUS`

Current AST branch inventory. Scenario rows name conditions; they do not claim isolated branch coverage. Focused receipts are in the implementation evidence; full regression remains pending.

| Branch | Source line and condition | Verification status |
|---|---|---|
| B1 | 26: `for _, market := range []StrategyMarket{StrategyMarketKR, StrategyMarketUS} {` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
| B2 | 30: `if !ok {` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
| B3 | 37: `if market == StrategyMarketUS {` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
| B4 | 48: `if market == StrategyMarketKR {` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
| B5 | 50: `} else {` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
| B6 | 58: `if err != nil {` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
| B7 | 63: `for _, market := range []StrategyMarket{StrategyMarketKR, StrategyMarketUS} {` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
| B8 | 66: `if refusal.Code != "" {` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
| B9 | 70: `if err != nil {` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
| B10 | 73: `if issuance.ExitPolicy == nil {` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
| B11 | 77: `if issuance.Entry.Market != string(market) \|\| issuance.Entry.EntryPrice != wantPrice \|\| issuance.Entry.QCandidate != result.Quantity \|\|` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
| B12 | 85: `if !markets[StrategyMarketKR] \|\| !markets[StrategyMarketUS] {` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
