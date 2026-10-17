# Function Logic Map: `TestDispatchDecisionBindsExactScopeExpiryGenerationAndProvenance`

- Source: `internal/protectionreadiness/dispatch_test.go` (8-33)
- Revision: current — HEAD `648df8ef`; source_sha256 `4be237f9eb604ce721dc4122fd58310780bc6d58d11c70bbe2cc56c581a7cd55`
- AST evidence: `ast.json` (`tools/logic-map` 추출기 출력)
- Risk scan: `risk-pattern-report.md` (ast-grep, `tools/logic-map/sgconfig.yml`)
- Extractor counts: AST branches 5 · returns 0 · calls 14
- 2026-09-25 task 5.1 갱신: 분기 표는 `analysis/harness/51_refresh.py` 가 ast.json 에서 생성함. 아래 산문 절은 a071 저자 기록(2026-08-04) 그대로이며 base 뒤 다른 change 가 이 함수에 더한 동작은 기술하지 않음.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| test fixture | exact immutable inputs | test contract | fail assertion |

## Branches and early returns

| Branch | AST anchor | Source text at anchor | Coverage at HEAD (untagged package suite, per-test runs) |
|---|---|---|---|
| B1 | if at 16:2 | `if !decision.Allowed \|\| decision.Generation == 0 \|\| decision.SnapshotID == "" {` | _test.go — 커버 계측 밖(분기 진입 여부 미측정) |
| B2 | if at 19:2 | `if decision.Provenance.AccountID != "acct" \|\| decision.Provenance.ProfileID != "production" \|\| decision...` | _test.go — 커버 계측 밖(분기 진입 여부 미측정) |
| B3 | if at 24:2 | `if got := snapshot.Dispatch(wrongAccount, readinessNow); got.Allowed \|\| got.Code != RefusalScopeMismatch {` | _test.go — 커버 계측 밖(분기 진입 여부 미측정) |
| B4 | if at 27:2 | `if got := snapshot.Dispatch(scope, decision.Provenance.ExpiresAt); got.Allowed \|\| got.Code != RefusalExpi...` | _test.go — 커버 계측 밖(분기 진입 여부 미측정) |
| B5 | if at 30:2 | `if got := snapshot.Dispatch(DispatchScope{AccountID: "acct", ProfileID: "production", Market: MarketKR}, re...` | _test.go — 커버 계측 밖(분기 진입 여부 미측정) |

## Author concept rows (2026-08-04 — N 번호는 AST 분기 번호가 아님)

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| N1 | test fixture | exact immutable inputs | test contract | fail assertion |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| tested production function | verify fail-closed behavior | no retry | CodeGraph + AST |

## State mutations and fallbacks

- No production mutation; test-only setup and assertions.

## Safety conclusion

- Safe edit boundary: update exact-scope expectations only
- High-risk impact: no (test)
