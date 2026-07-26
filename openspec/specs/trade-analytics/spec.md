# trade-analytics Specification

## Purpose
거래 비용 모델(구조·검증 게이트·MAX_RATE)·성과 원시 지표(동결 기록·실현 R)·분석 경로 격리 요구사항.

## Requirements

### Requirement: 거래 비용 모델
KRW·USD 거래의 수수료·거래세·환전 비용은 순수 함수 비용 모델로 계산되어야 한다(SHALL — StockOS `costs.py`의 **구조**(시장별 수수료 bps·거래세 bps·실질 본전 산식)와 **검증 게이트**(비수치·NaN·음수·상한 `MAX_RATE=0.05` 초과 거부)를 이식하되, KIS 수치·`KIS_*` 명명은 이식하지 않는다(SHALL NOT). override는 설정 주입 방식으로 재구현하고 test_costs(4)·test_costs_env_override(16)는 **검증 게이트·주입 구조**에 대해 이식한다(수치 단언은 Toss 보수값으로 재작성). 수치는 2b 실측으로 채우며, 실측 전에는 상한 이내의 과대 추정 보수값을 "미검증" provenance와 함께 사용한다(SHALL — 진입 측 보수 방향; 상한이 실질 본전 기준선의 폭주를 막는다). Guardian의 현금·비용 검증, exit 정책의 실질 본전, 성과의 비용 차감이 같은 모델을 사용한다(SHALL — 이중 정의 금지). 비용 모델은 청산 게이트로 적용되지 않는다(SHALL NOT — §0.3).

#### Scenario: 왕복 비용 계산
- **WHEN** KR 매수·매도 왕복 비용을 계산하면
- **THEN** 수수료·거래세가 반영된 결정적 값이 반환되고 이식된 구조 검증 테스트와 일치한다

#### Scenario: 상한 초과 설정 거부
- **WHEN** MAX_RATE를 넘는 비용률이 설정되면
- **THEN** 설정이 거부된다

### Requirement: 성과 원시 지표
거래 완결(청산) 시 다음이 journal에 기록되어야 한다(SHALL): 비용 차감 후 실현손익, **실현 R**(`실현손익 / (초기 위험 × 초기 수량)` — 총액·비용 차감 기준), 초기 위험, 보유 시간, 도달 exit 단계(ratchet 레벨·ladder rung). 실현 R은 exit 판정의 **가격 R**(`(관측가 − entry) / (entry − stop)` — 주당·gross)과 별개 지표이며 같은 이름을 쓰지 않는다(SHALL — 부분익절 후 둘은 달라진다). 기록은 CLOSED 트랜잭션에서 exit 상태의 스냅샷과 함께 **동결**되며(SHALL — 이후 비동기 작업이 원시 행을 다시 읽거나 갱신하지 않는다), 집계(승률·PF·MDD)는 원시 기록에서 파생 계산한다(SHALL). MFE/MAE는 시계열 소스 부재로 범위 밖(SHALL NOT — P3).

#### Scenario: 청산 완결 시 동결 기록
- **WHEN** 포지션이 CLOSED로 전이하면
- **THEN** 실현손익·실현 R·보유 시간·도달 exit 단계가 같은 트랜잭션에서 동결 기록된다

### Requirement: 분석 경로의 격리
성과 집계와 보존 기간 정리는 주문 실행 경로에서 격리되어야 한다(SHALL). 분석 작업의 실패·지연이 체결 반영·청산 처리를 지연시키거나 실행 상태를 되돌리거나 운영 모드 강화를 유발해서는 안 된다(SHALL NOT — 강화 트리거는 risk-management의 열거형뿐이다). 성과 테이블은 보존 기간 정책(기본 180일)을 갖는다(SHALL).

#### Scenario: 분석 계산 실패
- **WHEN** 성과 집계가 실패하면
- **THEN** 포지션 종결·운영 모드는 영향받지 않고 분석 재시도만 남는다

#### Scenario: 보존 기간 정리
- **WHEN** 180일 경과 기록의 정리가 수행되면
- **THEN** 주문 경로 트랜잭션과 경쟁하지 않는 비동기 작업으로 처리된다

### Requirement: 합성 R의 구분 집계

편입 포지션의 실현 R은 합성 분모(편입 시점 관측가 × default_stop_pct)에서 나오므로, 집계(승률·PF·ΣR 등)는 엔진 진입(실측 R)과 편입(합성 R)을 구분해 표기해야 한다(SHALL). 구분은 `positions.adoption_id IS NOT NULL` 명시 조인으로 하며(SHALL — trade_outcomes 스키마 무변경, 시간창 휴리스틱 매칭 금지 계약 준수), 혼합 집계를 낼 때는 두 모집단의 표본 수를 병기한다(SHALL).

#### Scenario: 혼합 계좌의 성과 집계
- **WHEN** 엔진 진입 3건과 편입 2건이 완결된 계좌의 집계를 조회하면
- **THEN** 실측 R 집계(n=3)와 합성 R 집계(n=2)가 구분 표기된다
