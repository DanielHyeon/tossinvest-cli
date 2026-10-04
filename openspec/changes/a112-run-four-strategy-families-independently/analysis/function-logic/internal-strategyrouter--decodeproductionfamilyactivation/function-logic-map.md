# Function Logic Map: `decodeProductionFamilyActivation`

- Source: `internal/strategyrouter/production_family_activation.go`
- Source SHA-256: `7842842bc1b1bf9d2be542624169502fcc846a4e60080841887f5072f4fb6507`
- Signature: `decodeProductionFamilyActivation(params=1, results=2)`
- Source range: `508:1`–`527:2`
- AST evidence: `ast.json` — **편집 뒤**(a112 8.8.4-B).
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- 판정 불변 — 같은 바이트가 같은 sentinel 로 거절.

## Branches and early returns

- Exact AST return nodes: `510:3, 517:3, 520:3, 524:3, 526:2`.

| Branch | AST kind | Source location | Meaning |
|---|---|---|---|
| B1 | if | 509:2 | 크기 0 또는 상한 초과 → `%w: manifest size …` |
| B2 | if | 516:2 | json 해석 실패(모르는 필드 포함) → `%w: manifest json: %w` |
| B3 | if | 519:2 | 문서 뒤 데이터 → `%w: trailing data …` |
| B4 | if | 523:2 | 정규 직렬화와 다름 → `%w: manifest bytes are not …canonical…` |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `len` | 509:5 |
| `len` | 509:23 |
| `fmt.Errorf` | 510:44 |
| `len` | 511:4 |
| `json.NewDecoder` | 513:13 |
| `bytes.NewReader` | 513:29 |
| `decoder.DisallowUnknownFields` | 514:2 |
| `decoder.Decode` | 516:12 |
| `fmt.Errorf` | 517:44 |
| `decoder.Decode` | 519:12 |
| `fmt.Errorf` | 520:44 |
| `json.Marshal` | 522:20 |
| `bytes.Equal` | 523:20 |
| `fmt.Errorf` | 524:44 |

## State mutations and fallbacks

- 상태 변경 없음.

## Safety conclusion

- 판정 · 수락 집합 불변 — 같은 입력이 같은 sentinel 로 거절된다(errors.Is 보존, `==` 비교 0 — grep). 메시지에 필드 이름만 더해진다.
