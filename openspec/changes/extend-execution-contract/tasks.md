# Tasks: extend-execution-contract

> [M]=Manager, [T]=Teammate(Opus). TDD, 체크박스는 산출물 커밋과 동일 커밋. 4판(접합부 사양화 — design.md D1~D9가 정본, 태스크는 전사).
> **경계**: 조건주문·보호주문·발동 주문·청산 예약 관련 코드 0줄(2c). 브로커 동작 가정에 openapi 인용 또는 `[미측정]` 태그가 없으면 리뷰 반려.
> **Pre-Edit 전문 선언 필수(upstream)**: `internal/orderintent/intent.go`, `internal/official/orders_write.go`, `internal/trading/`(엔진 진입점), `internal/app/engine/engine.go`(봉인·배선), `internal/journal/durability.go`(PrepareRequest 공개 계약 확장), `cmd/tossctl/flatten.go`(자체 배선 전환). 착수 전 범위·이유·무영향 근거 보고 → Manager 승인.
> 파일 앵커: `internal/execgw/{gateway,guardian,indoubt,classify,fingerprint,retry,symbolgate}.go`, `internal/journal/{schema,durability,dispatch,fills,recovery,resolution}.go`, `internal/brokerstate/derive.go`, `internal/reconcile/{mismatch,recovery}.go`, `internal/app/engine/{interlock,engine}.go`, `internal/flatten/{flatten,liquidate}.go`, `cmd/tossctl/flatten.go:223`(TradingService 소비자 — 7.4 자체 배선 전환 대상)

## 0. journal v5 스키마 [T]

- [x] 0.1 design D9 표의 전사(단일 원자 마이그레이션): `decisions`(account_ref·preimage_kind 포함)·`risk_reservations`(attempt_id 포함)·`spent_nonces`·`reconcile_states`·`execution_corrections` + `mutation_attempts` additive(decision_id·safety_class·generation·client_order_id·wire_body·serializer_version·replay_count·last_replay_at) + `fill_snapshots.filled_amount` + UNIQUE 제약 전부. 컬럼·제약은 표와 자구 일치
- [x] 0.2 직전 자동 백업·복원 절차·실패 무손상 테스트, v4→v5 전이·구버전 거부 스키마 계약 테스트

## 1. 결정 영속·멱등키

- [x] 1.1 [T][High-risk] **Pre-Edit 후** `orderintent.PlaceIntent`·`official.orderCreateV0/V1`·응답 파서에 `clientOrderId` 배선(현재 일반 주문 경로에 필드 없음). `CanonicalPlace` 미포함 — CLI confirm token 무변경 characterization
- [x] 1.2 [T] 키 유도 `f(decision_id, generation)` (≤36자, `[a-zA-Z0-9\-_]` — openapi 패턴), 결정론성 테스트. generation은 2a에서 0 고정
- [x] 1.3 [T][High-risk] **Pre-Edit 후** `PrepareRequest` 확장(공개 계약 — durability_test 갱신 목록 사전 열거): RECORDED에 decision_id·safety_class·generation·멱등키·canonical wire body·serializer_version 불변 영속
- [x] 1.4 [T][High-risk] 결정 영속·검증: 발급자가 Gateway 호출 전 `decisions` 기록(클래스별 preimage: RiskIntent/ReductionIntent), Gateway는 journal에서 읽은 preimage·키로 재검증(호출자 공급 값 금지). 바꿔치기·키 불일치 거부 테스트
- [x] 1.5 [T][High-risk] class↔형태 일치 검증: Gateway가 노출 증가 여부를 독립 계산, EXPOSURE_RAISING ⇔ raisesExposure 불일치 거부(BUY+RISK_REDUCING 거부 테스트). 한도 면제를 `KindCancel` 리터럴 → 검증된 class 기준으로 재작성(guardian.go:181-183)
- [x] 1.6 [T][High-risk] flatten 결정의 1급화: `decisionFor`가 ReductionIntent preimage를 journal에 기록 후 제출 — flatten 동작(취소→매도 saga·한도 미적용) 무변경 회귀 고정
- [ ] 1.7 [T] `Limits` fail-closed(항목별 configured·양수·유한·통화, 총 노출·일손실 추가) + journal NonceStore(소비를 MarkDispatchStarted 트랜잭션에 병합, 보존 ≥ 최대 TTL 불변식, 재시작 재사용 거부)

