# Branch Test Map: `qFinalKRRequest`

- Source: `internal/execgw/riskguardian_qfinal_test.go` (305-351, 편집 뒤). 편집 전 분기수 3 → 편집 뒤 3. RED 는 `analysis/red/`, GREEN 은 `analysis/verify/receipts.md`.

| Branch | Anchor | Scenario | Test | a137 |
|---|---|---|---|---|
| B1 | range 320 | 기존 동작 | `go test -tags tossos_testseams ./internal/execgw -run QFinal` | 무편집 분기 |
| B2 | if 324 | 기존 동작 | `go test -tags tossos_testseams ./internal/execgw -run QFinal` | 무편집 분기 |
| B3 | if 330 | 기존 동작 | `go test -tags tossos_testseams ./internal/execgw -run QFinal` | 무편집 분기 |
