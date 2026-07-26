# Tasks: extend-execution-contract

> [M]=Manager, [T]=Teammate(Opus). TDD, 체크박스는 산출물 커밋과 동일 커밋.
> **3판 경계**: 조건주문·보호주문·발동 주문·청산 예약은 어떤 태스크에도 없다(2c). 브로커 동작 서술에 openapi 인용 또는 `[미측정]` 태그가 없는 구현 가정은 리뷰에서 반려된다.
> **Pre-Edit 전문 선언 필수(upstream)**: `internal/orderintent/intent.go`, `internal/official/orders_write.go`, `internal/trading/`(엔진 진입점), `internal/app/engine/engine.go`(봉인). 착수 전 수정 범위·이유·CLI 무영향 근거 보고 → Manager 승인.
> 파일 앵커: `internal/execgw/{gateway,guardian,indoubt,classify,fingerprint,failclosed}.go`, `internal/journal/{schema,durability,dispatch,fills,recovery,resolution}.go`, `internal/official/orders_write.go`, `internal/orderintent/intent.go`, `internal/app/engine/{interlock,engine}.go`, `internal/brokerstate/`, `internal/flatten/flatten.go:625`(decisionFor)

## 0. journal v5 스키마 [T]

- [ ] 0.1 design.md D9 표의 전사: `decisions`·`risk_reservations`·`spent_nonces`·`execution_corrections` 테이블 + `mutation_attempts` additive 컬럼(client_order_id·wire_body·serializer_version·replay_count·last_replay_at) + UNIQUE 제약. **단일 원자 마이그레이션**, 컬럼·제약은 표와 자구 일치
- [ ] 0.2 마이그레이션 직전 자동 백업·복원 절차·실패 시 무손상 테스트, v4→v5 전이·구버전 거부 스키마 계약 테스트

## 1. 멱등키 배선

- [ ] 1.1 [T][High-risk] **Pre-Edit 후** `orderintent.PlaceIntent`·`official.orderCreateV0/V1`·응답 파서에 `clientOrderId` 추가 (현재 일반 주문 경로에 필드 없음 — 조건주문 경로에만 존재). `CanonicalPlace`에는 넣지 않음 — CLI confirm token 무변경을 characterization 테스트로 고정
- [ ] 1.2 [T] 결정적 키 생성 `deterministic(intent_id, generation)` (≤36자, `[a-zA-Z0-9\-_]`), 결정론성·충돌 회피 테스트
- [ ] 1.3 [T][High-risk] RECORDED 단계에 키 + canonical wire body + serializer_version 불변 영속. fsync 후 DISPATCH_STARTED 진행은 P1 규칙 유지

## 2. 멱등 재생 골격 (2b attestation 전 비활성)

- [ ] 2.1 [T][High-risk] Gateway 해소 전용 재생 진입점: 저장된 wire_body 재전송만 가능, 새 본문 구성 불가 — 정적 테스트로 두 번째 제출 문이 아님을 증명
- [ ] 2.2 [T][High-risk] 시간 규칙: `elapsed(dispatch_started_at) < TTL(10m 문서값) − margin(기본 60s)` 검사, 경계 초과 시 조회 폴백. margin은 설정 주입(2b 실측으로 조정)
- [ ] 2.3 [T][High-risk] 재생 실행·기록: 응답 orderId 회수→CONFIRMED, replay_count·last_replay_at 갱신, 상한(기본 2회) 초과·`idempotency-key-conflict` 시 조회 폴백. 활성화 플래그는 attestation 항목 확인 시에만 ON — 미검증 상태에서 재생 경로를 타지 않음을 테스트로 고정
- [ ] 2.4 [T] 재생과 nonce의 분리: 재생은 소비된 nonce의 attempt에 대한 정체 회수이며 nonce 재사용 거부에 걸리지 않음 — 테스트
- [ ] 2.5 [M] P1 아카이브 스펙·`indoubt.go` 주석의 "멱등성 키가 없으므로" 서술 정정 확인

## 3. 결정 계약

- [ ] 3.1 [T][High-risk] `decisions` 영속: safety_class·risk_preimage·risk_hash·client_order_id·limits·nonce·generation. 발급 인터페이스는 초안 유지(실구현 2d)
- [ ] 3.2 [T][High-risk] Gateway 재검증을 journal 기반으로: preimage 재계산 대조(호출자 공급 값 금지), 멱등키 결속 검증, 손절가 바꿔치기·키 불일치 거부 테스트
- [ ] 3.3 [T][High-risk] 한도 면제를 `KindCancel` 리터럴 → safety class 기준으로 재작성(`guardian.go:181-183`). EXPOSURE_RAISING 빈 스냅샷 거부, RISK_REDUCING 한도 미적용, 한도 초과 청산 통과 테스트
- [ ] 3.4 [T] `Limits` fail-closed: 항목별 configured 비트·양수·유한·통화 일치, 총 개방 노출·일일 손실 항목 추가
- [ ] 3.5 [T] journal 기반 NonceStore(`spent_nonces`) + 재시작 후 재사용 거부·보존 기간 테스트
- [ ] 3.6 [T] `flatten.decisionFor`에 RISK_REDUCING class 부여 — 기존 flatten 동작 무변경을 회귀 테스트로 고정

