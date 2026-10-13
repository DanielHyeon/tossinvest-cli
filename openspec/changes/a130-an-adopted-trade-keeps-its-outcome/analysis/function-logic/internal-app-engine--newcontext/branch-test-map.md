# Branch Test Map: `NewContext`

편집 뒤 좌표(4판 GREEN). 분기별 RED 는 따로 재지 않았음(no) — 4판 시험 전체가 base 에서 컴파일 RED(`analysis/red/red-4.txt`), 조작 행 모양은 행동 RED(red-2 · red-4). GREEN = 패키지 시험 통과(`analysis/verify/receipts.md`). Test 칸은 그 함수의 a130 시험 · 기존 시험 묶음(무편집 갈래는 기존 시험).

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:439` `if err != nil {` | `TestA130NewContextBindsTheHooksWithTheEngineLogger` · 기존 엔진 기동 시험 | no | yes |
| B2 | `:446` `if err != nil {` | `TestA130NewContextBindsTheHooksWithTheEngineLogger` · 기존 엔진 기동 시험 | no | yes |
| B3 | `:450` `if clk == nil {` | `TestA130NewContextBindsTheHooksWithTheEngineLogger` · 기존 엔진 기동 시험 | no | yes |
| B4 | `:464` `if opts.Publisher != nil {` | `TestA130NewContextBindsTheHooksWithTheEngineLogger` · 기존 엔진 기동 시험 | no | yes |
| B5 | `:472` `if err := recordGateSettings(auditLog, gate, cfg.Engine.Adoption, notifications,` | `TestA130NewContextBindsTheHooksWithTheEngineLogger` · 기존 엔진 기동 시험 | no | yes |
| B6 | `:487` `if err != nil {` | `TestA130NewContextBindsTheHooksWithTheEngineLogger` · 기존 엔진 기동 시험 | no | yes |
| B7 | `:494` `if err != nil {` | `TestA130NewContextBindsTheHooksWithTheEngineLogger` · 기존 엔진 기동 시험 | no | yes |
| B8 | `:503` `if err := bindApplyHooks(jrn, tradeOutcomeRefusalSink(opts.Logger)); err != nil ` | `TestA130NewContextBindsTheHooksWithTheEngineLogger` · 기존 엔진 기동 시험 | no | yes |
| B9 | `:522` `if err != nil {` | `TestA130NewContextBindsTheHooksWithTheEngineLogger` · 기존 엔진 기동 시험 | no | yes |
| B10 | `:537` `if gate.Enabled && guardian == nil && !opts.disableProductionGuardian {` | `TestA130NewContextBindsTheHooksWithTheEngineLogger` · 기존 엔진 기동 시험 | no | yes |
| B11 | `:539` `if factory == nil {` | `TestA130NewContextBindsTheHooksWithTheEngineLogger` · 기존 엔진 기동 시험 | no | yes |
| B12 | `:545` `if err != nil {` | `TestA130NewContextBindsTheHooksWithTheEngineLogger` · 기존 엔진 기동 시험 | no | yes |
| B13 | `:565` `if err != nil {` | `TestA130NewContextBindsTheHooksWithTheEngineLogger` · 기존 엔진 기동 시험 | no | yes |
| B14 | `:569` `if !automation.Verified {` | `TestA130NewContextBindsTheHooksWithTheEngineLogger` · 기존 엔진 기동 시험 | no | yes |
| B15 | `:574` `if err != nil {` | `TestA130NewContextBindsTheHooksWithTheEngineLogger` · 기존 엔진 기동 시험 | no | yes |
