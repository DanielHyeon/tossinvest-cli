# a146 — Guardian 상한 0 인 전략 진입은 그 진입만 건너뛴다 (guardian-cap-zero-skips-the-entry)

> **사용자 결정(2026-10-15, Manager 전달 원문):** 「Guardian 상한 0 → 레인 잠금 대신 「이번 진입만 건너뜀」」. 예산은 1/3 그대로(a139).
> 기각된 안: 잠금 상태를 원장 밖으로 · 메모리 제어 · 비동기 원장 쓰기(a112 5.3.3 durable latch 와 a136/a138 예약 tx 직렬화를 깸).
> 범위 밖: 남은 일일 손실에 맞춰 수량을 줄이는 것.
> **사람 결정(2026-10-18, 「권장대로」):** 범위는 「Guardian 상한 0 (위험 예산 또는 주문 금액 상한으로 1주도 안 되는 진입)」 — 주문 금액(notional) 상한 0 포함.
> F10 한계는 수용하되 execgw 결속 판정 이동 후속이 전략 레인 활성화(P2 S5/S6) 전에 착지해야 함. 범위 거절의 effective 잠금은 유지.

## Why

a139 리뷰(코드 읽기만 — 미검증)가 「건당 위험 예산이 한 주의 위험도 못 덮으면 Guardian 상한 0 → `EXISTING_GUARDIAN_CAP` → 첫 다리
`AUTHORITY_MISMATCH` → 레인이 첫 실패에 잠김(원장, 재시작으로 안 풀림)」을 잔여로 남겼다. a146 은 그 사슬을 **실제 경로로 먼저 재고**
(특성 시험 `a146_cap_zero_characterization_test.go`, AST `analysis/ast-evidence/`) 결정을 적용한다. 실측이 a139 의 서술을 두 군데 고친다(design F6 · F7):

1. **잠금은 감독자의 시장 잠금이고 원장 잠금이 아니다.** `latchMarket` 은 메모리 칸(`latched`)과 소비자 없는 fault 채널만 쓴다 — 원장의
   durable 잠금은 전략군 레인(`persistMarketLatches`)의 것이고 dispatch 오류는 거기 닿지 않는다. 잠긴 시장은 **프로세스 재시작까지** 평가되지 않는다.
2. **오늘 생산 감독자는 그 잠금 갈래에 오지 않는다.** 생산 감독자(`NewRefreshingPairedStrategyEntrySupervisor`)의 worker 는 전부
   권한 갱신 전용(`Effective: false`, `RefreshesAuthority: true`)이라 — **서명 활성화된 시장도 이 worker 로 돈다** — 주기 오류는 **삼켜 세어진다.** 대신 그 주기는
   Guardian 상한 0 인 범위에서 멈춰 **뒤 범위가 매 주기 평가되지 않는다**(결정적 순서 — 앞 범위가 그대로면 영원히 굶음). 뒤 범위가 있는 것은 서명된 4-가족
   활성화가 검증된 시장뿐이다 — 그렇지 않으면 `dispatchHandoffs` 가 시장 단위 handoff 하나만 돌려준다(strategy_dispatch_handoff.go:38). 잠금은 감독자 worker 의 `Effective`
   칸이 참일 때만 일어나고, 그 worker 를 `Run` 하는 생산 코드는 오늘 없다(방어 불변식으로 함께 다룸).

어느 쪽이든 「한 주도 못 사는 진입」은 결함이 아니라 정책 결과다. 결함으로 다루면 effective 시장은 통째로 잠기고, 오늘 생산은 같은 시장의 다른
범위를 굶기며 진짜 결함 계수(`SwallowedCycleErrors`)를 정책 결과로 채운다.

## What Changes

1. **분류** — 첫 다리 승인(`strategyFirstLegAdmissionBridge.admit`)이 precheck 거절 중 **`QFinalRefusal` 코드 `EXISTING_GUARDIAN_CAP` 이고 원인 사슬에
   `risk.ErrStrategyQuantityZero` 가 있는 것만**(타입 · 코드 · sentinel — 문구 아님) 새 코드 `GUARDIAN_CAP_ZERO` 와 엔진 타입 원인
   `*strategyEntrySkip`(시장 · 종목 · 원래 거절)으로 돌려준다. 그 밖의 `EXISTING_GUARDIAN_CAP`(정수 해석 실패 · FX/정책 오류 · `CalculateAdmission` 의 0 상한)은
   오늘처럼 `AUTHORITY_MISMATCH`.
