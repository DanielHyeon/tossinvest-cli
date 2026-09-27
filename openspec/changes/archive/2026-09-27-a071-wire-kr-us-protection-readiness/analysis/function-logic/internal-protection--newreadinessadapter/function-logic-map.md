# Function Logic Map: `NewReadinessAdapter`

- Source: `internal/protection/readiness_adapter.go` (69-77)
- Revision: current — HEAD `648df8ef`; source_sha256 `fed31b776d08c46f1cce728bb81740eee99734d3a489e6a4050df87c6e70af2a`
- AST evidence: `ast.json` (`tools/logic-map` 추출기 출력)
- Risk scan: `risk-pattern-report.md` (ast-grep, `tools/logic-map/sgconfig.yml`)
- Extractor counts: AST branches 1 · returns 2 · calls 4
- Exact AST return positions: 72:3, 76:2
- 2026-09-25 task 5.1 갱신: 분기 표는 `analysis/harness/51_refresh.py` 가 ast.json 에서 생성함. 아래 산문 절은 a071 저자 기록(2026-08-04) 그대로이며 base 뒤 다른 change 가 이 함수에 더한 동작은 기술하지 않음.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| provider/account/profile | non-nil provider and non-empty exact identity | engine assembly | constructor error |
| paired supervisor contracts | exact KR and US sealed contract, no defaulting | production supervisor assembly | constructor error; default adapter remains paired UNWIRED |

## Branches and early returns

| Branch | AST anchor | Source text at anchor | Coverage at HEAD (untagged package suite, per-test runs) |
|---|---|---|---|
| B1 | if at 71:2 | `if provider == nil \|\| accountID == "" \|\| profileID == "" {` | entered 1/48 |

## Author concept rows (2026-08-04 — N 번호는 AST 분기 번호가 아님)

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| N1 | provider/identity absent | none | constructor error | adapter constructor test |
| N2 | production contract corrupt/incomplete | none | constructor error | contract seal test |
| N3 | valid dependencies | immutable adapter seal | adapter | valid adapter test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| readiness adapter seal | bind account/profile and paired contract identities | pure/no retry | current HEAD |

## State mutations and fallbacks

- Constructor has no broker mutation and cannot turn a default snapshot WIRED.

## Safety conclusion

- Safe edit boundary: require explicit sealed contracts for production constructor; keep a separate default-UNWIRED constructor.
- High-risk impact: yes — an adapter is consumed at the exposure boundary.
