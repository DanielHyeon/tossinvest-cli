# trade-analytics Specification (delta)

## ADDED Requirements

### Requirement: 거래 비용 모델
KRW·USD 거래의 수수료·거래세·환전 비용은 순수 함수 비용 모델로 계산되어야 한다(SHALL — StockOS `costs.py`의 **구조**(수수료 bps·거래세 bps·시장 구분·실질 본전 산식)를 이식하되, **KIS 수치는 이식하지 않는다**(SHALL NOT — StockOS 기본값은 KIS 소매 수수료표·`KIS_*` override 체계다). 수치는 Toss 실측(2b)으로 채우며, 실측 전에는 과대 추정 보수값을 "미검증" provenance와 함께 사용한다(SHALL — 과소 추정은 진입 현금·본전 검증을 낙관적으로 만든다). Guardian의 현금·비용 검증, exit 정책의 실질 본전, 성과 지표의 비용 차감이 같은 모델을 사용한다(SHALL — 이중 정의 금지). 비용 모델은 청산 게이트로 적용되지 않는다(SHALL NOT — §0.3).

#### Scenario: 왕복 비용 계산
- **WHEN** KR 매수·매도 왕복 비용을 계산하면
- **THEN** 수수료·거래세가 반영된 결정적 값이 반환되고 구조 검증 테스트(StockOS test_costs·test_costs_env_override 이식 — 수치 단언은 Toss 값으로 재작성)와 일치한다

#### Scenario: 미검증 수치의 보수 적용
- **WHEN** 실측 전 비용 수치로 진입 판정을 수행하면
- **THEN** 과대 추정 값이 사용되어 판정이 보수적으로 기운다

### Requirement: 성과 원시 지표
거래 완결(청산) 시 다음 지표가 journal에 기록되어야 한다(SHALL): 비용 차감 후 실현손익, R 배수(실현손익 / 초기 위험), 보유 시간, 적용된 exit 정책 단계(도달 ratchet 레벨·ladder rung). 집계 지표(승률·profit factor·MDD)는 원시 기록에서 파생 계산한다(SHALL — 원시 데이터가 권위). MFE/MAE는 시장가 시계열 소스 부재로 범위 밖이다(SHALL NOT — P3 시세 스트림 도입 시).

#### Scenario: 청산 완결 시 기록
- **WHEN** 포지션이 CLOSED로 전이하면
- **THEN** 비용 차감 실현손익·R 배수·보유 시간·도달 exit 단계가 provenance와 연결되어 기록된다

### Requirement: 분석 경로의 격리
성과 지표의 계산·집계와 보존 기간 정리는 주문 실행 경로에서 격리되어야 한다(SHALL). 포지션 종결 트랜잭션은 종결 사실만 원자적으로 기록하고, 지표 계산·집계·삭제는 별도 비동기 작업으로 수행한다(SHALL). 분석 작업의 실패·지연이 체결 반영·청산 처리를 지연시키거나 실행 상태를 되돌리거나 **운영 모드 강화를 유발해서는 안 된다**(SHALL NOT — 모드 강화 트리거는 critical 알림 outbox에 한정된다). 성과 테이블은 보존 기간 정책(기본 180일)을 갖는다(SHALL).

#### Scenario: 분석 계산 실패
- **WHEN** 성과 지표 계산이 실패하면
- **THEN** 포지션 종결은 정상 완료되고 운영 모드는 변하지 않으며 분석 작업만 재시도 대상으로 남는다

#### Scenario: 보존 기간 정리
- **WHEN** 180일이 지난 성과 기록의 정리가 수행되면
- **THEN** 주문 경로의 DB 트랜잭션과 경쟁하지 않는 비동기 작업으로 처리된다
