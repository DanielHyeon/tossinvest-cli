# Function Logic Map: `Gateway.submit`

- Source: `internal/execgw/gateway.go` (466-763)
- Revision: current — HEAD `648df8ef`; source_sha256 `9601d6562e363a2a5c70f69eccd02e24b5c8d5e216412b8fb9ea6d213ef36832`
- AST evidence: `ast.json` (`tools/logic-map` 추출기 출력)
- Risk scan: `risk-pattern-report.md` (ast-grep, `tools/logic-map/sgconfig.yml`)
- Extractor counts: AST branches 57 · returns 44 · calls 89
- Exact AST return positions: 476:4, 484:5, 489:4, 507:3, 520:3, 526:3, 528:3, 536:3, 538:3, 546:3, 557:5, 565:5, 568:4, 570:3, 584:3, 589:3, 596:4, 605:3, 608:3, 614:4, 618:4, 626:3, 649:4, 652:4, 655:4, 659:4, 666:4, 674:5, 686:5, 690:6, 693:6, 696:6, 703:6, 709:4, 714:5, 719:5, 724:4, 726:3, 741:4, 745:4, 747:3, 757:3, 760:3, 762:2
- 2026-09-25 task 5.1 갱신: 분기 표는 `analysis/harness/51_refresh.py` 가 ast.json 에서 생성함. 아래 산문 절은 a071 저자 기록(2026-08-04) 그대로이며 base 뒤 다른 change 가 이 함수에 더한 동작은 기술하지 않음.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| function inputs | typed values | Gateway.submit | fail closed or test failure |

## Branches and early returns

