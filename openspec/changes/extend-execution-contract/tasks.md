# Tasks: extend-execution-contract

> [M]=Manager, [T]=Teammate. TDD, 체크박스는 산출물 커밋과 동일 커밋.
> 이 change는 강제 장치만 다룬다 — 판정 정책 수치는 add-core-domain. 테스트는 합성 GuardianDecision으로 가능하다.
> **설계 재작성본 기준**(design.md, review.md). 초판 설계(조건주문 = 새 MutationKind, 단건 수량 상한, 2-클래스)는 폐기됐다.
> **Pre-Edit 전문 선언 필수**: `internal/trading/conditional.go`, `internal/official/conditional_writes.go`, `internal/app/engine/engine.go`는 upstream 파일이다. 착수 전 수정 범위·이유·CLI 경로 무영향 근거를 보고하고 Manager 승인을 받는다.
> 파일 앵커: `internal/execgw/{gateway,guardian,indoubt,classify,fingerprint}.go`, `internal/journal/{schema,durability,dispatch,fills,lineage}.go`, `internal/official/{conditional_writes,conditional_reads,orders_write}.go`, `internal/app/engine/{interlock,engine,broker}.go`, `internal/reconcile/compare.go`, `internal/filldetect/detect.go`

## 0. 스키마 [T]

- [ ] 0.1 journal v5 **단일 원자 마이그레이션**: `conditional_intents`(유형·만료일·leg별 방향·트리거가·주문가), `conditional_attempts`(`conditional_order_id` 컬럼 — `broker_order_id` 아님), `expected_orders`(leg 단위·상태·tombstone), `risk_reservations`(진입·청산 양방향), `spent_nonces`, 그리고 attempt의 멱등키 컬럼. 키·FK·unique·append-only 여부 명시
- [ ] 0.2 마이그레이션 직전 자동 백업 + 복원 절차 + 실패 시 DB 무손상 테스트. 구버전 바이너리 거부와 v4→v5 전이의 스키마 계약 테스트. **구버전 실행은 롤백이 아님**을 문서화

## 1. 멱등키와 정체 회수

- [ ] 1.1 [T] Gateway가 attempt별 결정적 멱등키를 발행(≤36자, `[a-zA-Z0-9\-_]`)하고 RECORDED 단계에서 영속. 일반 주문 제출 본문에 실어 보내도록 배선 (`orderintent`→`official` 경로에 이미 필드 존재)
- [ ] 1.2 [T][High-risk] 멱등 재생 절차: 동일 본문·동일 키 재요청 → 응답에서 식별자 회수 → CONFIRMED 종결. **능력 검증 플래그가 꺼져 있으면 이 경로를 타지 않음**을 테스트로 고정
- [ ] 1.3 [T][High-risk] 재생 실패 분기: 유효 창 경과·`idempotency-key-conflict`·재생 미지원 → P1 조회 절차로 폴백. 각 분기 테스트
- [ ] 1.4 [T] 조회는 멱등키로 매칭할 수 없음을 계약으로 고정 (`Order`·`ConditionalOrderDetailResponse` 응답에 필드 없음) — 회귀 방지 테스트
- [ ] 1.5 [M] P1 아카이브 스펙의 "브로커 멱등성 키가 없으므로" 문장 정정이 델타에 반영됐는지 확인, `indoubt.go` 주석의 동일 서술 갱신

## 2. 조건주문 형제 수명주기

- [ ] 2.1 [T][High-risk] **Pre-Edit 후** `official.CancelConditionalOrder`·`ModifyConditionalOrder`가 `conditionalOrderId`를 반환하도록 확장 + `trading.ConditionalBroker` 인터페이스 확장 (현재 out 파라미터에 nil을 넘겨 성공한 취소·정정이 전부 IN_DOUBT로 간다)
- [ ] 2.2 [T][High-risk] **Pre-Edit 후** `trading.Service`에 Gateway 경유 엔진 진입점 추가, CLI 확인 토큰 경로 무변경을 characterization 테스트로 고정
- [ ] 2.3 [T][High-risk] 조건주문 attempt 수명주기: 선기록→DISPATCH_STARTED→분류, `conditional_order_id`에 저장. **조건주문 식별자가 일반 주문 추적 집합·로컬 미체결 목록에 들어가지 않음**을 테스트로 증명 (체결 감지 사이클·reconciliation 무영향)
- [ ] 2.4 [T] 조건주문 fingerprint를 저장된 leg 컬럼에서 재계산 — OTO(양방향 leg) 포함
- [ ] 2.5 [T][High-risk] 조건주문 등록 IN_DOUBT 해소: 멱등 재생 우선, 폴백은 조건주문 목록 미체결·종결 양쪽 pagination 완주 대조, 안정화 N회, 유일 매칭 실패 시 재제출 금지 + UNRESOLVED
- [ ] 2.6 [T][High-risk] 조건주문 정정 IN_DOUBT 해소: modify는 같은 식별자를 반환하므로 amend lineage가 이전되지 않는다 — leg의 트리거가·수량 재조회로 판정하는 별도 절차를 정의·구현
- [ ] 2.7 [T] 크래시 주입: 각 journal 커밋 전후 중단 → 재시작 시 중복 조건주문이 생기지 않음

