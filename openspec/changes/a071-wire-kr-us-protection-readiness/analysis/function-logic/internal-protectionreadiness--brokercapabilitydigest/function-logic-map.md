# Function Logic Map: `brokerCapabilityDigest`

- Source: `internal/protectionreadiness/types.go` (117-129)
- Revision: current — HEAD `648df8ef`; source_sha256 `782c26b8096f87efae82b074c4721281ff00ec938156678618fa5ea7f542ab1d`
- AST evidence: `ast.json` (`tools/logic-map` 추출기 출력)
- Risk scan: `risk-pattern-report.md` (ast-grep, `tools/logic-map/sgconfig.yml`)
- Extractor counts: AST branches 0 · returns 1 · calls 7
- Exact AST return positions: 128:2
- 2026-09-25 task 5.1 갱신: 분기 표는 `analysis/harness/51_refresh.py` 가 ast.json 에서 생성함. 아래 산문 절은 a071 저자 기록(2026-08-04) 그대로이며 base 뒤 다른 change 가 이 함수에 더한 동작은 기술하지 않음.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| broker capability | exact eight-field capability tuple | attested manifest/body | deterministic SHA-256 digest |

## Branches and early returns

| Branch | AST anchor | Source text at anchor | Coverage at HEAD (untagged package suite, per-test runs) |
|---|---|---|---|
| B1 | branchless happy path at 117:1 | `func brokerCapabilityDigest(capability brokerCapability) string {` | called 17/32 |

## Author concept rows (2026-08-04 — N 번호는 AST 분기 번호가 아님)

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| N1 | any capability field differs | none | different digest | dispatch substitution matrix |
| N2 | identical capability | none | identical digest | valid dispatch fixture |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `hashStrings`, `boolString`, `hexBytes` | unambiguous deterministic binding | no error/retry | CodeGraph + AST |

## State mutations and fallbacks

- Pure digest; no mutation or defaults.

## Safety conclusion

- Safe edit boundary: include every capability field in stable order
- High-risk impact: yes; dispatch authority binding
