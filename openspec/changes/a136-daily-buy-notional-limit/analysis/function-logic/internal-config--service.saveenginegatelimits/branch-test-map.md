# Branch Test Map: `SaveEngineGateLimits`

- Source: `internal/config/limits_io.go` (97-158, 편집 뒤)

| Branch | Anchor | Scenario | Test | a136 |
|---|---|---|---|---|
| B1 | if 101 | 기존 동작 | `go test ./internal/config -run 'Limit / TestA136'` | 무편집 분기 |
| B2 | if 105 | 기존 동작 | `go test ./internal/config -run 'Limit / TestA136'` | 무편집 분기 |
| B3 | if 111 | 기존 동작 | `go test ./internal/config -run 'Limit / TestA136'` | 무편집 분기 |
| B4 | switch 117 | 기존 동작 | `go test ./internal/config -run 'Limit / TestA136'` | 무편집 분기 |
| B5 | case 118 | 기존 동작 | `go test ./internal/config -run 'Limit / TestA136'` | 무편집 분기 |
| B6 | if 131 | 기존 동작 | `go test ./internal/config -run 'Limit / TestA136'` | 무편집 분기 |
| B7 | case 135 | 기존 동작 | `go test ./internal/config -run 'Limit / TestA136'` | 무편집 분기 |
| B8 | if 139 | 기존 동작 | `go test ./internal/config -run 'Limit / TestA136'` | 무편집 분기 |
| B9 | if 148 | 기존 동작 | `go test ./internal/config -run 'Limit / TestA136'` | 무편집 분기 |
| B10 | else 150 | 기존 동작 | `go test ./internal/config -run 'Limit / TestA136'` | 무편집 분기 |
| B11 | if 150 | 기존 동작 | `go test ./internal/config -run 'Limit / TestA136'` | 무편집 분기 |
| B12 | if 154 | 기존 동작 | `go test ./internal/config -run 'Limit / TestA136'` | 무편집 분기 |