## 3. 발동 주문 다리

- [ ] 3.1 [T][High-risk] 조건주문 상태 폴러: `GET /api/v1/conditional-orders` 주기 조회로 leg 상태·발동 주문 식별자 관측. **rate budget 명시**(체결 감지와 동일 그룹), 폴러 실패가 보호 경로를 막지 않음
- [ ] 3.2 [T] 예상 주문 레코드 생성·소비·종결: leg 단위, 발동 시 일회성 소비, 브로커측 자동 취소·만료 관측 시 종결, **체결 귀속 완료 후** 정리 + tombstone 보존
- [ ] 3.3 [T][High-risk] 발동 주문 편입: 관측된 발동 주문 식별자를 일반 추적 집합·체결 경로에 조건주문 lineage와 함께 등록
- [ ] 3.4 [T][High-risk] 순 보유수량·reconciliation 반영: 브로커측 손절 발동으로 포지션이 0이 되는 시나리오, 발동 주문이 외부 주문·미체결 불일치로 남지 않음
- [ ] 3.5 [T][High-risk] 거짓 매칭 방지: 만료·종결된 예상 주문이 운영자의 진짜 수동 매도를 흡수하지 않음을 테스트
- [ ] 3.6 [M] reconciliation 회귀 검토: **바뀌어야 할 단언과 불변인 단언을 사전에 열거**하고 그 목록에 대해서만 판정 (무조건 green 요구는 A3를 반만 고친 상태를 고정한다)

## 4. Safety Class

- [ ] 4.1 [T][High-risk] 3-클래스 분류 함수: EXPOSURE_RAISING / RISK_REDUCING / PROTECTION_WEAKENING. 판별자는 종류×방향이 아니라 **대상이 이 포지션의 보호인가**이며 예상 주문 레코드가 권위. 전 조합 표를 테스트로 고정
- [ ] 4.2 [T][High-risk] 클래스별 직렬화: 두 개의 독립 심볼 latch(EXPOSURE_RAISING·RISK_REDUCING 각 1건), PROTECTION_WEAKENING은 대상 조건주문 식별자 단위. RISK_REDUCING이 EXPOSURE_RAISING의 in-flight·IN_DOUBT에 막히지 않음
- [ ] 4.3 [T][High-risk] PROTECTION_WEAKENING 규칙: 원자적 교체의 일부이고 무보호 창이 측정·유계일 때만 허용, audit 필수
- [ ] 4.4 [T] §0.3 회귀 테스트: 진입 IN_DOUBT·UNRESOLVED·차단 latch·영구 불일치 각 상태에서 보호 제출과 청산이 통과

## 5. 청산 수량 예약

- [ ] 5.1 [T][High-risk] 가용 매도 수량 계산: `보유 − 미체결 SELL − 대기 조건주문 예약 − 유효 동시 예약`. **권위는 최근 브로커 스냅샷**, 로컬 파생을 상한 상향 근거로 쓰지 않음
- [ ] 5.2 [T][High-risk] 예약 삽입·해제와 반례 테스트: 보호 100 + 청산 100이 합계 200이 되지 않음, 동시 전량 청산 둘 중 하나만 성공
- [ ] 5.3 [T] 스냅샷 부재·stale 처리: 한계 초과 시 critical 알림 + 최근 값 보수 사용, 스냅샷 전무 시 로컬 매도 금지

## 6. 진입 측 위험 예약

