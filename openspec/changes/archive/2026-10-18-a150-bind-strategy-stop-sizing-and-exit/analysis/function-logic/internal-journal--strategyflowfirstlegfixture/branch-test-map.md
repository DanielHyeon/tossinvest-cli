# Branch Test Map: `strategyflowFirstLegFixture`

Current AST branch inventory. Scenario rows name conditions; they do not claim isolated branch coverage. Focused receipts are in the implementation evidence; full regression remains pending.

| Branch | Source line and condition | Verification status |
|---|---|---|
| B1 | 144: `result, err := strategyflow.AcceptedResultForJournalTest(descriptor, account, symbol, request.Campaign.CampaignID, 20, entryMinor, stopMinor, targetMinor)` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
| B2 | 148: `policy, err := QFinalPolicyVersion("engine.automation_gate/risk-policy-v1", request.Issue.Admission.TransactionID)` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
| B3 | 155: `policySnapshot, err := NewStrategyExitPolicySnapshot("")` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
| B4 | 160: `ActivationManifestDigest: request.Strategy.ActivationManifestDigest, CreatedAt: request.Strategy.CreatedAt, ExitPolicy: &policySnapshot})` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
| B5 | 164: `request.Strategy.Lineage = lineage` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
| B6 | 166: `provider, zone, stable := "XKRX_OFFICIAL", "Asia/Seoul", "KR-XKRX-2026-W14"` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
| B7 | 175: `})` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
