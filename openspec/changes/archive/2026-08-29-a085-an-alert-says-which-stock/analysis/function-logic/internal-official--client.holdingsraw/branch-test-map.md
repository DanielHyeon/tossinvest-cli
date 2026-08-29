# Branch Test Map: `Client.HoldingsRaw`

AST의 모든 분기를 1행씩 덮는다. 조건 열은 `internal/official/asset_reads.go`의 실제 소스 줄이고,
테스트 열과 판정 열은 `go test -covermode=count` 프로파일에서 **측정**한 값이다.
주장이 아니라 측정이므로 이 표는 덮이지 않은 분기를 숨길 수 없다.

> 미검증 분기: B1, B2, B3. 이 change가 바꾼 동작이 아니거나 테스트 하네스가 구성할 수 없는 상태다.
> 관측하지 않은 것을 관측했다고 적지 않는다 — 네 번째 독립 리뷰가 잡은 것이 그 행들이다.

| Branch | Condition | Covering test | Measured |
|---|---|---|---|
| B1 | (119) `if` — if symbol != "" { | 없음 | **측정: 어떤 테스트도 이 줄을 실행하지 않는다** |
| B2 | (123) `if` — if err := c.getAcct(ctx, "/api/v1/holdings", q, &raw); err != nil { | 없음 | **측정: 어떤 테스트도 이 줄을 실행하지 않는다** |
| B3 | (127) `range` — for _, item := range raw.Items { | 없음 | **측정: 어떤 테스트도 이 줄을 실행하지 않는다** |
