# Function Logic Map: `currentPair (시험 도우미)`

- Source: `internal/strategyprojection/projection_test.go`
- Source SHA-256: `44ffddfa4f011fdfb8484929cb28babca7a8367a43daeec90c1042667ed74dd0`
- Signature: `currentPair(params=1, results=1)`
- Source range: `113:1`–`142:2`
- AST evidence: `ast.json` — **편집 뒤**(a112 7.3).
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- 시장 레코드 값은 편집 전과 같다.

## Branches and early returns

- Exact AST return nodes: `141:2`.

| Branch | AST kind | Source location | Meaning |
|---|---|---|---|
| B1 | range | 116:2 | KR · US 시장 레코드 만들기 |
| B2 | if | 138:2 | 만든 스냅숏이 Validate 를 통과 |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `t.Helper` | 114:2 |
| `projectionNow.Add` | 117:15 |
| `string` | 118:37 |
| `digest` | 118:53 |
| `string` | 118:72 |
| `string` | 119:34 |
| `string` | 120:36 |
| `string` | 122:59 |
| `digest` | 123:23 |
| `string` | 123:46 |
| `DormantSnapshot` | 136:14 |
| `Validate` | 138:12 |
| `t.Fatal` | 139:3 |

## State mutations and fallbacks

- 시험 코드.

## Safety conclusion

- 시험 코드 — 생산 경로 없음.
