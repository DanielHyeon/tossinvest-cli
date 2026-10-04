# Function Logic Map: `validateProductionFamilyActivation`

- Source: `internal/strategyrouter/production_family_activation.go`
- Source SHA-256: `7842842bc1b1bf9d2be542624169502fcc846a4e60080841887f5072f4fb6507`
- Signature: `validateProductionFamilyActivation(params=2, results=2)`
- Source range: `539:1`–`621:2`
- AST evidence: `ast.json` — **편집 뒤**(a112 8.8.4-B).
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- 판정 불변 — 복합 결속의 결정식은 앞 판 OR 의 항과 같다(`len(fields) != 0`).

## Branches and early returns

- Exact AST return nodes: `556:3, 572:3, 576:3, 604:4, 609:4, 613:4, 618:3, 620:2`.

| Branch | AST kind | Source location | Meaning |
|---|---|---|---|
| B1 | if | 543:2 | 몸통 결속(복합) → `%w: body binding: <어긋난 필드 전부>` |
| B2 | if | 563:2 | 수명(복합) → `%w: lifetime: <어긋난 항목 전부>` |
| B3 | if | 575:2 | 만료 → familyActivationRemaining 의 Expired 오류 그대로 |
| B4 | range | 593:2 | 서술자 순회 |
| B5 | if | 596:3 | 서술자 필드(복합) → `%w: descriptors[lane_id=…]: <어긋난 필드>` |
| B6 | if | 608:3 | effective ON 인데 desired ON 아님 → 이유 |
| B7 | if | 612:3 | 중복 레인 → `%w: descriptors: duplicate lane_id …` |
| B8 | if | 617:2 | 네 레인이 아님 → `%w: descriptors: N of M lanes` |

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

## State mutations and fallbacks

- 상태 변경 없음.

## Safety conclusion

- 판정 · 수락 집합 불변 — 같은 입력이 같은 sentinel 로 거절된다(errors.Is 보존, `==` 비교 0 — grep). 메시지에 필드 이름만 더해진다.
- 복합 결속은 분기 하나 그대로 두고 그 안에서 필드별 비교를 모은다(Q-B1=(c)); 항은 전부 순수 비교 · 부작용 없는 검사라 무조건 평가해도 판정이 같다(2026-10-04 단락 평가 전제 확인).
