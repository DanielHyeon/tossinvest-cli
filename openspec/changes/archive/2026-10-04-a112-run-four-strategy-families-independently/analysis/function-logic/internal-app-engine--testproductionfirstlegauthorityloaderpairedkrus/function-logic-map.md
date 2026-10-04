# Function Logic Map: `TestProductionFirstLegAuthorityLoaderPairedKRUS`

- Source: `internal/app/engine/strategy_account_first_leg_authority_test.go`
- Source SHA-256: `47f555f84382d2768f30dbce87c834c0ad832c28664c595b0779b7c295b1ce1e`
- Signature: `TestProductionFirstLegAuthorityLoaderPairedKRUS(params=1, results=0)`
- Source range: `20:1`–`85:2`
- AST evidence: `ast.json` — **편집 뒤**(a112 5.2.2.2).
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- 시험 코드 — 생산 동작 없음(비례 원칙: FLM 의무의 무거운 규율 대상 아님, 게이트 모양만 채움)

## Branches and early returns

- Exact AST return nodes: ``.

| Branch | AST kind | Source location | Meaning |
|---|---|---|---|
| B1 | range | 26:2 | 시험 본문 분기(편집 불변) |
| B2 | if | 30:3 | 시험 본문 분기(편집 불변) |
| B3 | if | 37:3 | 시험 본문 분기(편집 불변) |
| B4 | if | 48:3 | 시험 본문 분기(편집 불변) |
| B5 | else | 50:10 | 시험 본문 분기(편집 불변) |
| B6 | if | 58:2 | 시험 본문 분기(편집 불변) |
| B7 | range | 63:2 | 시험 본문 분기(편집 불변) |
| B8 | if | 66:3 | 시험 본문 분기(편집 불변) |
| B9 | if | 70:3 | 시험 본문 분기(편집 불변) |
| B10 | if | 74:3 | 시험 본문 분기(편집 불변) |
| B11 | if | 82:2 | 시험 본문 분기(편집 불변) |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `newStrategyRiskLoaderFixture` | 21:17 |
| `riskFixture.loader.collect` | 23:14 |
| `context.Background` | 23:41 |
| `riskFixture.results.forMarket` | 27:13 |
| `strategyproposal.ProductionBatchAuthorityForTest` | 28:12 |
| `string` | 28:80 |
| `batch.For` | 29:20 |
| `t.Fatal` | 31:4 |
| `strategyaccount.AuthorityForTest` | 43:23 |
| `now.Add` | 43:85 |
| `now.Add` | 43:108 |
| `a112ScopedAccount` | 47:18 |
| `openTestJournal` | 54:7 |
| `clock.NewFake` | 55:15 |
| `execgw.NewRiskGuardian` | 56:19 |
| `risk.DefaultPolicy` | 57:11 |
| `costs.DefaultModel` | 57:40 |
| `t.Fatal` | 59:3 |
| `newProductionStrategyFirstLegAuthorityLoader` | 61:12 |
| `routeReadySchedulePair` | 61:81 |
| `entries.authority.Proposal` | 64:13 |
| `proposals.forMarket` | 64:13 |
| `validateStrategyFirstLegResult` | 65:24 |
| `t.Fatalf` | 67:4 |
| `loader.collectStrategyFirstLegAuthority` | 69:20 |
| `context.Background` | 69:60 |
| `t.Fatalf` | 71:4 |
| `string` | 74:31 |
| `len` | 75:4 |
| `len` | 75:60 |
| `guardian.PolicyVersion` | 76:44 |
| `t.Fatalf` | 78:4 |
| `t.Fatalf` | 83:3 |

## State mutations and fallbacks

- 시험 fixture 만 바꾼다.

## Safety conclusion

- High-risk 아님. 편집은 fixture 가 생산 모양(범위별 계좌 권한)을 갖게 할 뿐 — 단언 불변.
