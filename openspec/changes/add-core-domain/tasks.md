# Tasks: add-core-domain

> [M]=Manager, [T]=Teammate. TDD, 체크박스는 산출물 커밋과 동일 커밋.
> **선행 필수**: `extend-execution-contract` 완료. 4·5절은 그 계약(조건주문 Gateway 편입·발동 주문 귀속·safety class·위험 예약) 없이 안전하게 구현할 수 없다.
> **StockOS 경로 선행조건**: `/mnt/D/project/axipient/stockos`를 읽을 수 있어야 한다. 격리 worktree에 마운트되지 않았으면 착수 전 보고할 것.
> 파일 앵커: `internal/execgw/guardian.go`(결정 계약), `internal/app/engine/{interlock,engine,broker}.go`(배선·official-only 증명), `internal/trading/conditional.go`+`internal/official/conditional_writes.go`(조건주문), `internal/reconcile/compare.go`(로컬 포지션 파생), `internal/journal/{schema,fills}.go`
> upstream 파일 수정 예정 없음. 발생 시 Pre-Edit 전문 선언 후 Manager 승인.

## 1. 비용·수량·손절 (StockOS 순수 로직 이식)

- [ ] 1.1 [T] `internal/costs`: KRW/USD 수수료·거래세 비용 모델 (StockOS costs·cost_model 이식). 등급 배수·비용 bps를 개별 값으로 열거, provenance 주석, 미검증 시 과대 추정. test_costs 케이스 이식
- [ ] 1.2 [T][High-risk] `internal/risk` 손절 계약: No Stop = No Trade, long-only(매수 진입 `stop < entry`, 매도는 보유수량 이하 reduce-only, short 금지) — test_target_stop_contract(29) 이식
- [ ] 1.3 [T][High-risk] 위험 기반 수량: `floor(위험예산 × 등급배수 / (entry − stop))`, stop 폭 0 이하면 수량 0(fail-closed) — test_a090(36) 이식
- [ ] 1.4 [T] 구조적 RR 계산(measured-move, cap 6.0, 계산 불가 시 None) + 최소 RR 1.5 미달·계산 불가 거부 — test_structural_rr(14) 이식

## 2. Guardian 판정

- [ ] 2.1 [T][High-risk] 판정 체인 골격(순수 함수): 고정 순서·첫 실패 정지·reason-code 통합, 순서 표 문서화 — test_guardian(20) 이식. **이식 범위 명시**: 제외(KIS·LLM·capital stage·미국장 진입 시간창), 보류 3항목(레버리지/인버스, ETF/ETN, 당일 재진입 쿨다운)은 임의 판단 금지 — Manager 확인 후 결정
- [ ] 2.2 [T][High-risk] 총계 한도 계산 계약: 일일 손실·총 개방 노출의 권위 데이터·통화 정규화(USD→KRW 시세 staleness)·거래일 경계(시장별 TZ/DST)·실현/미실현·비용 반영·외부 거래·미체결 매수와 위험 예약 포함·자본 분모 시점. stale/미지 시 fail-closed
- [ ] 2.3 [T][High-risk] 한도 판정 구현: 주문 크기·총 개방 노출·일일 손실(절대액+%, 자본 0 이하 즉시 차단)·중복/재진입 — 보수 기본값 + provenance
- [ ] 2.4 [T] 정책 수치 provenance 계약: 출처·검증 상태 주석 규약, "Toss 검증됨" 전환은 실계좌 결과에만 근거, 변경 시 audit

## 3. 운영 모드·kill switch

- [ ] 3.1 [T] journal v6 마이그레이션 (1): 운영 모드·전환 이력 테이블 (키·제약 명시, 백업·복원 절차, 스키마 계약 테스트)
- [ ] 3.2 [T][High-risk] 모드 축 구현: NORMAL/ENTRY_BLOCKED/EXIT_ONLY/HALT_ALL 의미, HALT_ALL에서 위험 감소 mutation 통과, kill switch와의 우선순위 조합표 산출물
- [ ] 3.3 [T][High-risk] 방향 비대칭 승인: 보수 방향 전환은 자동·즉시·durable, 완화·해제만 사람 승인 + audit. 손실 한도 도달 시 승인 대기 없이 전환되는 테스트
- [ ] 3.4 [T] 재시작 유지·알림 연동: 모드·kill switch 영속 복구, 전환 critical 알림

## 4. GuardianDecision 발급·게이트 배선

- [ ] 4.1 [T][High-risk] 발급자: 체인 ALLOW → 결정 계약(주문 해시·RiskIntent 해시·한도 스냅샷·만료·nonce) 변환
- [ ] 4.2 [T][High-risk] 위험 예약 연동: 진입 결정은 예약과 같은 트랜잭션에서 발급. 동시 다심볼 한도 초과 방지 race 테스트
- [ ] 4.3 [T][High-risk] 위험 감소 발급 경로: 보호·청산 의도는 빈 한도 스냅샷, 진입 한도 면제, 모드·kill switch는 적용. 주문 한도 초과 청산 통과 테스트
- [ ] 4.4 [T][High-risk] 엔진 Gateway 구성: 엔진 Context에 ExecutionGateway 배선 (현재 `execgw.New`는 `cmd/tossctl/flatten.go`에만 존재), EntryGate·Resolver·NonceStore·예약 저장소 연결
- [ ] 4.5 [T][High-risk] Guardian 주입: 인터록이 감사한 설정 한도 **단일 출처**에서 Guardian 구성, 주입된 결정 한도가 그것과 같음을 테스트
- [ ] 4.6 [T] 미충족 조합 통합 테스트 + 게이트 상태 audit·구조적 로그

