# Branch Test Map: `spliceGateLimits`

- Source: `internal/config/limits_io.go` (207-249, 편집 뒤)

| Branch | Anchor | Scenario | Test | a136 |
|---|---|---|---|---|
| B1 | if 209 | 기존 동작 | `go test ./internal/config -run 'Splice / Limit / TestA136'` | 무편집 분기 |
| B2 | range 216 | 기존 동작 | `go test ./internal/config -run 'Splice / Limit / TestA136'` | 무편집 분기 |
| B3 | if 218 | 기존 동작 | `go test ./internal/config -run 'Splice / Limit / TestA136'` | 무편집 분기 |
| B4 | if 221 | 기존 동작 | `go test ./internal/config -run 'Splice / Limit / TestA136'` | 무편집 분기 |
| B5 | if 222 | 기존 동작 | `go test ./internal/config -run 'Splice / Limit / TestA136'` | 무편집 분기 |
| B6 | if 225 | 기존 동작 | `go test ./internal/config -run 'Splice / Limit / TestA136'` | 무편집 분기 |
| B7 | else 227 | 기존 동작 | `go test ./internal/config -run 'Splice / Limit / TestA136'` | 무편집 분기 |
| B8 | if 227 | 기존 동작 | `go test ./internal/config -run 'Splice / Limit / TestA136'` | 무편집 분기 |
| B9 | if 233 | 기존 동작 | `go test ./internal/config -run 'Splice / Limit / TestA136'` | 무편집 분기 |
| B10 | if 236 | 기존 동작 | `go test ./internal/config -run 'Splice / Limit / TestA136'` | 무편집 분기 |
| B11 | if 244 | 기존 동작 | `go test ./internal/config -run 'Splice / Limit / TestA136'` | 무편집 분기 |
