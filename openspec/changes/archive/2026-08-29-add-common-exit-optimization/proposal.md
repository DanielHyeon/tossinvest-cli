## Why

TossOS에는 StockOS에서 검증·운영 중인 공통 익절 사다리와 고점 추종 보호선 정책을 선택하는 운영 화면이 없고, 엔진이 직접 연 포지션과 외부 매수 후 편입한 포지션이 명시적인 정책 스냅샷 없이 기본 RATCHET에 고정된다. 모바일/VPN 접속을 열기 전에, 보유 포지션의 청산 정책과 그 변경 권한을 먼저 명시적이고 감사 가능한 계약으로 만들어야 한다.

## What Changes

- StockOS의 `COMMON_LADDER_HYBRID_50`, `COMMON_LADDER_RUNNER`, `COMMON_LADDER_BALANCED` 수치와 잔량 기준 부분익절 의미를 TossOS의 decimal 기반 exit evaluator로 이식한다.
- HYBRID_50의 마지막 잔량은 고정 전량익절 대신 고점 대비 6.5% trailing 보호선을 사용하며 보호선은 절대 내려가지 않는다.
- 승인된 공통 정책은 새로 관리가 시작되는 포지션에 정책 ID로 스냅샷한다. 기존 포지션은 배포나 공통값 변경만으로 재해석하지 않는다.
- 외부 구매 포지션도 편입 시점의 관측 가격·합성 손절선과 함께 같은 정책을 스냅샷한다.
- 콘솔에 `/optimization` 메뉴를 추가해 등록된 정책의 수치·효과·현재 승인 상태를 표시하고, 세션+CSRF를 통과한 사람의 POST만 공통 정책을 변경할 수 있게 한다.
- 공통 정책 미설정은 기존 RATCHET 동작을 그대로 보존한다. 설정 저장은 다음 엔진 기동부터 반영하며 주문·게이트·거래 토글을 변경하지 않는다.
- 정책 변경 전후 값을 audit에 남기고, 알 수 없는 정책이나 약화되는/불완전한 정책 블록은 저장·기동 단계에서 fail-closed로 거부한다.

## Capabilities

### New Capabilities

- `common-exit-policy`: 등록 정책, 공통 기본 승인, 포지션 정책 스냅샷, 외부 편입 적용과 고점 추종 보호선의 계약.

### Modified Capabilities

- `exit-policy`: 새 포지션의 정책 선택·스냅샷과 HYBRID_50/runner 평가를 기존 즉시 청산·단조 보호선 불변식에 결합한다.
- `operator-console`: 최적화 조회·승인 화면과 CSRF 보호 설정 경로를 추가한다.
- `engine-safety`: 정책 설정이 LIVE 권한이나 기존 위험 축소 우선순위를 넓히지 않도록 기동 검증과 감사 요구를 추가한다.

## Impact

- `internal/exitpolicy`: 공통 정책 레지스트리와 HYBRID_50 runner 보호선 평가.
- `internal/journal`: additive schema migration과 `exit_states.policy_id` 스냅샷.
- `internal/config`: `engine.exit_policy.common_policy`의 raw load, 검증, surgical save.
- `internal/app/engine`: 신규/외부 편입 포지션의 정책 선택과 포지션별 정책 해석.
- `internal/console`, `cmd/tossctl`: `/optimization` 화면, 저장 seam, audit wiring.
- OpenSpec `exit-policy`, `operator-console`, `engine-safety` 요구사항과 회귀 테스트.
PM bootstrap exception: the current TossOS portfolio contains only the completed
SDD-adoption story, so `add-common-exit-optimization` is temporarily registered in
`docs/pm/portfolio/_registry.yaml` until a product story hierarchy is established.
