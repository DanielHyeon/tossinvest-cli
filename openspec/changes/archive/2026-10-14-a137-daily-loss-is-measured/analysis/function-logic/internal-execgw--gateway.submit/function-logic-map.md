# Function Logic Map: `Gateway.submit`

- Source: `internal/execgw/gateway.go` (466-762, 편집 뒤)
- Qualified: `Gateway.submit`
- AST evidence: `ast.json`(편집 뒤, `source_sha256` b4b020607839345a…) · `ast-base.json`(편집 전, base `a000cf29`, `source_sha256` 96264a796aa63f4a…, 편집 전 분기수 57) — 현재 branches 57 · returns 44
- Risk scan: `risk-pattern-report.md`

**역할.** 변이 하나의 제출 시퀀스(봉인).

**a137 편집.** B46(712) 펜스 거절을 `notSent(strategyFenceRejection(err))` 로 — 문구 · 사유(ReasonStrategyDispatchFenced)는 같고, 원인이 일일 손실 표지면 그 cause 만 이어 붙임. 분기 수 불변(57), 줄 하나.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| plan · ref | 봉인된 계획 · 결정 | place | 기존 |

## Branches and early returns

| Branch | Condition (소스 줄 인용) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if strategyCrossedTransport \|\| returnedErr == nil \|\| preTransportClaimFinalizationAttempted {` (if 475) | — | 분기 본문 참조 | `go test ./internal/execgw`(기존 제출 시험 전체) |
| B2 | `if preTransportClaim == nil && plan.strategy != nil && !preTransportClaimAttempted &&` (if 478) | — | 분기 본문 참조 | `go test ./internal/execgw`(기존 제출 시험 전체) |
| B3 | `if claimRejected != nil {` (if 482) | — | 분기 본문 참조 | `go test ./internal/execgw`(기존 제출 시험 전체) |
| B4 | `if preTransportClaim == nil {` (if 488) | — | 분기 본문 참조 | `go test ./internal/execgw`(기존 제출 시험 전체) |
| B5 | `if !errors.As(returnedErr, &rejected) {` (if 492) | — | 분기 본문 참조 | `go test ./internal/execgw`(기존 제출 시험 전체) |
| B6 | `if err != nil {` (if 506) | — | 분기 본문 참조 | `go test ./internal/execgw`(기존 제출 시험 전체) |
| B7 | `if !g.claimSymbol(symbolKey) {` (if 519) | — | 분기 본문 참조 | `go test ./internal/execgw`(기존 제출 시험 전체) |
| B8 | `if rejected, err := g.checkSymbolFree(ctx, plan); err != nil {` (if 525) | — | 분기 본문 참조 | `go test ./internal/execgw`(기존 제출 시험 전체) |
| B9 | `} else if rejected != nil {` (else 527) | — | 분기 본문 참조 | `go test ./internal/execgw`(기존 제출 시험 전체) |
| B10 | `} else if rejected != nil {` (if 527) | — | 분기 본문 참조 | `go test ./internal/execgw`(기존 제출 시험 전체) |
| B11 | `if rejected, err := g.checkDecisionUnspent(ctx, prep, decision); err != nil {` (if 535) | — | 분기 본문 참조 | `go test ./internal/execgw`(기존 제출 시험 전체) |
| B12 | `} else if rejected != nil {` (else 537) | — | 분기 본문 참조 | `go test ./internal/execgw`(기존 제출 시험 전체) |
| B13 | `} else if rejected != nil {` (if 537) | — | 분기 본문 참조 | `go test ./internal/execgw`(기존 제출 시험 전체) |
| B14 | `if err != nil {` (if 545) | — | 분기 본문 참조 | `go test ./internal/execgw`(기존 제출 시험 전체) |
| B15 | `if plan.strategy != nil && preTransportClaim == nil && !preTransportClaimAttempted {` (if 552) | — | 분기 본문 참조 | `go test ./internal/execgw`(기존 제출 시험 전체) |
| B16 | `if claimRejected != nil {` (if 555) | — | 분기 본문 참조 | `go test ./internal/execgw`(기존 제출 시험 전체) |
| B17 | `if plan.strategy != nil && preTransportClaim != nil {` (if 561) | — | 분기 본문 참조 | `go test ./internal/execgw`(기존 제출 시험 전체) |
| B18 | `if err := g.refusePreparedClaimedStrategyPreTransport(ctx, attempt, *preTransportClaim, rejected); err != nil {` (if 564) | — | 분기 본문 참조 | `go test ./internal/execgw`(기존 제출 시험 전체) |
| B19 | `if rejected != nil {` (if 583) | — | 분기 본문 참조 | `go test ./internal/execgw`(기존 제출 시험 전체) |
| B20 | `if rejected := g.checkEntry(plan); rejected != nil {` (if 588) | — | 분기 본문 참조 | `TestA137AFenceCausedByTheDailyLossEscalationIsAnAccountBlock`(checkEntry 표지) |
| B21 | `if plan.preflight != nil {` (if 594) | — | 분기 본문 참조 | `go test ./internal/execgw`(기존 제출 시험 전체) |
| B22 | `if rejected := plan.preflight(ctx); rejected != nil {` (if 595) | — | 분기 본문 참조 | `go test ./internal/execgw`(기존 제출 시험 전체) |
| B23 | `if loadRejected != nil {` (if 604) | — | 분기 본문 참조 | `go test ./internal/execgw`(기존 제출 시험 전체) |
| B24 | `if rejected := g.checkDecision(decision, ref, plan, g.clk.Now()); rejected != nil {` (if 607) | — | 분기 본문 참조 | `go test ./internal/execgw`(기존 제출 시험 전체) |
| B25 | `if plan.strategy != nil {` (if 610) | — | 분기 본문 참조 | `go test ./internal/execgw`(기존 제출 시험 전체) |
| B26 | `if rejected != nil {` (if 613) | — | 분기 본문 참조 | `go test ./internal/execgw`(기존 제출 시험 전체) |
| B27 | `if rejected := g.checkClaimedStrategyLease(ctx, decision, prep.ClientOrderID, plan); rejected != nil {` (if 617) | — | 분기 본문 참조 | `go test ./internal/execgw`(기존 제출 시험 전체) |
| B28 | `if rejected := g.checkReservation(ctx, decision); rejected != nil {` (if 625) | — | 분기 본문 참조 | `go test ./internal/execgw`(기존 제출 시험 전체) |
| B29 | `if plan.strategy != nil {` (if 636) | — | 분기 본문 참조 | `go test ./internal/execgw`(기존 제출 시험 전체) |
| B30 | `if rejected != nil {` (if 648) | — | 분기 본문 참조 | `go test ./internal/execgw`(기존 제출 시험 전체) |
| B31 | `if rejected := g.checkDecision(fresh, ref, plan, now); rejected != nil {` (if 651) | — | 분기 본문 참조 | `go test ./internal/execgw`(기존 제출 시험 전체) |
| B32 | `if fresh.ClientOrderID != prep.ClientOrderID {` (if 654) | — | 분기 본문 참조 | `go test ./internal/execgw`(기존 제출 시험 전체) |
| B33 | `if _, rejected := g.checkProtection(dctx, plan, protectionCheckpoint); rejected != nil {` (if 658) | — | 분기 본문 참조 | `go test ./internal/execgw`(기존 제출 시험 전체) |
| B34 | `if rejected := g.checkEntry(plan); rejected != nil {` (if 665) | — | 분기 본문 참조 | `TestA137AFenceCausedByTheDailyLossEscalationIsAnAccountBlock`(checkEntry 표지) |
| B35 | `if plan.strategy == nil {` (if 672) | — | 분기 본문 참조 | `go test ./internal/execgw`(기존 제출 시험 전체) |
| B36 | `if rejected := g.checkReservation(dctx, fresh); rejected != nil {` (if 673) | — | 분기 본문 참조 | `go test ./internal/execgw`(기존 제출 시험 전체) |
| B37 | `if err := dctx.Err(); err != nil {` (if 685) | — | 분기 본문 참조 | `go test ./internal/execgw`(기존 제출 시험 전체) |
| B38 | `if plan.strategy != nil {` (if 688) | — | 분기 본문 참조 | `go test ./internal/execgw`(기존 제출 시험 전체) |
| B39 | `if plan.strategy.finalAuthorityCheck == nil {` (if 689) | — | 분기 본문 참조 | `go test ./internal/execgw`(기존 제출 시험 전체) |
| B40 | `if err := plan.strategy.finalAuthorityCheck(dctx); err != nil {` (if 692) | — | 분기 본문 참조 | `go test ./internal/execgw`(기존 제출 시험 전체) |
| B41 | `if rejected := g.checkReservation(dctx, fresh); rejected != nil {` (if 695) | — | 분기 본문 참조 | `go test ./internal/execgw`(기존 제출 시험 전체) |
| B42 | `if requireTransport == nil {` (if 699) | — | 분기 본문 참조 | `go test ./internal/execgw`(기존 제출 시험 전체) |
| B43 | `if err := requireTransport(dctx, plan.strategy.lease); err != nil {` (if 702) | — | 분기 본문 참조 | `go test ./internal/execgw`(기존 제출 시험 전체) |
| B44 | `if plan.strategy != nil && !g.skipStrategyEntryABAForTest {` (if 711) | — | 분기 본문 참조 | `go test ./internal/execgw`(기존 제출 시험 전체) |
| B45 | `} else {` (else 716) | — | 분기 본문 참조 | `go test ./internal/execgw`(기존 제출 시험 전체) |
| B46 | `if err := g.withStrategyEntryGateAuthority(dctx, plan.strategy.entryGateAuthority,` (if 712) | — | 분기 본문 참조 | `TestA137AFenceCausedByTheDailyLossEscalationIsAnAccountBlock` · `TestA137OtherFencesBehaveAsBefore` · `TestA137TheFinalGateFenceMarksOnlyTheDailyLossChange`(strategyFenceRejection) |
| B47 | `if err := call(); err != nil {` (if 717) | — | 분기 본문 참조 | `go test ./internal/execgw`(기존 제출 시험 전체) |
| B48 | `if tracker == nil {` (if 722) | — | 분기 본문 참조 | `go test ./internal/execgw`(기존 제출 시험 전체) |
| B49 | `if plan.strategy == nil {` (if 728) | — | 분기 본문 참조 | `go test ./internal/execgw`(기존 제출 시험 전체) |
| B50 | `} else if g.dispatchStrategyVerified != nil {` (else 730) | — | 분기 본문 참조 | `go test ./internal/execgw`(기존 제출 시험 전체) |
| B51 | `} else if g.dispatchStrategyVerified != nil {` (if 730) | — | 분기 본문 참조 | `go test ./internal/execgw`(기존 제출 시험 전체) |
| B52 | `} else {` (else 732) | — | 분기 본문 참조 | `go test ./internal/execgw`(기존 제출 시험 전체) |
| B53 | `if err != nil {` (if 735) | — | 분기 본문 참조 | `go test ./internal/execgw`(기존 제출 시험 전체) |
| B54 | `if errors.Is(err, journal.ErrNonceSpent) {` (if 739) | — | 분기 본문 참조 | `go test ./internal/execgw`(기존 제출 시험 전체) |
| B55 | `if plan.strategy != nil && !strategyCrossedTransport {` (if 743) | — | 분기 본문 참조 | `go test ./internal/execgw`(기존 제출 시험 전체) |
| B56 | `if res.Final == journal.StateConfirmed {` (if 755) | — | 분기 본문 참조 | `go test ./internal/execgw`(기존 제출 시험 전체) |
| B57 | `if res.Err != nil {` (if 758) | — | 분기 본문 참조 | `go test ./internal/execgw`(기존 제출 시험 전체) |

Exact AST return positions: 476 · 484 · 489 · 507 · 520 · 526 · 528 · 536 · 538 · 546 · 557 · 565 · 568 · 570 · 584 · 589 · 596 · 605 · 608 · 614 · 618 · 626 · 649 · 652 · 655 · 659 · 666 · 674 · 686 · 690 · 693 · 696 · 703 · 709 · 714 · 718 · 723 · 725 · 740 · 744 · 746 · 756 · 759 · 761

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `strategyFenceRejection` | 펜스 거절(a137) | 표지 이어 붙임 | AST calls |
| `g.withStrategyEntryGateAuthority` | 최종 관문 | 거절 → B46 | AST calls |

## State mutations and fallbacks

- attempt 기록(기존).

## Safety conclusion

- Safe edit boundary: 펜스 거절 값 하나.
- High-risk impact: yes(주문 제출) — 브로커 바이트 0 유지, 매도 경로 무관.
