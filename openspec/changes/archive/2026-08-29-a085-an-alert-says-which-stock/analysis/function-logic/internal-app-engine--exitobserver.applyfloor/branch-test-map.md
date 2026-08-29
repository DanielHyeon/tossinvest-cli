# Branch Test Map: `ExitObserver.applyFloor`

AST의 모든 분기를 1행씩 덮는다. 조건 열은 `internal/app/engine/exitloop.go`의 실제 소스 줄이고,
테스트 열과 판정 열은 `go test -covermode=count` 프로파일에서 **측정**한 값이다.
주장이 아니라 측정이므로 이 표는 덮이지 않은 분기를 숨길 수 없다.

> 미검증 분기: B4, B5, B6. 이 change가 바꾼 동작이 아니거나 테스트 하네스가 구성할 수 없는 상태다.
> 관측하지 않은 것을 관측했다고 적지 않는다 — 네 번째 독립 리뷰가 잡은 것이 그 행들이다.

| Branch | Condition | Covering test | Measured |
|---|---|---|---|
| B1 | (1404) `if` — if o.opts.Floor == nil { | `TestAReJudgementNeverWithholdsAStop` | 측정: 아래 테스트가 이 줄을 실행한다 |
| B2 | (1408) `if` — if err != nil { | `TestAReJudgementNeverWithholdsAStop` | 측정: 아래 테스트가 이 줄을 실행한다 |
| B3 | (1416) `if` — if !applies { | `TestAReJudgementNeverWithholdsAStop` | 측정: 아래 테스트가 이 줄을 실행한다 |
| B4 | (1420) `if` — if err != nil { | 패키지 suite (기존 커버리지) | 측정: 이 change의 테스트는 닿지 않는다. 패키지 전체 suite가 덮는다 |
| B5 | (1423) `if` — if cmp >= 0 { | 패키지 suite (기존 커버리지) | 측정: 이 change의 테스트는 닿지 않는다. 패키지 전체 suite가 덮는다 |
| B6 | (1427) `if` — if err != nil { | 패키지 suite (기존 커버리지) | 측정: 이 change의 테스트는 닿지 않는다. 패키지 전체 suite가 덮는다 |
