## ADDED Requirements

### Requirement: 전략 설정과 실행 권한은 분리해 표시된다
콘솔은 a050의 `strategy-runtime` 카테고리에서 전략 파라미터, lane desired/effective 상태, 자동 기동과 LIVE 주문 승인을 별도 section과 별도 action으로 제공해야 하며 (SHALL), 이를 한 번에 활성화하는 control을 제공해서는 안 된다 (MUST NOT).

#### Scenario: 새 설치
- **WHEN** lane 설정이 처음 표시된다
- **THEN** lane desired와 auto-start 기본값은 OFF이고 LIVE 주문은 별도 미승인 상태다

#### Scenario: 전략 설정 저장
- **WHEN** 운영자가 lane 파라미터만 저장한다
- **THEN** lane, auto-start와 LIVE approval 상태는 바뀌지 않고 적용 시점과 restart 필요 여부를 설명한다

#### Scenario: lane ON 요청
- **WHEN** 운영자가 lane desired state를 ON으로 바꾼다
- **THEN** Guardian, protection, reconciliation과 LIVE approval 결과에 따라 effective 상태와 첫 refusal reason을 별도로 표시한다

### Requirement: 확정되지 않은 lane 값은 기본값으로 꾸미지 않는다
모든 lane field는 label, 쉬운 설명, default, desired/effective, 단위·범위, source/version과 적용 시점을 가져야 한다 (SHALL). proposal-freeze가 끝나지 않은 field는 `미구성 / 읽기 전용`이어야 한다 (MUST).

#### Scenario: 첫 lane 미동결
- **WHEN** StockOS source policy·시장·상수가 아직 동결되지 않았다
- **THEN** UI는 숫자 0이나 추정값을 표시하지 않고 미구성 사유와 선행 문서를 안내한다
