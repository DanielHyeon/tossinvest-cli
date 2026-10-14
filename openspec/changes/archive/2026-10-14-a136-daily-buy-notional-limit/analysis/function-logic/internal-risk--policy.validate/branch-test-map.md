# Branch Test Map: `Validate`

- Source: `internal/risk/input.go` (257-309, 편집 뒤)

| Branch | Anchor | Scenario | Test | a136 |
|---|---|---|---|---|
| B1 | if 259 | 기존 동작 | `go test ./internal/risk -run 'Validate / Policy / TestA136'` | 무편집 분기 |
| B2 | range 262 | 기존 동작 | `TestA136PolicyValidationRefusesAnUnusableDailyBuyCeiling` · contract_test mixed currencies | 루프 대상에 새 항목 |
| B3 | if 271 | 기존 동작 | `go test ./internal/risk -run 'Validate / Policy / TestA136'` | 무편집 분기 |
| B4 | if 277 | 기존 동작 | `go test ./internal/risk -run 'Validate / Policy / TestA136'` | 무편집 분기 |
| B5 | if 280 | 기존 동작 | `go test ./internal/risk -run 'Validate / Policy / TestA136'` | 무편집 분기 |
| B6 | if 286 | 기존 동작 | `go test ./internal/risk -run 'Validate / Policy / TestA136'` | 무편집 분기 |
| B7 | if 289 | 기존 동작 | `go test ./internal/risk -run 'Validate / Policy / TestA136'` | 무편집 분기 |
| B8 | if 294 | 기존 동작 | `go test ./internal/risk -run 'Validate / Policy / TestA136'` | 무편집 분기 |
| B9 | if 297 | 기존 동작 | `go test ./internal/risk -run 'Validate / Policy / TestA136'` | 무편집 분기 |
| B10 | if 301 | 기존 동작 | `go test ./internal/risk -run 'Validate / Policy / TestA136'` | 무편집 분기 |
| B11 | if 305 | 기존 동작 | `go test ./internal/risk -run 'Validate / Policy / TestA136'` | 무편집 분기 |