## 4. 진입 측 위험 예약

- [ ] 4.1 [T][High-risk] 예약 트랜잭션: 스냅샷 밖 수집 → 안에서 as-of·staleness 검증 → 삽입. 기존 journal 메서드 재진입 금지, 락 순서 문서화. 재수집 상한(3회)·데드라인·초과 시 fail-closed
- [ ] 4.2 [T][High-risk] 해제 규칙: 브로커 종결 상태(FILLED/CANCELED/REJECTED/NOT_DISPATCHED/FAILED_CONFIRMED — **미체결 만료 포함**) / nonce 미소비 만료 / 운영자. nonce 소비 후 만료 미해제, UNRESOLVED 보존. 각 경로 + 만료 주문 누수 방지 테스트
- [ ] 4.3 [T] 일일 손실 예약의 거래일 경계 소멸(시장별, P1 clock) + 재시작 시 고아 예약 회수
- [ ] 4.4 [T] decimal 문자열 산술(float 누적 금지) — 소수점 수량 케이스 테스트
- [ ] 4.5 [T] 동시성 검증: `SetMaxOpenConns(1)`이 문장을 직렬화하므로 `-race`만으로 불충분 — **as-of 조건이 실제 재검증을 유발하는지** 직접 검증(스냅샷 버전 어긋남 주입)

## 5. RECONCILE 상태

- [ ] 5.1 [T][High-risk] 상태 정의·전이: 진입 조건(조회 불가·stale·수량 불일치·식별자 상충), journal 영속, 해제(재조회 일치+원인 기록)
- [ ] 5.2 [T][High-risk] 행동 제한: 진입·수량 확대 거부, 확정 하한 수량의 위험 축소 허용 — §0.3 회귀 테스트(RECONCILE 중 청산이 하한까지 통과)

## 6. 브로커 취급 정정

- [ ] 6.1 [T] opaque 식별자 규칙: 빈 값 거부·원문 저장·변환 금지·계좌 스코프·생성 후 round-trip 확인·상충 시 RECONCILE. 형태 검증 코드 부재를 정적으로 확인
- [ ] 6.2 [T][High-risk] `internal/brokerstate` 파생을 문서화된 OrderStatus 10개로 확장(openapi 인용 주석), 미지 값 UNKNOWN_BROKER_STATE 유지. CANCEL_REJECTED/REPLACE_REJECTED 별도 레코드 인지 — 귀속 실패는 외부 분류가 아닌 RECONCILE `[형태 미측정 — 2b]`
- [ ] 6.3 [T] EXECUTION_CORRECTION: 수량 동일+평균가/금액 변경 관측을 정정 이벤트로 기록, 수량 재반영 없음. dedup 키에 가격·시각 미포함 확인 테스트

## 7. 총계 한도 계산 계약

- [ ] 7.1 [T][High-risk] 계약 산출물(문서+타입): 권위 데이터·평가 가격·통화 정규화·환율 staleness·실현/미실현 범위·비용 반영·거래일 경계·예약 합산·외부 거래 취급. 수치는 2d — placeholder 없이 fail-closed 기본
- [ ] 7.2 [T] stale·미지 입력의 fail-closed 구현·테스트

## 8. 엔진 배선·인터록

- [ ] 8.1 [T][High-risk] **Pre-Edit 후** 엔진 프로필 Gateway 구성(journal·EntryGate·Resolver·NonceStore·예약 저장소), `runInterlock`과의 순서 결정 문서화 (현재 `execgw.New`는 flatten CLI에만 존재)
- [ ] 8.2 [T][High-risk] **Pre-Edit 후** `Context.TradingService` 봉인 — mutation 메서드 노출 제거, 정적 테스트, CLI 무영향 characterization
- [ ] 8.3 [T][High-risk] 인터록 강화: 한도 항목별 fail-closed·거래 정책(매도+실행)·Guardian 단일 출처(EXPOSURE_RAISING 한정)·Gateway 구성 확인. 미충족 조합 전수 통합 테스트
- [ ] 8.4 [T] 게이트 상태 audit·구조적 로그 갱신

## 9. 완료 게이트 [M]

- [ ] 9.1 diff 리뷰: upstream 수정이 Pre-Edit 승인 범위 내, CLI confirm token 무변경, 조건주문 코드 0줄 확인
- [ ] 9.2 `go test ./... -race` 독립 재실행 green (P1 1370+ 회귀 없음)
- [ ] 9.3 crash 주입·동시성·예약 누수 테스트 결과 확인, `issues.md` 검토
- 9.4 (게이트 명령 자체) `make gate CHANGE=extend-execution-contract` 통과 후 완료 선언
