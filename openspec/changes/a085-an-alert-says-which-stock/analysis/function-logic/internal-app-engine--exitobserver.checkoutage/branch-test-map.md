# Branch Test Map: `ExitObserver.checkOutage`

AST의 모든 분기를 1행씩 덮는다. 조건 열은 `internal/app/engine/exitloop.go`의 실제 소스 줄이고,
테스트 열과 판정 열은 `go test -covermode=count` 프로파일에서 **측정**한 값이다.
주장이 아니라 측정이므로 이 표는 덮이지 않은 분기를 숨길 수 없다.

> 미검증 분기: B1, B2, B3, B4, B5. 이 change가 바꾼 동작이 아니거나 테스트 하네스가 구성할 수 없는 상태다.
> 관측하지 않은 것을 관측했다고 적지 않는다 — 네 번째 독립 리뷰가 잡은 것이 그 행들이다.

| Branch | Condition | Covering test | Measured |
|---|---|---|---|
| B1 | (769) `if` — if since.IsZero() { | 패키지 suite (기존 커버리지) | 측정: 이 change의 테스트는 닿지 않는다. 패키지 전체 suite가 덮는다 |
| B2 | (772) `if` — if o.clk.Now().Sub(since) < o.outageAfter() { | 패키지 suite (기존 커버리지) | 측정: 이 change의 테스트는 닿지 않는다. 패키지 전체 suite가 덮는다 |
| B3 | (775) `if` — if o.outageRaised { | 패키지 suite (기존 커버리지) | 측정: 이 change의 테스트는 닿지 않는다. 패키지 전체 suite가 덮는다 |
| B4 | (793) `if` — if o.opts.Escalate == nil \|\| strings.TrimSpace(o.opts.AccountRef) == "" { | 없음 | **측정: 어떤 테스트도 이 줄을 실행하지 않는다** |
| B5 | (798) `if` — if err != nil { | 패키지 suite (기존 커버리지) | 측정: 이 change의 테스트는 닿지 않는다. 패키지 전체 suite가 덮는다 |