- [ ] 6.1 [T][High-risk] 예약 트랜잭션: 스냅샷은 트랜잭션 **밖에서** 수집, 안에서 as-of·staleness 검증 후 삽입, 불충족 시 롤백·재수집. 트랜잭션 스코프 API와 락 순서 명시, 기존 journal 메서드 재진입 금지
- [ ] 6.2 [T][High-risk] 예약 수명주기: 체결·취소 확정 / 제출 실패 확정 / **nonce 미소비 상태의** 만료에서만 해제. nonce 소비 후 만료는 해제하지 않음. IN_DOUBT 유지, UNRESOLVED는 운영자 해소로만
- [ ] 6.3 [T] 일일 손실 예약의 거래일 경계 소멸 + 재시작 시 고아 예약 회수
- [ ] 6.4 [T] 동시성 검증: `SetMaxOpenConns(1)`이 이미 모든 문장을 직렬화하므로 `-race`만으로는 무의미 — **스냅샷 as-of 조건이 실제 재검증을 유발하는지**를 직접 검증하는 테스트를 작성

## 7. 총계 한도 계산 계약

- [ ] 7.1 [T][High-risk] 계산 계약 산출물: 권위 데이터, 미체결·조건주문의 평가 가격, 통화 정규화·환율 권위·staleness, 실현/미실현 범위, 비용 반영, 시장별 거래일 경계, 예약 합산, 외부 거래 취급. **수치는 대상 아님**
- [ ] 7.2 [T] stale·미지 입력의 fail-closed 구현과 테스트

## 8. 결정 계약

- [ ] 8.1 [T] `RiskIntent` 타입 + canonical 직렬화·해시(정책 버전 포함), 결정론성 테스트
- [ ] 8.2 [T][High-risk] preimage를 결정과 함께 journal에 영속, Gateway는 **journal에서 읽은** preimage로 재검증 (호출자 공급 값 사용 금지). 손절가 바꿔치기 거부 테스트
- [ ] 8.3 [T][High-risk] 결정에 safety class 추가 + 한도 면제를 **class 기준으로 재작성**(현재 `KindCancel` 리터럴 기준). 조건주문 취소가 한도에 걸리지 않음을 테스트
- [ ] 8.4 [T][High-risk] `Limits` fail-closed: 항목별 configured 비트, 양수·유한·통화 일치, 총 개방 노출·일일 손실 항목 추가. EXPOSURE_RAISING의 빈 스냅샷 거부
- [ ] 8.5 [T] journal 기반 `NonceStore` + 재시작 후 재사용 거부 테스트, 보존 기간

## 9. 엔진 배선·인터록

- [ ] 9.1 [T][High-risk] **Pre-Edit 후** 엔진 프로필에 ExecutionGateway 구성 (현재 `execgw.New`는 `cmd/tossctl/flatten.go`에만 존재하고 엔진 Context에 Gateway 필드가 없다). EntryGate·Resolver·NonceStore·예약 저장소 연결, `runInterlock`과의 순서 결정
- [ ] 9.2 [T][High-risk] 엔진 컨텍스트가 조건주문 mutation 메서드를 가진 값을 노출하지 않도록 봉인 + 정적 테스트. CLI 표면 무영향
- [ ] 9.3 [T][High-risk] `RequiredEndpoints()` 확장: 조건주문 등록·취소·정정 + **조건주문 목록 조회**(해소 입력) + **매도가능수량 조회**(수량 예약 입력)
- [ ] 9.4 [T][High-risk] 거래 정책 검증(매도·조건주문·실주문 실행) + Gateway 구성 확인을 인터록에 추가
- [ ] 9.5 [T][High-risk] 한도 단일 출처: Guardian을 감사된 설정 한도에서 구성, 동등성 검증은 **EXPOSURE_RAISING 결정에만** 적용
- [ ] 9.6 [T] 미충족 조합 전수 통합 테스트 + 게이트 상태 audit·구조적 로그

## 10. 완료 게이트 [M]

- [ ] 10.1 diff 리뷰: upstream 수정이 Pre-Edit 승인 범위 내인지, CLI 경로 무영향 확인
- [ ] 10.2 `go test ./... -race` 독립 재실행 green (P1 회귀 없음)
- [ ] 10.3 체결 감지·reconciliation이 조건주문 존재 하에서 정상 동작함을 직접 확인 (A1·A2 회귀 방지)
- [ ] 10.4 crash 주입·동시성 테스트 결과 확인, `issues.md` 기록 검토
- 10.5 (게이트 명령 자체) `make gate CHANGE=extend-execution-contract` 통과 후 완료 선언
