# Function Logic Map: `Gateway.submit`

- Source: `internal/execgw/gateway.go` (466-762)
- Qualified function: `Gateway.submit`
- Revision: `current` (a165 base `34e12967e`, `source_sha256` b4b020607839345a…)
- AST evidence: `ast.json` — AST branches 57, 반환 44, 호출 89
- Risk scan: `risk-pattern-report.md`
- 실행 칸 측정: 편집 전 기준선 `go test -count=1 -coverprofile` (execgw·official 패키지 전체, journal 은 `-run 'Classify|SendTracker|Dispatch'`) — 분기 줄에서 시작하는 본문 블록의 실행 수. 시험 이름별 귀속은 GREEN 재추출에서 함.

## 역할 (a165)

게이트웨이 순서(결정 읽기 → Prepare → 거절 → DISPATCH_STARTED → 직전 재검증 → 브로커 1회 → 정착)의 본체. a165 는 이 함수의 dispatch 클로저 안에 재전송 루프를 넣는다(design D1).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| plan mutationPlan | place/cancel/amend 계획 | Gateway.place·Cancel·Amend | prepareRequest 오류면 기록 없이 반환 |
| ref GuardianDecision | 결정 id | journal | 읽기 실패는 refusePrepared 로 기록 |
| plan.strategy | nil 또는 paired KR/US first-leg 능력 | PlaceClaimedStrategy | nil 이면 일반 경로 |

## Branches and early returns

AST 열거 순서 그대로. Condition 칸은 AST 좌표 줄의 원문.

Exact AST return positions: 476:4, 484:5, 489:4, 507:3, 520:3, 526:3, 528:3, 536:3, 538:3, 546:3, 557:5, 565:5, 568:4, 570:3, 584:3, 589:3, 596:4, 605:3, 608:3, 614:4, 618:4, 626:3, 649:4, 652:4, 655:4, 659:4, 666:4, 674:5, 686:5, 690:6, 693:6, 696:6, 703:6, 709:4, 714:5, 718:5, 723:4, 725:3, 740:4, 744:4, 746:3, 756:3, 759:3, 761:2