## 5. 포지션 원장

- [ ] 5.1 [T] journal v6 마이그레이션 (2): Position·position-instance·조정 이벤트 테이블
- [ ] 5.2 [T][High-risk] 상태기계 전이표: `(상태, 주문 역할, 누적 delta, lineage, 브로커 포지션)` 완전표 — 즉시 전량체결·OPENING 종료·SCALING·정정 lineage 승계·외부 포지션·매도 귀속·CLOSED→새 인스턴스. 체결 delta는 부호 없으므로 intent side에서 방향 재도출
- [ ] 5.3 [T][High-risk] 단일 권위 배선: `internal/position`을 `reconcile.LocalStateFromJournal`과 같은 질의 위의 투영으로 구현 — 두 번째 포지션 계산을 만들지 않음. 두 조회 결과 일치 테스트
- [ ] 5.4 [T][High-risk] 조정 이벤트: reconcile 불일치를 append-only 조정으로 반영(직접 덮어쓰기 금지), 근거·분류 기록, 청산 수량은 `min(로컬, 계좌)`, 차단 해제 조건
- [ ] 5.5 [T][High-risk] 체결 반영 원자성: 체결 스냅샷·Position 투영·보호 목표 수량·작업 등록을 한 트랜잭션으로, 재처리 멱등성 키. 커밋 직후 크래시 테스트
- [ ] 5.6 [T] provenance lineage: 결정 스냅샷 → intent → attempt → fill → position → 보호·발동 주문 → 청산 참조 연결, 단일 질의 재구성 테스트
- [ ] 5.7 [T] `docs/aggregates.md`: 경계·이벤트 흐름 문서

## 6. 보호주문 saga

- [ ] 6.1 [T] journal v6 마이그레이션 (3): ProtectionSaga 상태·이력 테이블
- [ ] 6.2 [T][High-risk] saga 상태 전이표 산출물: DESIRED→RECORDED→DISPATCHED→(ACTIVE|AMBIGUOUS)→(REPLACING|CANCELING|DEGRADED)→CLOSED, 상태별 허용 mutation·timeout·재시작 행동 + crash point 매트릭스
- [ ] 6.3 [T][High-risk] saga 구현: 체결 감지 → 보호 제출(Gateway 경유, stop-first) → 확인 → ACTIVE. 무보호 노출 SLO(기본 10s) 측정·critical 알림
- [ ] 6.4 [T][High-risk] 재시작 복구: 미완 saga 감지, AMBIGUOUS는 해소 후 판단(재제출 금지), 미보호 포지션 critical 알림 — crash/restart 테스트
- [ ] 6.5 [T][High-risk] 수량 정합: transient invariant + 최대 허용 시간, 원자적 정정 우선(`ModifyConditionalOrder`), 미검증 시 취소-후-재등록 폴백, 보호 합계 ≤ 보유수량, CLOSED 시 보호 취소
- [ ] 6.6 [T][High-risk] 폴백 전환 조건: 명시적 미지원 응답만 폴백, timeout·5xx·유실은 조회 해소 우선, 확인 실패 시 진입 잔량 취소 + 운영자 승인 긴급 청산. synthetic은 공격적 limit + 한계 문서화
- [ ] 6.7 [T] 손절 즉시성 회귀 테스트: 진입 IN_DOUBT·UNRESOLVED·차단 latch·EXIT_ONLY·HALT_ALL·영구 불일치 각 상태에서 보호·청산 경로 무영향 (§0.3)

## 7. 성과·tracer

- [ ] 7.1 [T] journal v6 마이그레이션 (4): 성과 테이블 (보존 기간 180일 정책)
- [ ] 7.2 [T] 성과 원시 지표: 청산 완결 시 비용 차감 실현손익·R 배수·보유 시간 기록, 집계는 파생 계산 (MFE/MAE는 범위 밖 — 시세 소스 부재)
- [ ] 7.3 [T] 분석 경로 격리: 종결 트랜잭션은 종결 사실만, 계산·집계·retention은 outbox 비동기. 분석 실패가 실행 상태를 되돌리지 않음을 테스트
- [ ] 7.4 [T] tracer slice: 심볼 1개·limit·최소 수량의 진입→보호→청산 실행기. 계좌·시장·최대 notional·가격 freshness·중단 기준을 파라미터로 명시. httptest 통합 테스트로 검증 — **실전 실행은 이 change 밖**(attestation + 게이트 ON + 사용자 승인)
- [ ] 7.5 [T] live 검증 미완료 시 자동화 게이트가 계속 OFF임을 나타내는 명시적 상태·산출물

## 8. 완료 게이트 [M]

- [ ] 8.1 diff 리뷰(upstream 무수정 확인)·Pre-Edit/race/crash 확인
- [ ] 8.2 `go test ./... -race` 독립 재실행 green
- [ ] 8.3 reconcile·filldetect 기존 테스트 회귀 없음 확인 (5.3·5.4가 건드리는 영역)
- 8.4 (게이트 명령 자체) `make gate CHANGE=add-core-domain` 통과 후 완료 선언
- 8.5 (사용자 확인 후) archive · tracer 실전 실행은 verify change와 함께
