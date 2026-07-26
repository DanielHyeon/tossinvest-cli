# Tasks: verify-execution-capability

> 선행: harden-execution-base 완료(됨). 사용자 협조 필수. 이 change의 측정 결과가 2c `add-protection-orders` 스펙 작성의 입력이다 — **2c는 2.5~2.9 완료 전에 작성하지 않는다.**

## 1. Soak·실측 [T]

- [x] 1.1 조회 전용 soak 도구 작성 (mutation transport 컴파일 제외, 실 hostname POST 불가 가드 상속)
- [ ] 1.2 자격증명 무인 갱신 soak 연속 3일+ 실행·기록 (사용자 환경)
- [ ] 1.3 rate limit 실측 → retry matrix·폴링 SLO 수치 확정 반영
- [ ] 1.4 attestation 파일 생성 (만료·계좌 식별·성공 endpoint 집합·속성 결과) + 엔진 기동 인터록 연동 확인 — endpoint 집합은 `engine.RequiredEndpoints()`와 drift 가드 테스트로 동기화한다(2a가 목록을 확장하면 자동 강제; 조건주문 endpoint는 2c가 추가)

- [ ] 1.5 [T] **안내형 실계좌 검증 도구** `tossctl verify run`: 2.1·2.2·2.5·2.7·2.8의 측정을 단계 목록으로 구동 — 각 mutation은 flatten-all 패턴의 TTY typed-confirmation(자동화 플래그 금지), 최소 수량·즉시 취소, 단계별 증거 JSONL 기록(→ attestation 속성 입력), 중단·재개 가능(조건주문 존속 확인은 재실행 2회 구조), 비용은 검증 주문 자체의 execution.commission에서 수집. 자동 테스트가 아니라 운영자 도구다(테스트는 httptest 전용, testenv 가드 상속). 유효 창 경계(2.7)는 의도적 이중 주문 절차임을 단계 안내문에 명시하고 기본 생략(--include-ttl-edge 옵트인)

## 2. 실계좌 검증 [M+사용자]

- [ ] 2.1 주문 status enum 실측 fixture 수집 → 상태 파생 표 보강. **CANCEL_REJECTED/REPLACE_REJECTED "별도 주문 레코드"의 실제 형태**(목록 조회 노출 여부·원주문 링크 유무) 관측 포함 — 2a 브로커 상태 파생과 2c 귀속 규칙의 입력
- [ ] 2.2 실계좌 주문 경로 1회성 검증 절차 실행(사용자): 최소 수량·limit-only·즉시 취소, 매도 경계(부분/전량/보유초과), KR cancel/amend
- [ ] 2.3 flatten-all `--dry-run` 리허설 1회 실행·기록
- [ ] 2.4 약관·자동화 허용 범위 검토 기록 + 계정 정지 시 포지션 방침 문서화
- [ ] 2.5 **조건주문 능력 검증**(사용자): 시장·유형별 등록·조회·취소·정정(신규 ID 발급·기존 ID 무효화 — openapi 문서 확인), 프로세스 종료 후 존속, 발동 관측과 `triggeredOrderId` 노출 지연, 정규장 밖 동작, 만료, OCO sibling 취소 시점, 부분체결 잔량, **SINGLE+MARKET 손절의 실동작**(OCO/OTO는 LIMIT 전용 — openapi), 조건주문 예약이 매도가능수량에 반영되는지, 조건주문과 일반 매도 동시 제출의 거부 의미 — 최소 수량으로 수행, 결과를 ProtectiveCapability 속성(시장·수량 종류·조건 유형·발동 주문 유형·세션·modify 의미·triggeredOrderId 노출·검증 시각·증거 digest)으로 기록
- [ ] 2.6 2.5 결과로 보호 유형 확정(기본 가설: SINGLE+MARKET 손절 단독, 익절은 로컬 청산 — "한 심볼에 브로커측 매도 청구권 1개" 불변식) + 미검증 시장·유형의 자동 진입 금지 목록 산출
- [ ] 2.7 **멱등키 실동작 검증**(사용자): 일반 주문·조건주문 각각 동일 `clientOrderId`·동일 본문 재요청이 이전 결과를 재반환하는지, 유효 창(문서상 10분)과 **안전 마진 산정을 위한 왕복 지연 관측**, 키의 계좌 스코프, 본문 상이 시 `idempotency-key-conflict` — 최소 수량·즉시 취소. **주의**: 유효 창 경계 확인은 의도적으로 두 번째 라이브 주문을 만드는 절차다 — 양쪽 즉시 취소 범위로 한정하고, 사용자가 거부하면 창 경계는 미검증으로 남기고 재생은 보수 마진(TTL/2)으로만 허용하거나 비활성 유지
- [ ] 2.8 매도가능수량 의미 실측: 담보·미결제·미체결 매도·조건주문 예약이 `sellableQuantity`에 반영되는 방식 — 2c 청산 예약 공식의 입력
- [ ] 2.9 실측 비용표(수수료·거래세 bps) 수집 → 2d 비용 모델의 Toss 검증값 입력

## 3. 완료 게이트 [M]

- [ ] 3.1 attestation 유효 확인 + 검증 기록 docs/ 반영, 미검증 속성의 명시적 목록(= 자동 경로 금지 목록) 산출
- [ ] 3.2 `make gate CHANGE=verify-execution-capability`
