# execution-verification Specification (delta)

## ADDED Requirements

### Requirement: Capability Attestation 생성
공식 API 무인 운영 전제는 실측으로 검증되어 attestation으로 기록되어야 한다(SHALL): 자격증명 무인 갱신 연속 3일 이상 soak, rate limit 실측치, 주문·체결·잔고 조회 완전성 확인 결과를 만료 시각·계좌 식별·성공 endpoint 집합과 함께 로컬 durable 파일로 기록한다. soak 도구는 조회 전용이며 mutation transport를 포함해서는 안 된다(SHALL NOT). 미충족 항목이 있으면 attestation은 생성되지 않는다(SHALL).

#### Scenario: soak 성공 후 attestation 생성
- **WHEN** 3일 soak과 실측 항목이 모두 통과하면
- **THEN** 만료 시각·계좌·endpoint 집합을 담은 attestation 파일이 생성되고 엔진 기동 인터록 검증을 통과한다

#### Scenario: soak 도구의 mutation 시도
- **WHEN** soak 도구 코드에 mutation 호출이 포함되면
- **THEN** transport 가드 테스트가 실패한다

### Requirement: 실계좌 주문 경로 검증 기록
자동화 게이트 활성화 전에 실계좌 1회성 검증(사용자 실행·승인)이 완료·기록되어야 한다(SHALL): 최소 수량·limit-only·즉시 취소 규칙으로 매도 경계(부분/전량/보유초과)와 KR cancel/amend를 확인하고, flatten-all `--dry-run` 리허설을 1회 수행한다. 결과와 status enum 실측 fixture는 docs/에 기록되어 상태 파생 표를 보강한다(SHALL).

#### Scenario: 검증 미완료 상태의 게이트 활성화 시도
- **WHEN** 실계좌 검증 기록 없이 자동화 게이트를 켜려 하면
- **THEN** attestation 부재로 엔진 기동이 거부된다
