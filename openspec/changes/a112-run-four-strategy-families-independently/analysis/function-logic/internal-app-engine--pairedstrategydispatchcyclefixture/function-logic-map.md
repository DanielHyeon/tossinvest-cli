# Function Logic Map: `pairedStrategyDispatchCycleFixture`

- Source: `internal/app/engine/strategy_dispatch_cycle_test.go` (226-277)
- Function: `pairedStrategyDispatchCycleFixture` in package `engine`
- File SHA-256: `10f38fe4e88c3e076e8c88b1cc4764c847fb6052835e3be2595c85f34e9b1464`
- Pinned revision: `current` — the AST and the SHA-256 above are this worktree's file.
- AST evidence: `ast.json` — AST branches 7.
- Risk scan: `risk-pattern-report.md`.

Exact AST return nodes: `275:108, 276:2`.

## Inputs and invariants

시험 픽스처다. 생산 코드가 아니지만 이 change 가 편집했으므로 증거를 남긴다.

두 시장의 제안·계좌·위험·일정 권한을 세우고 공유 dispatch 주기를 만든다.
태스크 8.8.2 가 그 생성자에 `proposals` 를 넘기도록 한 줄을 바꿨다.

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `t` | non-nil | 시험 런타임 | `t.Fatal` |
| 배선의 보호 세대 | 상수 9 (`strategyDispatchGatewaySpy`) | 같은 파일의 spy | 이 값이 바뀌면 ProtectionReady 하한 시험이 다른 것을 잰다 |

**이 픽스처가 증명하지 못하는 것.** 활성화는 여기서 영값이다. 하한 결속을 재는
시험은 `kr.activation` 을 자기가 채우고 `cycle.proposals` 를 다시 넣는다 —
픽스처가 대신 채우면 모든 시험이 활성화된 상태만 보게 된다.

## Branches and early returns

| Branch | AST kind | Source location | Required test disposition |
|---|---|---|---|
| B1 | range | 233:2 | 픽스처 조립 분기. 실패하면 `t.Fatal` 로 시험이 멈춘다 |
| B2 | if | 237:3 | 픽스처 조립 분기. 실패하면 `t.Fatal` 로 시험이 멈춘다 |
| B3 | if | 243:3 | 픽스처 조립 분기. 실패하면 `t.Fatal` 로 시험이 멈춘다 |
| B4 | if | 252:3 | 픽스처 조립 분기. 실패하면 `t.Fatal` 로 시험이 멈춘다 |
| B5 | else | 254:10 | 픽스처 조립 분기. 실패하면 `t.Fatal` 로 시험이 멈춘다 |
| B6 | if | 262:2 | 픽스처 조립 분기. 실패하면 `t.Fatal` 로 시험이 멈춘다 |
| B7 | if | 268:2 | 픽스처 조립 분기. 실패하면 `t.Fatal` 로 시험이 멈춘다 |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `t.Helper` | 227:2 |
| `newStrategyRiskLoaderFixture` | 228:17 |
| `riskFixture.loader.collect` | 230:14 |
| `context.Background` | 230:41 |
| `riskFixture.results.forMarket` | 234:13 |
| `strategyproposal.ProductionBatchAuthorityForTest` | 235:12 |
| `string` | 235:80 |
| `batch.For` | 236:20 |
| `t.Fatal` | 238:4 |
| `strategyaccount.AuthorityForTest` | 250:15 |
| `now.Add` | 250:77 |
| `now.Add` | 250:100 |
| `strings.Repeat` | 251:15 |
| `pairedDispatchSchedule` | 258:14 |
| `clock.NewFake` | 259:15 |
| `journal.Open` | 260:12 |
| `context.Background` | 260:25 |
| `filepath.Join` | 260:69 |
| `t.TempDir` | 260:83 |
| `journal.FixedFSProber` | 261:13 |
| `t.Fatal` | 263:3 |
| `t.Cleanup` | 265:2 |
| `j.Close` | 265:25 |
| `execgw.NewRiskGuardian` | 266:19 |
| `risk.DefaultPolicy` | 267:11 |
| `costs.DefaultModel` | 267:40 |
| `t.Fatal` | 269:3 |
| `newProductionStrategyFirstLegAuthorityLoader` | 271:12 |
| `newStrategyFirstLegAdmissionBridge` | 272:14 |
| `newStrategyDispatchCycle` | 274:11 |

## State mutations and fallbacks

- 임시 디렉터리에 원장을 열고 `t.Cleanup` 으로 닫는다.
- fallback 없음. 조립에 실패하면 시험이 멈춘다.

## Safety conclusion

- Safe edit boundary: 시험 전용. 생산 바이너리에 들어가지 않는다.
- High-risk impact: no — 다만 이 픽스처가 재는 대상은 High-risk 경로이므로,
  배선의 상수를 바꾸면 그것을 인용하는 시험이 다른 것을 재게 된다.
