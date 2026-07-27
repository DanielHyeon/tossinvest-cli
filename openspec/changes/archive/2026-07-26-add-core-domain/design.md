# Design: add-core-domain

> 2026-07-26 3판. 판 이력·발견 108건(1판 70 + 2판 38)은 `review.md`. 선행 2a는 GATE PASS·archive — 이 change는 그 메인 스펙 위의 **판단 정책**이며, 재정의·재서술 금지(참조만). 브로커·StockOS 주장에 인용 또는 `[미측정]`/`[미검증]` 태그 필수.

## Context

2a가 만든 것: 결정=journal 참조(클래스별 preimage), `f(decision_id, generation)` 멱등키, 위험 예약(**결정 FK 필수** — 예약은 영속된 결정을 요구한다), RECONCILE 영속·확정 하한, 총계 계산 계약, `internal/riskcalc`, 엔진 Gateway 배선·인터록, `ExposureLimiter`, `DefaultDecisionTTL=60s`. 발급자가 없어 자동 진입은 구조적으로 불가능한 상태다.

2c(브로커측 보호)는 2b 측정 후다. **이 change의 손절은 로컬 판정이다** — 그래서 t0부터 기준선이 존재해야 하고(진입 손절가), 관측 두절은 보호 부재와 같으므로 자동 강화로 탈출하며, 게이트 ON은 인터록 조항(ProtectionReady)으로 기계적으로 막는다.

## Goals / Non-Goals

**Goals**: Guardian 체인·발급자(원자 발급), 운영 모드(모드×클래스·EntryGate 투영), 비용 모델(구조+검증 게이트), 포지션 투영·조정·reconciliation 재배선(SHALL 보존), exit 정책(t0 기준선·워터마크 래칫·ladder·pending 수명주기), 성과(동결 기록), tracer 코드.

**Non-Goals**: 보호주문·조건주문(2c), 구조적 RR 계산·등급배수·신호 트레일·시간 종료·백테스트(P3), MFE/MAE(P3), 전략(P3), 웹(P4+), tracer 실전 실행(verify 트랙).

## Decisions

### D1. 원자 발급 — 결정과 예약은 한 트랜잭션

실장 제약: 예약의 `decision_id`는 NOT NULL FK이고 `checkDecisionReservable`이 영속된 결정을 읽는다 — "예약 후 결정"은 불가능하고, "결정 후 예약"은 크래시 창을 만든다. 해소: **신규 journal 원자 API** `RecordDecisionAndReserve`(결정 삽입+예약 검증·삽입을 하나의 `BEGIN IMMEDIATE`로). 예약 거부 시 결정도 롤백 — 제출 가능한 고아 결정이 없다. Gateway는 EXPOSURE_RAISING 결정 제출 시 HELD 예약을 검증한다(engine-safety delta — 권위 주장의 강제 지점). 실패 reason 분화: LIMIT_REACHED / SNAPSHOT_RECOLLECTION_EXHAUSTED / VERSION_CONFLICT / DECISION_EXPIRED.

TTL은 실장 60초(`execgw.DefaultDecisionTTL` — 제출 경로의 buying-power 왕복 생존을 위해 선택된 값). 신선도 논증: 재수집(예산 10초)은 체인을 재실행하지 않으며, 모드·latch 변화는 Gateway 제출 시 EntryGate 재검사가 잡고, 나머지 체인 입력의 신선도 상한은 TTL이다.

### D2. 체인 — 신호 입력 0, 경계값 ≥

입력은 의도 필드·브로커 스냅샷·journal 상태만. 최소 RR은 **신규 검사**(원본 체인에 없음 — 전략 계층 소관이던 것을 의도 산술로 승격), 순서 권위는 TossOS 표(원본 상대 순서 보존 SHALL 폐기 — 매핑은 `docs/guardian-chain.md` 참고 산출물). 재진입 쿨다운 30분·당일 최대 2회 `[미검증]`. allowlist가 상품 클래스 차단의 구조 대체. 총계 경계값은 ≥(도달=차단 — 예약 실장과 일치), 주문 단위는 포함 상한(issues.md 판정 유지).

