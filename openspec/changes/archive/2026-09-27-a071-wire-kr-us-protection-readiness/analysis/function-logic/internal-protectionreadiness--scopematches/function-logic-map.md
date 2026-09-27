# Function Logic Map: `scopeMatches`

- Source: `internal/protectionreadiness/attestation.go` (126-131)
- Revision: current — HEAD `648df8ef`; source_sha256 `130e1174f65d407f68b78273c6e3f8c5cc1c373e6fd33274aa9b93e9d118e502`
- AST evidence: `ast.json` (`tools/logic-map` 추출기 출력)
- Risk scan: `risk-pattern-report.md` (ast-grep, `tools/logic-map/sgconfig.yml`)
- Extractor counts: AST branches 0 · returns 1 · calls 0
- Exact AST return positions: 127:2
- 2026-09-25 task 5.1 갱신: 분기 표는 `analysis/harness/51_refresh.py` 가 ast.json 에서 생성함. 아래 산문 절은 a071 저자 기록(2026-08-04) 그대로이며 base 뒤 다른 change 가 이 함수에 더한 동작은 기술하지 않음.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| signed body and runtime scope | every authority field exactly equal | signed body plus sealed runtime manifest | false produces typed scope mismatch |

## Branches and early returns

| Branch | AST anchor | Source text at anchor | Coverage at HEAD (untagged package suite, per-test runs) |
|---|---|---|---|
| B1 | branchless happy path at 126:1 | `func scopeMatches(body attestationBody, scope runtimeScope) bool {` | called 18/32 |

## Author concept rows (2026-08-04 — N 번호는 AST 분기 번호가 아님)

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| N1 | any account/profile/market/order/session/quantity bounds/trigger/replace/broker/tool/build/evidence field differs | none | false | exact scope matrix |
| N2 | all fields equal | none | true | valid signed fixture |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| none | direct exact equality including broker struct | no error/retry | CodeGraph + AST |

## State mutations and fallbacks

- Pure comparison; no defaults, coercion, or mutation.

## Safety conclusion

- Safe edit boundary: exact attestation/runtime intersection only
- High-risk impact: yes; authority must fail closed
