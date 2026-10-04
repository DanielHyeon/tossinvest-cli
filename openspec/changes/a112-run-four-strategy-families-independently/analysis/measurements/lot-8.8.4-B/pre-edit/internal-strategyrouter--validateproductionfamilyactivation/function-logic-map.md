# Function Logic Map (편집 전): `validateProductionFamilyActivation`

- Source: `internal/strategyrouter/production_family_activation.go`
- Source SHA-256: `c8e2efb9b1a8243bcec000f0c2fa9e96bd8576c97aa35a694a1e6c053407288f`
- Signature: `validateProductionFamilyActivation(params=2, results=2)`
- Source range: `522:1`–`584:2`
- AST evidence: `ast.json` — 편집 **전**(HEAD f473d815).

## Inputs and invariants

- 편집 계획: a112 8.8.4 로트 B 항목 1(Q-B1=(c)): 맨 sentinel 반환마다 필드명 %w 래핑(errors.Is 보존). 복합 결속은 분기 그대로 필드-diff 목록으로 한 번 반환. :455 읽기 실패는 underlying err 를 %w 사슬로 보존하고 digest 불일치만 필드명(두 종류 분리).

## Branches and early returns

- Exact AST return nodes: `533:3`, `541:3`, `545:3`, `567:4`, `572:4`, `576:4`, `581:3`, `583:2`.

| Branch | AST kind | Source location | Condition |
|---|---|---|---|
| B1 | if | 525:2 | `if body.SchemaVersion != productionFamilyActivationSchema // body.Domain != productionFamilyActivationDomain //` |
| B2 | if | 539:2 | `if !okApproved // !okIssued // !okExpires // issued.Before(approved) // issued.After(now) //` |
| B3 | if | 544:2 | `if _, err := familyActivationRemaining(expires, now); err != nil {` |
| B4 | range | 562:2 | `for _, descriptor := range body.Descriptors {` |
| B5 | if | 564:3 | `if !known // table.Family != descriptor.Family // table.Horizon != descriptor.Horizon //` |
| B6 | if | 571:3 | `if descriptor.Effective == StateOn && descriptor.Desired != StateOn {` |
| B7 | if | 575:3 | `if _, duplicate := state[key]; duplicate {` |
| B8 | if | 580:2 | `if len(state) != len(want) {` |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `validMarket` | 526:60 |
| `productionRouteIdentity` | 532:4 |
| `productionRouteTime` | 535:26 |
| `productionRouteTime` | 536:22 |
| `productionRouteTime` | 537:24 |
| `config.ObservedAt.UTC` | 538:9 |
| `issued.Before` | 539:47 |
| `issued.After` | 539:74 |
| `issued.Before` | 540:4 |
| `expires.Sub` | 540:30 |
| `familyActivationRemaining` | 544:15 |
| `productionRouteDescriptors` | 553:10 |
| `make` | 561:11 |
| `len` | 561:72 |
| `validDesiredState` | 566:5 |
| `validDesiredState` | 566:47 |
| `len` | 580:5 |
| `len` | 580:19 |

## Safety conclusion

- 판정 · 수락 집합 불변 — 같은 입력이 같은 sentinel 로 거절된다(errors.Is). 메시지에 필드명이 더해질 뿐. 미선언(Undeclared)이 맨 앞인 순서 불변.
