# Tasks: extend-execution-contract

> [M]=Manager, [T]=Teammate. TDD, 체크박스는 산출물 커밋과 동일 커밋.
> 이 change는 강제 장치만 다룬다 — 판정 정책 수치는 add-core-domain. 모든 테스트는 합성 GuardianDecision으로 가능하다.
> **Pre-Edit 전문 선언 필수**: `internal/trading/conditional.go`는 upstream 파일이다. 1.4 착수 전 수정 범위·이유·CLI 경로 무영향 근거를 보고하고 Manager 승인을 받는다.
> 파일 앵커: `internal/execgw/{gateway,guardian,indoubt,classify}.go`, `internal/journal/{schema,mutation,fills,lineage}.go`, `internal/app/engine/{interlock,engine,broker}.go`, `internal/trading/conditional.go`, `internal/official/conditional_writes.go`

## 1. 조건주문의 Gateway 편입

- [ ] 1.1 [T] `journal.MutationKind`에 조건주문 등록·취소·정정 추가 + 기존 attempt 테이블이 새 kind를 수용하는지 스키마 계약 테스트 (additive, v4 호환)
- [ ] 1.2 [T] 조건주문 canonical 해시·fingerprint(계좌·심볼·방향·트리거가·수량·유형·제출 창) — 일반 주문 해시와 충돌하지 않음을 테스트
- [ ] 1.3 [T][High-risk] `internal/official`에 조건주문 조회(미체결·종결, pagination) 확인 — 없으면 additive 신규 파일로 추가하고 `issues.md`에 근거 기록. 응답에 조건주문 식별자·발동 주문번호가 실리는지 확인해 design.md Open Question 해소
- [ ] 1.4 [T][High-risk] **Pre-Edit 후** 엔진용 조건주문 제출 경로: `trading.Service`에 Gateway 경유 진입점 추가, CLI 경로(확인 토큰) 무변경을 characterization 테스트로 고정
- [ ] 1.5 [T][High-risk] Gateway에 조건주문 mutation 배선: 선기록→DISPATCH_STARTED→분류. httptest로 성공·거부·타임아웃 경로
- [ ] 1.6 [T][High-risk] 조건주문 IN_DOUBT 해소: fingerprint 대조, 미체결·종결 양쪽 pagination 완주, 안정화 N회, 유일 매칭 실패 시 재제출 금지 + UNRESOLVED_IN_DOUBT
- [ ] 1.7 [T] 크래시 주입 테스트: 각 journal 커밋 전후에서 프로세스 중단 → 재시작 시 중복 조건주문이 생기지 않음을 증명

## 2. 발동 주문의 체결 귀속

- [ ] 2.1 [T] journal v5 마이그레이션 (1): `expected_orders` 테이블 — 키·FK·unique 제약 명시, 마이그레이션 전 자동 백업, v4→v5 전이 및 구버전 거부 계약 테스트
- [ ] 2.2 [T] 조건주문 등록 확정 시 예상 주문 기록 (조건주문 식별자·심볼·방향·최대 수량·연결 대상), 취소·종결 시 정리
- [ ] 2.3 [T][High-risk] 체결 귀속 규칙 확장: 로컬 intent 없는 브로커 주문을 예상 주문과 먼저 대조, 매칭 시 lineage 귀속. **매칭 실패 시에만** 외부 주문
- [ ] 2.4 [T][High-risk] 순 보유수량 계산에 귀속 규칙 반영 — 브로커측 손절 발동으로 포지션이 0이 되는 시나리오 테스트
- [ ] 2.5 [T][High-risk] reconciliation 회귀 방지: 기존 외부 주문 판정 테스트 전부 green 유지 확인, 실제 외부 주문이 여전히 외부로 분류됨을 테스트

## 3. Mutation Safety Class

