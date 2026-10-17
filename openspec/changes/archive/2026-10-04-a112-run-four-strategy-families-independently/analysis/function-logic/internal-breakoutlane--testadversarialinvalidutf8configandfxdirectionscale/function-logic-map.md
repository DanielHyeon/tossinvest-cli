# Function Logic Map: `TestAdversarialInvalidUTF8ConfigAndFXDirectionScale (시험)`

- Source: `internal/breakoutlane/evaluator_test.go`
- Source SHA-256: `878c75713f0d714c0a6569f8f7b314a7f9796611f38d28fb50939f8138982b3e`
- Signature: `TestAdversarialInvalidUTF8ConfigAndFXDirectionScale(params=1, results=0)`
- Source range: `307:1`–`325:2`
- AST evidence: `ast.json` — **편집 뒤**(a112 B2).
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- 역방향 봉인은 주어진 모양 그대로 봉인돼야 정규화된다.

## Branches and early returns

- Exact AST return nodes: ``.

| Branch | AST kind | Source location | Meaning |
|---|---|---|---|
| B1 | if | 311:2 | 이 시험 자신의 판정 갈래(UTF-8 거절 · 역방향 digest 없음 거절 · 봉인된 역방향 정규화) |
| B2 | if | 317:2 | 이 시험 자신의 판정 갈래(UTF-8 거절 · 역방향 digest 없음 거절 · 봉인된 역방향 정규화) |
| B3 | if | 322:2 | 이 시험 자신의 판정 갈래(UTF-8 거절 · 역방향 digest 없음 거절 · 봉인된 역방향 정규화) |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `fixtureInput` | 308:7 |
| `string` | 309:13 |
| `Evaluate` | 311:10 |
| `d.Refusal` | 311:28 |
| `t.Fatal` | 312:3 |
| `d.Refusal` | 312:11 |
| `NewFXSeal` | 317:15 |
| `t.Fatal` | 318:3 |
| `FXSealDigest` | 320:15 |
| `NewFXSeal` | 321:12 |
| `t.Fatalf` | 323:3 |

## State mutations and fallbacks

- 시험 코드.

## Safety conclusion

- 시험 코드 — 생산 경로 없음.
