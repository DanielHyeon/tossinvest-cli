# Function Logic Map: `validateProductionFamilyActivation`

- Source: `internal/strategyrouter/production_family_activation.go`
- Source SHA-256: `bac16c04b38d49c479a381fb326d7dd066e5525997322619d8ceb496e3ddb0d8`
- Signature: `validateProductionFamilyActivation(params=2, results=2)`
- Source range: `540:1`–`624:2`
- AST evidence: `ast.json` — **편집 뒤**(a112 8.5-R).
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- 판정 불변 — 복합 결속의 결정식은 앞 판 OR 의 항과 같다(`len(fields) != 0`).

## Branches and early returns

- Exact AST return nodes: `557:3, 573:3, 577:3, 607:4, 612:4, 616:4, 621:3, 623:2`.

| Branch | AST kind | Source location | Meaning |
|---|---|---|---|
| B1 | if | 544:2 | 몸통 결속(복합) → `%w: body binding: <어긋난 필드 전부>` |
| B2 | if | 564:2 | 수명(복합) → `%w: lifetime: <어긋난 항목 전부>` |
| B3 | if | 576:2 | 만료 → familyActivationRemaining 의 Expired 오류 그대로 |
| B4 | range | 596:2 | 서술자 순회 |
| B5 | if | 599:3 | 서술자 필드(복합) → `%w: descriptors[i]: <어긋난 필드>`(lane_id 원문 없음) |
| B6 | if | 611:3 | effective ON 인데 desired ON 아님 → `descriptors[i]: effective ON without desired ON` |
| B7 | if | 615:3 | 중복 레인 → `%w: descriptors[i]: duplicate lane_id`(원문 없음) |
| B8 | if | 620:2 | 네 레인이 아님 → `%w: descriptors: N of M lanes` |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `failedFields` | 544:15 |
| `validMarket` | 548:57 |
| `productionRouteIdentity` | 555:24 |
| `len` | 556:5 |
| `fmt.Errorf` | 557:15 |
| `strings.Join` | 557:92 |
| `productionRouteTime` | 559:26 |
| `productionRouteTime` | 560:22 |
| `productionRouteTime` | 561:24 |
| `config.ObservedAt.UTC` | 562:9 |
| `failedFields` | 564:15 |
| `issued.Before` | 568:46 |
| `issued.After` | 569:45 |
| `issued.Before` | 570:50 |
| `expires.Sub` | 571:39 |
| `len` | 572:5 |
| `fmt.Errorf` | 573:15 |
| `strings.Join` | 573:88 |
| `familyActivationRemaining` | 576:15 |
| `productionRouteDescriptors` | 585:10 |
| `make` | 593:11 |
| `len` | 593:72 |
| `failedFields` | 599:16 |
| `validDesiredState` | 604:27 |
| `validDesiredState` | 605:29 |
| `len` | 606:6 |
| `fmt.Errorf` | 607:16 |
| `strings.Join` | 607:103 |
| `fmt.Errorf` | 612:16 |
| `fmt.Errorf` | 616:16 |
| `len` | 620:5 |
| `len` | 620:19 |
| `fmt.Errorf` | 621:15 |
| `len` | 621:103 |
| `len` | 621:115 |

## State mutations and fallbacks

- 상태 변경 없음.

## Safety conclusion

- 판정 · 수락 집합 불변 — 메시지 문자열만 바뀐다(sentinel 그대로, 배타성은 `a112_activation_error_fields_test.go` `TestEveryActivationRefusalNamesItsFieldAndKeepsItsSentinel` 가 모양마다 잰다).

0.5 응답 로트: 같은 파일의 주석 · 한 글자 편집(본문 구조 불변)으로 파일 SHA 만 바뀜
