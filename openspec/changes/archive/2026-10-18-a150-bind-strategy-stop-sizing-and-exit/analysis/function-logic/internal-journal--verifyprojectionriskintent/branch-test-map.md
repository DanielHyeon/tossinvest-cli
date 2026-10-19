# Branch Test Map: `verifyProjectionRiskIntent`

Current AST branch inventory. Scenario rows name conditions; they do not claim isolated branch coverage. Focused receipts are in the implementation evidence; full regression remains pending.

| Branch | Source line and condition | Verification status |
|---|---|---|
| B1 | 107: `if !canonical \|\| quantity != risk.Quantity \|\| quantityErr != nil \|\| qFinal == 0 \|\| qFinal > terms.Quantity() \|\| !marked \|\|` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
| B2 | 113: `if schema == strategyflowRiskBindingSchemaVersion \|\| schema == strategyflowRiskBindingSchemaVersionV4 {` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
| B3 | 121: `} else if schema != strategyflowRiskBindingSchemaVersionV2 {` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
| B4 | 118: `if !entryOK \|\| !stopOK \|\| !targetOK {` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
| B5 | 121: `} else if schema != strategyflowRiskBindingSchemaVersionV2 {` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
| B6 | 124: `if risk.EntryPrice != entry \|\| risk.StopPrice != stop \|\| risk.TargetPrice != target {` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
