# Function Logic Map: `failedFields (새 함수)`

- Source: `internal/strategyrouter/production_family_activation.go`
- Source SHA-256: `7842842bc1b1bf9d2be542624169502fcc846a4e60080841887f5072f4fb6507`
- Signature: `failedFields(params=1, results=1)`
- Source range: `630:1`–`638:2`
- AST evidence: `ast.json` — **편집 뒤**(a112 8.8.4-B).
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- 결정은 호출자의 `len(fields) != 0` 하나 — 이 함수가 비면 결속이 통과하므로(B3) 이름 수집이 곧 판정 입력이다.

## Branches and early returns

- Exact AST return nodes: `637:2`.

| Branch | AST kind | Source location | Meaning |
|---|---|---|---|
| B1 | range | 632:2 | 검사 순회 |
| B2 | if | 633:3 | 실패한 검사의 이름을 모음 |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `append` | 634:13 |

## State mutations and fallbacks

- 상태 변경 없음.

## Safety conclusion

- 판정 입력이다 — 변이 B3 이 그것을 잡는다.
