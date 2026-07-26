# exit-policy Specification (delta)

## ADDED Requirements

### Requirement: Baseline Ratchet — 기준선 단조 상승
보유 포지션의 보호 기준선은 R 배수 트리거로 단계 상승해야 한다(SHALL — StockOS `exit/baseline_ratchet.py` 이식, 사용자 요구 "일정 수익 시 기준선을 위로 올리며 손익 극대화"). 판정은 순수 함수이며 주문·IO에 접촉하지 않는다(SHALL NOT). 단계와 기본값(provenance 주석, §0.9 보수 방향으로만 변경):

| 트리거(도달 R) | 조치 | 새 기준선(R) |
|---|---|---|
| +0.4R | HALF_RISK | 스톱을 −0.5R로 상승 |
| +0.8R | BREAKEVEN | 실질 본전(비용 차감 `break_even_sell_price`) |
| +1.0R | PARTIAL | 40% 부분익절 발의 |
| +1.2R | PARTIAL_LOCK | +0.3R |
| +2.0R | PROFIT_LOCK | +0.8R |

**기준선은 단조 상승만 한다(SHALL NOT 하강)** — 관측 가격이 되돌아가도 이미 오른 기준선은 유지된다. 본전은 명목 진입가가 아니라 비용 차감 실질 본전이다(SHALL — costs 모델 결합). R의 분모는 진입 시 확정된 초기 위험(entry − stop)이며 이후 변하지 않는다(SHALL).

#### Scenario: 수익 진행에 따른 단계 상승
- **WHEN** 포지션이 +0.8R에 도달하면
- **THEN** 기준선이 실질 본전으로 상승하고, 이후 가격이 +0.5R로 되돌아도 기준선은 내려가지 않는다

#### Scenario: 부분익절 발의
- **WHEN** 포지션이 +1.0R에 도달하면
- **THEN** 40% 부분익절이 RISK_REDUCING 의도로 발의된다 (제출·수량 정합은 실행 계층 소관)

### Requirement: Profit Ladder — multi-rung ratchet
수익 목표는 다단 rung으로 관리되어야 한다(SHALL — StockOS `profit_ladder.py` 이식): rung 목표 도달 → 부분익절 발의 → 보호선 승격 → 다음 rung 활성. 판정 시점 필드(활성 rung·승격된 보호선)와 체결 시점 필드(누적 부분익절 비율·완결)는 분리되며(SHALL — 판정은 관측에서, 체결 반영은 실제 fill에서만 이동), 상태는 journal에 영속되어 재시작 후 유지된다(SHALL). 백테스트·동일 관측 창에서 목표와 스톱이 동시에 걸리면 STOP_FIRST 보수 모델을 기본으로 한다(SHALL).

#### Scenario: rung 도달과 승격
- **WHEN** 활성 rung의 목표가에 도달하면
- **THEN** 부분익절이 발의되고 보호선이 해당 rung의 잠금 수준으로 승격되며 다음 rung이 활성화된다

#### Scenario: 재시작 후 ladder 유지
- **WHEN** rung 2가 활성인 상태에서 재시작되면
- **THEN** 활성 rung·승격된 보호선·누적 익절 비율이 journal에서 복원된다

#### Scenario: 판정·체결 분리
- **WHEN** 부분익절이 발의됐지만 체결이 아직 없으면
- **THEN** 누적 부분익절 비율은 변하지 않고, 체결 반영 시에만 이동한다

### Requirement: 판정과 액추에이션의 경계
exit 정책은 판정만 산출한다(SHALL): 기준선 상승·부분익절 발의는 의도(ReductionIntent 계열)로 표현되고, 실제 브로커측 보호주문 교체·청산 제출·수량 정합은 보호주문 실행 계층의 소관이다(이 change의 범위가 아니다). 브로커측 보호가 아직 없는 구성에서 기준선은 로컬 판정 상태이며, 기준선 하회 관측은 청산 발의(RISK_REDUCING)를 산출한다(SHALL). 신호 기반 트레일(이동평균·VWAP·CVD·추세선)과 시간 종료는 이 capability의 범위가 아니다(P3 신호·스케줄러).

#### Scenario: 기준선 하회 관측
- **WHEN** 관측 가격이 현재 기준선을 하회하면
- **THEN** 전량(잔여) 청산이 RISK_REDUCING 의도로 발의된다

#### Scenario: 비용은 청산 게이트가 아니다
- **WHEN** 청산 발의의 예상 비용이 어떤 버퍼를 초과하더라도
- **THEN** 청산 발의는 차단되지 않는다 (§0.3 — StockOS SELL_COST_BUFFER는 이식하지 않는다)
