# Branch Test Map: `Gateway.submit`

- Source: `internal/execgw/gateway.go` (466-762)
- AST branches 57 — 실행 칸은 편집 전 패키지 스위트 coverprofile 측정(시험 이름 귀속은 GREEN 재추출 때)

| Branch | Coordinate | Scenario | Baseline executed | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `if at 475:3` | if strategyCrossedTransport \|\| returnedErr == nil \|\| preTransportClaimFinalizationAttempted { | 실행됨 — 패키지 스위트 count=1 | 해당 없음 — 편집 전 기준선(0.4) | 해당 없음 — 편집 전 기준선(0.4) |
| B2 | `if at 478:3` | if preTransportClaim == nil && plan.strategy != nil && !preTransportClaimAttempted && | 측정 불가(본문 블록 없음) | 해당 없음 — 편집 전 기준선(0.4) | 해당 없음 — 편집 전 기준선(0.4) |
| B3 | `if at 482:4` | if claimRejected != nil { | 미실행 — 패키지 스위트 count=0 | 해당 없음 — 편집 전 기준선(0.4) | 해당 없음 — 편집 전 기준선(0.4) |
| B4 | `if at 488:3` | if preTransportClaim == nil { | 실행됨 — 패키지 스위트 count=1 | 해당 없음 — 편집 전 기준선(0.4) | 해당 없음 — 편집 전 기준선(0.4) |
| B5 | `if at 492:3` | if !errors.As(returnedErr, &rejected) { | 미실행 — 패키지 스위트 count=0 | 해당 없음 — 편집 전 기준선(0.4) | 해당 없음 — 편집 전 기준선(0.4) |
| B6 | `if at 506:2` | if err != nil { | 미실행 — 패키지 스위트 count=0 | 해당 없음 — 편집 전 기준선(0.4) | 해당 없음 — 편집 전 기준선(0.4) |
| B7 | `if at 519:2` | if !g.claimSymbol(symbolKey) { | 실행됨 — 패키지 스위트 count=1 | 해당 없음 — 편집 전 기준선(0.4) | 해당 없음 — 편집 전 기준선(0.4) |
| B8 | `if at 525:2` | if rejected, err := g.checkSymbolFree(ctx, plan); err != nil { | 실행됨 — 패키지 스위트 count=1 | 해당 없음 — 편집 전 기준선(0.4) | 해당 없음 — 편집 전 기준선(0.4) |
| B9 | `else at 527:9` | } else if rejected != nil { | 실행됨 — 패키지 스위트 count=1 | 해당 없음 — 편집 전 기준선(0.4) | 해당 없음 — 편집 전 기준선(0.4) |
| B10 | `if at 527:9` | } else if rejected != nil { | 실행됨 — 패키지 스위트 count=1 | 해당 없음 — 편집 전 기준선(0.4) | 해당 없음 — 편집 전 기준선(0.4) |
| B11 | `if at 535:2` | if rejected, err := g.checkDecisionUnspent(ctx, prep, decision); err != nil { | 미실행 — 패키지 스위트 count=0 | 해당 없음 — 편집 전 기준선(0.4) | 해당 없음 — 편집 전 기준선(0.4) |
| B12 | `else at 537:9` | } else if rejected != nil { | 실행됨 — 패키지 스위트 count=1 | 해당 없음 — 편집 전 기준선(0.4) | 해당 없음 — 편집 전 기준선(0.4) |
| B13 | `if at 537:9` | } else if rejected != nil { | 실행됨 — 패키지 스위트 count=1 | 해당 없음 — 편집 전 기준선(0.4) | 해당 없음 — 편집 전 기준선(0.4) |
| B14 | `if at 545:2` | if err != nil { | 실행됨 — 패키지 스위트 count=1 | 해당 없음 — 편집 전 기준선(0.4) | 해당 없음 — 편집 전 기준선(0.4) |
| B15 | `if at 552:3` | if plan.strategy != nil && preTransportClaim == nil && !preTransportClaimAttempted { | 미실행 — 패키지 스위트 count=0 | 해당 없음 — 편집 전 기준선(0.4) | 해당 없음 — 편집 전 기준선(0.4) |
| B16 | `if at 555:4` | if claimRejected != nil { | 미실행 — 패키지 스위트 count=0 | 해당 없음 — 편집 전 기준선(0.4) | 해당 없음 — 편집 전 기준선(0.4) |
| B17 | `if at 561:3` | if plan.strategy != nil && preTransportClaim != nil { | 미실행 — 패키지 스위트 count=0 | 해당 없음 — 편집 전 기준선(0.4) | 해당 없음 — 편집 전 기준선(0.4) |
| B18 | `if at 564:4` | if err := g.refusePreparedClaimedStrategyPreTransport(ctx, attempt, *preTransportClaim, rejected); err != nil { | 미실행 — 패키지 스위트 count=0 | 해당 없음 — 편집 전 기준선(0.4) | 해당 없음 — 편집 전 기준선(0.4) |
| B19 | `if at 583:2` | if rejected != nil { | 실행됨 — 패키지 스위트 count=1 | 해당 없음 — 편집 전 기준선(0.4) | 해당 없음 — 편집 전 기준선(0.4) |
| B20 | `if at 588:2` | if rejected := g.checkEntry(plan); rejected != nil { | 실행됨 — 패키지 스위트 count=1 | 해당 없음 — 편집 전 기준선(0.4) | 해당 없음 — 편집 전 기준선(0.4) |
| B21 | `if at 594:2` | if plan.preflight != nil { | 실행됨 — 패키지 스위트 count=1 | 해당 없음 — 편집 전 기준선(0.4) | 해당 없음 — 편집 전 기준선(0.4) |
| B22 | `if at 595:3` | if rejected := plan.preflight(ctx); rejected != nil { | 실행됨 — 패키지 스위트 count=1 | 해당 없음 — 편집 전 기준선(0.4) | 해당 없음 — 편집 전 기준선(0.4) |
| B23 | `if at 604:2` | if loadRejected != nil { | 실행됨 — 패키지 스위트 count=1 | 해당 없음 — 편집 전 기준선(0.4) | 해당 없음 — 편집 전 기준선(0.4) |
| B24 | `if at 607:2` | if rejected := g.checkDecision(decision, ref, plan, g.clk.Now()); rejected != nil { | 실행됨 — 패키지 스위트 count=1 | 해당 없음 — 편집 전 기준선(0.4) | 해당 없음 — 편집 전 기준선(0.4) |
| B25 | `if at 610:2` | if plan.strategy != nil { | 미실행 — 패키지 스위트 count=0 | 해당 없음 — 편집 전 기준선(0.4) | 해당 없음 — 편집 전 기준선(0.4) |
| B26 | `if at 613:3` | if rejected != nil { | 미실행 — 패키지 스위트 count=0 | 해당 없음 — 편집 전 기준선(0.4) | 해당 없음 — 편집 전 기준선(0.4) |
| B27 | `if at 617:3` | if rejected := g.checkClaimedStrategyLease(ctx, decision, prep.ClientOrderID, plan); rejected != nil { | 미실행 — 패키지 스위트 count=0 | 해당 없음 — 편집 전 기준선(0.4) | 해당 없음 — 편집 전 기준선(0.4) |
| B28 | `if at 625:2` | if rejected := g.checkReservation(ctx, decision); rejected != nil { | 실행됨 — 패키지 스위트 count=1 | 해당 없음 — 편집 전 기준선(0.4) | 해당 없음 — 편집 전 기준선(0.4) |
| B29 | `if at 636:3` | if plan.strategy != nil { | 미실행 — 패키지 스위트 count=0 | 해당 없음 — 편집 전 기준선(0.4) | 해당 없음 — 편집 전 기준선(0.4) |
| B30 | `if at 648:3` | if rejected != nil { | 미실행 — 패키지 스위트 count=0 | 해당 없음 — 편집 전 기준선(0.4) | 해당 없음 — 편집 전 기준선(0.4) |
| B31 | `if at 651:3` | if rejected := g.checkDecision(fresh, ref, plan, now); rejected != nil { | 실행됨 — 패키지 스위트 count=1 | 해당 없음 — 편집 전 기준선(0.4) | 해당 없음 — 편집 전 기준선(0.4) |
| B32 | `if at 654:3` | if fresh.ClientOrderID != prep.ClientOrderID { | 미실행 — 패키지 스위트 count=0 | 해당 없음 — 편집 전 기준선(0.4) | 해당 없음 — 편집 전 기준선(0.4) |
| B33 | `if at 658:3` | if _, rejected := g.checkProtection(dctx, plan, protectionCheckpoint); rejected != nil { | 실행됨 — 패키지 스위트 count=1 | 해당 없음 — 편집 전 기준선(0.4) | 해당 없음 — 편집 전 기준선(0.4) |
| B34 | `if at 665:3` | if rejected := g.checkEntry(plan); rejected != nil { | 미실행 — 패키지 스위트 count=0 | 해당 없음 — 편집 전 기준선(0.4) | 해당 없음 — 편집 전 기준선(0.4) |
| B35 | `if at 672:3` | if plan.strategy == nil { | 실행됨 — 패키지 스위트 count=1 | 해당 없음 — 편집 전 기준선(0.4) | 해당 없음 — 편집 전 기준선(0.4) |
| B36 | `if at 673:4` | if rejected := g.checkReservation(dctx, fresh); rejected != nil { | 실행됨 — 패키지 스위트 count=1 | 해당 없음 — 편집 전 기준선(0.4) | 해당 없음 — 편집 전 기준선(0.4) |
| B37 | `if at 685:4` | if err := dctx.Err(); err != nil { | 미실행 — 패키지 스위트 count=0 | 계획(D10 w2·w3·w4): 재호출마다 B37 이 `winCtx` 로 다시 섬 — 호출자 취소·창 마감·`cancel()` 뒤 재전송 0 | GREEN 재추출 때 |
| B38 | `if at 688:4` | if plan.strategy != nil { | 미실행 — 패키지 스위트 count=0 | 해당 없음 — 편집 전 기준선(0.4) | 해당 없음 — 편집 전 기준선(0.4) |
| B39 | `if at 689:5` | if plan.strategy.finalAuthorityCheck == nil { | 미실행 — 패키지 스위트 count=0 | 해당 없음 — 편집 전 기준선(0.4) | 해당 없음 — 편집 전 기준선(0.4) |
| B40 | `if at 692:5` | if err := plan.strategy.finalAuthorityCheck(dctx); err != nil { | 미실행 — 패키지 스위트 count=0 | 해당 없음 — 편집 전 기준선(0.4) | 해당 없음 — 편집 전 기준선(0.4) |
| B41 | `if at 695:5` | if rejected := g.checkReservation(dctx, fresh); rejected != nil { | 미실행 — 패키지 스위트 count=0 | 해당 없음 — 편집 전 기준선(0.4) | 해당 없음 — 편집 전 기준선(0.4) |
| B42 | `if at 699:5` | if requireTransport == nil { | 미실행 — 패키지 스위트 count=0 | 해당 없음 — 편집 전 기준선(0.4) | 해당 없음 — 편집 전 기준선(0.4) |
| B43 | `if at 702:5` | if err := requireTransport(dctx, plan.strategy.lease); err != nil { | 미실행 — 패키지 스위트 count=0 | 해당 없음 — 편집 전 기준선(0.4) | 해당 없음 — 편집 전 기준선(0.4) |
| B44 | `if at 711:3` | if plan.strategy != nil && !g.skipStrategyEntryABAForTest { | 미실행 — 패키지 스위트 count=0 | 해당 없음 — 편집 전 기준선(0.4) | 해당 없음 — 편집 전 기준선(0.4) |
| B45 | `else at 716:10` | } else { | 측정 불가(본문 블록 없음) | 해당 없음 — 편집 전 기준선(0.4) | 해당 없음 — 편집 전 기준선(0.4) |
| B46 | `if at 712:4` | if err := g.withStrategyEntryGateAuthority(dctx, plan.strategy.entryGateAuthority, | 측정 불가(본문 블록 없음) | 해당 없음 — 편집 전 기준선(0.4) | 해당 없음 — 편집 전 기준선(0.4) |
| B47 | `if at 717:4` | if err := call(); err != nil { | 미실행 — 패키지 스위트 count=0 | 계획(D10 w4): 재호출의 `call()` 오류 → `notSent` → 래퍼가 첫 409 AMBIGUOUS 로 정착(Failed 아님) | GREEN 재추출 때 |
| B48 | `if at 722:3` | if tracker == nil { | 미실행 — 패키지 스위트 count=0 | 해당 없음 — 편집 전 기준선(0.4) | 해당 없음 — 편집 전 기준선(0.4) |
| B49 | `if at 728:2` | if plan.strategy == nil { | 실행됨 — 패키지 스위트 count=1 | 계획(D10 w5·W3·W4·변이 ctx②): 유일한 편집 자리 — 정책 nil·비취소는 원 `dispatch` 그대로(바이트 동일 골든), `DispatchVerified` 첫 인자 = `ctx`(AST 구조 시험) | GREEN 재추출 때 |
| B50 | `else at 730:9` | } else if g.dispatchStrategyVerified != nil { | 미실행 — 패키지 스위트 count=0 | 계획: 편집 0 — strategy 갈래는 래퍼 없음(취소는 strategy 없음) | GREEN 재추출 때 |
| B51 | `if at 730:9` | } else if g.dispatchStrategyVerified != nil { | 미실행 — 패키지 스위트 count=0 | 해당 없음 — 편집 전 기준선(0.4) | 해당 없음 — 편집 전 기준선(0.4) |
| B52 | `else at 732:9` | } else { | 측정 불가(본문 블록 없음) | 계획: 편집 0 — strategy 갈래는 래퍼 없음 | GREEN 재추출 때 |
| B53 | `if at 735:2` | if err != nil { | 실행됨 — 패키지 스위트 count=1 | 해당 없음 — 편집 전 기준선(0.4) | 해당 없음 — 편집 전 기준선(0.4) |
| B54 | `if at 739:3` | if errors.Is(err, journal.ErrNonceSpent) { | 미실행 — 패키지 스위트 count=0 | 해당 없음 — 편집 전 기준선(0.4) | 해당 없음 — 편집 전 기준선(0.4) |
| B55 | `if at 743:3` | if plan.strategy != nil && !strategyCrossedTransport { | 미실행 — 패키지 스위트 count=0 | 해당 없음 — 편집 전 기준선(0.4) | 해당 없음 — 편집 전 기준선(0.4) |
| B56 | `if at 755:2` | if res.Final == journal.StateConfirmed { | 실행됨 — 패키지 스위트 count=1 | 해당 없음 — 편집 전 기준선(0.4) | 해당 없음 — 편집 전 기준선(0.4) |
| B57 | `if at 758:2` | if res.Err != nil { | 실행됨 — 패키지 스위트 count=1 | 해당 없음 — 편집 전 기준선(0.4) | 해당 없음 — 편집 전 기준선(0.4) |