## 2. 멱등 재생 골격 (2b attestation 전 비활성)

- [ ] 2.1 [T][High-risk] 자기 방어 진입점: 입력은 attempt id뿐, 저장 wire_body 외 전송 불가(새 본문 구성 API 부재를 정적 증명). 진입점 자신이 검증: state==IN_DOUBT·attestation 플래그·**회당** `elapsed < TTL−margin`(기본 60s 설정 주입)·상한 2회·최소 간격
- [ ] 2.2 [T][High-risk] 응답 규칙(dispatch 분류기 사용 금지): 2xx+식별자→CONFIRMED / `409 request-in-progress`→대기·상한 미소비(openapi) / `422 idempotency-key-conflict`→FAILED 금지·UNRESOLVED+critical(openapi) / 유실→기록 후 조회 폴백. 각 분기 테스트
- [ ] 2.3 [T][High-risk] 복구 호출 그래프: 재시작 복구가 순서 소유 — 미종결 순회 → IN_DOUBT·적격 시 Gateway 재생 → 부적격·실패 시 조회 폴백. Resolver 무-writer 불변식 유지 확인
- [ ] 2.4 [T][High-risk] 조회 폴백 정정: **orderId dedup 후 유일성 판정**(PARTIAL_FILLED가 OPEN·CLOSED 양쪽 그룹 — openapi status param 인용), 관측 창 중 동일 심볼 mutation 개입 시 delta 교차확인 무효→자동 FAILED 금지·UNRESOLVED. 부분 체결+응답 유실 시나리오 테스트
- [ ] 2.5 [M] 정정 확인: P1 아카이브 스펙 "멱등키 없음" 서술, `indoubt.go:9-12`·`retry.go:8-10,77`의 동일 근거 문구, Retry Matrix MODIFIED 반영

## 3. 진입 측 위험 예약

- [ ] 3.1 [T][High-risk] 예약 트랜잭션: 스냅샷 밖 수집→안 as-of·staleness 검증→삽입, journal 메서드 재진입 금지·락 순서 문서화, 재수집 상한 3회·데드라인·초과 fail-closed. decimal 산술(소수점 케이스)
- [ ] 3.2 [T][High-risk] 해제: 파생된 종결 상태(5.1 재작성이 **선행 의존**)와 같은 트랜잭션 원자 해제 / nonce 미소비 만료 / 운영자 해제 경로(UNKNOWN·UNRESOLVED의 유일 출구 — 근거·audit) / 거래일 소멸. "만료 추정 해제 금지" 테스트(CLOSED+fill0+무취소=UNKNOWN 유지 시 예약 보존+운영자 알림)
- [ ] 3.3 [T] 재시작 고아 예약 회수 + as-of 조건이 실제 재검증을 유발하는 동시성 테스트(`SetMaxOpenConns(1)` 하에서 `-race`만으로는 불충분 — 스냅샷 버전 어긋남 주입)

## 4. RECONCILE 상태

- [ ] 4.1 [T][High-risk] `reconcile_states` 영속 + 권위 관계 구현: journal이 권위, EntryGate reconcile 계열 래치·Tracker 상태는 기동 시 재구성되는 투영(mismatch.go의 메모리 상태 이전). 재시작 차단 유지 테스트
- [ ] 4.2 [T][High-risk] 확정 하한 공식 구현: `max(0, min(신선 보유, 신선 매도가능) − 로컬 미체결 SELL)`, 부재 시 0, 로컬 파생으로 상향 금지, 매도가능은 하향 방향만 `[의미 미측정 — 2b 2.8]`. flatten 무영향(자체 신선 조회) 회귀 테스트 — §0.3

## 5. 브로커 취급 정정

