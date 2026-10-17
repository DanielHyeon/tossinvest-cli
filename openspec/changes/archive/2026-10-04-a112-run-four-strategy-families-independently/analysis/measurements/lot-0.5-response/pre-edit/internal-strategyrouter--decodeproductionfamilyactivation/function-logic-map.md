# Function Logic Map (편집 전): `decodeProductionFamilyActivation`

- Source: `internal/strategyrouter/production_family_activation.go`
- Source SHA-256: `57123f814187d9201ec6999d6375adc649b579c89beb6c4731f36cc51243b005`
- Signature: `decodeProductionFamilyActivation(params=1, results=2)`
- Source range: `509:1`–`528:2`
- AST evidence: `ast.json` — 편집 **전**(HEAD 4cbcfb36).

## Inputs and invariants

- 편집 계획: 0.5 리뷰 유지#4=보안#2: 해석 거절의 둘째 %w 를 %v 로(사슬에 sentinel 하나 — 로트 자기 규칙 복원). 분기 불변.

## Branches and early returns

- Exact AST return nodes: `511:3`, `518:3`, `521:3`, `525:3`, `527:2`.

| Branch | AST kind | Source location | Condition |
|---|---|---|---|
| B1 | if | 510:2 | `if len(data) == 0 // len(data) > productionFamilyActivationMaximumBytes {` |
| B2 | if | 517:2 | `if err := decoder.Decode(&manifest); err != nil {` |
| B3 | if | 520:2 | `if err := decoder.Decode(&struct{}{}); err != io.EOF {` |
| B4 | if | 524:2 | `if err != nil // !bytes.Equal(canonical, data) {` |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `len` | 510:5 |
| `len` | 510:23 |
| `fmt.Errorf` | 511:44 |
| `len` | 512:4 |
| `json.NewDecoder` | 514:13 |
| `bytes.NewReader` | 514:29 |
| `decoder.DisallowUnknownFields` | 515:2 |
| `decoder.Decode` | 517:12 |
| `fmt.Errorf` | 518:44 |
| `decoder.Decode` | 520:12 |
| `fmt.Errorf` | 521:44 |
| `json.Marshal` | 523:20 |
| `bytes.Equal` | 524:20 |
| `fmt.Errorf` | 525:44 |

## Safety conclusion

- High-risk 인접(활성화 적재) — 판정 · sentinel 불변, json 오류가 사슬의 둘째 신원이 되지 않게만 한다.
