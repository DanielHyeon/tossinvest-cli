# Tasks: add-core-domain

> [M]=Manager, [T]=Teammate(Opus). TDD, 체크박스는 산출물 커밋과 동일 커밋(**경로 스테이징만** — `git add -A` 금지). 3판 — design D1~D8 정본, 태스크는 전사.
> **경계**: 보호주문·조건주문 코드 0줄(2c — ProtectionReady 표지는 미충족 상수로만 존재). 신호 계층 입력 0(P3). 브로커·StockOS 가정에 인용 또는 `[미측정]`/`[미검증]` 태그 없으면 반려. StockOS 절대 경로 `/mnt/D/project/axipient/stockos` 읽기 가능이 이식 선행조건.
> upstream 수정 예정 없음(대상 파일 전부 TossOS 생성). 발생 시 Pre-Edit 전문 → Manager 승인.
> 파일 앵커: `internal/journal/{decision,reservations,reconcile_states,fills,execution_contract}.go`, `internal/execgw/{guardian,gateway,retry,symbolgate,issue}.go`, `internal/riskcalc/`, `internal/reconcile/{compare,mismatch}.go`, `internal/app/engine/{engine,interlock,reads}.go`, `cmd/tossctl/soak_test.go:392`(drift guard), StockOS `{guardian.py,costs.py,exit/baseline_ratchet.py,exit/protected_stop_candidate.py,profit_ladder.py}`

## 0. journal v6 스키마 [T]

- [x] 0.1 design D7 표의 전사(단일 원자 마이그레이션): `positions`(entry_decision_id FK)·`position_adjustments`(expected_prev)·`operating_modes`·`exit_states`(entry/initial_stop/initial_risk/high_water/policy_kind/pending 3컬럼, baseline NOT NULL)·`exit_events`·`trade_outcomes` + 제약 자구 일치. 백업·복원·전이·구버전 거부 계약 테스트(2a 패턴)
- [x] 0.2 [T][High-risk] 원자 발급 API `RecordDecisionAndReserve`: 결정 삽입+예약 검증·삽입 한 트랜잭션, 거부 시 전체 롤백(고아 결정 없음 테스트), 실패 reason 매핑 명시 — LIMIT_REACHED←ErrReservationLimitExceeded / SNAPSHOT_RECOLLECTION_EXHAUSTED←ErrRecollectionExhausted(내부 재시도로 stale·superseded는 종단에서 여기 수렴) / VERSION_CONFLICT←단발 Reserve의 stale·superseded / DECISION_EXPIRED←만료(신규 sentinel 필요 — 현재 ErrInvalidRequest)
- [x] 0.3 [T][High-risk] tx-scoped apply hook: 체결 반영 트랜잭션이 주입된 투영·exit 적용 함수를 tx-scope에서 호출 — journal 공개 API 설계 문서화, hook 밖에서 taken_ratio·pending을 쓸 수 없음을 테스트

## 1. 비용 모델

- [x] 1.1 [T] `internal/costs`: 구조(시장별 수수료·거래세 bps·실질 본전) + 검증 게이트(비수치·NaN·음수·MAX_RATE=0.05 초과 거부) 이식, KIS 수치·명명 미이식, 설정 주입 override, placeholder ≤ 상한·"미검증" provenance. test_costs+env_override(16)를 게이트·주입 구조에 이식
- [x] 1.2 [T] 청산 게이트 미적용·본전 폭주 방지 테스트(§0.3, 상한 초과 설정 거부)

## 2. Guardian 판정 체인

- [x] 2.1 [T][High-risk] 체인 골격·reason enum(STOP_MISSING·STOP_NOT_BELOW_ENTRY·TARGET_NOT_ABOVE_ENTRY·INVALID_TARGET_STOP·TARGET_BELOW_BREAK_EVEN·INVALID_ORDER_SIZE·MAX_ORDER_EXCEEDED·INSUFFICIENT_CASH·SYMBOL_NOT_ALLOWED·MIN_RR_NOT_MET·재진입·총계·중복)·`docs/guardian-chain.md`(이식/신규/제외/구조대체 분류표) — test_guardian(20) 범위 내 이식
- [x] 2.2 [T][High-risk] 손절 계약·수량(배수 1.0)·최소 RR 2.0(신규·순수 산술·provenance §22 lock)·long-only — test_target_stop_contract(29)·test_a090(36) 범위 내
- [x] 2.3 [T][High-risk] 크기·현금(비용 포함·본전 미달 target 거부)·allowlist·재진입(쿨다운 30분·당일 2회 `[미검증]`)·총 노출·일손실(riskcalc 소비·경계 ≥·자본 0 즉시)·중복 — 각 검사 분리 커밋 가능
- [x] 2.4 [T] 진입 latch 통합(401/403·SLO·RECONCILE·recovery) — 기존 EntryGate 사유 매핑, 중복 판정 없음

## 3. 운영 모드

- [x] 3.1 [T][High-risk] 3모드×클래스 표 구현(`operating_modes` 영속·재시작 복원·보수 우선), **EntryGate 투영 강제**(전환=영속+latch 투영, landed checkEntry 소비 — 봉인 시퀀스 무변경 확인)
- [x] 3.2 [T][High-risk] 방향 비대칭: 보수 자동·즉시(트리거→목적 상태 전부 ENTRY_BLOCKED: 일손실·401/403·critical outbox·exit 관측 두절; 분석 실패 비트리거·HALT_ALL 비자동 테스트), 완화는 승인+audit
- [x] 3.3 [T] 전환 알림·구조적 로그·모드 스냅샷 배선(Gateway·Guardian·flatten 동일 뷰)

