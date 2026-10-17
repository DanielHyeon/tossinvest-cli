# Function Logic Map: `TestDispatchRejectsCorruptAndFutureIssuedSnapshots`

- Source: `internal/protectionreadiness/dispatch_test.go` (35-45)
- Revision: current — HEAD `648df8ef`; source_sha256 `4be237f9eb604ce721dc4122fd58310780bc6d58d11c70bbe2cc56c581a7cd55`
- AST evidence: `ast.json` (`tools/logic-map` 추출기 출력)
- Risk scan: `risk-pattern-report.md` (ast-grep, `tools/logic-map/sgconfig.yml`)
- Extractor counts: AST branches 1 · returns 0 · calls 8
- 2026-09-25 task 5.1 갱신: 분기 표는 `analysis/harness/51_refresh.py` 가 ast.json 에서 생성함. 아래 산문 절은 a071 저자 기록(2026-08-04) 그대로이며 base 뒤 다른 change 가 이 함수에 더한 동작은 기술하지 않음.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| valid sealed KR snapshot then mutated provenance | mutation invalidates market seal | test fixture | assert state-corrupt and denied |

## Branches and early returns

| Branch | AST anchor | Source text at anchor | Coverage at HEAD (untagged package suite, per-test runs) |
|---|---|---|---|
| B1 | if at 42:2 | `if got := snapshot.Dispatch(scope, readinessNow); got.Code != RefusalStateCorrupt \|\| got.Allowed {` | _test.go — 커버 계측 밖(분기 진입 여부 미측정) |

## Author concept rows (2026-08-04 — N 번호는 AST 분기 번호가 아님)

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| N1 | sealed expiry changed without resealing | test-only value mutation | `RefusalStateCorrupt` | named test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `Assess`, `Dispatch` | create then validate sealed snapshot | fail assertion | CodeGraph + AST |

## State mutations and fallbacks

- No production mutation; corrupts a local snapshot copy only.

## Safety conclusion

- Safe edit boundary: corruption rejection assertion only
- High-risk impact: no (test)
