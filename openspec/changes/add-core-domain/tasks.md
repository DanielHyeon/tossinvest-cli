# Tasks: add-core-domain

> [M]=Manager, [T]=Teammate(Opus). TDD, 체크박스는 산출물 커밋과 동일 커밋(경로 스테이징만). 2판 — design D1~D8 정본, 태스크는 전사.
> **경계**: 보호주문·조건주문 코드 0줄(2c). 신호 계층 입력(세션 구조·등급) 0건(P3). 브로커 가정에 openapi 인용 또는 `[미측정]` 태그 없으면 반려. StockOS 절대 경로 `/mnt/D/project/axipient/stockos` 읽기 가능이 이식 태스크 선행조건.
> upstream 수정 예정 없음(엔진·journal·reconcile은 TossOS 생성 파일). 발생 시 Pre-Edit 전문 → Manager 승인.
> 파일 앵커: `internal/execgw/guardian.go`(ExposureLimiter·결정 계약), `internal/journal/{decision,reservations,reconcile_states,execution_contract}.go`, `internal/riskcalc/`, `internal/reconcile/compare.go`(LocalStateFromJournal), `internal/app/engine/{engine,interlock}.go`, StockOS `packages/trading/stockos_trading/{guardian.py,costs.py,exit/baseline_ratchet.py,profit_ladder.py,tradeplan/contract.py}`

## 0. journal v6 스키마 [T]

- [ ] 0.1 design D7 표의 전사(단일 원자 마이그레이션): `positions`·`position_adjustments`·`operating_modes`·`exit_states`·`trade_outcomes` + UNIQUE·FK·CHECK 자구 일치. 직전 백업·복원·전이·구버전 거부 계약 테스트(2a 0.2 패턴 재사용)

## 1. 비용 모델

- [ ] 1.1 [T] `internal/costs`: 구조 이식(시장별 수수료 bps·거래세 bps·실질 본전 산식) — **KIS 수치·`KIS_*` override 미이식**, 과대 추정 placeholder + "미검증" provenance, 설정 주입. test_costs(4)+test_costs_env_override(16) 구조 이식(수치 단언은 Toss 보수값으로 재작성 명시)
- [ ] 1.2 [T] 청산 게이트 미적용 보장: 비용 모델을 소비하는 어떤 경로도 SELL을 차단하지 않음을 테스트로 고정(§0.3 — SELL_COST_BUFFER 미이식 확인)

## 2. Guardian 판정 체인

- [ ] 2.1 [T][High-risk] 체인 골격(순수 함수)·reason-code enum·`docs/guardian-chain.md`(TossOS 순서 ↔ StockOS evaluate_guardian 매핑 표, 이식/제외/구조대체 열거) — test_guardian(20) 중 범위 내 케이스 이식
- [ ] 2.2 [T][High-risk] No Stop=No Trade·위험 기반 수량(배수 1.0 고정)·최소 RR 2.0(순수 산술·provenance §22 lock)·long-only reduce-only — test_target_stop_contract(29)·test_a090(36) 범위 내 이식
- [ ] 2.3 [T][High-risk] 한도·규칙 판정: 주문 크기, 현금·비용(실질 본전 미달 target 거부), 심볼 allowlist, 중복·재진입 쿨다운(StockOS same_day_reentry 이식·수치 provenance), 총 노출·일손실(riskcalc 계산 계약 소비, 자본 0 이하 즉시 차단)
- [ ] 2.4 [T] 게이트 latch 통합: 401/403·SLO 위반·RECONCILE·recovery 미완료를 체인 입력으로 — 기존 EntryGate 사유와 중복 판정 없이 매핑

## 3. 운영 모드

- [ ] 3.1 [T][High-risk] 모드×클래스 표 구현(`operating_modes` 영속, 현재=최신 행, 재시작 복원): HALT_ALL=RISK_REDUCING 허용·PROTECTION_WEAKENING 열 예약, kill switch 동시 적용 시 보수 우선. 표 전 조합 테스트
- [ ] 3.2 [T][High-risk] 방향 비대칭 승인: 보수 자동·즉시(트리거 열거형: 일손실·401/403·critical outbox — 분석 outbox 비트리거 테스트), 완화는 사람 승인+audit
- [ ] 3.3 [T] 전환 critical 알림·구조적 로그, Gateway·Guardian·flatten이 같은 모드 스냅샷을 읽는 배선

## 4. Guardian 발급자·게이트 연결

