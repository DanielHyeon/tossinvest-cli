# Branch Test Map: `NewContext(params=2, results=2)`

factory에서 borrowed budget 읽어 gateway에 전달, 오류를 기존 startup refusal 경로로 전파. owner만 close, nil OFF 유지. 시험 TestA166GatewayReplayBinding 및 기존 root fixture.

| 분기 | AST 위치와 조건 원문 | 결과 범위 |
|---|---|---|
| B1 | if L453: `if err != nil {` | 선택 범위 외 기존 분기 개별 통과는 주장하지 않음 |
| B2 | if L460: `if err != nil {` | 선택 범위 외 기존 분기 개별 통과는 주장하지 않음 |
| B3 | if L464: `if clk == nil {` | 선택 범위 외 기존 분기 개별 통과는 주장하지 않음 |
| B4 | if L478: `if opts.Publisher != nil {` | 선택 범위 외 기존 분기 개별 통과는 주장하지 않음 |
| B5 | if L486: `if err := recordGateSettings(auditLog, gate, cfg.Engine.Adoption, notifications,` | 선택 범위 외 기존 분기 개별 통과는 주장하지 않음 |
| B6 | if L501: `if err != nil {` | 선택 범위 외 기존 분기 개별 통과는 주장하지 않음 |
| B7 | if L508: `if err != nil {` | 선택 범위 외 기존 분기 개별 통과는 주장하지 않음 |
| B8 | if L517: `if err := bindApplyHooks(jrn, tradeOutcomeRefusalSink(opts.Logger)); err != nil {` | 선택 범위 외 기존 분기 개별 통과는 주장하지 않음 |
| B9 | if L524: `if opts.BudgetFactory != nil {` | 선택 범위 외 기존 분기 개별 통과는 주장하지 않음 |
| B10 | if L526: `if err != nil {` | 선택 범위 외 기존 분기 개별 통과는 주장하지 않음 |
| B11 | if L545: `if err != nil {` | 선택 범위 외 기존 분기 개별 통과는 주장하지 않음 |
| B12 | if L560: `if gate.Enabled && guardian == nil && !opts.disableProductionGuardian {` | 선택 범위 외 기존 분기 개별 통과는 주장하지 않음 |
| B13 | if L562: `if factory == nil {` | 선택 범위 외 기존 분기 개별 통과는 주장하지 않음 |
| B14 | if L568: `if err != nil {` | 선택 범위 외 기존 분기 개별 통과는 주장하지 않음 |
| B15 | if L588: `if err != nil {` | 선택 범위 외 기존 분기 개별 통과는 주장하지 않음 |
| B16 | if L592: `if !automation.Verified {` | 선택 범위 외 기존 분기 개별 통과는 주장하지 않음 |
| B17 | if L597: `if err != nil {` | 선택 범위 외 기존 분기 개별 통과는 주장하지 않음 |
