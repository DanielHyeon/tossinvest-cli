# Branch Test Map: `productionStrategyFirstLegAuthorityLoader.collectStrategyFirstLegAuthority`

Current AST branch inventory. Scenario rows name conditions; they do not claim isolated branch coverage. Focused receipts are in the implementation evidence; full regression remains pending.

| Branch | Source line and condition | Verification status |
|---|---|---|
| B1 | 295: `if loader == nil \|\| ctx == nil \|\| loader.clk == nil \|\| loader.journal == nil \|\| loader.guardian == nil {` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
| B2 | 298: `if loader.exitPolicy == nil {` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
| B3 | 303: `if err != nil \|\| exitPolicy != expected {` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
| B4 | 309: `if !riskAuthority.snapshot.Ready \|\| !fx.snapshot.Ready \|\| !account.snapshot.Ready \|\|` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
| B5 | 316: `if !scoped {` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
| B6 | 320: `if result.Lineage.Identity != accepted.result.Lineage.Identity \|\| result.ExecutionTerms.Identity() != accepted.result.ExecutionTerms.Identity() {` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
| B7 | 325: `if !proposal.familyActivation().Verified() && len(proposal.entries) != 1 {` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
| B8 | 332: `if !keyed { // 봉인이 정규화를 보장하므로 도달 불가 — 바뀌면 범위 거절이 아니라 결함(리뷰 A #4)` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
| B9 | 338: `if !riskScoped {` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
| B10 | 340: `if !scopeLocal {` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
| B11 | 346: `if !accountScoped {` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
| B12 | 353: `if !currencyKnown {` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
| B13 | 357: `if err := riskBundle.Validate(scope); err != nil \|\| scope.AccountID != result.Lineage.AccountRef \|\|` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
| B14 | 362: `if err != nil \|\| cas.Claimed \|\| cas.State != "FLAT" && cas.State != "CLOSED" \|\| cas.Generation < 0 \|\|` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
| B15 | 369: `for _, entry := range entries {` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
| B16 | 380: `if !entryOK \|\| !stopOK \|\| !targetOK {` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
| B17 | 389: `if readCtx == nil \|\| readCtx.Err() != nil \|\| now.IsZero() \|\| now.After(accountAuthority.FreshUntil()) {` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
| B18 | 393: `if versionErr != nil {` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
