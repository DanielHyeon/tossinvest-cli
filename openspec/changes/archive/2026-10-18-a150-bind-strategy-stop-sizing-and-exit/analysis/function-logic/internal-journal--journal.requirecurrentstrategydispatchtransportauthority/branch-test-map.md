# Branch Test Map: `Journal.RequireCurrentStrategyDispatchTransportAuthority`

Current AST branch inventory. Scenario rows name conditions; they do not claim isolated branch coverage. Focused receipts are in the implementation evidence; full regression remains pending.

| Branch | Source line and condition | Verification status |
|---|---|---|
| B1 | 1660: `if j == nil \|\| j.db == nil \|\| ctx == nil \|\| !validStrategyDispatchIdentity(request.LeaseID) \|\|` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
| B2 | 1666: `if err != nil {` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
| B3 | 1670: `if err := requireCurrentStrategyDispatchOwner(ctx, tx, request.OwnerEpoch, request.FencingToken); err != nil {` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
| B4 | 1674: `if err != nil {` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
| B5 | 1677: `if lease.State != StrategyDispatchLeaseSubmitting \|\| lease.Disposition != StrategyDispatchReservationReserved \|\|` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
| B6 | 1683: `if err != nil {` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
| B7 | 1686: `if blocked \|\| state == positioncampaign.CampaignReconcile \|\| state == positioncampaign.CampaignClosed {` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
| B8 | 1690: `if err != nil {` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
| B9 | 1693: `if !protection.Legacy && (protection.Risk.AccountRef != lease.AccountRef \|\| !strings.EqualFold(protection.Risk.Market, string(lease.Market)) \|\| protection.Risk.Symbol != lease.Symbol) {` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
| B10 | 1696: `if err := requireStrategyProtectionPositionInTx(ctx, tx, lease.CampaignID, protection); err != nil {` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
