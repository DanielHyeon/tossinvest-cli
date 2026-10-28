# Function Logic Map: `Gateway.submit`

- Source: `internal/execgw/gateway.go` (470-766)
- Qualified function: `Gateway.submit`
- Revision: `current` (a165 base `34e12967e`, `source_sha256` c304e32a62e42759…; GREEN 뒤 AST 재추출 2026-10-29)
- AST evidence: `ast.json` — AST branches 57, 반환 44, 호출 89
- Risk scan: `risk-pattern-report.md`
- 실행 칸 측정: 편집 전 기준선 `go test -count=1 -coverprofile` (execgw·official·trading·flatten 패키지 전체, journal `-run 'Classify|SendTracker|Dispatch'`, app/engine `-run 'Clear|Cancel'` — journal·engine 은 부분 실행이라 count=0 이 「미실행」을 뜻하지 않을 수 있음, cmd/tossctl 은 미측정) — 분기 줄에서 시작하는 본문 블록의 실행 수. 시험 이름별 귀속은 GREEN 재추출에서 함.

## 역할 (a165)

게이트웨이 순서(결정 읽기 → Prepare → 거절 → DISPATCH_STARTED → 직전 재검증 → 브로커 1회 → 정착)의 본체. a165 는 dispatch 클로저를 고치지 않고, B49 갈래가 `DispatchVerified` 에 넘기는 함수만 바깥 래퍼로 감싼다(design D5, 사람 결정 P2-2).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| plan mutationPlan | place/cancel/amend 계획 | Gateway.place·Cancel·Amend | prepareRequest 오류면 기록 없이 반환 |
| ref GuardianDecision | 결정 id | journal | 읽기 실패는 refusePrepared 로 기록 |
| plan.strategy | nil 또는 paired KR/US first-leg 능력 | PlaceClaimedStrategy | nil 이면 일반 경로 |

## Branches and early returns

AST 열거 순서 그대로. Condition 칸은 AST 좌표 줄의 원문.

Exact AST return positions: 480:4, 488:5, 493:4, 511:3, 524:3, 530:3, 532:3, 540:3, 542:3, 550:3, 561:5, 569:5, 572:4, 574:3, 588:3, 593:3, 600:4, 609:3, 612:3, 618:4, 622:4, 630:3, 653:4, 656:4, 659:4, 663:4, 670:4, 678:5, 690:5, 694:6, 697:6, 700:6, 707:6, 713:4, 718:5, 722:5, 727:4, 729:3, 744:4, 748:4, 750:3, 760:3, 763:3, 765:2

