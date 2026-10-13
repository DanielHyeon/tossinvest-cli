# CodeGraph baseline (a130)

- 도구: codegraph **1.6.0** · 기준: base `233d5d9c` 워킹트리(편집 0) · 2026-10-13 · 원문 `raw/`

| 심볼 | 정의 | 호출자(생산) | 호출자(시험) | 비고 |
|---|---|---|---|---|
| `journal.freezeTradeOutcomeTx` | trade_outcomes.go:129 | `ApplyExitFill` 1 | 0 | 반환값을 버리는 자리 apply_hook.go:1023 (AST) |
| `journal.computeTradeOutcome` | :163 | `freezeTradeOutcomeTx` · `Journal.BackfillTradeOutcome` | 0 | |
| `journal.adoptedRoundTripLegs` | :488 | `computeTradeOutcome` 1 | 0 | 피호출: `adoptedBasis` · `engineSellOrders` · `sumSellFills` |
| `journal.Journal.BackfillTradeOutcome` | :843 | **0** | 3(adoption_outcome_test 2 · trade_outcomes_test 1) | 생산 호출자 0 — a130 도구가 첫 생산 호출자 |
| `journal.ApplyExitFill` | apply_hook.go:961 | CodeGraph 「호출자 없음」 | — | 함수 값으로 쓰임: `gateway.go:191` `Exit: journal.ApplyExitFill` (HEAD grep) · 시험 `outcomeFixture` 등 |
| `journal.Journal.runApplyHooks` | apply_hook.go:264 | `RecordFill`(fills.go) · `backfillConfirmedStrategyFillTx`(strategy_dispatch_runtime.go) | 3 | |
| `engine.bindApplyHooks` | gateway.go:187 | `NewContext`(engine.go:503) | 14 | 시그니처 유지(시험 14곳) |
| `engine.NewContext` | engine.go:437 | `assembleEngine`(cmd/tossctl/engine_assembly.go:14) · `New` | — | 생산 `Options.Logger = logger`(engine_assembly.go:24) |

## affected

- 기본 판별식: 시험 1건(`auth-helper/tests/test_cli.py` — Go 아님, CLAUDE.md 하네스 주석의 알려진 결함).
- `--filter '*_test.go'`: **1109** 줄(엔진 패키지 167 · 콘솔 다수 · 저널 · cmd). 「영향 시험 없음」이 아니다 — 전체 패키지 시험으로 회귀를 잰다.