### D3. 모드 — 3모드, EntryGate 투영으로 강제

NORMAL / ENTRY_BLOCKED / HALT_ALL. EXIT_ONLY 삭제 — ENTRY_BLOCKED와 행동이 동일해 §0.7 사다리의 무의미 단계였다(2c가 실제 차이를 만들면 재도입). 강제 지점: 모드 전환 = journal 영속 + EntryGate 계좌 latch 투영(landed `checkEntry`가 제출 시 재검사 — 봉인 시퀀스 무변경). HALT_ALL의 추가 규칙(PROTECTION_WEAKENING 거부)은 landed `RecordDecision`이 이미 강제. 자동 강화 목적 상태는 전부 ENTRY_BLOCKED(메인 스펙 "신규 진입 차단" 정합 — HALT_ALL 자동 진입 없음).

### D4. Position 투영·reconciliation 재배선 — SHALL 보존

투영 = 체결+조정 이벤트(인스턴스·시장·decimal·`entry_decision_id`). reconciliation MODIFIED는 메인 스펙 SHALL **전문 보존** 위에 델타만 추가(2판이 삭제한 스냅샷 순서·폐기·lineage 키·오차 0·안정화·상태표·30초·카운터 리셋 복원). 조정은 compare-and-append(기대 이전 값·watermark 재검증·불일치 폐기). 자동 해제는 조정 반영 후 일치 + 신규 release cause(ADJUSTMENT_APPLIED — 2a 저장소의 cause 상수 집합 확장). 외부 포지션은 조정으로 편입하되 exit 대상 아님(+알림). 비교는 심볼 수준 합산(보유 스냅샷 market 차원 `[미측정]`).

### D5. Exit 정책 — t0 기준선, 워터마크, pending, 정책 단일

핵심 수정 4: (1) **t0 기준선 = 진입 손절가**(NOT NULL — "+0.4R 전 무손절" 구멍 봉합), (2) **`high_water` 워터마크**가 R 프로브(원본 `high_since_entry` 복원 — 레벨 단조화, 관측-최고가의 한계 명시), (3) **pending 수명주기**(레벨/rung당 1회·미해소 억제·크래시 복원 — 원본 `pending_order_id` 복원), (4) **포지션당 정책 하나**(RATCHET|LADDER — 원본 `exit/policy_assignment.py` DEFAULT_ASSIGNMENT 구조 복원; 기본값 RATCHET, LADDER는 설정 지정). 후보 합성(`compute_protected_stop` max, strict >)·입력 검증·분모 규칙(누적=초기, rung=잔여)·rung 기본 세트(`[미검증]` KOSPI 튜닝) 이식. STOP_FIRST SHALL 폐기(OHLC 입력 부재 — P3). MAX_RATE 비용 상한 이식(본전 기준선 폭주 방지). 관측: 최신가 1점·기본 5초·§0.4 내·체결 감지 SLO 양보, **두절 60초 → ENTRY_BLOCKED 자동 강화**(무기한 무손절 금지). 확정 하한 캡 시 잔여 pending 유지·알림. 가격 R vs 실현 R 명명 분리.

### D6. 비용 — 구조+검증 게이트 이식, KIS 수치 금지

override는 설정 주입으로 재구현(KIS_* 명명 제거), test_costs_env_override는 검증 게이트 구조에 이식. placeholder ≤ MAX_RATE.

### D7. journal v6 — 단일 원자 (태스크는 전사)

전 decimal TEXT. 직전 백업, ErrSchemaTooNew.