| Branch | Coordinate | Condition | Mutation/side effect | Return/error | Executed (baseline) |
|---|---|---|---|---|---|
| B1 | `if at 479:3` | if strategyCrossedTransport \|\| returnedErr == nil \|\| preTransportClaimFinalizationAttempted { | 아래 「State mutations」 | 소스 줄 475 이하 | 실행됨 — 패키지 스위트 count=1 |
| B2 | `if at 482:3` | if preTransportClaim == nil && plan.strategy != nil && !preTransportClaimAttempted && | 아래 「State mutations」 | 소스 줄 478 이하 | 측정 불가(본문 블록 없음) |
| B3 | `if at 486:4` | if claimRejected != nil { | 아래 「State mutations」 | 소스 줄 482 이하 | 미실행 — 패키지 스위트 count=0 |
| B4 | `if at 492:3` | if preTransportClaim == nil { | 아래 「State mutations」 | 소스 줄 488 이하 | 실행됨 — 패키지 스위트 count=1 |
| B5 | `if at 496:3` | if !errors.As(returnedErr, &rejected) { | 아래 「State mutations」 | 소스 줄 492 이하 | 미실행 — 패키지 스위트 count=0 |
| B6 | `if at 510:2` | if err != nil { | 아래 「State mutations」 | 소스 줄 506 이하 | 미실행 — 패키지 스위트 count=0 |
| B7 | `if at 523:2` | if !g.claimSymbol(symbolKey) { | 아래 「State mutations」 | 소스 줄 519 이하 | 실행됨 — 패키지 스위트 count=1 |
| B8 | `if at 529:2` | if rejected, err := g.checkSymbolFree(ctx, plan); err != nil { | 아래 「State mutations」 | 소스 줄 525 이하 | 실행됨 — 패키지 스위트 count=1 |
| B9 | `else at 531:9` | } else if rejected != nil { | 아래 「State mutations」 | 소스 줄 527 이하 | 실행됨 — 패키지 스위트 count=1 |
| B10 | `if at 531:9` | } else if rejected != nil { | 아래 「State mutations」 | 소스 줄 527 이하 | 실행됨 — 패키지 스위트 count=1 |
| B11 | `if at 539:2` | if rejected, err := g.checkDecisionUnspent(ctx, prep, decision); err != nil { | 아래 「State mutations」 | 소스 줄 535 이하 | 미실행 — 패키지 스위트 count=0 |
| B12 | `else at 541:9` | } else if rejected != nil { | 아래 「State mutations」 | 소스 줄 537 이하 | 실행됨 — 패키지 스위트 count=1 |
| B13 | `if at 541:9` | } else if rejected != nil { | 아래 「State mutations」 | 소스 줄 537 이하 | 실행됨 — 패키지 스위트 count=1 |
| B14 | `if at 549:2` | if err != nil { | 아래 「State mutations」 | 소스 줄 545 이하 | 실행됨 — 패키지 스위트 count=1 |
| B15 | `if at 556:3` | if plan.strategy != nil && preTransportClaim == nil && !preTransportClaimAttempted { | 아래 「State mutations」 | 소스 줄 552 이하 | 미실행 — 패키지 스위트 count=0 |
| B16 | `if at 559:4` | if claimRejected != nil { | 아래 「State mutations」 | 소스 줄 555 이하 | 미실행 — 패키지 스위트 count=0 |
| B17 | `if at 565:3` | if plan.strategy != nil && preTransportClaim != nil { | 아래 「State mutations」 | 소스 줄 561 이하 | 미실행 — 패키지 스위트 count=0 |
| B18 | `if at 568:4` | if err := g.refusePreparedClaimedStrategyPreTransport(ctx, attempt, *preTransportClaim, rejected); err != nil { | 아래 「State mutations」 | 소스 줄 564 이하 | 미실행 — 패키지 스위트 count=0 |
| B19 | `if at 587:2` | if rejected != nil { | 아래 「State mutations」 | 소스 줄 583 이하 | 실행됨 — 패키지 스위트 count=1 |
| B20 | `if at 592:2` | if rejected := g.checkEntry(plan); rejected != nil { | 아래 「State mutations」 | 소스 줄 588 이하 | 실행됨 — 패키지 스위트 count=1 |
| B21 | `if at 598:2` | if plan.preflight != nil { | 아래 「State mutations」 | 소스 줄 594 이하 | 실행됨 — 패키지 스위트 count=1 |
| B22 | `if at 599:3` | if rejected := plan.preflight(ctx); rejected != nil { | 아래 「State mutations」 | 소스 줄 595 이하 | 실행됨 — 패키지 스위트 count=1 |
| B23 | `if at 608:2` | if loadRejected != nil { | 아래 「State mutations」 | 소스 줄 604 이하 | 실행됨 — 패키지 스위트 count=1 |
| B24 | `if at 611:2` | if rejected := g.checkDecision(decision, ref, plan, g.clk.Now()); rejected != nil { | 아래 「State mutations」 | 소스 줄 607 이하 | 실행됨 — 패키지 스위트 count=1 |
| B25 | `if at 614:2` | if plan.strategy != nil { | 아래 「State mutations」 | 소스 줄 610 이하 | 미실행 — 패키지 스위트 count=0 |
| B26 | `if at 617:3` | if rejected != nil { | 아래 「State mutations」 | 소스 줄 613 이하 | 미실행 — 패키지 스위트 count=0 |
| B27 | `if at 621:3` | if rejected := g.checkClaimedStrategyLease(ctx, decision, prep.ClientOrderID, plan); rejected != nil { | 아래 「State mutations」 | 소스 줄 617 이하 | 미실행 — 패키지 스위트 count=0 |
| B28 | `if at 629:2` | if rejected := g.checkReservation(ctx, decision); rejected != nil { | 아래 「State mutations」 | 소스 줄 625 이하 | 실행됨 — 패키지 스위트 count=1 |
| B29 | `if at 640:3` | if plan.strategy != nil { | 아래 「State mutations」 | 소스 줄 636 이하 | 미실행 — 패키지 스위트 count=0 |
| B30 | `if at 652:3` | if rejected != nil { | 아래 「State mutations」 | 소스 줄 648 이하 | 미실행 — 패키지 스위트 count=0 |
| B31 | `if at 655:3` | if rejected := g.checkDecision(fresh, ref, plan, now); rejected != nil { | 아래 「State mutations」 | 소스 줄 651 이하 | 실행됨 — 패키지 스위트 count=1 |
| B32 | `if at 658:3` | if fresh.ClientOrderID != prep.ClientOrderID { | 아래 「State mutations」 | 소스 줄 654 이하 | 미실행 — 패키지 스위트 count=0 |
| B33 | `if at 662:3` | if _, rejected := g.checkProtection(dctx, plan, protectionCheckpoint); rejected != nil { | 아래 「State mutations」 | 소스 줄 658 이하 | 실행됨 — 패키지 스위트 count=1 |
| B34 | `if at 669:3` | if rejected := g.checkEntry(plan); rejected != nil { | 아래 「State mutations」 | 소스 줄 665 이하 | 미실행 — 패키지 스위트 count=0 |
| B35 | `if at 676:3` | if plan.strategy == nil { | 아래 「State mutations」 | 소스 줄 672 이하 | 실행됨 — 패키지 스위트 count=1 |
| B36 | `if at 677:4` | if rejected := g.checkReservation(dctx, fresh); rejected != nil { | 아래 「State mutations」 | 소스 줄 673 이하 | 실행됨 — 패키지 스위트 count=1 |
| B37 | `if at 689:4` | if err := dctx.Err(); err != nil { | 아래 「State mutations」 | 소스 줄 685 이하 | 미실행 — 패키지 스위트 count=0 |
| B38 | `if at 692:4` | if plan.strategy != nil { | 아래 「State mutations」 | 소스 줄 688 이하 | 미실행 — 패키지 스위트 count=0 |
| B39 | `if at 693:5` | if plan.strategy.finalAuthorityCheck == nil { | 아래 「State mutations」 | 소스 줄 689 이하 | 미실행 — 패키지 스위트 count=0 |
| B40 | `if at 696:5` | if err := plan.strategy.finalAuthorityCheck(dctx); err != nil { | 아래 「State mutations」 | 소스 줄 692 이하 | 미실행 — 패키지 스위트 count=0 |
| B41 | `if at 699:5` | if rejected := g.checkReservation(dctx, fresh); rejected != nil { | 아래 「State mutations」 | 소스 줄 695 이하 | 미실행 — 패키지 스위트 count=0 |
| B42 | `if at 703:5` | if requireTransport == nil { | 아래 「State mutations」 | 소스 줄 699 이하 | 미실행 — 패키지 스위트 count=0 |
| B43 | `if at 706:5` | if err := requireTransport(dctx, plan.strategy.lease); err != nil { | 아래 「State mutations」 | 소스 줄 702 이하 | 미실행 — 패키지 스위트 count=0 |
| B44 | `if at 715:3` | if plan.strategy != nil && !g.skipStrategyEntryABAForTest { | 아래 「State mutations」 | 소스 줄 711 이하 | 미실행 — 패키지 스위트 count=0 |
| B45 | `else at 720:10` | } else { | 아래 「State mutations」 | 소스 줄 716 이하 | 측정 불가(본문 블록 없음) |
| B46 | `if at 716:4` | if err := g.withStrategyEntryGateAuthority(dctx, plan.strategy.entryGateAuthority, | 아래 「State mutations」 | 소스 줄 712 이하 | 측정 불가(본문 블록 없음) |
| B47 | `if at 721:4` | if err := call(); err != nil { | 아래 「State mutations」 | 소스 줄 717 이하 | 미실행 — 패키지 스위트 count=0 |
| B48 | `if at 726:3` | if tracker == nil { | 아래 「State mutations」 | 소스 줄 722 이하 | 미실행 — 패키지 스위트 count=0 |
| B49 | `if at 732:2` | if plan.strategy == nil { | 아래 「State mutations」 | 소스 줄 732 이하 | A165 GREEN 실행됨 |
| B50 | `else at 734:9` | } else if g.dispatchStrategyVerified != nil { | 아래 「State mutations」 | 소스 줄 730 이하 | 미실행 — 패키지 스위트 count=0 |
| B51 | `if at 734:9` | } else if g.dispatchStrategyVerified != nil { | 아래 「State mutations」 | 소스 줄 730 이하 | 미실행 — 패키지 스위트 count=0 |
| B52 | `else at 736:9` | } else { | 아래 「State mutations」 | 소스 줄 732 이하 | 측정 불가(본문 블록 없음) |
| B53 | `if at 739:2` | if err != nil { | 아래 「State mutations」 | 소스 줄 735 이하 | 실행됨 — 패키지 스위트 count=1 |
| B54 | `if at 743:3` | if errors.Is(err, journal.ErrNonceSpent) { | 아래 「State mutations」 | 소스 줄 739 이하 | 미실행 — 패키지 스위트 count=0 |
| B55 | `if at 747:3` | if plan.strategy != nil && !strategyCrossedTransport { | 아래 「State mutations」 | 소스 줄 743 이하 | 미실행 — 패키지 스위트 count=0 |
| B56 | `if at 759:2` | if res.Final == journal.StateConfirmed { | 아래 「State mutations」 | 소스 줄 755 이하 | 실행됨 — 패키지 스위트 count=1 |
| B57 | `if at 762:2` | if res.Err != nil { | 아래 「State mutations」 | 소스 줄 758 이하 | 실행됨 — 패키지 스위트 count=1 |

## Calls and live bindings

`plan.call` 은 클로저 `call` 안의 한 자리(708)에서만 불리고, `call` 은 B44(`withStrategyEntryGateAuthority` 의 fn, 712) 또는 B45(717) 중 정확히 한 갈래에서 한 번 불린다. 함수 전체에 for/range/select/go 가 0 이다 — 즉 지금은 attempt 당 브로커 호출이 정확히 1회다. 결과는 725 의 `classifyMutation(tracker.State(), result, callErr)` 한 자리에서 분류된다.

AST 호출 이름 전체: `ReasonCode`, `attempt.DispatchStrategyVerified`, `attempt.DispatchVerified`, `attempt.ID`, `call`, `classifyMutation`, `dctx.Err`, `err.Error`, `errors.As`, `errors.Is`, `errors.Join`, `errors.New`, `fmt.Errorf`, `g.checkClaimedStrategyLease`, `g.checkDecision`, `g.checkDecisionUnspent`, `g.checkEntry`, `g.checkProtection`, `g.checkReservation`, `g.checkSymbolFree`, `g.claimSymbol`, `g.clk.Now`, `g.dispatchStrategyVerified`, `g.journal.Prepare`, `g.loadClaimedStrategyLeaseBinding`, `g.loadDecision`, `g.prepareRequest`, `g.refuseClaimedStrategyPreTransport`, `g.refusePreparedClaimedStrategyPreTransport`, `g.refuseWithStrategySettlement`, `g.releaseSymbol`, `g.roundTripFor`, `g.withStrategyEntryGateAuthority`, `joinStrategyPreTransportRefusal`, `journal.WithSendTracker`, `notSent`, `plan.call`, `plan.preflight`, `plan.strategy.finalAuthorityCheck`, `refusePrepared`, `reject`, `requireTransport`, `strategyFenceRejection`, `strategySubmittingRefusal`, `tracker.State`.

## State mutations and fallbacks

- journal Prepare(RECORDED) · DispatchVerified/DispatchStrategyVerified(DISPATCH_STARTED → 정착) · 심볼 claim/release. 재시도가 생기면 같은 attempt·같은 nonce·같은 clientOrderId 로 브로커 호출이 늘 뿐 journal 행은 늘지 않는다(design D1).

## a165 편집 계획 (3판 보완 — 사람 결정 P2-2 바깥 래퍼, design D5)

| 위치 | 오늘 | a165 뒤 | 근거(AST) |
|---|---|---|---|
| B29–B48 dispatch 클로저(636–726) | 결정 재읽기 B30·checkDecision B31·키 B32·protection B33·entry B34·reservation B36 → `call` 의 B37(685) `dctx.Err()` → `plan.call`(708) 1회 → `classifyMutation`(725) | **편집 0.** 정책이 있는 취소에서만 래퍼가 같은 클로저 값을 `winCtx` 로 최대 2회 더 부른다 | 클로저 본문 그대로 |
| B49(732) `if plan.strategy == nil` → 733 | `attempt.DispatchVerified(ctx, dispatch, g.roundTripFor(plan))` | `attempt.DispatchVerified(ctx, g.withCancelRefusal(plan, dispatch), g.roundTripFor(plan))` — 유일한 편집 | 취소 계획은 strategy nil(`Gateway.Cancel` 408–421, 손으로 읽음) → 취소는 이 갈래로만 옴 |
| B50/B51(730–731)·B52(732–733) | strategy 정착 | **편집 0**(`dispatch` 그대로) | strategy 는 PlaceClaimedStrategy 만 |
| B53–B57 정착 뒤 반환 | — | 편집 0 | — |

- **B37 의 재적용**: 재호출은 클로저 전체를 다시 지나므로 각 재전송의 `plan.call` 바로 앞에 B37 이 다시 선다. 그 `dctx` 는 `winCtx`(=`context.WithDeadline(dctx, wallStart+8s)`, 래퍼가 `defer cancel()`)라
  창 마감·cancel·호출자 취소가 여기서 걸려 B47(717) `notSent` 가 된다.
- **정착 ctx 분리**: 정착은 `DispatchVerified` 가 자기 `ctx` 로 한다(번들 `internal-journal--attempt.dispatchverified` B5·B7·B9·B11–B13·B15·B17·B19 전부 `ctx`). 래퍼는 `DispatchOutcome` 값만 돌려주므로 `winCtx` 는 정착에 닿지 않는다.
- **정책 nil 동일성**: `withCancelRefusal` 은 정책 nil 이거나 `plan.kind != KindCancel` 이면 받은 `dispatch` 를 그대로 돌려준다 — B49 의 인자 값이 오늘과 같고 B29–B48·B50–B57 편집 0.
- **창 시작점(P3)**: 래퍼가 원 클로저를 처음 부르기 **전**에 잰다(첫 전송의 B30–B36 시간이 창에 듦 — 보수적).
- GREEN 뒤 이 번들의 `ast.json` 을 다시 뽑음(B49 갈래의 호출 식 변화, 분기 수 57 유지 확인).
- 이 표와 BTM 의 B37·B47·B49·B50·B52 행 「RED observed」 계획 칸은 1.3 2차 처분(2026-10-28)에서 손으로 적었다 — `measure/gen_flm.py` 로 재생성하면 지워지므로 재생성 뒤 다시 옮길 것.

## Safety conclusion

- a165 편집 경계(3판 보완, P2-2): dispatch 클로저(B29–B48, 636–726)·클로저 밖 거절 순서(B6–B28)·정착(B53–B57)은 바꾸지 않는다. 편집은 B49 갈래 729 의 인자 식 하나(`dispatch` → `g.withCancelRefusal(plan, dispatch)`)뿐이고, 정책 nil 이거나 취소가 아니면 그 식의 값이 원 `dispatch` 다. 3판의 `sendOnce` 추출은 폐기. High-risk: yes.