## 4. Guardian 발급자

- [x] 4.1 [T][High-risk] 발급자: 체인 ALLOW → `RecordDecisionAndReserve` → Gateway 참조 전달. RiskIntent/ReductionIntent 구성, TTL 60s(실장 상수), 예약 거부 시 고아 결정 없음·reason 기록 테스트
- [x] 4.2 [T][High-risk] `ExposureLimiter` 구현(감사 한도 단일 출처·Set 비트 일치)·엔진 Guardian 주입·인터록 조합 테스트 — small_live 5필드로 **조항 1–5 통과 + 조항 6(ProtectionReady)이 유일한 거부 사유**임을 검증(게이트 ON 완전 통과는 2c 후)
- [x] 4.3 [T] 발급 race(동시 다심볼 합산 한도)·발급-제출 사이 모드 강화 시 EntryGate 거부 테스트

## 5. Gateway·인터록 확장 (engine-safety delta 전사)

- [x] 5.1 [T][High-risk] Gateway의 EXPOSURE_RAISING HELD 예약 검증(예약 없는 진입 결정 거부) — RISK_REDUCING 비요구·flatten 무영향 회귀
- [x] 5.2 [T][High-risk] 인터록 조항 6(ProtectionReady — 미충족 상수, 게이트 ON 거부)·가격 조회를 `engine.RequiredEndpoints()`에 추가(soak 목록·retry matrix의 QueryPrice@15s는 **이미 landed** — 실제 델타는 engine 목록+drift guard 통과 확인뿐)

## 6. 포지션 투영·reconciliation

- [x] 6.1 [T][High-risk] `internal/position` 투영(인스턴스·시장·decimal·entry_decision_id)·상태기계 전이표(design 산출물로 표 명문화 후 전 행 테스트)
- [x] 6.2 [T][High-risk] 조정 이벤트 compare-and-append(기대 이전 값·watermark 재검증·불일치 폐기·재수집 테스트)
- [x] 6.3 [T][High-risk] `LocalStateFromJournal` 재배선(투영 소비·심볼 합산)·외부 포지션 편입+exit 제외+알림·release cause ADJUSTMENT_APPLIED 추가(2a cause 상수 확장)·해제 규칙(조정 후 일치=자동, 영구=운영자) — **바뀌어야 할 단언 사전 열거**
- [x] 6.4 [T] provenance 단일 질의(명시 참조 조인: 결정→intent→attempt→fill→position→exit_events→청산) + `docs/aggregates.md` 신설

## 7. Exit 정책 (사용자 요구 — 손익 극대화)

- [x] 7.1 [T][High-risk] `internal/exitpolicy` ratchet 이식: t0 기준선=진입 손절, high_water 프로브, 후보 합성(max·strict >)·입력 검증, 트리거 표 — test_baseline_ratchet 이식(+high_since_entry 케이스) + **3중 단조 property**(기준선·레벨·워터마크)
- [x] 7.2 [T][High-risk] ladder 이식: rung 기본 세트(`[미검증]` provenance)·정책 검증(목표 단조·잠금 비감소)·분모 규칙(누적=초기, rung=잔여) — test_profit_ladder 이식(float→decimal 재작성 명시)
- [x] 7.3 [T][High-risk] pending 수명주기: 레벨/rung당 1회·미해소 억제·거부 시 재무장·크래시 복원(중복발의·미재발의 양방향 테스트), policy_kind 단일
- [ ] 7.4 [T][High-risk] 판정 루프: 최신가 관측(기본 5초·§0.4 예산·SLO 양보·**`RecordSuccess(QueryPrice)` 배선** — 미배선 시 게이트 ON에서 전 진입이 QUERY_STALE 차단됨)·두절 에스컬레이션(15초=landed 쿼리 staleness 진입 차단 → 60초=ENTRY_BLOCKED 모드 강화)·발의(t0 하회 전량·40% 부분·rung)·`exit_events` 기록 — Guardian 위험 감소 경로 경유
- [ ] 7.5 [T][High-risk] 확정 하한 캡 상호작용(잔여 pending·해제 후 재발의·알림)·진입 attempt 지연 창 알림 — §0.3 회귀
- [ ] 7.6 [T] 발의→체결 반영(apply hook에서 taken_ratio·pending 해소) httptest end-to-end

## 8. 성과·tracer

- [ ] 8.1 [T] `trade_outcomes` CLOSED 트랜잭션 동결 기록(실현 R 명명 분리)·집계 파생·180일 비동기 정리(실패 비전파·모드 비강화)
- [ ] 8.2 [T] tracer slice(allowlist 1심볼·LIMIT·최소 수량·파라미터 명시) httptest 검증 — 실전은 verify 트랙 + 인터록 조항 6이 게이트 차단

## 9. 완료 게이트 [M]

- [ ] 9.1 diff 리뷰: upstream 무수정·조건주문/보호 0줄·신호 입력 0·태그 전수·6.3 사전 열거 준수
- [ ] 9.2 `go test ./... -race -count=1` 독립 재실행 green (1785+ 회귀 없음)
- [ ] 9.3 property·crash·race·pending 양방향 테스트 확인, `issues.md` 검토
- 9.4 (게이트 명령 자체) `make gate CHANGE=add-core-domain` 통과 후 완료 선언
- 9.5 (사용자 확인 후) archive — 게이트 ON은 2b attestation + 2c(ProtectionReady) 후에만