2. **전달** — `deliverEachStrategyHandoff` 는 그 원인을 보면 그 범위만 건너뛰고 다음 범위로 간다. 범위 거절과 달리 **주기 오류로 모으지 않는다**
   (정책 결과 — 아래 3 이 흔적을 남김). 스킵만 있던 주기는 오류 없이 끝남 → effective 는 잠기지 않고 refreshOnly 는 삼킴 계수가 오르지 않는다.
3. **흔적** — 감독자 `runMarket` 이 주기 ctx 에 기록자를 싣고(a137 측정 불가와 같은 모양), 스킵마다 시장 스냅숏 `GuardianCapZeroSkips` 를 세고 첫 원인을
   `FirstGuardianCapZeroSkip` 에 남기며, 경고 줄 `engine.strategy_entry_skipped` 를 **(계좌 일 키, 시장, 종목)마다 한 번** 낸다(5 초 주기마다가 아님). 계좌 값 없음.
4. **spec** — `risk-management` 에 요구 하나 추가(「Guardian 상한 0 인 전략 진입은 그 진입만 건너뛴다」).

## What Does NOT Change

- 사이징(`risk` · `execgw` 무편집 — 예산 1/3 · min 규칙 · `ErrStrategyQuantityZero` 그대로), precheck 순서, 체인, reason code 집합.
- 범위 거절(버킷 고갈)의 보고: 건너뛴 거절은 지금처럼 주기 오류로 돌아오고 effective 시장을 잠근다 — **별도 발견**(design F8 · D7), 사람 결정(2026-10-18)으로 유지.
- 일일 손실 한도 · 측정 불가(a137), 중앙 무결성, 그 밖의 precheck · 발급 · Gateway · 원장 거절은 오늘과 같다.
- 매도 · 손절 · 익절 · 비상 청산 · 대사 · 체결 감지 — 편집 함수는 전부 첫 다리(매수) 진입 경로에만 있다(design F9).
- 원장 스키마 · 설정 키 · 서명 매니페스트. 잠금의 저장 위치(사용자가 기각한 안 — 건드리지 않음). 건너뛴 진입의 브로커 호출은 0 이다 — 다만 같은 주기의
  뒤 범위가 이제 평가되므로 그 범위의 정상 주문은 같은 주기에 나갈 수 있다(의도, design D5).
- **명시 한계(design F10 · D5, freeze 리뷰 P1-1):** 상한 0 은 그 뒤의 Guardian 판정(정책 버전 · 주간 예약 결속 · 투영 · 버킷 · 체인 · 노출)을 오늘도 가린다. a146 뒤에는
  그 자리에 동시에 있는 결함이 잠금/삼킴 대신 조용한 스킵이 되고, 상한이 1 이상이 되는 주기에 드러난다(주문 0). 결속 판정을 앞으로 옮기는 안은 사람 결정 잔여.

## Impact

- Affected specs: `risk-management`(ADDED 1).
- Affected code(예정): `internal/app/engine/strategy_first_leg_admission.go`(`admit` · 코드 상수) · `strategy_market_handoff_delivery.go`(`deliverEachStrategyHandoff`) ·
  `strategy_entry_supervisor.go`(`runMarket` 한 줄 · `Snapshot` · `NewStrategyEntrySupervisor` · `NewRefreshingPairedStrategyEntrySupervisor` 배선 한 줄 · 옵션/스냅숏 칸) ·
  새 파일 `a146_entry_skip.go`(타입 · 기록자 · 알림), `internal/obs/event.go`(이벤트 상수). 기존 시험 `TestAnExistingGuardianCapRefusalIsNotAScopeRefusal` 은 새 동작으로 고침.
- 행동: 생산(refreshOnly) — 상한 0 범위 뒤의 범위가 같은 주기에 평가된다(굶음 해소 — 각 범위는 여전히 전체 Guardian 판정을 지난다), 그 주기의 shadow 관측이
  더 이상 무효화되지 않는다(`strategyCycleWithShadow` B3), 삼킴 계수가 정책 결과로 오르지 않는다. effective — 상한 0 이 시장을 잠그지 않는다.
- 위험 등급: **High-risk**(Guardian · 전략 진입 · 레인 잠금 분류). FLM/BTM · Pre-Edit · RED 선행 · 변이 원장 · 적대 리뷰.
- 배포: 엔진 이미지 교체. 설정 · 서명 변경 불요. 현재 생산 레인은 DORMANT(서명 매니페스트 0)라 실제 동작 변화는 활성화 뒤에 나타난다.
