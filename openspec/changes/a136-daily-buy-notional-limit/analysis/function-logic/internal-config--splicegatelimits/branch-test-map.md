# Branch Test Map: `spliceGateLimits`

- Source: `internal/config/limits_io.go` (217-259, 편집 뒤)

| Branch | Anchor | Scenario | Test | a136 |
|---|---|---|---|---|
| B1 | if 219 | 기존 동작 | `go test ./internal/config -run 'Splice / Limit / TestA136'` | 무편집 분기 |
| B2 | range 226 | 기존 동작 | `go test ./internal/config -run 'Splice / Limit / TestA136'` | 무편집 분기 |
| B3 | if 228 | 기존 동작 | `go test ./internal/config -run 'Splice / Limit / TestA136'` | 무편집 분기 |
| B4 | if 231 | 기존 동작 | `go test ./internal/config -run 'Splice / Limit / TestA136'` | 무편집 분기 |
| B5 | if 232 | 기존 동작 | `go test ./internal/config -run 'Splice / Limit / TestA136'` | 무편집 분기 |
| B6 | if 235 | 기존 동작 | `go test ./internal/config -run 'Splice / Limit / TestA136'` | 무편집 분기 |
| B7 | else 237 | 기존 동작 | `go test ./internal/config -run 'Splice / Limit / TestA136'` | 무편집 분기 |
| B8 | if 237 | 기존 동작 | `go test ./internal/config -run 'Splice / Limit / TestA136'` | 무편집 분기 |
| B9 | if 243 | 기존 동작 | `go test ./internal/config -run 'Splice / Limit / TestA136'` | 무편집 분기 |
| B10 | if 246 | 기존 동작 | `go test ./internal/config -run 'Splice / Limit / TestA136'` | 무편집 분기 |
| B11 | if 254 | 기존 동작 | `go test ./internal/config -run 'Splice / Limit / TestA136'` | 무편집 분기 |