| 테이블 | 컬럼 |
|---|---|
| `positions` | id TEXT PK, account_ref·market·symbol NOT NULL, instance_seq INTEGER NOT NULL, entry_decision_id TEXT REFERENCES decisions(id) (외부 편입은 NULL), state CHECK(FLAT\|OPENING\|OPEN\|SCALING\|CLOSING\|CLOSED), quantity·avg_price TEXT NOT NULL, opened_at, closed_at. UNIQUE(account_ref, market, symbol, instance_seq) |
| `position_adjustments` | id PK, position_id NOT NULL FK, kind CHECK(EXTERNAL\|MANUAL\|UNKNOWN), expected_prev_quantity TEXT NOT NULL, prev/new quantity·avg_price, broker_as_of NOT NULL, evidence, created_at — append-only |
| `operating_modes` | id PK, account_ref NOT NULL, mode CHECK(NORMAL\|ENTRY_BLOCKED\|HALT_ALL), cause·actor CHECK(AUTO\|OPERATOR), created_at — append-only, 현재=최신 행 |
| `exit_states` | position_id TEXT PK FK, policy_kind CHECK(RATCHET\|LADDER), entry_price·initial_stop·initial_risk TEXT NOT NULL, baseline_price TEXT NOT NULL, high_water TEXT NOT NULL, ratchet_level CHECK(NONE\|HALF_RISK\|BREAKEVEN\|PARTIAL_LOCK\|PROFIT_LOCK), active_rung INTEGER, taken_ratio_total TEXT NOT NULL DEFAULT '0', pending_action TEXT, pending_level TEXT (RATCHET 레벨 또는 LADDER rung 인덱스), pending_intent_id TEXT, completed INTEGER DEFAULT 0, updated_at |
| `exit_events` | id PK, position_id NOT NULL FK, observed_price·high_water·baseline_after TEXT, level_after·action·proposed_intent_id, created_at — append-only 판정 이력(provenance 조인 경로) |
| `trade_outcomes` | position_id PK FK, realized_pnl_after_costs·realized_r·initial_risk·initial_quantity TEXT, held_seconds INTEGER, exit_ratchet_level·exit_rung, closed_at — CLOSED 트랜잭션에서 동결, 보존 180일 |

ladder rung 정책·allowlist·쿨다운·한도는 config(audit 계약 적용). **원자 apply hook**: 체결 반영 트랜잭션이 주입된 투영·exit 적용 함수를 tx-scope에서 호출(taken_ratio·pending 해소는 여기서만).

### D8. tracer

allowlist 심볼 1개·LIMIT·최소 수량 진입→ratchet 판정→청산 end-to-end, httptest 검증. 실전 실행은 verify 트랙 — 그리고 인터록 조항 6(ProtectionReady)이 게이트 ON을 기계적으로 막는다.

## Risks / Trade-offs

- [ratchet·ladder·쿨다운 수치 미검증] → 보수 기본값+provenance+§0.9 잠금, tracer·2b 피드백
- [관측 표본 사이 트리거 누락] → 상방은 워터마크가 복구(레벨 단조화). **하방 표본 누락**(표본 사이 기준선 하회 후 반등)은 다음 관측에서만 잡힌다 — 브로커측 보호 부재 구간의 잔존 리스크로 명명하며 2c의 브로커 상주 stop이 이 창을 닫는다
- [2c 전 로컬 손절의 크래시 무력] → 인터록 조항 6 + 관측 두절 자동 강화 + verify 트랙 한정 tracer
- [reconciliation 재배선 회귀] → 바뀌어야 할 단언 사전 열거
- [과대 추정 비용의 조기 본전 청산] → MAX_RATE 상한 + 2b 실측 교체(수익 최적화는 실측 후)

## Migration Plan

v6 단일 원자(D7 표). 직전 백업 → 실패 시 복원.

## Open Questions

- 운영 파라미터 수치(사용자 확정 대기 — 미확정 시 small_live 5필드 전체 집합)
- ratchet·rung 수치의 Toss 적합성 — tracer 결과로 재검토(보수 방향만)
