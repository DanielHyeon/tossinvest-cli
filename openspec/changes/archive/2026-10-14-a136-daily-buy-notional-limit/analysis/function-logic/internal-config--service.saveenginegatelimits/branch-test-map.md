# Branch Test Map: `SaveEngineGateLimits`

- Source: `internal/config/limits_io.go` (98-163, 편집 뒤)

| Branch | Anchor | Scenario | Test | a136 |
|---|---|---|---|---|
| B1 | if 102 | 기존 동작 | `go test ./internal/config -run 'Limit / TestA136'` | 무편집 분기 |
| B2 | if 106 | 기존 동작 | `go test ./internal/config -run 'Limit / TestA136'` | 무편집 분기 |
| B3 | if 112 | 기존 동작 | `go test ./internal/config -run 'Limit / TestA136'` | 무편집 분기 |
| B4 | switch 118 | 기존 동작 | `go test ./internal/config -run 'Limit / TestA136'` | 무편집 분기 |
| B5 | case 119 | 기존 동작 | `go test ./internal/config -run 'Limit / TestA136'` | 무편집 분기 |
| B6 | if 131 | 기존 동작 | `go test ./internal/config -run 'Limit / TestA136'` | 무편집 분기 |
| B7 | if 135 | 기존 동작 | `go test ./internal/config -run 'Limit / TestA136'` | 무편집 분기 |
| B8 | case 139 | 기존 동작 | `go test ./internal/config -run 'Limit / TestA136'` | 무편집 분기 |
| B9 | if 143 | 기존 동작 | `go test ./internal/config -run 'Limit / TestA136'` | 무편집 분기 |
| B10 | if 152 | 기존 동작 | `go test ./internal/config -run 'Limit / TestA136'` | 무편집 분기 |
| B11 | else 154 | 기존 동작 | `go test ./internal/config -run 'Limit / TestA136'` | 무편집 분기 |
| B12 | if 154 | 기존 동작 | `go test ./internal/config -run 'Limit / TestA136'` | 무편집 분기 |
| B13 | if 159 | 기존 동작 | `go test ./internal/config -run 'Limit / TestA136'` | 무편집 분기 |
