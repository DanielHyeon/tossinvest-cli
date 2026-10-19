## Why

a160은 저장된 캠페인 상태를 일관되게 조회하지만 전체 계획 원문과 초기 예산, lane 위험 상태를 복원하지 않는다. 사용자 요청에 따라 초기 예산뿐 아니라 캡처 시점의 체결·예약 위험액과 차단·중복 처리 기록을 함께 보존하는 오프라인 기반을 마련한다.

## What Changes

- continuation/reversal/weekly의 KR·US 계획 및 native RiskState를 단방향으로 내보내고, 버전이 있는 제한된 크기의 자료로 직렬화한다.
- 원래 계획 digest·배분 상한·통화·FX 근거와 lane별 위험 기록의 의미를 그대로 보존한다. 파일 close/reopen 후에는 거래용 객체가 아닌 독립된 검사 자료로 복원한다.
- 복원은 외부에서 지정한 content digest 및 scope와 결속한다. 손상·누락·알 수 없는 버전·불완전 입력은 실패시키고 0이나 신뢰 인장을 만들어 채우지 않는다.
- 현재 위험액은 **캡처 시점에 제공된 상태 값**이다. 현재 실계좌 원장과의 일치·최신성·전이 이력의 완전성은 검증되지 않은 것으로 표시한다. weekly 별도 ReservationState와 취소 이력은 이번 RiskState 복원에 포함되지 않음을 명시한다.

## Capabilities

### New Capabilities

- `campaign-state-evidence`: 비권한 오프라인 계획·위험 상태의 제한된 직렬화와 검사 자료 복원.

### Modified Capabilities

없음. a160의 원장 조회는 계속 full-plan-unavailable을 반환한다.

## Impact

세 lane 패키지의 신규 단방향 export와 stdlib 기반의 작은 공통 codec 및 테스트, OpenSpec/PM만 대상으로 한다. 기존 계획·FX 생성자, 위험 연산, admission, 원장 schema, engine/CLI/API/UI는 변경하지 않는다. 기존 증거·a157·실거래·운영 설정은 보존한다. 실제 admission과 원자적인 계획 저장, fill/cancel 위험 원장 복구와 주문용 객체 재구성은 후속 단계다.
