# Branch Test Map: `ReconcileDriver.RunOnce`

AST의 모든 분기를 1행씩 덮는다. 조건 열은 `internal/app/engine/reconcileloop.go`의 실제 소스 줄이고,
테스트 열과 판정 열은 `go test -covermode=count` 프로파일에서 **측정**한 값이다.
주장이 아니라 측정이므로 이 표는 덮이지 않은 분기를 숨길 수 없다.

> 미검증 분기: B1, B2, B3, B4, B5, B6, B7, B8. 이 change가 바꾼 동작이 아니거나 테스트 하네스가 구성할 수 없는 상태다.
> 관측하지 않은 것을 관측했다고 적지 않는다 — 네 번째 독립 리뷰가 잡은 것이 그 행들이다.

| Branch | Condition | Covering test | Measured |
|---|---|---|---|
| B1 | (408) `if` — if !ok { | 패키지 suite (기존 커버리지) | 측정: 이 change의 테스트는 닿지 않는다. 패키지 전체 suite가 덮는다 |
| B2 | (414) `if` — if err != nil { | 패키지 suite (기존 커버리지) | 측정: 이 change의 테스트는 닿지 않는다. 패키지 전체 suite가 덮는다 |
| B3 | (425) `if` — if err != nil && cycle.Err == nil { | 없음 | **측정: 어떤 테스트도 이 줄을 실행하지 않는다** |
| B4 | (431) `if` — if err != nil && cycle.Err == nil { | 없음 | **측정: 어떤 테스트도 이 줄을 실행하지 않는다** |
| B5 | (437) `if` — if err := d.opts.Tracker.Refresh(ctx); err != nil { | 없음 | **측정: 어떤 테스트도 이 줄을 실행하지 않는다** |
| B6 | (438) `if` — if cycle.Err == nil { | 없음 | **측정: 어떤 테스트도 이 줄을 실행하지 않는다** |
| B7 | (447) `if` — if err != nil { | 패키지 suite (기존 커버리지) | 측정: 이 change의 테스트는 닿지 않는다. 패키지 전체 suite가 덮는다 |
| B8 | (448) `if` — if cycle.Err == nil { | 패키지 suite (기존 커버리지) | 측정: 이 change의 테스트는 닿지 않는다. 패키지 전체 suite가 덮는다 |