| Branch | Coordinate | Condition | Mutation/side effect | Return/error | Executed (baseline) |
|---|---|---|---|---|---|
| B1 | `if at 475:3` | if strategyCrossedTransport \|\| returnedErr == nil \|\| preTransportClaimFinalizationAttempted { | 아래 「State mutations」 | 소스 줄 475 이하 | 실행됨 — 패키지 스위트 count=1 |
| B2 | `if at 478:3` | if preTransportClaim == nil && plan.strategy != nil && !preTransportClaimAttempted && | 아래 「State mutations」 | 소스 줄 478 이하 | 측정 불가(본문 블록 없음) |
| B3 | `if at 482:4` | if claimRejected != nil { | 아래 「State mutations」 | 소스 줄 482 이하 | 미실행 — 패키지 스위트 count=0 |
| B4 | `if at 488:3` | if preTransportClaim == nil { | 아래 「State mutations」 | 소스 줄 488 이하 | 실행됨 — 패키지 스위트 count=1 |
| B5 | `if at 492:3` | if !errors.As(returnedErr, &rejected) { | 아래 「State mutations」 | 소스 줄 492 이하 | 미실행 — 패키지 스위트 count=0 |
| B6 | `if at 506:2` | if err != nil { | 아래 「State mutations」 | 소스 줄 506 이하 | 미실행 — 패키지 스위트 count=0 |
| B7 | `if at 519:2` | if !g.claimSymbol(symbolKey) { | 아래 「State mutations」 | 소스 줄 519 이하 | 실행됨 — 패키지 스위트 count=1 |
| B8 | `if at 525:2` | if rejected, err := g.checkSymbolFree(ctx, plan); err != nil { | 아래 「State mutations」 | 소스 줄 525 이하 | 실행됨 — 패키지 스위트 count=1 |
| B9 | `else at 527:9` | } else if rejected != nil { | 아래 「State mutations」 | 소스 줄 527 이하 | 실행됨 — 패키지 스위트 count=1 |
| B10 | `if at 527:9` | } else if rejected != nil { | 아래 「State mutations」 | 소스 줄 527 이하 | 실행됨 — 패키지 스위트 count=1 |
| B11 | `if at 535:2` | if rejected, err := g.checkDecisionUnspent(ctx, prep, decision); err != nil { | 아래 「State mutations」 | 소스 줄 535 이하 | 미실행 — 패키지 스위트 count=0 |
| B12 | `else at 537:9` | } else if rejected != nil { | 아래 「State mutations」 | 소스 줄 537 이하 | 실행됨 — 패키지 스위트 count=1 |
| B13 | `if at 537:9` | } else if rejected != nil { | 아래 「State mutations」 | 소스 줄 537 이하 | 실행됨 — 패키지 스위트 count=1 |
| B14 | `if at 545:2` | if err != nil { | 아래 「State mutations」 | 소스 줄 545 이하 | 실행됨 — 패키지 스위트 count=1 |
| B15 | `if at 552:3` | if plan.strategy != nil && preTransportClaim == nil && !preTransportClaimAttempted { | 아래 「State mutations」 | 소스 줄 552 이하 | 미실행 — 패키지 스위트 count=0 |
| B16 | `if at 555:4` | if claimRejected != nil { | 아래 「State mutations」 | 소스 줄 555 이하 | 미실행 — 패키지 스위트 count=0 |
| B17 | `if at 561:3` | if plan.strategy != nil && preTransportClaim != nil { | 아래 「State mutations」 | 소스 줄 561 이하 | 미실행 — 패키지 스위트 count=0 |
| B18 | `if at 564:4` | if err := g.refusePreparedClaimedStrategyPreTransport(ctx, attempt, *preTransportClaim, rejected); err != nil { | 아래 「State mutations」 | 소스 줄 564 이하 | 미실행 — 패키지 스위트 count=0 |
| B19 | `if at 583:2` | if rejected != nil { | 아래 「State mutations」 | 소스 줄 583 이하 | 실행됨 — 패키지 스위트 count=1 |
| B20 | `if at 588:2` | if rejected := g.checkEntry(plan); rejected != nil { | 아래 「State mutations」 | 소스 줄 588 이하 | 실행됨 — 패키지 스위트 count=1 |
| B21 | `if at 594:2` | if plan.preflight != nil { | 아래 「State mutations」 | 소스 줄 594 이하 | 실행됨 — 패키지 스위트 count=1 |
| B22 | `if at 595:3` | if rejected := plan.preflight(ctx); rejected != nil { | 아래 「State mutations」 | 소스 줄 595 이하 | 실행됨 — 패키지 스위트 count=1 |
| B23 | `if at 604:2` | if loadRejected != nil { | 아래 「State mutations」 | 소스 줄 604 이하 | 실행됨 — 패키지 스위트 count=1 |
| B24 | `if at 607:2` | if rejected := g.checkDecision(decision, ref, plan, g.clk.Now()); rejected != nil { | 아래 「State mutations」 | 소스 줄 607 이하 | 실행됨 — 패키지 스위트 count=1 |
| B25 | `if at 610:2` | if plan.strategy != nil { | 아래 「State mutations」 | 소스 줄 610 이하 | 미실행 — 패키지 스위트 count=0 |
| B26 | `if at 613:3` | if rejected != nil { | 아래 「State mutations」 | 소스 줄 613 이하 | 미실행 — 패키지 스위트 count=0 |
| B27 | `if at 617:3` | if rejected := g.checkClaimedStrategyLease(ctx, decision, prep.ClientOrderID, plan); rejected != nil { | 아래 「State mutations」 | 소스 줄 617 이하 | 미실행 — 패키지 스위트 count=0 |
| B28 | `if at 625:2` | if rejected := g.checkReservation(ctx, decision); rejected != nil { | 아래 「State mutations」 | 소스 줄 625 이하 | 실행됨 — 패키지 스위트 count=1 |
| B29 | `if at 636:3` | if plan.strategy != nil { | 아래 「State mutations」 | 소스 줄 636 이하 | 미실행 — 패키지 스위트 count=0 |
| B30 | `if at 648:3` | if rejected != nil { | 아래 「State mutations」 | 소스 줄 648 이하 | 미실행 — 패키지 스위트 count=0 |
| B31 | `if at 651:3` | if rejected := g.checkDecision(fresh, ref, plan, now); rejected != nil { | 아래 「State mutations」 | 소스 줄 651 이하 | 실행됨 — 패키지 스위트 count=1 |
| B32 | `if at 654:3` | if fresh.ClientOrderID != prep.ClientOrderID { | 아래 「State mutations」 | 소스 줄 654 이하 | 미실행 — 패키지 스위트 count=0 |
| B33 | `if at 658:3` | if _, rejected := g.checkProtection(dctx, plan, protectionCheckpoint); rejected != nil { | 아래 「State mutations」 | 소스 줄 658 이하 | 실행됨 — 패키지 스위트 count=1 |
| B34 | `if at 665:3` | if rejected := g.checkEntry(plan); rejected != nil { | 아래 「State mutations」 | 소스 줄 665 이하 | 미실행 — 패키지 스위트 count=0 |
| B35 | `if at 672:3` | if plan.strategy == nil { | 아래 「State mutations」 | 소스 줄 672 이하 | 실행됨 — 패키지 스위트 count=1 |
| B36 | `if at 673:4` | if rejected := g.checkReservation(dctx, fresh); rejected != nil { | 아래 「State mutations」 | 소스 줄 673 이하 | 실행됨 — 패키지 스위트 count=1 |
| B37 | `if at 685:4` | if err := dctx.Err(); err != nil { | 아래 「State mutations」 | 소스 줄 685 이하 | 미실행 — 패키지 스위트 count=0 |
| B38 | `if at 688:4` | if plan.strategy != nil { | 아래 「State mutations」 | 소스 줄 688 이하 | 미실행 — 패키지 스위트 count=0 |
| B39 | `if at 689:5` | if plan.strategy.finalAuthorityCheck == nil { | 아래 「State mutations」 | 소스 줄 689 이하 | 미실행 — 패키지 스위트 count=0 |
| B40 | `if at 692:5` | if err := plan.strategy.finalAuthorityCheck(dctx); err != nil { | 아래 「State mutations」 | 소스 줄 692 이하 | 미실행 — 패키지 스위트 count=0 |
| B41 | `if at 695:5` | if rejected := g.checkReservation(dctx, fresh); rejected != nil { | 아래 「State mutations」 | 소스 줄 695 이하 | 미실행 — 패키지 스위트 count=0 |
| B42 | `if at 699:5` | if requireTransport == nil { | 아래 「State mutations」 | 소스 줄 699 이하 | 미실행 — 패키지 스위트 count=0 |
| B43 | `if at 702:5` | if err := requireTransport(dctx, plan.strategy.lease); err != nil { | 아래 「State mutations」 | 소스 줄 702 이하 | 미실행 — 패키지 스위트 count=0 |
| B44 | `if at 711:3` | if plan.strategy != nil && !g.skipStrategyEntryABAForTest { | 아래 「State mutations」 | 소스 줄 711 이하 | 미실행 — 패키지 스위트 count=0 |
| B45 | `else at 716:10` | } else { | 아래 「State mutations」 | 소스 줄 716 이하 | 측정 불가(본문 블록 없음) |
| B46 | `if at 712:4` | if err := g.withStrategyEntryGateAuthority(dctx, plan.strategy.entryGateAuthority, | 아래 「State mutations」 | 소스 줄 712 이하 | 측정 불가(본문 블록 없음) |
| B47 | `if at 717:4` | if err := call(); err != nil { | 아래 「State mutations」 | 소스 줄 717 이하 | 미실행 — 패키지 스위트 count=0 |
| B48 | `if at 722:3` | if tracker == nil { | 아래 「State mutations」 | 소스 줄 722 이하 | 미실행 — 패키지 스위트 count=0 |
| B49 | `if at 728:2` | if plan.strategy == nil { | 아래 「State mutations」 | 소스 줄 728 이하 | 실행됨 — 패키지 스위트 count=1 |
| B50 | `else at 730:9` | } else if g.dispatchStrategyVerified != nil { | 아래 「State mutations」 | 소스 줄 730 이하 | 미실행 — 패키지 스위트 count=0 |
| B51 | `if at 730:9` | } else if g.dispatchStrategyVerified != nil { | 아래 「State mutations」 | 소스 줄 730 이하 | 미실행 — 패키지 스위트 count=0 |
| B52 | `else at 732:9` | } else { | 아래 「State mutations」 | 소스 줄 732 이하 | 측정 불가(본문 블록 없음) |
| B53 | `if at 735:2` | if err != nil { | 아래 「State mutations」 | 소스 줄 735 이하 | 실행됨 — 패키지 스위트 count=1 |
| B54 | `if at 739:3` | if errors.Is(err, journal.ErrNonceSpent) { | 아래 「State mutations」 | 소스 줄 739 이하 | 미실행 — 패키지 스위트 count=0 |
| B55 | `if at 743:3` | if plan.strategy != nil && !strategyCrossedTransport { | 아래 「State mutations」 | 소스 줄 743 이하 | 미실행 — 패키지 스위트 count=0 |
| B56 | `if at 755:2` | if res.Final == journal.StateConfirmed { | 아래 「State mutations」 | 소스 줄 755 이하 | 실행됨 — 패키지 스위트 count=1 |
| B57 | `if at 758:2` | if res.Err != nil { | 아래 「State mutations」 | 소스 줄 758 이하 | 실행됨 — 패키지 스위트 count=1 |

## Calls and live bindings

`plan.call` 은 클로저 `call` 안의 한 자리(708)에서만 불리고, `call` 은 B44(`withStrategyEntryGateAuthority` 의 fn, 712) 또는 B45(717) 중 정확히 한 갈래에서 한 번 불린다. 함수 전체에 for/range/select/go 가 0 이다 — 즉 지금은 attempt 당 브로커 호출이 정확히 1회다. 결과는 725 의 `classifyMutation(tracker.State(), result, callErr)` 한 자리에서 분류된다.

AST 호출 이름 전체: `ReasonCode`, `attempt.DispatchStrategyVerified`, `attempt.DispatchVerified`, `attempt.ID`, `call`, `classifyMutation`, `dctx.Err`, `err.Error`, `errors.As`, `errors.Is`, `errors.Join`, `errors.New`, `fmt.Errorf`, `g.checkClaimedStrategyLease`, `g.checkDecision`, `g.checkDecisionUnspent`, `g.checkEntry`, `g.checkProtection`, `g.checkReservation`, `g.checkSymbolFree`, `g.claimSymbol`, `g.clk.Now`, `g.dispatchStrategyVerified`, `g.journal.Prepare`, `g.loadClaimedStrategyLeaseBinding`, `g.loadDecision`, `g.prepareRequest`, `g.refuseClaimedStrategyPreTransport`, `g.refusePreparedClaimedStrategyPreTransport`, `g.refuseWithStrategySettlement`, `g.releaseSymbol`, `g.roundTripFor`, `g.withStrategyEntryGateAuthority`, `joinStrategyPreTransportRefusal`, `journal.WithSendTracker`, `notSent`, `plan.call`, `plan.preflight`, `plan.strategy.finalAuthorityCheck`, `refusePrepared`, `reject`, `requireTransport`, `strategyFenceRejection`, `strategySubmittingRefusal`, `tracker.State`.

## State mutations and fallbacks

- journal Prepare(RECORDED) · DispatchVerified/DispatchStrategyVerified(DISPATCH_STARTED → 정착) · 심볼 claim/release. 재시도가 생기면 같은 attempt·같은 nonce·같은 clientOrderId 로 브로커 호출이 늘 뿐 journal 행은 늘지 않는다(design D1).

## Safety conclusion

- 편집 경계는 dispatch 클로저(636–726) 안. 클로저 밖 거절 순서(B6–B28)와 정착(B49–B57)은 바꾸지 않는다. High-risk: yes.
