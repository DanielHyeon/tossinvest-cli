# Function Logic Map: `Context.Close`

- Source: `internal/app/engine/engine.go` (606-616)
- Revision: current — HEAD `648df8ef`; source_sha256 `d0cb8011c0186beeedface3cb4ba47dc81b3a79202b9ab9650dda99a44a347d7`
- AST evidence: `ast.json` (`tools/logic-map` 추출기 출력)
- Risk scan: `risk-pattern-report.md` (ast-grep, `tools/logic-map/sgconfig.yml`)
- Extractor counts: AST branches 2 · returns 3 · calls 1
- Exact AST return positions: 608:3, 613:3, 615:2
- 2026-09-25 task 5.1 갱신: 분기 표는 `analysis/harness/51_refresh.py` 가 ast.json 에서 생성함. 아래 산문 절은 a071 저자 기록(2026-08-04) 그대로이며 base 뒤 다른 change 가 이 함수에 더한 동작은 기술하지 않음.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| context journal | nil, open, or already closed | Context ownership | idempotent cleanup |

## Branches and early returns

| Branch | AST anchor | Source text at anchor | Coverage at HEAD (untagged package suite, per-test runs) |
|---|---|---|---|
| B1 | if at 607:2 | `if c == nil {` | NOT entered 0/439 |
| B2 | if at 612:2 | `if j != nil {` | entered 54/439 |

## Author concept rows (2026-08-04 — N 번호는 AST 분기 번호가 아님)

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| N1 | nil/already closed | none | nil | close idempotence test |
| N2 | owned journal | clear field then close journal | journal close error | engine lifecycle test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `Journal.Close` | release sole durable Context-owned store | no broker operation | CodeGraph + AST |

## State mutations and fallbacks

- Cleanup only; no mutation transport.

## Safety conclusion

- Safe edit boundary: retain nil-safe idempotent journal close only.
- High-risk impact: low, but resource leaks can affect restart durability.