| Branch | AST anchor | Source text at anchor | Coverage at HEAD (untagged package suite, per-test runs) |
|---|---|---|---|
| B1 | if at 475:3 | `if strategyCrossedTransport \|\| returnedErr == nil \|\| preTransportClaimFinalizationAttempted {` | entered 31/222 |
| B2 | if at 478:3 | `if preTransportClaim == nil && plan.strategy != nil && !preTransportClaimAttempted &&` | NOT entered 0/222 |
| B3 | if at 482:4 | `if claimRejected != nil {` | NOT entered 0/222 |
| B4 | if at 488:3 | `if preTransportClaim == nil {` | entered 87/222 |
| B5 | if at 492:3 | `if !errors.As(returnedErr, &rejected) {` | NOT entered 0/222 |
| B6 | if at 506:2 | `if err != nil {` | NOT entered 0/222 |
| B7 | if at 519:2 | `if !g.claimSymbol(symbolKey) {` | entered 2/222 |
| B8 | if at 525:2 | `if rejected, err := g.checkSymbolFree(ctx, plan); err != nil {` | entered 1/222 |
| B9 | else at 527:9 | `} else if rejected != nil {` | else-if: 본문 블록 없음 |
| B10 | if at 527:9 | `} else if rejected != nil {` | else-if: 본문 블록 없음 |
| B11 | if at 535:2 | `if rejected, err := g.checkDecisionUnspent(ctx, prep, decision); err != nil {` | NOT entered 0/222 |
| B12 | else at 537:9 | `} else if rejected != nil {` | else-if: 본문 블록 없음 |
| B13 | if at 537:9 | `} else if rejected != nil {` | else-if: 본문 블록 없음 |
| B14 | if at 545:2 | `if err != nil {` | entered 1/222 |
| B15 | if at 552:3 | `if plan.strategy != nil && preTransportClaim == nil && !preTransportClaimAttempted {` | NOT entered 0/222 |
| B16 | if at 555:4 | `if claimRejected != nil {` | NOT entered 0/222 |
| B17 | if at 561:3 | `if plan.strategy != nil && preTransportClaim != nil {` | NOT entered 0/222 |
| B18 | if at 564:4 | `if err := g.refusePreparedClaimedStrategyPreTransport(ctx, attempt, *preTransportClaim, rejected); err != n...` | NOT entered 0/222 |
| B19 | if at 583:2 | `if rejected != nil {` | entered 4/222 |
| B20 | if at 588:2 | `if rejected := g.checkEntry(plan); rejected != nil {` | entered 5/222 |
| B21 | if at 594:2 | `if plan.preflight != nil {` | entered 3/222 |
| B22 | if at 595:3 | `if rejected := plan.preflight(ctx); rejected != nil {` | entered 1/222 |
| B23 | if at 604:2 | `if loadRejected != nil {` | entered 1/222 |
| B24 | if at 607:2 | `if rejected := g.checkDecision(decision, ref, plan, g.clk.Now()); rejected != nil {` | entered 10/222 |
| B25 | if at 610:2 | `if plan.strategy != nil {` | NOT entered 0/222 |
| B26 | if at 613:3 | `if rejected != nil {` | NOT entered 0/222 |
| B27 | if at 617:3 | `if rejected := g.checkClaimedStrategyLease(ctx, decision, prep.ClientOrderID, plan); rejected != nil {` | NOT entered 0/222 |
| B28 | if at 625:2 | `if rejected := g.checkReservation(ctx, decision); rejected != nil {` | entered 4/222 |
| B29 | if at 636:3 | `if plan.strategy != nil {` | NOT entered 0/222 |
| B30 | if at 648:3 | `if rejected != nil {` | NOT entered 0/222 |
| B31 | if at 651:3 | `if rejected := g.checkDecision(fresh, ref, plan, now); rejected != nil {` | entered 1/222 |
| B32 | if at 654:3 | `if fresh.ClientOrderID != prep.ClientOrderID {` | NOT entered 0/222 |
| B33 | if at 658:3 | `if _, rejected := g.checkProtection(dctx, plan, protectionCheckpoint); rejected != nil {` | entered 1/222 |
| B34 | if at 665:3 | `if rejected := g.checkEntry(plan); rejected != nil {` | NOT entered 0/222 |
| B35 | if at 672:3 | `if plan.strategy == nil {` | entered 86/222 |
| B36 | if at 673:4 | `if rejected := g.checkReservation(dctx, fresh); rejected != nil {` | entered 1/222 |
| B37 | if at 685:4 | `if err := dctx.Err(); err != nil {` | NOT entered 0/222 |
| B38 | if at 688:4 | `if plan.strategy != nil {` | NOT entered 0/222 |
| B39 | if at 689:5 | `if plan.strategy.finalAuthorityCheck == nil {` | NOT entered 0/222 |
| B40 | if at 692:5 | `if err := plan.strategy.finalAuthorityCheck(dctx); err != nil {` | NOT entered 0/222 |
| B41 | if at 695:5 | `if rejected := g.checkReservation(dctx, fresh); rejected != nil {` | NOT entered 0/222 |
| B42 | if at 699:5 | `if requireTransport == nil {` | NOT entered 0/222 |
| B43 | if at 702:5 | `if err := requireTransport(dctx, plan.strategy.lease); err != nil {` | NOT entered 0/222 |
| B44 | if at 711:3 | `if plan.strategy != nil && !g.skipStrategyEntryABAForTest {` | NOT entered 0/222 |
| B45 | else at 717:10 | `} else {` | entered 85/222 |
| B46 | if at 712:4 | `if err := g.withStrategyEntryGateAuthority(dctx, plan.strategy.entryGateAuthority,` | NOT entered 0/222 |
| B47 | if at 718:4 | `if err := call(); err != nil {` | NOT entered 0/222 |
| B48 | if at 723:3 | `if tracker == nil {` | NOT entered 0/222 |
| B49 | if at 729:2 | `if plan.strategy == nil {` | entered 87/222 |
| B50 | else at 731:9 | `} else if g.dispatchStrategyVerified != nil {` | else-if: 본문 블록 없음 |
| B51 | if at 731:9 | `} else if g.dispatchStrategyVerified != nil {` | else-if: 본문 블록 없음 |
| B52 | else at 733:9 | `} else {` | NOT entered 0/222 |
| B53 | if at 736:2 | `if err != nil {` | entered 1/222 |
| B54 | if at 740:3 | `if errors.Is(err, journal.ErrNonceSpent) {` | NOT entered 0/222 |
| B55 | if at 744:3 | `if plan.strategy != nil && !strategyCrossedTransport {` | NOT entered 0/222 |
| B56 | if at 756:2 | `if res.Final == journal.StateConfirmed {` | entered 31/222 |
| B57 | if at 759:2 | `if res.Err != nil {` | entered 56/222 |

## Author concept rows (2026-08-04 — N 번호는 AST 분기 번호가 아님)

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| N1 | mapped AST control flow | bounded to function | typed return | affected regression |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| mapped dependencies | preserve function contract | caller handles error | CodeGraph + AST |

## State mutations and fallbacks

- No authority broadening; current behavior is covered by focused tests.

## Safety conclusion

- Safe edit boundary: Gateway.submit only.
- High-risk impact: reviewed and regression-tested.