- [x] 5.1 [T][High-risk] `brokerstate` 파생 **재작성**(확장 아님): 문서화된 10값 OrderStatus 우선순위 표(openapi 인용 주석), canceledAt·수량·lineage 모순·미지 값 UNKNOWN fail-closed 유지. 기존 테스트 중 바뀌어야 할 단언 사전 열거(무조건 green 금지). CANCEL_REJECTED/REPLACE_REJECTED 별도 레코드 인지 — 귀속 실패는 RECONCILE `[형태 미측정 — 2b 2.1]`
- [x] 5.2 [T] opaque 식별자: 공백 검사 후 원문 저장(위반 3개소 수정 — classify.go:149·resolution.go:42,47,126·indoubt.go:512,516), 바이트 동일 비교, round-trip 실패→IN_DOUBT(MarkAcked 후·Settle 전 배치), 상충→RECONCILE. 정규식·prefix 검증 미추가(리뷰 항목)
- [x] 5.3 [T] EXECUTION_CORRECTION: filldetect payload에 `filledAmount` 추가, RecordFill 동일 `BEGIN IMMEDIATE` 내 정정 이벤트+스냅샷 갱신, 반복 poll 멱등 테스트. 5.2 잔여 원문 저장 위반 3곳(`lineage.go:118`·`fills.go:173/385/393`·`filldetect/payload.go:84`) 동반 수정 — issues.md Manager 배정

## 6. 총계 계산 계약

- [x] 6.1 [T][High-risk] 계약 산출물(문서+타입) — `internal/riskcalc`(leaf, stdlib만, 순수 함수·주입 입력): design·스펙의 구조 결정 전사(자동 진입 LIMIT 전용·gross long·실현 손실 기준·환율 staleness fail-closed·외부 거래는 브로커 스냅샷 경유) + **D6a 보수 기본값 전사**(계좌 스냅샷 10초·환율 60초 — 2d는 보수 방향 또는 사람 승인·audit로만 대체). stale·미지 fail-closed 테스트

## 7. 엔진 배선·인터록 (순서 고정: 7.1→7.2→7.3→7.4→7.5)

- [ ] 7.1 [T] 계좌 해석 무조건화: 게이트 OFF에서도 엔진 기동 시 계좌 해석(`GET /api/v1/accounts`) — interlock 내부에서 분리
- [ ] 7.2 [T][High-risk] journal 편입: 엔진 프로필이 journal open — 파일시스템 allowlist·무결성 검사가 기동 조건이 됨을 문서화, config-dir 격리 보존, tmpfs 테스트는 격리 경로
- [ ] 7.3 [T][High-risk] **Pre-Edit 후** Gateway 구성(journal·EntryGate journal-투영 재구성·해소기·NonceStore·예약 저장소) → `runInterlock`은 구성 후 실행
- [ ] 7.4 [T][High-risk] **Pre-Edit 후** `Context.TradingService` 봉인 + flatten을 자체 배선으로 전환(엔진 Context 미의존) — 정적 봉인 테스트, flatten P1 동작 characterization
- [ ] 7.5 [T][High-risk] 인터록 강화: 한도 항목별 fail-closed·거래 정책(매도+실행)·단일 출처(EXPOSURE_RAISING 한정)·Gateway 구성 확인(**round-trip용 Orders 배선 포함** — issues.md)·키 미지원 transport 거부. 미충족 조합 전수 통합 테스트 + audit·구조적 로그

## 8. 완료 게이트 [M]

- [ ] 8.1 diff 리뷰: Pre-Edit 승인 범위, CLI confirm token 무변경, 조건주문 코드 0줄, 브로커 가정 태그 전수 확인
- [ ] 8.2 `go test ./... -race` 독립 재실행 green — 기존 단언 변경은 1.1(clientOrderId 배선)·1.3(PrepareRequest·durability_test)·5.1(파생 재작성)·5.2(원문 저장)·7.4(봉인 — seal_test.go:21 "TradingService is deliberately not caught" 주석 **반전** 포함, engine_test·precheck_test·wts_isolation_test)의 **사전 열거 목록만** 허용, 그 외 회귀 없음
- [ ] 8.3 crash 주입·동시성·예약 누수·재생 경계 테스트 결과 확인, `issues.md` 검토
- 8.4 (게이트 명령 자체) `make gate CHANGE=extend-execution-contract` 통과 후 완료 선언
