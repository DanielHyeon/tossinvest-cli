# Change: adopt-external-positions

## Why

사용자 결정(2026-07-26): **사용자가 수동 매수한 종목에도 익절·손절·기준선 수익 극대화(래칫)를 자동 적용**할 것. 현행 계약(2d)은 반대다 — exit-policy·position-ledger 메인 스펙이 "entry 결정 없는 포지션(외부·수동 편입)은 exit 정책의 대상이 아니다"라고 명시하고 알림만 발송한다. 근거는 "R·기준선의 근거(진입 손절가)가 없다"였다. 이 change는 그 근거를 **편입 결정**으로 만들어 공급한다.

## What Changes

- **편입 = 영속된 결정**: 무결정 보유를 발견하면 ADOPTION class 결정(전용 preimage: 심볼·수량·비용기준·합성 손절·관측 시각)을 journal에 영속하고, 포지션의 `entry_decision_id`가 이 결정을 가리킨다. "결정이 정당화하지 않는 포지션은 exit 대상이 아니다"라는 불변식은 **그대로** — 편입은 예외를 뚫는 것이 아니라 결정을 만들어 불변식을 충족시키는 것이다(decide→persist→execute 유지).
- **합성 t0**: EntryPrice는 브로커 평균단가(모르면 편입 시점 관측가 — 어느 쪽을 썼는지 결정 preimage에 기록), InitialStop은 `EntryPrice × (1 − adoption.default_stop_pct)`. 이후 래칫·ladder·부분익절·관측 루프는 엔진 진입 포지션과 **동일 코드 경로**로 적용된다.
- **편입 조건**: Guardian 인터록 활성 + RECONCILE 아님 + 신선한 보유 확인. 제외 목록(`adoption.exclude_symbols`, 기본 빈 목록 — 사용자 결정: 전부 관리)에 있으면 편입하지 않고 기존 알림 경로 유지.
- **알림 대체**: "외부 포지션 발견" 알림은 "편입 완료(편입가·합성 손절 포함)" 이벤트로 대체된다. 제외·편입 실패 시에만 기존 알림.

## Non-Goals

- 매수(진입) 자동화 변경 — 편입은 이미 존재하는 보유에 보호를 붙일 뿐이다
- 손절 즉시성·사이징 규칙 변경(§0.3·§0.9 무접촉 — 편입은 보호 **추가**이므로 보수 방향)
- 브로커측 보호주문 — 그것은 2c의 소관이며, 편입 포지션도 2c의 보호 계약을 동일하게 받는다

## Capabilities

### New Capabilities

(없음)

### Modified Capabilities

- `exit-policy`: "t0 기준선" 요구사항 — 외부 포지션 비대상 조항을 편입 계약으로 대체 + 편입 요구사항 추가
- `position-ledger`: "Position 투영과 단일 권위" — entry_decision_id NULL 조항에 편입 경로 반영

## Impact

- Affected code: `internal/journal`(ADOPTION 결정 class·preimage), 편입 파이프라인(projection 관측→결정→exit_state open), `internal/app/engine`(관측 루프 통합), config(default_stop_pct·exclude_symbols)
- **실효 시점**: 게이트 ON 이후(엔진 가동 중에만 편입·관측이 돈다). 코드는 지금 landed 가능 — 2c 완료·게이트 ON과 함께 활성화
- §0 검토: 편입은 무보호 보유에 손절을 **추가**하는 보수 방향 변경. 편입 포지션의 자동 매도는 long-only 감축(허용 방향)이며 Guardian 인터록 하에서만. flatten·kill switch는 편입 포지션을 동일하게 덮는다
- `adoption.default_stop_pct` 기본값은 보수적으로 정하고 출처(StockOS 대응 계약 또는 산정 근거)를 기록한다 — 임의 수치 금지
