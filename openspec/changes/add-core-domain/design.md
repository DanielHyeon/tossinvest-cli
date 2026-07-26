# Design: add-core-domain

> 2026-07-26 2판(동결 해제). 1판의 2라운드 리뷰 25건(`review.md` 후반부)과 사용자 추가 요구(손익 극대화 exit 정책)를 반영. 선행: 2a `extend-execution-contract` GATE PASS·archive — 이 change는 그 메인 스펙(결정 계약·예약·RECONCILE·계산 계약) **위에서** 판단 정책을 구현하며, 그 요구를 재정의하지 않는다.

## Context

2a가 만든 것: 결정은 journal 참조(클래스별 preimage 재검증), `f(decision_id, generation)` 멱등키, 진입 측 위험 예약(원자 트랜잭션·brokerterminal 해제), RECONCILE 영속 상태·확정 하한, 총계 계산 계약(LIMIT 전용·gross long·실현 손실·staleness 10s/60s), `internal/riskcalc`, 엔진 Gateway 배선·인터록 5조항, `ExposureLimiter` 요구. **발급자(Guardian)는 아직 없다** — 자동 진입은 구조적으로 불가능한 상태다. 이 change가 발급자·포지션·exit 정책·성과를 채운다.

2c(보호주문)는 2b 측정 후 작성된다. 따라서 이 change의 범위에서 **브로커측 보호주문은 존재하지 않는다** — 손절·기준선은 로컬 판정이고, 발동은 청산 발의(RISK_REDUCING)다. 브로커 상주 보호는 2c가 얹는다.

## Goals / Non-Goals

**Goals**: Guardian 판정 체인·발급자, 운영 모드(모드×클래스 표), 비용 모델(구조 이식·수치 2b), 포지션 투영·조정 이벤트·reconciliation 재배선, exit 정책(baseline ratchet·profit ladder — 사용자 요구), 성과 원시 지표, tracer slice 코드.

**Non-Goals**: 보호주문·조건주문 일체(2c), 구조적 RR 계산·등급배수·신호 트레일·시간 종료(P3 — 입력 생산자 부재), MFE/MAE(P3), 전략·후보·스케줄러(P3), 웹(P4+), tracer 실전 실행(attestation+승인 후 verify 트랙).

## Decisions

### D1. 체인은 사전 검사, 예약이 총계의 권위

Guardian 체인은 스냅샷 위 순수 함수로 reason-code를 산출한다. 총계 한도의 최종 권위는 2a 예약 트랜잭션이며, 발급 절차는 체인 ALLOW → 예약(as-of 재검증) → 결정 영속·발급. 예약 거부 = RESERVATION_CONFLICT. 재수집 시 체인 재실행 없음 — 결정 만료 5초가 입력 신선도를 담보한다. (1판 리뷰 B1 해소.)

체인 순서는 TossOS 정의가 권위이고 이식 검사에 한해 StockOS 상대 순서를 보존한다 — 매핑 표를 `docs/guardian-chain.md` 산출물로 남긴다(B3 해소: 원 체인에는 미이식 단계가 섞여 있어 "원 순서 보존" 전체 주장은 성립 불가).

### D2. 신호 계층 입력이 필요한 검사는 없다

구조적 RR 계산(세션 고가·VWAP)과 등급배수는 P3 신호 계층이 생산한다 — 이 change의 체인은 의도 필드·브로커 스냅샷·journal 상태만 소비한다(B2 해소: 명세대로 ALLOW 불가였던 문제). 최소 RR은 의도의 순수 산술, 기본 2.0(StockOS §22 lock — 1.5는 최저 티어 값이라 기각, C6). 수량 배수는 1.0 고정(보수 하한). 레버리지/인버스·ETF/ETN 차단은 분류 소스 부재로 **심볼 allowlist**가 구조 대체(D3 보류 3항목 해소; 재진입 쿨다운은 이식 — 순수 로직).

### D3. 운영 모드 = 모드×클래스 표

risk-management 델타의 표가 정본. HALT_ALL은 "전면 중단"이 아니라 EXPOSURE_RAISING·PROTECTION_WEAKENING 거부 + RISK_REDUCING 허용(A1·C1·C3 해소 — 2-클래스 어휘 폐기). 자동 강화 트리거는 열거형(일손실·401/403·critical outbox — 분석 outbox 제외, C5). journal 영속·방향 비대칭 승인.

### D4. Position은 투영, reconciliation은 재배선

`internal/position`은 journal 체결 이벤트+조정 이벤트의 투영(심볼·시장·인스턴스·평균단가, decimal). `reconcile.LocalStateFromJournal`을 이 투영 소비로 재배선한다 — "같은 질의 위의 얇은 투영" 주장(1판)은 `NetPositions`의 형태(심볼→순수량, float, 시장 무차원)로는 성립하지 않았다(B4). 따라서 **reconciliation capability를 MODIFIED로 선언**한다(A9 해소 — 조정 이벤트가 해제 의미를 바꾸는 것도 명시: 자동 해제는 조정 반영 후 재조회 일치, 영구 승격은 운영자만).

방향 재도출은 intent side — 이 범위의 모든 체결은 로컬 intent가 있다(조건주문 없음). 발동 주문 방향은 2c가 expected_orders에서 정의(A8 스코프 정리).