- [ ] 4.1 [T][High-risk] 발급자: 체인 ALLOW → 예약 트랜잭션(2a `ReserveWithRecollection`) → `RecordDecision`(RiskIntent preimage·멱등키·한도 스냅샷·만료 5s·nonce) → 발급. 예약 거부=RESERVATION_CONFLICT, 재수집 시 체인 재실행 없음 테스트
- [ ] 4.2 [T][High-risk] 위험 감소 발급 경로(ReductionIntent·한도 없음)·`ExposureLimiter` 구현(인터록 단일 출처 통과), 엔진 배선에 Guardian 주입 — 인터록 전 조합 green 통합 테스트
- [ ] 4.3 [T] 발급 race: 동시 다심볼 발급이 합산 한도를 못 뚫음(2a 예약 테스트 위에 발급자 레벨 재검증)

## 5. 포지션 투영·reconciliation 재배선

- [ ] 5.1 [T][High-risk] `internal/position` 투영: 체결+조정 이벤트 → 심볼·시장·인스턴스·평균단가(decimal), 상태기계 전이표(즉시 전량체결·OPENING 종료·SCALING·lineage 승계·CLOSED 종결·불허 전이→RECONCILE) — 표 전 행 테스트
- [ ] 5.2 [T][High-risk] 조정 이벤트: append-only·근거·분류, 투영 수렴, provenance 구분
- [ ] 5.3 [T][High-risk] `reconcile.LocalStateFromJournal` 재배선(투영 소비) + 해제 규칙(조정 반영 후 재조회 일치=자동+원인 기록, 영구 승격=운영자만) — **바뀌어야 할 기존 단언 사전 열거**, 그 외 회귀 없음
- [ ] 5.4 [T] 체결 반영 원자성: 스냅샷·투영·exit 상태 트리거 동일 트랜잭션, 커밋 직후 crash 테스트
- [ ] 5.5 [T] provenance 단일 질의 재구성(결정→intent→attempt→fill→position→exit 판정→청산) + `docs/aggregates.md` 갱신

## 6. Exit 정책 (사용자 요구 — 손익 극대화)

- [ ] 6.1 [T][High-risk] `internal/exitpolicy` baseline ratchet 이식: R 트리거 표(0.4/0.8/1.0/1.2/2.0)·실질 본전(costs 결합)·**단조 상승 불변식**(property 테스트: 임의 가격 시퀀스에서 기준선 비감소) — test_baseline_ratchet 이식 + provenance
- [ ] 6.2 [T][High-risk] profit ladder 이식: rung 전이·판정/체결 필드 분리·STOP_FIRST — test_profit_ladder(전) + test_exit_strategy 범위 내 케이스(ladder·ratchet·breakeven 액션만; 신호 트레일·TIME_EXIT 케이스 제외 목록 명시)
- [ ] 6.3 [T][High-risk] `exit_states` 영속·재시작 복원, 판정 루프: 보유 심볼 최신가 관측(주기·§0.4 예산 명시, 실패=기준선 유지 보류) → 판정 → ReductionIntent 발의(부분익절 40%·기준선 하회 전량) — Guardian 위험 감소 경로 경유
- [ ] 6.4 [T] 발의→체결 반영 루프: 부분익절 체결 시 taken_ratio 이동(체결 시점 필드), 잔여 수량 재계산 — httptest end-to-end

## 7. 성과·tracer

- [ ] 7.1 [T] `trade_outcomes` 기록(CLOSED 트랜잭션은 사실만, 계산·집계·180일 정리는 비동기 — 실패 비전파·모드 비강화 테스트), 집계는 파생
- [ ] 7.2 [T] tracer slice: allowlist 심볼 1개·LIMIT·최소 수량 진입→ratchet→청산 실행기(계좌·notional 상한·신선도·중단 기준 파라미터), httptest 통합 검증 — **실전 실행은 verify 트랙**(attestation+게이트 ON+사용자 승인; 브로커측 보호 부재 구간이므로 2c 전 자동 진입 게이트 ON 금지 명시)

## 8. 완료 게이트 [M]

- [ ] 8.1 diff 리뷰: upstream 무수정, 보호주문·조건주문 0줄, 신호 입력 0건, 태그 전수, 5.3 사전 열거 준수
- [ ] 8.2 `go test ./... -race -count=1` 독립 재실행 green (1785+ 회귀 없음)
- [ ] 8.3 property·crash·race 테스트 확인, `issues.md` 검토
- 8.4 (게이트 명령 자체) `make gate CHANGE=add-core-domain` 통과 후 완료 선언
- 8.5 (사용자 확인 후) archive — 게이트 ON은 2b attestation + 2c 완료 후에만
