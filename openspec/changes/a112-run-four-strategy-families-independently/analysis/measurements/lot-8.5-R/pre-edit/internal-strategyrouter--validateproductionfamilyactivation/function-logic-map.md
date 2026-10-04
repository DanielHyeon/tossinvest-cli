# Function Logic Map (편집 전): `validateProductionFamilyActivation`

- Source: `internal/strategyrouter/production_family_activation.go`
- Source SHA-256: `7842842bc1b1bf9d2be542624169502fcc846a4e60080841887f5072f4fb6507`
- Signature: `validateProductionFamilyActivation(params=2, results=2)`
- Source range: `539:1`–`621:2`
- AST evidence: `ast.json` — 편집 **전**(a112 8.5 응답 로트(HEAD 178cc196)).

## Inputs and invariants

- 편집 계획: 응답 로트 ③(P2-b): 서술자 거절 셋(B5 · B6 · B7)의 메시지에서 lane_id 원문을 빼고 위치 descriptors[%d] 로 — 개행 · 임의 문자열이 오류에 실리지 않게. 분기 불변.

## Branches and early returns

- Exact AST return nodes: `556:3`, `572:3`, `576:3`, `604:4`, `609:4`, `613:4`, `618:3`, `620:2`.

| Branch | AST kind | Source location | Condition |
|---|---|---|---|
| B1 | if | 543:2 | `if fields := failedFields(` |
| B2 | if | 563:2 | `if fields := failedFields(` |
| B3 | if | 575:2 | `if _, err := familyActivationRemaining(expires, now); err != nil {` |
| B4 | range | 593:2 | `for _, descriptor := range body.Descriptors {` |
| B5 | if | 596:3 | `if fields := failedFields(` |
| B6 | if | 608:3 | `if descriptor.Effective == StateOn && descriptor.Desired != StateOn {` |
| B7 | if | 612:3 | `if _, duplicate := state[key]; duplicate {` |
| B8 | if | 617:2 | `if len(state) != len(want) {` |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `failedFields` | 543:15 |
| `validMarket` | 547:57 |
| `productionRouteIdentity` | 554:24 |
| `len` | 555:5 |
| `fmt.Errorf` | 556:15 |
| `strings.Join` | 556:92 |
| `productionRouteTime` | 558:26 |
| `productionRouteTime` | 559:22 |
| `productionRouteTime` | 560:24 |
| `config.ObservedAt.UTC` | 561:9 |
| `failedFields` | 563:15 |
| `issued.Before` | 567:46 |
| `issued.After` | 568:45 |
| `issued.Before` | 569:50 |
| `expires.Sub` | 570:39 |
| `len` | 571:5 |
| `fmt.Errorf` | 572:15 |
| `strings.Join` | 572:88 |
| `familyActivationRemaining` | 575:15 |
| `productionRouteDescriptors` | 584:10 |
| `make` | 592:11 |
| `len` | 592:72 |
| `failedFields` | 596:16 |
| `validDesiredState` | 601:27 |
| `validDesiredState` | 602:29 |
| `len` | 603:6 |
| `fmt.Errorf` | 604:16 |
| `strings.Join` | 604:123 |
| `fmt.Errorf` | 609:16 |
| `fmt.Errorf` | 613:16 |
| `len` | 617:5 |
| `len` | 617:19 |
| `fmt.Errorf` | 618:15 |
| `len` | 618:103 |
| `len` | 618:115 |

## Safety conclusion

- 판정 · 수락 집합 불변 — 메시지 문자열만 바뀐다(sentinel 그대로).
