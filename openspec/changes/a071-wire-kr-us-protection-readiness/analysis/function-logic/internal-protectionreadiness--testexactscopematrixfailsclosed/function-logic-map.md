# Function Logic Map: `TestExactScopeMatrixFailsClosed`

- Source: `internal/protectionreadiness/attestation_test.go` (42-74)
- Revision: current — HEAD `648df8ef`; source_sha256 `2fa5888a4648fb4546bddf1970c4f8f22f7ab68b6953e7ed687a4c10fcc2fb48`
- AST evidence: `ast.json` (`tools/logic-map` 추출기 출력)
- Risk scan: `risk-pattern-report.md` (ast-grep, `tools/logic-map/sgconfig.yml`)
- Extractor counts: AST branches 2 · returns 0 · calls 11
- 2026-09-25 task 5.1 갱신: 분기 표는 `analysis/harness/51_refresh.py` 가 ast.json 에서 생성함. 아래 산문 절은 a071 저자 기록(2026-08-04) 그대로이며 base 뒤 다른 change 가 이 함수에 더한 동작은 기술하지 않음.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| signed KR fixture plus one mutated runtime field | one exact field differs from attested scope | signed attestation fixture | assert typed scope mismatch and UNWIRED |

## Branches and early returns

| Branch | AST anchor | Source text at anchor | Coverage at HEAD (untagged package suite, per-test runs) |
|---|---|---|---|
| B1 | range at 64:2 | `for _, test := range tests {` | _test.go — 커버 계측 밖(분기 진입 여부 미측정) |
| B2 | if at 69:4 | `if got.State != Unwired \|\| got.Code != RefusalScopeMismatch {` | _test.go — 커버 계측 밖(분기 진입 여부 미측정) |

## Author concept rows (2026-08-04 — N 번호는 AST 분기 번호가 아님)

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| N1 | each account/profile/market/order/session/quantity/trigger/replace/tool/build/evidence/broker field substitution | test-only fixture mutation | `RefusalScopeMismatch` | table row subtest |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `Assess` | project sealed market verdict | no retry; pure evaluation | CodeGraph + AST |

## State mutations and fallbacks

- No production mutation; each subtest creates a fresh signed fixture and mutates only its runtime scope.

## Safety conclusion

- Safe edit boundary: exact-scope matrix expectations only
- High-risk impact: no (test)
- High-risk impact: no (test)
