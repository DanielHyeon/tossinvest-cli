# Branch Test Map: `TestVerifyStrategyflowRiskBindingV2Compatibility`

Current AST branch inventory. Scenario rows name conditions; they do not claim isolated branch coverage. Focused receipts are in the implementation evidence; full regression remains pending.

| Branch | Source line and condition | Verification status |
|---|---|---|
| B1 | 80: `result, err := strategyflow.AcceptedResultForJournalTest(descriptor, "acct-v2-us", "AAPL", "campaign-v2-us", 20, "5000", "4505", "6001")` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
| B2 | 84: `projection, err := strategyflow.ProjectAccepted(result)` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
| B3 | 88: `policy, err := QFinalPolicyVersion("engine.automation_gate/risk-policy-v1", "risk-v2-us")` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
| B4 | 94: `risk, err = exactCanonicalRiskIntent(risk)` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
| B5 | 102: `identity, err := strategyflowDecisionIdentity(record)` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
| B6 | 107: `payload, err := json.Marshal(record)` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
| B7 | 111: `golden, err := os.ReadFile("testdata/a150-baseline-v2.json")` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
| B8 | 118: `IssuedAt: created, ExpiresAt: created.Add(time.Minute)}).build()` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
| B9 | 121: `}` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
| B10 | 124: `}` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
