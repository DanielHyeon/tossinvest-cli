# Function Logic Map: `TestReadinessAdapterRejectsInvalidMarketWithoutCallingProvider`

- Source: `internal/protection/readiness_adapter_test.go` (47-59)
- Revision: current — HEAD `648df8ef`; source_sha256 `5bb01c42670517355a9684e7be85225d6ec96b46e94a337293707383513df5c5`
- AST evidence: `ast.json` (`tools/logic-map` 추출기 출력)
- Risk scan: `risk-pattern-report.md` (ast-grep, `tools/logic-map/sgconfig.yml`)
- Extractor counts: AST branches 3 · returns 0 · calls 8
- 2026-09-25 task 5.1 갱신: 분기 표는 `analysis/harness/51_refresh.py` 가 ast.json 에서 생성함. 아래 산문 절은 a071 저자 기록(2026-08-04) 그대로이며 base 뒤 다른 change 가 이 함수에 더한 동작은 기술하지 않음.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| test fixture | exact immutable inputs | test contract | fail assertion |

## Branches and early returns

| Branch | AST anchor | Source text at anchor | Coverage at HEAD (untagged package suite, per-test runs) |
|---|---|---|---|
| B1 | if at 50:2 | `if err != nil {` | _test.go — 커버 계측 밖(분기 진입 여부 미측정) |
| B2 | if at 53:2 | `if _, refusal := adapter.Check(context.Background(), ReadinessRequest{Market: "cn", OrderType: "LIMIT", Qua...` | _test.go — 커버 계측 밖(분기 진입 여부 미측정) |
| B3 | if at 56:2 | `if provider.calls != 0 {` | _test.go — 커버 계측 밖(분기 진입 여부 미측정) |

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
