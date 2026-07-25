# trade-analytics Specification (delta)

## ADDED Requirements

### Requirement: 거래 비용 모델
KRW·USD 거래의 수수료·거래세·환전 비용은 순수 함수 비용 모델로 계산되어야 하며(SHALL), 수치는 provenance 주석(StockOS costs.py·cost_model.py 유래, Toss 실측 검증 상태)과 함께 설정 주입으로 교체 가능해야 한다. Guardian의 현금·비용 검증과 성과 지표의 비용 차감이 같은 모델을 사용한다(SHALL — 이중 정의 금지).

#### Scenario: 왕복 비용 계산
- **WHEN** KR 매수·매도 왕복 비용을 계산하면
- **THEN** 수수료(bps)·거래세(bps)가 반영된 결정적 값이 반환되고 테스트 케이스(StockOS test_costs 이식)와 일치한다

### Requirement: 성과 원시 지표
거래 완결(청산) 시 다음 지표가 journal에 기록되어야 한다(SHALL): 비용 차감 후 실현손익, R 배수(실현손익/초기 위험), 보유 시간, MFE/MAE(보유 중 최대 유리/불리 이동 — filldetect 가격 스냅샷 기반, 관측 간격의 한계를 필드로 명시). 집계 지표(승률·profit factor·MDD)는 원시 기록에서 파생 계산한다(SHALL — 원시 데이터가 권위).

#### Scenario: 청산 완결 시 기록
- **WHEN** 포지션이 CLOSED로 전이하면
- **THEN** 비용 차감 실현손익·R 배수·MFE/MAE가 provenance와 연결되어 기록된다

#### Scenario: MFE 관측 한계 명시
- **WHEN** 가격 폴링 간격이 15초인 구간의 MFE를 기록하면
- **THEN** 관측 해상도 필드가 함께 저장되어 과대 해석을 방지한다
