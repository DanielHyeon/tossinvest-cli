# Branch Test Map: `TestOnlyTheListedFilesCanNameTheChaseVerdict`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | repository walk must find a credible file count | this guard | existing behavior | GREEN |
| B2 | every discovered Go file is parsed | this guard | existing behavior | GREEN |
| B3 | parser failure is fatal | this guard | existing behavior | GREEN |
| B4 | files without verdict names are skipped | this guard | existing behavior | GREEN |
| B5 | unlisted HTTP API verdict reader is detected | this guard | RED: `cmd/tossctl/httpapi_reader.go` unlisted | GREEN: explicit read-only entry |
| B6 | every allowlist permission is checked for continued use | this guard | existing behavior | GREEN |
| B7 | stale allowlist entries are rejected | this guard | existing behavior | GREEN |
| Order separation | reader cannot also reach an order verb | `TestNoFileThatReadsTheVerdictCanAlsoPlaceAnOrder` | existing guard evaluated new file | GREEN |
