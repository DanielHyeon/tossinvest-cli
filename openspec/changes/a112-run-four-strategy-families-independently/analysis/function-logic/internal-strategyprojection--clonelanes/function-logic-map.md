# Function Logic Map: `cloneLanes`

- Source: `internal/strategyprojection/lanes.go`
- Source SHA-256: `9b6b15b0ebe8e74fcd582a8a6d6abed626706542ba9706e5ba551a3c302315a4`
- Signature: `cloneLanes(params=1, results=1)`
- Source range: `434:1`–`451:2`
- AST evidence: `ast.json` — **편집 뒤**(a112 7.3.1 SHADOW).
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- 복사 시험 `TestTheShadowVocabularyAndTheWireShape`.

## Branches and early returns

- Exact AST return nodes: `436:3, 450:2`.

| Branch | AST kind | Source location | Meaning |
|---|---|---|---|
| B1 | if | 435:2 | nil |
| B2 | range | 439:2 | 레인 순회 |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `make` | 438:9 |
| `len` | 438:39 |
| `clonePointer` | 440:17 |
| `cloneString` | 441:23 |
| `clonePointer` | 442:44 |
| `clonePointer` | 442:72 |
| `clonePointer` | 442:98 |
| `cloneString` | 443:18 |
| `cloneString` | 444:46 |
| `clonePointer` | 444:79 |
| `cloneTime` | 445:43 |
| `cloneTime` | 445:70 |
| `cloneString` | 446:46 |
| `cloneString` | 446:80 |
| `clonePointer` | 447:24 |

## State mutations and fallbacks

- 새 슬라이스.

## Safety conclusion

- 읽기 전용 복사.
