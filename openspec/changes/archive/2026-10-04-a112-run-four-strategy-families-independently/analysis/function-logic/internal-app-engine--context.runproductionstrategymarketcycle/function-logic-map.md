# Function Logic Map: `runProductionStrategyMarketCycle`

- Source: `internal/app/engine/strategy_entry_supervisor.go`
- Source SHA-256: `9cb510c1c9c7f44ac109de5fd8ddac8e6c559adb4d654b1acd65dce19cc9e577`
- Signature: `Context.runProductionStrategyMarketCycle(params=3, results=1)`
- Source range: `516:1`–`579:2`
- AST evidence: `ast.json` — **편집 뒤**(a112 7.3.1 SHADOW).
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- 본문의 shadow 타입 식은 Args[5] 하나 · 그 유일한 호출은 forMarket — 타입 규칙 핀 `TestTheMarketCycleCarriesShadowOnlyAsTheLastEvaluateArgument`.

## Branches and early returns

- Exact AST return nodes: `519:3, 543:3, 553:3, 556:3, 578:2`.

| Branch | AST kind | Source location | Meaning |
|---|---|---|---|
| B1 | if | 518:2 | refresh 오류 |
| B2 | if | 542:2 | 레인 런타임 오류 |
| B3 | if | 548:2 | evaluate 오류(durable latch) |
| B4 | if | 555:2 | dispatch 없는 조립 → nil |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `c.refreshPairedStrategyEntryProductionAssembly` | 517:16 |
| `c.productionStrategyLanes` | 541:16 |
| `lanes.evaluate` | 548:12 |
| `restore.Activation.Generation` | 549:3 |
| `fresh.schedule.forMarket` | 549:3 |
| `familyActivation` | 550:3 |
| `fresh.proposals.forMarket` | 550:3 |
| `strategyLaneInputs` | 551:3 |
| `fresh.proposals.forMarket` | 551:36 |
| `fresh.shadow.forMarket` | 552:3 |
| `dispatchStrategyMarketHandoffs` | 578:9 |
| `dispatchHandoffs` | 578:72 |
| `fresh.proposals.forMarket` | 578:72 |

## State mutations and fallbacks

- 레인 관측 · dispatch(편집 전과 같음).

## Safety conclusion

- High-risk(시장 주기) — dispatch 문장 · 원천 · 순서 무변경(:169-175 핀 · 차등 dispatch 시험). dispatch 앞 shadow 일은 값 운반 하나.
