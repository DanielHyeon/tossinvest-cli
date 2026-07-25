# Tasks: verify-execution-capability

> 선행: harden-execution-base 태스크 2.10(reason-code)·4.2(attestation 검증) 완료 후 착수 가능. 사용자 협조 필수.

## 1. Soak·실측 [T]

- [x] 1.1 조회 전용 soak 도구 작성 (mutation transport 컴파일 제외, 실 hostname POST 불가 가드 상속)
- [ ] 1.2 자격증명 무인 갱신 soak 연속 3일+ 실행·기록 (사용자 환경)
- [ ] 1.3 rate limit 실측 → retry matrix·폴링 SLO 수치 확정 반영
- [ ] 1.4 attestation 파일 생성 (만료·계좌 식별·성공 endpoint 집합) + 엔진 기동 인터록 연동 확인 — endpoint 집합은 조건주문 등록·취소·정정을 포함해야 하며, `extend-execution-contract` 6.1이 `RequiredEndpoints()`를 확장하면 drift 가드 테스트가 이를 강제한다

## 2. 실계좌 검증 [M+사용자]

- [ ] 2.1 주문 status enum 실측 fixture 수집 → 상태 파생 표 보강
- [ ] 2.2 실계좌 주문 경로 1회성 검증 절차 실행(사용자): 최소 수량·limit-only·즉시 취소, 매도 경계(부분/전량/보유초과), KR cancel/amend
- [ ] 2.3 flatten-all `--dry-run` 리허설 1회 실행·기록
- [ ] 2.4 약관·자동화 허용 범위 검토 기록 + 계정 정지 시 포지션 방침 문서화
- [ ] 2.5 **조건주문 능력 검증**(사용자): 시장·유형별 등록·조회·취소·정정, 프로세스 종료 후 존속, 발동 관측과 발동 주문의 식별 가능성, 정규장 밖 동작, 만료, OCO sibling 취소, 부분체결 잔량, 정정 원자성, 보유수량 예약 — 최소 수량으로 수행하고 결과를 속성별로 기록
- [ ] 2.6 2.5 결과로 SINGLE 단독 vs OCO 시작 유형 결정 + 미검증 시장·유형의 자동 진입 금지 목록 산출
- [ ] 2.7 **멱등키 실동작 검증**(사용자): 일반 주문·조건주문 각각에 대해 동일 `clientOrderId`·동일 본문 재요청이 이전 결과를 그대로 반환하는지, 유효 창이 문서상 10분과 일치하는지, 키가 계좌 스코프인지, 본문이 다르면 `idempotency-key-conflict`가 오는지 — 최소 수량·즉시 취소로 수행. 결과를 attestation 속성으로 기록하며, 미검증이면 멱등 재생 경로는 비활성 유지

## 3. 완료 게이트 [M]

- [ ] 3.1 attestation 유효 확인 + 검증 기록 docs/ 반영
- [ ] 3.2 `make gate CHANGE=verify-execution-capability`
