# Change: verify-execution-capability

## Why

harden-execution-base의 자동화 게이트는 capability attestation 없이 켤 수 없다(기동 인터록). attestation을 생성하는 검증 활동들은 wall-clock(수일 soak)과 사용자 협조(실계좌·약관)에 묶여 있어 코드 change와 분리한다 — 코드는 먼저 gate·archive되고, 이 change가 완료되어야 실전 자동 주문이 열린다.

## What Changes

- 공식 API capability soak: 자격증명 무인 갱신 연속 3일+, rate limit 실측, 주문·체결·잔고 조회 완전성 — 결과를 로컬 durable attestation(만료·계좌·성공 endpoint 집합)으로 기록
- 공식 주문 status enum 실측 확정: 실계좌 응답 fixture 수집으로 상태 파생 표 보강 (미지 값 fail-closed는 이미 코드 계약)
- 실계좌 주문 경로 1회성 검증(사용자 실행): 최소 수량·limit-only·즉시 취소로 매도 경계(부분/전량/보유초과)·KR cancel/amend 확인 + flatten-all `--dry-run` 리허설 1회
- 토스 Open API 약관·자동화 허용 범위 검토 기록, 계정 정지 시 포지션 처리 방침

## Capabilities

### New Capabilities

- `execution-verification`: capability attestation 생성(soak·실측)과 실계좌 주문 경로 검증 기록 — engine-safety의 기동 인터록이 소비하는 산출물의 생성 계약

### Modified Capabilities

(없음)

## Impact

- Affected code: soak 도구(cmd 또는 tools/, 조회 전용 — mutation transport 미포함), fixture 추가
- 완료 전까지 자동화 게이트 ON 불가(엔진 기동 인터록이 attestation을 요구)
- 사용자 액션 필요: soak 실행 환경 제공, 실계좌 검증 승인·실행, 약관 확인
