# Function Logic Map: `Context.OfficialClientForTest`

- Source: `internal/app/engine/export_test.go` (26-26)
- Revision: current — HEAD `648df8ef`; source_sha256 `62739eb840b4e75533064134cf00ce6d02dd581e50cf46cf47c0495f225d9a81`
- AST evidence: `ast.json` (`tools/logic-map` 추출기 출력)
- Risk scan: `risk-pattern-report.md` (ast-grep, `tools/logic-map/sgconfig.yml`)
- Extractor counts: AST branches 0 · returns 1 · calls 0
- Exact AST return positions: 26:62
- 2026-09-25 task 5.1 갱신: 분기 표는 `analysis/harness/51_refresh.py` 가 ast.json 에서 생성함. 아래 산문 절은 a071 저자 기록(2026-08-04) 그대로이며 base 뒤 다른 change 가 이 함수에 더한 동작은 기술하지 않음.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| test Context | initialized engine context | engine assembly | nil only when test built an invalid context |

## Branches and early returns

| Branch | AST anchor | Source text at anchor | Coverage at HEAD (untagged package suite, per-test runs) |
|---|---|---|---|
| B1 | branchless happy path at 26:1 | `func (c *Context) OfficialClientForTest() *official.Client { return c.official }` | _test.go — 커버 계측 밖(분기 진입 여부 미측정) |

## Author concept rows (2026-08-04 — N 번호는 AST 분기 번호가 아님)

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| N1 | accessor called | none | returns sealed official client | existing official transport isolation tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| none | direct test-only field access | no error/retry | CodeGraph + AST |

## State mutations and fallbacks

- No mutation; `_test.go` accessor is absent from production binaries.

## Safety conclusion

- Safe edit boundary: retain read-only test seam
- High-risk impact: no (test-only)
