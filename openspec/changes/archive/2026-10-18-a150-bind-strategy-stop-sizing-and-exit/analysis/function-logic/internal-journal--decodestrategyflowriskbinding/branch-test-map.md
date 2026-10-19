# Branch Test Map: `decodeStrategyflowRiskBinding`

Current AST branch inventory. Scenario rows name conditions; they do not claim isolated branch coverage. Focused receipts are in the implementation evidence; full regression remains pending.

| Branch | Source line and condition | Verification status |
|---|---|---|
| B1 | 197: `if err := decoder.Decode(&record); err != nil {` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
| B2 | 200: `if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
| B3 | 204: `if err != nil \|\| string(canonical) != payload {` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
| B4 | 207: `if record.SchemaVersion != strategyflowRiskBindingSchemaVersionV2 && record.SchemaVersion != strategyflowRiskBindingSchemaVersion && record.SchemaVersion != strategyflowRiskBindingSchemaVersionV4 {` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
| B5 | 210: `if record.SchemaVersion == strategyflowRiskBindingSchemaVersionV4 {` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
| B6 | 217: `} else if record.ExitPolicy != nil {` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
| B7 | 211: `if record.ExitPolicy == nil {` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
| B8 | 214: `if err := validateStrategyExitPolicySnapshot(*record.ExitPolicy); err != nil {` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
| B9 | 217: `} else if record.ExitPolicy != nil {` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
| B10 | 220: `if record.ActivationManifestDigest == "" \|\| strings.TrimSpace(record.ActivationManifestDigest) != record.ActivationManifestDigest {` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
