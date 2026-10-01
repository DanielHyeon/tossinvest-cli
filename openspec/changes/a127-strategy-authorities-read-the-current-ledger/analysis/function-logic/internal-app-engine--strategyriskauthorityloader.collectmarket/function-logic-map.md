# Function Logic Map: `strategyRiskAuthorityLoader.collectMarket`

- Source: `internal/app/engine/strategy_risk_authority.go` (`187`–`231`)
- Qualified: `strategyRiskAuthorityLoader.collectMarket`
- AST evidence: `ast.json` (`source_sha256` f7ad67b8ad584c17…) — **편집 전**(base `de3b4f65` 의 바이트, 커버리지 `analysis/impl/coverage-pre-edit.out`)
- Risk scan: `risk-pattern-report.md`
- AST branches 9 · return 4 · 호출 18

**역할.** 결과 권한의 범위마다 `riskbucket.LoadProductionRiskSnapshotAuthority` 를 부르고 원인을 운반. a127: config 리터럴에 원장 스키마 주입 값 `journal.SchemaVersion`(상수 선택자)을 넣음(D2).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 결과 · FX 권한 | ready | 상위 로더 | B1 · B2 거절 |
| 범위 owner 키 | 유효 | 계보 | B5 거짓이면 그 범위 거절 |

## Branches and early returns

> 분기 표는 `analysis/harness/branch_table.py` 가 ast · 소스 · 커버리지로 만들었다. 「창의 return」은 위치, 「진입 실측」은 그 줄로 시작하는 커버리지 블록.

| Branch | 종류 | 조건 (원문) | 창의 return | 진입 실측 |
|---|---|---|---|---|
| B1 | if | `:193` `if !result.ready {` | :194 | 예 |
| B2 | if | `:196` `if !fx.snapshot.Ready \|\| !fx.read.valid {` | :197 | 예 |
| B3 | if | `:200` `if market == StrategyMarketUS {` | — | 예 |
| B4 | range | `:206` `for _, scoped := range result.results() {` | — | 예 |
| B5 | if | `:210` `if keyed {` | — | 예 |
| B6 | switch | `:217` `switch {` | — | — |
| B7 | case | `:218` `case err != nil:` | — | 예 |
| B8 | case | `:221` `case string(scope.Market) == string(market) && scope.AccountID == loader.accountID &&` | — | — |
| B9 | case | `:224` `default:` | :230 | 아니오 |

## Calls and live bindings

`strategyOwnerKeyOf` · `riskbucket.LoadProductionRiskSnapshotAuthority`(config 리터럴) · `bundle.Scope` · `bundle.Entries` · `strategyRiskMarketFromScopes`.

## State mutations and fallbacks

없음 — 범위 권한 값을 만듦.

## Safety conclusion

- **Safe edit boundary**: 편집 전 — a127 은 config 리터럴에 필드 하나만 더할 예정.
- **High-risk impact**: yes — 위험 권한 배선.
