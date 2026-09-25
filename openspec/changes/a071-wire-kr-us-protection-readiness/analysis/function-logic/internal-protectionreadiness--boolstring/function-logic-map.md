# Function Logic Map: `boolString`

- Source: `internal/protectionreadiness/types.go` (131-136)
- Revision: current — HEAD `648df8ef`; source_sha256 `782c26b8096f87efae82b074c4721281ff00ec938156678618fa5ea7f542ab1d`
- AST evidence: `ast.json` (`tools/logic-map` 추출기 출력)
- Risk scan: `risk-pattern-report.md` (ast-grep, `tools/logic-map/sgconfig.yml`)
- Extractor counts: AST branches 1 · returns 2 · calls 0
- Exact AST return positions: 133:3, 135:2
- 2026-09-25 task 5.1 갱신: 분기 표는 `analysis/harness/51_refresh.py` 가 ast.json 에서 생성함. 아래 산문 절은 a071 저자 기록(2026-08-04) 그대로이며 base 뒤 다른 change 가 이 함수에 더한 동작은 기술하지 않음.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| boolean | true or false | broker capability field | canonical lowercase token |

## Branches and early returns

| Branch | AST anchor | Source text at anchor | Coverage at HEAD (untagged package suite, per-test runs) |
|---|---|---|---|
| B1 | if at 132:2 | `if value {` | entered 17/32 |

## Author concept rows (2026-08-04 — N 번호는 AST 분기 번호가 아님)

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| N1 | true | none | `"true"` | digest fixture |
| N2 | false | none | `"false"` | capability substitution |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| none | canonical conversion | no error/retry | CodeGraph + AST |

## State mutations and fallbacks

- Pure conversion.

## Safety conclusion

- Safe edit boundary: stable lowercase tokens
- High-risk impact: no
