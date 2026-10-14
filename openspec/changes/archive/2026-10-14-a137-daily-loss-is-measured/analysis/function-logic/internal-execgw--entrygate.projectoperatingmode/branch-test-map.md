# Branch Test Map: `EntryGate.ProjectOperatingMode`

- Source: `internal/execgw/modegate.go` (35-73, 편집 뒤). 편집 전 분기수 6 → 편집 뒤 7. RED 는 `analysis/red/`, GREEN 은 `analysis/verify/receipts.md`.

| Branch | Anchor | Scenario | Test | a137 |
|---|---|---|---|---|
| B1 | if 37 | 기존 동작 | `go test ./internal/execgw -run Mode` | 무편집 분기 |
| B2 | if 40 | 기존 동작 | `go test ./internal/execgw -run Mode` | 무편집 분기 |
| B3 | if 49 | 기존 동작 | `go test ./internal/execgw -run Mode` | 무편집 분기 |
| B4 | if 54 | a137 편집 | `TestA137OnlyTheAutomaticDailyLossModeIsAQuietAccountBlock`(해제 후 표지 없음) · 기존 모드 투영 시험 | 편집 분기 |
| B5 | if 56 | 기존 동작 | `go test ./internal/execgw -run Mode` | 무편집 분기 |
| B6 | if 65 | a137 편집 | `TestA137OnlyTheAutomaticDailyLossModeIsAQuietAccountBlock`(사람 전이 · 다른 trigger · HALT_ALL 은 표지 없음) | 편집 분기 |
| B7 | if 70 | 기존 동작 | `go test ./internal/execgw -run Mode` | 무편집 분기 |