### D5. Exit 정책 — 이식 대상과 경계 (사용자 요구)

`internal/exitpolicy` 순수 판정 모듈:

- **baseline ratchet**(StockOS `exit/baseline_ratchet.py` — Decimal·주문 무접촉 이식): R 트리거 0.4/0.8/1.0/1.2/2.0 → 스톱 −0.5R/실질 본전/부분 40%/+0.3R/+0.8R. **단조 상승 불변식**(§0.9 정합).
- **profit ladder**(`profit_ladder.py`): multi-rung, 판정 시점 필드(활성 rung·승격 보호선)와 체결 시점 필드(누적 익절 비율·완결) 분리 보존, STOP_FIRST 보수 모델.
- 본전 = `break_even_sell_price`(costs 결합). 발의는 ReductionIntent 계열 의도로만 — 제출·수량 정합은 실행 계층(2a Gateway·2c).
- 가격 관측 입력: 보유 심볼의 시세 조회(quote — 엔진 read 경로에 이미 존재)를 exit 판정 주기로 사용하며, §0.4 rate budget 안에서 주기를 명시하고 관측 실패는 판정 보류(기준선 유지 — fail-safe)다. 이것은 MFE/MAE용 시계열 저장이 아니라 판정용 최신가 1점이다(P3 이관 결정과 충돌하지 않음).
- 제외: EMA9/VWAP/CVD/추세선 트레일·TIME_EXIT/EOD·limit-up hold(P3), SELL_COST_BUFFER(§0.3 위반 — 미이식, C4).

### D6. 비용 모델 — 구조 이식, 수치는 2b

`internal/costs`: StockOS `costs.py`의 구조(시장별 수수료 bps·거래세 bps·실질 본전 산식)만 이식, KIS 수치·`KIS_*` override 미이식(C7). 실측 전 과대 추정 placeholder + "미검증" provenance. 청산 게이트 적용 금지(C4). 테스트: test_costs(4)+test_costs_env_override(16) 구조 이식, 수치 단언은 Toss 값 재작성(D5-리뷰).

### D7. journal v6 — 단일 원자 마이그레이션 (태스크는 전사)

전 decimal TEXT. 직전 백업·복원, 구버전 ErrSchemaTooNew.

| 테이블 | 컬럼(요지) |
|---|---|
| `positions` | id PK, account_ref·market·symbol NOT NULL, instance_seq INTEGER NOT NULL, state CHECK(FLAT\|OPENING\|OPEN\|SCALING\|CLOSING\|CLOSED), quantity·avg_price TEXT, opened_at/closed_at. UNIQUE(account_ref, market, symbol, instance_seq) |
| `position_adjustments` | id PK, position_id FK, kind CHECK(EXTERNAL\|MANUAL\|UNKNOWN), prev/new quantity·avg_price, broker_as_of, evidence, created_at — append-only |
| `operating_modes` | id PK, account_ref, mode CHECK(NORMAL\|ENTRY_BLOCKED\|EXIT_ONLY\|HALT_ALL), cause, actor CHECK(AUTO\|OPERATOR), created_at — append-only 이력, 현재 = 최신 행 |
| `exit_states` | position_id PK FK, ratchet_level CHECK(NONE\|HALF_RISK\|BREAKEVEN\|PARTIAL_LOCK\|PROFIT_LOCK), baseline_price TEXT NOT NULL DEFAULT '', active_rung INTEGER, taken_ratio_total TEXT NOT NULL DEFAULT '0', completed INTEGER DEFAULT 0, updated_at |
| `trade_outcomes` | position_id PK FK, realized_pnl_after_costs·r_multiple·initial_risk TEXT, held_seconds INTEGER, exit_ratchet_level·exit_rung, closed_at — 보존 180일 |

Guardian 설정(allowlist·쿨다운·한도)은 config — journal이 아니다(운영 설정 audit는 기존 계약).

### D8. tracer slice

하드코딩 심볼 1개(=allowlist)·LIMIT·최소 수량의 진입→ratchet 판정→청산 end-to-end 실행기. 계좌·시장·최대 notional·가격 신선도·중단 기준을 파라미터로 명시. httptest 통합 검증이 완료 조건이고 실전 실행은 attestation+게이트 ON+사용자 승인 후 verify 트랙(변경 없음).

## Risks / Trade-offs

- [ratchet·ladder 수치가 Toss 시장 미검증] → 보수 기본값+provenance, §0.9 잠금. tracer·2b가 실측 피드백
- [exit 판정의 시세 관측이 rate budget 경쟁] → 주기 명시·§0.4 내, 실패는 기준선 유지(보류)
- [reconciliation 재배선의 회귀] → 바뀌어야 할 단언 사전 열거(무조건 green 금지)
- [브로커측 보호 부재 구간(2c 전)의 크래시] → 로컬 기준선은 프로세스 사망 시 무력 — **이 구간에서 자동 진입을 켜지 않는다**(게이트 OFF 유지가 전제이며, 게이트 ON은 2c 이후 + attestation. proposal에 명시)

## Migration Plan

journal v6 단일 원자(D7 표). 직전 백업 → 실패 시 복원.

## Open Questions

- 운영 파라미터 수치(사용자 확정 대기 — 미확정 시 small_live)
- ratchet 트리거·잠금 수치의 Toss 적합성 — tracer 결과로 재검토(보수 방향만)
