## ADDED Requirements

### Requirement: 시장 스케줄은 desired와 effective를 구분해 설명한다
콘솔은 a050의 `strategy-runtime > 시장·일정` section에서 scheduler와 auto-start의 desired/effective 상태, 시장·세션 범위, calendar version/updated-at, 다음 전환 시각과 typed decision reason을 표시해야 한다 (SHALL).
시장·세션 범위와 운영 reason은 server-defined option으로만 선택해야 하며 (SHALL), 임의 문자열·시간·휴장일 입력 control을 제공해서는 안 된다 (MUST NOT).

#### Scenario: 새 설치
- **WHEN** scheduler 저장값이 없다
- **THEN** scheduler OFF, auto-start OFF, 선택 시장 없음, 정규장만 허용을 기본값과 쉬운 설명으로 표시한다

#### Scenario: 장 닫힘
- **WHEN** desired는 ON이지만 시장이 휴장이다
- **THEN** effective는 WAIT_MARKET이고 다음 세션과 함께 exit/reconcile은 계속됨을 설명한다

#### Scenario: API 예산 대기
- **WHEN** scheduler decision이 BUDGET_DEFERRED다
- **THEN** 사용자 OFF와 구분된 대기 사유 및 safety budget을 침범하지 않는다는 설명을 표시한다

### Requirement: exchange calendar는 운영 근거로 읽기 전용이다
calendar version, source와 updated-at은 표시해야 하지만 (SHALL), 최초 범위에서 사용자가 임의 휴장일을 입력하는 control을 제공해서는 안 된다 (MUST NOT).

#### Scenario: stale calendar
- **WHEN** calendar freshness gate가 실패한다
- **THEN** entry effective 상태는 fail-closed이고 stale reason과 갱신 방법을 표시한다
