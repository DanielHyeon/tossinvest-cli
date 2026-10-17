# Branch Test Map: `DefaultPolicy`

- Source: `internal/risk/input.go` (239-254)

| Branch | Anchor | Scenario | Test |
|---|---|---|---|
| B1 | 직선(분기 없음) | 정상 경로 | `go test ./internal/risk -run 'DefaultPolicy|A138'` · `TestA138DefaultPolicyCarriesTheOpenPositionMaximum` |
