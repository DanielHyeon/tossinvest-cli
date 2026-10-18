# Branch Test Map: `TestQFinalCampaignFirstLegRejectsTamperWithoutWritesPairedKRUS`

Current AST branch inventory. Scenario rows name conditions; they do not claim isolated branch coverage. Focused receipts are in the implementation evidence; full regression remains pending.

| Branch | Source line and condition | Verification status |
|---|---|---|
| B1 | 248: `for _, descriptor := range []strategyflow.Descriptor{descriptors[0], descriptors[1]} {` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
| B2 | 250: `for _, test := range []struct {` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
| B3 | 270: `if err != nil {` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
| B4 | 277: `if err != nil {` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
| B5 | 281: `if err != nil {` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
| B6 | 290: `if _, err := rig.guardian.PrecheckQFinalCampaignFirstLeg(request); err == nil {` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
| B7 | 293: `if rig.collections != 0 \|\| decisionOnDisk(t, rig.journal, decisionID) {` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
