# Branch Test Map: `seedExistingRiskReservation`

- Source: `internal/journal/risk_bucket_admission_test.go` (406-422)

| Branch | Anchor | Scenario | Test |
|---|---|---|---|
| B1 | if 413 | 분기 동작 | `go test ./internal/journal -run A066` · `TestA066StaleUsageRuleEdgesAtBothAdmissionSites` |
| B2 | if 416 | 분기 동작 | `go test ./internal/journal -run A066` · `TestA066StaleUsageRuleEdgesAtBothAdmissionSites` |
| B3 | if 419 | 분기 동작 | `go test ./internal/journal -run A066` · `TestA066StaleUsageRuleEdgesAtBothAdmissionSites` |
