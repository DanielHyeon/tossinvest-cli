# Tasks: verify-execution-capability

> 선행: harden-execution-base 태스크 2.10(reason-code)·4.2(attestation 검증) 완료 후 착수 가능. 사용자 협조 필수.

## 1. Soak·실측 [T]

- [ ] 1.1 조회 전용 soak 도구 작성 (mutation transport 컴파일 제외, 실 hostname POST 불가 가드 상속)
- [ ] 1.2 자격증명 무인 갱신 soak 연속 3일+ 실행·기록 (사용자 환경)
- [ ] 1.3 rate limit 실측 → retry matrix·폴링 SLO 수치 확정 반영
- [ ] 1.4 attestation 파일 생성 (만료·계좌 식별·성공 endpoint 집합) + 엔진 기동 인터록 연동 확인

## 2. 실계좌 검증 [M+사용자]

- [ ] 2.1 주문 status enum 실측 fixture 수집 → 상태 파생 표 보강
- [ ] 2.2 실계좌 주문 경로 1회성 검증 절차 실행(사용자): 최소 수량·limit-only·즉시 취소, 매도 경계(부분/전량/보유초과), KR cancel/amend
- [ ] 2.3 flatten-all `--dry-run` 리허설 1회 실행·기록
- [ ] 2.4 약관·자동화 허용 범위 검토 기록 + 계정 정지 시 포지션 방침 문서화

## 3. 완료 게이트 [M]

- [ ] 3.1 attestation 유효 확인 + 검증 기록 docs/ 반영
- [ ] 3.2 `make gate CHANGE=verify-execution-capability`
