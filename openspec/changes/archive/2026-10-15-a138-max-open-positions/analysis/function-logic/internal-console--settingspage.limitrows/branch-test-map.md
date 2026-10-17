# Branch Test Map: `LimitRows`

- Source: `internal/console/settings_limits.go` (61-72)

| Branch | Anchor | Scenario | Test |
|---|---|---|---|
| B1 | 직선(분기 없음) | 정상 경로 | `go test ./internal/console -run 'A136|A138|Limit'` · `TestA138TheScreenSaysWhatTheOpenPositionMaximumIs` |
