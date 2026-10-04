# Function Logic Map (편집 전): `decodeProductionFamilyActivation`

- Source: `internal/strategyrouter/production_family_activation.go`
- Source SHA-256: `c8e2efb9b1a8243bcec000f0c2fa9e96bd8576c97aa35a694a1e6c053407288f`
- Signature: `decodeProductionFamilyActivation(params=1, results=2)`
- Source range: `492:1`–`510:2`
- AST evidence: `ast.json` — 편집 **전**(HEAD f473d815).

## Inputs and invariants

- 편집 계획: a112 8.8.4 로트 B 항목 1(Q-B1=(c)): 맨 sentinel 반환마다 필드명 %w 래핑(errors.Is 보존). 복합 결속은 분기 그대로 필드-diff 목록으로 한 번 반환. :455 읽기 실패는 underlying err 를 %w 사슬로 보존하고 digest 불일치만 필드명(두 종류 분리).

## Branches and early returns

- Exact AST return nodes: `494:3`, `500:3`, `503:3`, `507:3`, `509:2`.

| Branch | AST kind | Source location | Condition |
|---|---|---|---|
| B1 | if | 493:2 | `if len(data) == 0 // len(data) > productionFamilyActivationMaximumBytes {` |
| B2 | if | 499:2 | `if err := decoder.Decode(&manifest); err != nil {` |
| B3 | if | 502:2 | `if err := decoder.Decode(&struct{}{}); err != io.EOF {` |
| B4 | if | 506:2 | `if err != nil // !bytes.Equal(canonical, data) {` |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `len` | 493:5 |
| `len` | 493:23 |
| `json.NewDecoder` | 496:13 |
| `bytes.NewReader` | 496:29 |
| `decoder.DisallowUnknownFields` | 497:2 |
| `decoder.Decode` | 499:12 |
| `decoder.Decode` | 502:12 |
| `json.Marshal` | 505:20 |
| `bytes.Equal` | 506:20 |

## Safety conclusion

- 판정 · 수락 집합 불변 — 같은 입력이 같은 sentinel 로 거절된다(errors.Is). 메시지에 필드명이 더해질 뿐. 미선언(Undeclared)이 맨 앞인 순서 불변.
