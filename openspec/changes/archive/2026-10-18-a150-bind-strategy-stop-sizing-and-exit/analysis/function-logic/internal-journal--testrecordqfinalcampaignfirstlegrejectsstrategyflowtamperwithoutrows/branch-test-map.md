# Branch Test Map: `TestRecordQFinalCampaignFirstLegRejectsStrategyflowTamperWithoutRows`

Current AST branch inventory. Scenario rows name conditions; they do not claim isolated branch coverage. Focused receipts are in the implementation evidence; full regression remains pending.

| Branch | Source line and condition | Verification status |
|---|---|---|
| B1 | 224: `for _, descriptor := range []strategyflow.Descriptor{descriptors[0], descriptors[1]} {` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
| B2 | 226: `for _, test := range []struct {` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
| B3 | 247: `if reflect.DeepEqual(before, request.Strategy.Lineage) {` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
| B4 | 250: `if _, err := j.RecordQFinalCampaignFirstLeg(context.Background(), request); err == nil {` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
| B5 | 253: `for _, table := range []string{"decisions", "risk_reservations", "strategy_decision_lineage", "strategy_attempt_lineage", "position_campaigns", "position_campaign_claims", "campaign_legs", "risk_bucket_final_decisions", "risk_bucket_owners", "risk_bucket_reservations", "strategy_first_leg_bindings", "strategy_dispatch_leases"} {` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
| B6 | 254: `if got := countRiskBucketRows(t, j, table); got != 0 {` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
