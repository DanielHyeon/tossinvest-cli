## Provenance

- StockOS `optimization`의 공통 정책 레지스트리에서 BALANCED, RUNNER, HYBRID_50 세 ID와 decimal 수치를 확인했다.
- StockOS runtime 기본 `exit_trail_stop_pct`는 6.5%이며 HYBRID_50의 T4 runner gap으로 사용한다.
- StockOS A168은 외부 편입 RUNNER에 대해 보호선만 승격하고 자동 부분익절을 하지 않는다.

## TossOS hard evidence

- `internal/exitpolicy/ladder.go`의 기존 `DefaultLadderPolicy`는 legacy `default_v1` BALANCED 의미이며 `EvaluateLadder`는 rung 승격 후 breach를 판정한다.
- `internal/app/engine/exitloop.go`는 현재 신규 자체/편입 state를 RATCHET으로만 열고 LADDER observer 전체에 단일 policy를 주입한다.
- `internal/journal/exit_state.go`와 `internal/journal/adoption.go`에는 policy ID snapshot 컬럼이 없고 현재 schema version은 8이다.
- console state-changing route는 `session0(mutating(...))`로 session과 CSRF를 함께 강제한다.
- CodeGraph impact는 `OpenExitState` 81 symbols, `EvaluateLadder` 26 symbols, `mergeEngine`은 직접 함수 범위로 확인됐다.

## Frozen safety boundary

- 기존 활성 state는 config 변경으로 rebind하지 않는다.
- policy save는 broker, order, journal writer, automation gate, trading toggle capability를 받지 않는다.
- 기존 proposal → journal → arm → Guardian reduce-only → execution gateway 경로는 변경하지 않는다.
- LIVE gate 변경, 엔진 기동, 실제 주문은 이 change의 검증 범위에서 수행하지 않는다.
