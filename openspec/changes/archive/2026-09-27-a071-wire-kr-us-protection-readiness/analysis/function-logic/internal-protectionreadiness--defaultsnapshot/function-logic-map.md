# Function Logic Map: `DefaultSnapshot`

- Source: `internal/protectionreadiness/readiness.go` (27-35)
- Revision: current — HEAD `648df8ef`; source_sha256 `d95f486a1cef853f912c9ab6330896f930cdc6b73f2efdce5aa138bd0510337c`
- AST evidence: `ast.json` (`tools/logic-map` 추출기 출력)
- Risk scan: `risk-pattern-report.md` (ast-grep, `tools/logic-map/sgconfig.yml`)
- Extractor counts: AST branches 0 · returns 1 · calls 3
- Exact AST return positions: 34:2
- 2026-09-25 task 5.1 갱신: 분기 표는 `analysis/harness/51_refresh.py` 가 ast.json 에서 생성함. 아래 산문 절은 a071 저자 기록(2026-08-04) 그대로이며 base 뒤 다른 change 가 이 함수에 더한 동작은 기술하지 않음.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| paired release | exact KR and US | compiled release | both UNWIRED |

## Branches and early returns

| Branch | AST anchor | Source text at anchor | Coverage at HEAD (untagged package suite, per-test runs) |
|---|---|---|---|
| B1 | branchless happy path at 27:1 | `func DefaultSnapshot() ReadinessSnapshot {` | called 29/32 |

## Author concept rows (2026-08-04 — N 번호는 AST 분기 번호가 아님)

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| N1 | branchless construction | local seals only | paired snapshot | default assembly test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| market/global seal helpers | immutable defaults | no I/O | AST |

## State mutations and fallbacks

- Constructs both market refusals; no provider, broker, toggle or approval mutation.

## Safety conclusion

- Safe edit boundary: add independent market seals.
- High-risk impact: no; defaults remain UNWIRED.
