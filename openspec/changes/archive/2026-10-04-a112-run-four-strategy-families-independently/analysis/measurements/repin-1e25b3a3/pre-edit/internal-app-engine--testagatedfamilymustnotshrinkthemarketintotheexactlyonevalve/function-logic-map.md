# Function Logic Map (편집 전): `TestAGatedFamilyMustNotShrinkTheMarketIntoTheExactlyOneValve`

- Source: `internal/app/engine/a112_family_gate_test.go`
- Source SHA-256: `fcc45e1e7f7f9730dfaaacbfd37f73782053ab729f13b1f210af8699417f2702`
- Signature: `TestAGatedFamilyMustNotShrinkTheMarketIntoTheExactlyOneValve(params=1, results=0)`
- Source range: `465:1`–`497:2`
- AST evidence: `ast.json` — 편집 **전**(a112 base 재고정 1e25b3a3 — 8.5 응답 로트(6f5b0df6) 편집 전(178cc196)).

## Inputs and invariants

- 편집 계획: 8.5 응답 로트 ⑦: 판정 활성화 carry 단언 추가(B10/B12 — 보이스 3 P2-1); collectOverflowing 은 configure 가변 인자(관문 아래 실행).

## Branches and early returns

- Exact AST return nodes: 없음.

| Branch | AST kind | Source location | Condition |
|---|---|---|---|
| B1 | range | 468:2 | `for _, lane := range runtime.lanesFor(StrategyMarketKR) {` |
| B2 | if | 469:3 | `if lane.Key().Family != strategyrouter.FamilyContinuation {` |
| B3 | if | 472:3 | `if _, locked := lane.Fail("a measured fault", true); !locked {` |
| B4 | if | 477:2 | `if latched != 1 {` |
| B5 | if | 482:2 | `if authority.snapshot.Ready {` |
| B6 | if | 487:2 | `if authority.snapshot.Reason != StrategyProposalFamilyGateClosed {` |
| B7 | if | 490:2 | `if authority.snapshot.RefusedCount != authority.snapshot.RoutedCount {` |
| B8 | if | 494:2 | `if _, handedOff := authority.dispatchHandoff().Single(); handedOff {` |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `familyGateFixture` | 466:25 |
| `runtime.lanesFor` | 468:23 |
| `lane.Key` | 469:6 |
| `lane.Fail` | 472:19 |
| `t.Fatal` | 473:4 |
| `t.Fatalf` | 478:3 |
| `collectUnderGate` | 480:15 |
| `t.Fatalf` | 483:3 |
| `len` | 485:68 |
| `t.Fatalf` | 488:3 |
| `t.Fatalf` | 491:3 |
| `Single` | 494:21 |
| `authority.dispatchHandoff` | 494:21 |
| `t.Fatal` | 495:3 |

## Safety conclusion

- 시험 코드 — 생산 경로 없음.
