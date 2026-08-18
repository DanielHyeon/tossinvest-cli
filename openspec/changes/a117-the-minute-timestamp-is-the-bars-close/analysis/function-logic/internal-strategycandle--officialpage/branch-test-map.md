# Branch Test Map: `officialPage`

- Source: `internal/strategycandle/adapter_test.go`, SHA-256 `d0f227759181a0727ba87ed06ac6f5d163bafd048e5429a2df7778510bb717dc`; branch IDs follow `ast.json` (regenerated 2026-08-18 after the edit).
- AST counts: branches 5, returns 1, defers 0, go statements 0. Source range `15:1-40:2`.
- Test-fixture bundle: this function is test-only; it has no production branch to hold.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | the fake server routes by path | `TestOfficialDTOToStrategyMarketAdapterIsLosslessAndIdentityBound` | the pre-shift literals made it fail with `outside_regular_session: 2026-07-31T09:00:00+09:00` | green |
| B2 | the token arm answers the OAuth call | `TestOfficialDTOToStrategyMarketAdapterIsLosslessAndIdentityBound` | not edited | green |
| B3 | the candles arm serves the five shifted labels | `TestOfficialDTOToStrategyMarketAdapterIsLosslessAndIdentityBound` | the pre-shift literals made it fail as above | green |
| B4 | unknown path | not-applicable: defensive arm, never requested | not edited | not-applicable |
| B5 | reader error | not-applicable: the fake server always answers | not edited | not-applicable |

Verification: `go test ./... -count=1` green on 2026-08-18 (9,425 tests, 102 packages, exit 0).
