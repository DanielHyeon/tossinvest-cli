> 골격 초안(ADDED) — a171 design 에서 「Retry Matrix 산출물」·「IN_DOUBT 해소」 MODIFIED 로 다시 쓴다.

## ADDED Requirements

### Requirement: 진입 접수 결과 불명의 창 안 재전송

엔진 게이트웨이는 노출 증가 지정가 접수의 결과가 불명일 때, 최초 전송부터 8초 창 안에서만, 미체결·종결 조회로 접수가 보이지 않을 때에 한해 같은 `clientOrderId`·같은 본문으로 같은 attempt 안에서 최대 2회 재전송할 수 있어야 한다(SHALL — 세부 조건은 design 에서 확정).

#### Scenario: 창 밖 결과 불명
- **WHEN** 접수 결과가 불명이고 최초 전송부터 8초가 지났으면
- **THEN** 재전송하지 않고 IN_DOUBT 해소 절차로 넘긴다
