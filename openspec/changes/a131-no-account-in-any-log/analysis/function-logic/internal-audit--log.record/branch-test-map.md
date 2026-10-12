# Branch Test Map: `Log.Record`

- Source: `internal/audit/audit.go` (173-208) — 편집 전(base)

| Branch | Anchor | Scenario | Test | a131 |
|---|---|---|---|---|
| B1 | if at 174:2 | `if l == nil {` | a131 무편집 분기 — `go test ./internal/audit/` 전체 | 무편집 |
| B2 | if at 177:2 | `if e.At.IsZero() {` | a131 무편집 분기 — `go test ./internal/audit/` 전체 | 무편집 |
| B3 | else at 179:9 | `} else {` | a131 무편집 분기 — `go test ./internal/audit/` 전체 | 무편집 |
| B4 | if at 182:2 | `if strings.TrimSpace(e.Subject) == "" {` | a131 무편집 분기 — `go test ./internal/audit/` 전체 | 무편집 |
| B5 | if at 187:2 | `if err != nil {` | a131 무편집 분기 — `go test ./internal/audit/` 전체 | 무편집 |
| B6 | if at 196:2 | `if err != nil {` | a131 무편집 분기 — `go test ./internal/audit/` 전체 | 무편집 |
| B7 | if at 201:2 | `if _, err := f.Write(line); err != nil {` | a131 무편집 분기 — `go test ./internal/audit/` 전체 | 무편집 |
| B8 | if at 204:2 | `if err := f.Sync(); err != nil {` | a131 무편집 분기 — `go test ./internal/audit/` 전체 | 무편집 |

## a131 시험

- `TestA131AuditRecordReplacesOnlyAccountText` — 센티널 계좌가 detail 에서 치환, action·setting·old·new·subject·at 보존
- `TestA131RecordChangeStillDeduplicates`
