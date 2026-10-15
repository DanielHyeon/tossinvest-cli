# Branch Test Map: `LimitsDigest`

- Source: `internal/execgw/riskguardian.go` (328-336)

| Branch | Anchor | Scenario | Test |
|---|---|---|---|
| B1 | 직선(분기 없음) | 정상 경로 | `go test ./internal/execgw -run 'LimitsDigest|A136|A138'` · `TestA138TheLimitsDigestBindsTheOpenPositionMaximum` |