- [ ] 3.1 [T] safety class 분류 함수 (EXPOSURE_RAISING / RISK_REDUCING) — 모든 mutation kind × 방향 조합의 표를 테스트로 고정
- [ ] 3.2 [T][High-risk] 클래스별 직렬화: EXPOSURE_RAISING 심볼당 1건 유지, RISK_REDUCING은 in-flight·IN_DOUBT에 차단되지 않음. RISK_REDUCING은 대상 식별자 단위 직렬화
- [ ] 3.3 [T][High-risk] oversell 방지 수량 상한: 미해소 진입 존재 시 `min(확정 보유수량, 매도가능수량)`, 계좌 조회 stale 시 확정 보유수량으로만 제한
- [ ] 3.4 [T] §0.3 회귀 테스트: 진입 IN_DOUBT·UNRESOLVED_IN_DOUBT·진입 차단 latch·영구 불일치 각 상태에서 보호 제출과 청산이 통과함을 검증

## 4. 결정 계약 강화

- [ ] 4.1 [T] `RiskIntent` 타입 + canonical 직렬화·해시 (정책 버전 포함), 결정론성 테스트
- [ ] 4.2 [T][High-risk] `GuardianDecision`에 RiskIntent 해시 추가 + Gateway 재검증. 손절가 바꿔치기 제출 거부 테스트
- [ ] 4.3 [T][High-risk] `Limits` fail-closed: 필수 항목별 configured 비트, 양수·유한·통화 일치 검증, 총 개방 노출·일일 손실 항목 추가. **빈 스냅샷 전체**는 위험 감소 mutation을 계속 통과시킴을 테스트로 고정
- [ ] 4.4 [T] 위험 감소 발급자 계약: 해당 클래스 결정은 빈 한도 스냅샷. 주문 한도 초과 청산이 통과하는 테스트
- [ ] 4.5 [T] journal v5 마이그레이션 (2): `spent_nonces` 테이블 + journal 기반 `NonceStore`, 재시작 후 재사용 거부 테스트, 보존 기간 정책

## 5. 원자적 위험 예약

- [ ] 5.1 [T] journal v5 마이그레이션 (3): `risk_reservations` 테이블 (키·상태·만료·연결 nonce)
- [ ] 5.2 [T][High-risk] 예약 API: `BEGIN IMMEDIATE` 안에서 노출·현금 조회 → 한도 대조 → 예약 삽입. 동시 다심볼 결정 중 하나만 성공하는 race 테스트(`-race`)
- [ ] 5.3 [T][High-risk] 예약 수명주기: 체결·취소 확정 / 만료 / 제출 실패 확정에서 해제, IN_DOUBT는 해소 전까지 유지. 각 경로 테스트
- [ ] 5.4 [T] 고아 예약 회수: 재시작 시 만료된 예약 정리, 미해소 IN_DOUBT 연결 예약은 보존

## 6. 게이트 인터록 강화

- [ ] 6.1 [T][High-risk] `RequiredEndpoints()`에 조건주문 등록·취소·정정 추가 — attestation에 조건주문이 없으면 기동 거부
- [ ] 6.2 [T][High-risk] 거래 정책 검증: 매도·조건주문·실주문 실행이 모두 허용되지 않으면 기동 거부 (naked long 방지)
- [ ] 6.3 [T][High-risk] 한도 단일 출처: Guardian을 감사된 설정 한도에서 구성하고, 주입된 결정 한도가 그것과 같음을 테스트
- [ ] 6.4 [T] 미충족 조합 전수 통합 테스트: 한도 항목별 누락 × attestation 상태 × 정책 조합의 기동 거부 매트릭스
- [ ] 6.5 [T] 게이트 상태 audit·구조적 로그에 새 전제조건 반영

## 7. 완료 게이트 [M]

- [ ] 7.1 diff 리뷰: upstream 수정은 1.4의 Pre-Edit 승인 범위 내인지, CLI 조건주문 경로 무영향 확인
- [ ] 7.2 `go test ./... -race` 독립 재실행 green (P1 1308 테스트 회귀 없음)
- [ ] 7.3 crash 주입·race 테스트 결과 확인, `issues.md` 기록 검토
- 7.4 (게이트 명령 자체) `make gate CHANGE=extend-execution-contract` 통과 후 완료 선언
