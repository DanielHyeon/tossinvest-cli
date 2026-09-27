# Branch Test Map: Gateway.submit

- Source: `internal/execgw/gateway.go` (466-763); current — HEAD `648df8ef`
- 시험 칸은 측정값: 무태그 `go test -c -cover` 바이너리를 시험 함수마다 따로 돌린 커버 프로필에서 그 분기 본문 블록을 실행한 시험(`analysis/harness/51_matrix.sh`). `_test.go` 함수와 base 함수는 계측 밖이라 그 사실을 행에 적음.

| Branch | AST anchor | Scenario (source text) | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | if at 475:3 | `if strategyCrossedTransport \|\| returnedErr == nil \|\| preTransportClaimFinalizationAttempted {` | `TestKRDriftDoesNotBlockValidUSAndDispatchDriftCallsNoBroker` · `TestReductionNeverReadsReadinessProvider` (+29) | 5.1 에서 재실행 안 함 | 블록 475.95-477.4 을 시험 31개가 실행, 전부 PASS |
| B2 | if at 478:3 | `if preTransportClaim == nil && plan.strategy != nil && !preTransportClaimAttempted &&` | 없음 | 5.1 에서 재실행 안 함 | 블록 479.50-482.28 을 무태그 패키지 시험 어느 것도 실행하지 않음 |
| B3 | if at 482:4 | `if claimRejected != nil {` | 없음 | 5.1 에서 재실행 안 함 | 블록 482.28-485.5 을 무태그 패키지 시험 어느 것도 실행하지 않음 |
| B4 | if at 488:3 | `if preTransportClaim == nil {` | `TestKRDriftDoesNotBlockValidUSAndDispatchDriftCallsNoBroker` · `TestNonIntegralOrUnsafeProtectionQuantityStopsBeforeProviderAndBroker` (+85) | 5.1 에서 재실행 안 함 | 블록 488.31-490.4 을 시험 87개가 실행, 전부 PASS |
| B5 | if at 492:3 | `if !errors.As(returnedErr, &rejected) {` | 없음 | 5.1 에서 재실행 안 함 | 블록 492.41-494.4 을 무태그 패키지 시험 어느 것도 실행하지 않음 |
| B6 | if at 506:2 | `if err != nil {` | 없음 | 5.1 에서 재실행 안 함 | 블록 506.16-508.3 을 무태그 패키지 시험 어느 것도 실행하지 않음 |
| B7 | if at 519:2 | `if !g.claimSymbol(symbolKey) {` | `TestConcurrentMutationsOnOneSymbolSerialise` · `TestNonceIsOneShotUnderRace` | 5.1 에서 재실행 안 함 | 블록 519.31-522.3 을 시험 2개가 실행, 전부 PASS |
| B8 | if at 525:2 | `if rejected, err := g.checkSymbolFree(ctx, plan); err != nil {` | `TestJournalFailureBlocksSubmission` | 5.1 에서 재실행 안 함 | 블록 525.63-527.3 을 시험 1개가 실행, 전부 PASS |
| B9 | else at 527:9 | `} else if rejected != nil {` | — | 5.1 에서 재실행 안 함 | 같은 좌표의 if 분기(다음 행)가 평가된다는 것이 곧 이 else 로 들어옴 — 그 행의 측정을 볼 것 |
| B10 | if at 527:9 | `} else if rejected != nil {` | — | 5.1 에서 재실행 안 함 | 같은 좌표의 if 분기(다음 행)가 평가된다는 것이 곧 이 else 로 들어옴 — 그 행의 측정을 볼 것 |
| B11 | if at 535:2 | `if rejected, err := g.checkDecisionUnspent(ctx, prep, decision); err != nil {` | 없음 | 5.1 에서 재실행 안 함 | 블록 535.78-537.3 을 무태그 패키지 시험 어느 것도 실행하지 않음 |
| B12 | else at 537:9 | `} else if rejected != nil {` | — | 5.1 에서 재실행 안 함 | 같은 좌표의 if 분기(다음 행)가 평가된다는 것이 곧 이 else 로 들어옴 — 그 행의 측정을 볼 것 |
| B13 | if at 537:9 | `} else if rejected != nil {` | — | 5.1 에서 재실행 안 함 | 같은 좌표의 if 분기(다음 행)가 평가된다는 것이 곧 이 else 로 들어옴 — 그 행의 측정을 볼 것 |
| B14 | if at 545:2 | `if err != nil {` | `TestAReusedIntentIDIsARecoveryAndNotASecondIntent` | 5.1 에서 재실행 안 함 | 블록 545.16-548.3 을 시험 1개가 실행, 전부 PASS |
| B15 | if at 552:3 | `if plan.strategy != nil && preTransportClaim == nil && !preTransportClaimAttempted {` | 없음 | 5.1 에서 재실행 안 함 | 블록 552.86-555.28 을 무태그 패키지 시험 어느 것도 실행하지 않음 |
| B16 | if at 555:4 | `if claimRejected != nil {` | 없음 | 5.1 에서 재실행 안 함 | 블록 555.28-558.5 을 무태그 패키지 시험 어느 것도 실행하지 않음 |
| B17 | if at 561:3 | `if plan.strategy != nil && preTransportClaim != nil {` | 없음 | 5.1 에서 재실행 안 함 | 블록 561.55-564.114 을 무태그 패키지 시험 어느 것도 실행하지 않음 |
| B18 | if at 564:4 | `if err := g.refusePreparedClaimedStrategyPreTransport(ctx, attempt, *preTransportClaim, rejected); err != n...` | 없음 | 5.1 에서 재실행 안 함 | 블록 564.114-567.5 을 무태그 패키지 시험 어느 것도 실행하지 않음 |
| B19 | if at 583:2 | `if rejected != nil {` | `TestKRDriftDoesNotBlockValidUSAndDispatchDriftCallsNoBroker` · `TestNonIntegralOrUnsafeProtectionQuantityStopsBeforeProviderAndBroker` (+2) | 5.1 에서 재실행 안 함 | 블록 583.21-585.3 을 시험 4개가 실행, 전부 PASS |
| B20 | if at 588:2 | `if rejected := g.checkEntry(plan); rejected != nil {` | `TestATighteningBetweenIssuanceAndSubmissionIsRefusedAtTheGateway` · `TestAnAutomaticTighteningRefusesTheNextPlace` (+3) | 5.1 에서 재실행 안 함 | 블록 588.53-590.3 을 시험 5개가 실행, 전부 PASS |
| B21 | if at 594:2 | `if plan.preflight != nil {` | `TestExpiryIsRecheckedOnAFreshRowBeforeTheBrokerCall` · `TestGatewayRefusesFailClosedBranchesBeforeDispatch` (+1) | 5.1 에서 재실행 안 함 | 블록 594.27-595.55 을 시험 3개가 실행, 전부 PASS |
| B22 | if at 595:3 | `if rejected := plan.preflight(ctx); rejected != nil {` | `TestGatewayRefusesFailClosedBranchesBeforeDispatch` | 5.1 에서 재실행 안 함 | 블록 595.55-597.4 을 시험 1개가 실행, 전부 PASS |
| B23 | if at 604:2 | `if loadRejected != nil {` | `TestGuardianRefusalsNeverReachBroker` | 5.1 에서 재실행 안 함 | 블록 604.25-606.3 을 시험 1개가 실행, 전부 PASS |
| B24 | if at 607:2 | `if rejected := g.checkDecision(decision, ref, plan, g.clk.Now()); rejected != nil {` | `TestAmendRaisingQuantityIsLimitChecked` · `TestCancelWithAnEntryDecisionIsRefused` (+8) | 5.1 에서 재실행 안 함 | 블록 607.84-609.3 을 시험 10개가 실행, 전부 PASS |
| B25 | if at 610:2 | `if plan.strategy != nil {` | 없음 | 5.1 에서 재실행 안 함 | 블록 610.26-613.22 을 무태그 패키지 시험 어느 것도 실행하지 않음 |
| B26 | if at 613:3 | `if rejected != nil {` | 없음 | 5.1 에서 재실행 안 함 | 블록 613.22-615.4 을 무태그 패키지 시험 어느 것도 실행하지 않음 |
| B27 | if at 617:3 | `if rejected := g.checkClaimedStrategyLease(ctx, decision, prep.ClientOrderID, plan); rejected != nil {` | 없음 | 5.1 에서 재실행 안 함 | 블록 617.104-619.4 을 무태그 패키지 시험 어느 것도 실행하지 않음 |
| B28 | if at 625:2 | `if rejected := g.checkReservation(ctx, decision); rejected != nil {` | `TestAReleasedReservationNoLongerAuthorisesTheEntry` · `TestAnEntryDecisionWithNoHeldReservationIsRefused` (+2) | 5.1 에서 재실행 안 함 | 블록 625.68-627.3 을 시험 4개가 실행, 전부 PASS |
| B29 | if at 636:3 | `if plan.strategy != nil {` | 없음 | 5.1 에서 재실행 안 함 | 블록 636.27-640.4 을 무태그 패키지 시험 어느 것도 실행하지 않음 |
| B30 | if at 648:3 | `if rejected != nil {` | 없음 | 5.1 에서 재실행 안 함 | 블록 648.22-650.4 을 무태그 패키지 시험 어느 것도 실행하지 않음 |
| B31 | if at 651:3 | `if rejected := g.checkDecision(fresh, ref, plan, now); rejected != nil {` | `TestExpiryIsRecheckedOnAFreshRowBeforeTheBrokerCall` | 5.1 에서 재실행 안 함 | 블록 651.74-653.4 을 시험 1개가 실행, 전부 PASS |
| B32 | if at 654:3 | `if fresh.ClientOrderID != prep.ClientOrderID {` | 없음 | 5.1 에서 재실행 안 함 | 블록 654.48-657.4 을 무태그 패키지 시험 어느 것도 실행하지 않음 |
| B33 | if at 658:3 | `if _, rejected := g.checkProtection(dctx, plan, protectionCheckpoint); rejected != nil {` | `TestKRDriftDoesNotBlockValidUSAndDispatchDriftCallsNoBroker` | 5.1 에서 재실행 안 함 | 블록 658.90-660.4 을 시험 1개가 실행, 전부 PASS |
| B34 | if at 665:3 | `if rejected := g.checkEntry(plan); rejected != nil {` | 없음 | 5.1 에서 재실행 안 함 | 블록 665.54-667.4 을 무태그 패키지 시험 어느 것도 실행하지 않음 |
| B35 | if at 672:3 | `if plan.strategy == nil {` | `TestKRDriftDoesNotBlockValidUSAndDispatchDriftCallsNoBroker` · `TestReductionNeverReadsReadinessProvider` (+84) | 5.1 에서 재실행 안 함 | 블록 672.27-673.68 을 시험 86개가 실행, 전부 PASS |
| B36 | if at 673:4 | `if rejected := g.checkReservation(dctx, fresh); rejected != nil {` | `TestGatewayLastMomentQFinalBarrierRefusesHoldReleaseAfterInitialAdmissionCheck` | 5.1 에서 재실행 안 함 | 블록 673.68-675.5 을 시험 1개가 실행, 전부 PASS |
| B37 | if at 685:4 | `if err := dctx.Err(); err != nil {` | 없음 | 5.1 에서 재실행 안 함 | 블록 685.37-687.5 을 무태그 패키지 시험 어느 것도 실행하지 않음 |
| B38 | if at 688:4 | `if plan.strategy != nil {` | 없음 | 5.1 에서 재실행 안 함 | 블록 688.28-689.49 을 무태그 패키지 시험 어느 것도 실행하지 않음 |
| B39 | if at 689:5 | `if plan.strategy.finalAuthorityCheck == nil {` | 없음 | 5.1 에서 재실행 안 함 | 블록 689.49-691.6 을 무태그 패키지 시험 어느 것도 실행하지 않음 |
| B40 | if at 692:5 | `if err := plan.strategy.finalAuthorityCheck(dctx); err != nil {` | 없음 | 5.1 에서 재실행 안 함 | 블록 692.67-694.6 을 무태그 패키지 시험 어느 것도 실행하지 않음 |
| B41 | if at 695:5 | `if rejected := g.checkReservation(dctx, fresh); rejected != nil {` | 없음 | 5.1 에서 재실행 안 함 | 블록 695.69-697.6 을 무태그 패키지 시험 어느 것도 실행하지 않음 |
| B42 | if at 699:5 | `if requireTransport == nil {` | 없음 | 5.1 에서 재실행 안 함 | 블록 699.32-701.6 을 무태그 패키지 시험 어느 것도 실행하지 않음 |
| B43 | if at 702:5 | `if err := requireTransport(dctx, plan.strategy.lease); err != nil {` | 없음 | 5.1 에서 재실행 안 함 | 블록 702.71-704.6 을 무태그 패키지 시험 어느 것도 실행하지 않음 |
| B44 | if at 711:3 | `if plan.strategy != nil && !g.skipStrategyEntryABAForTest {` | 없음 | 5.1 에서 재실행 안 함 | 블록 711.61-713.49 을 무태그 패키지 시험 어느 것도 실행하지 않음 |
| B45 | else at 717:10 | `} else {` | `TestKRDriftDoesNotBlockValidUSAndDispatchDriftCallsNoBroker` · `TestReductionNeverReadsReadinessProvider` (+83) | 5.1 에서 재실행 안 함 | 블록 717.9-718.33 을 시험 85개가 실행, 전부 PASS |
| B46 | if at 712:4 | `if err := g.withStrategyEntryGateAuthority(dctx, plan.strategy.entryGateAuthority,` | 없음 | 5.1 에서 재실행 안 함 | 블록 713.49-716.5 을 무태그 패키지 시험 어느 것도 실행하지 않음 |
| B47 | if at 718:4 | `if err := call(); err != nil {` | 없음 | 5.1 에서 재실행 안 함 | 블록 718.33-721.5 을 무태그 패키지 시험 어느 것도 실행하지 않음 |
| B48 | if at 723:3 | `if tracker == nil {` | 없음 | 5.1 에서 재실행 안 함 | 블록 723.21-725.4 을 무태그 패키지 시험 어느 것도 실행하지 않음 |
| B49 | if at 729:2 | `if plan.strategy == nil {` | `TestKRDriftDoesNotBlockValidUSAndDispatchDriftCallsNoBroker` · `TestReductionNeverReadsReadinessProvider` (+85) | 5.1 에서 재실행 안 함 | 블록 729.26-731.3 을 시험 87개가 실행, 전부 PASS |
| B50 | else at 731:9 | `} else if g.dispatchStrategyVerified != nil {` | — | 5.1 에서 재실행 안 함 | 같은 좌표의 if 분기(다음 행)가 평가된다는 것이 곧 이 else 로 들어옴 — 그 행의 측정을 볼 것 |
| B51 | if at 731:9 | `} else if g.dispatchStrategyVerified != nil {` | — | 5.1 에서 재실행 안 함 | 같은 좌표의 if 분기(다음 행)가 평가된다는 것이 곧 이 else 로 들어옴 — 그 행의 측정을 볼 것 |
| B52 | else at 733:9 | `} else {` | 없음 | 5.1 에서 재실행 안 함 | 블록 733.8-735.3 을 무태그 패키지 시험 어느 것도 실행하지 않음 |
| B53 | if at 736:2 | `if err != nil {` | `TestConcurrentMutationsOnOneSymbolSerialise` | 5.1 에서 재실행 안 함 | 블록 736.16-740.44 을 시험 1개가 실행, 전부 PASS |
| B54 | if at 740:3 | `if errors.Is(err, journal.ErrNonceSpent) {` | 없음 | 5.1 에서 재실행 안 함 | 블록 740.44-743.4 을 무태그 패키지 시험 어느 것도 실행하지 않음 |
| B55 | if at 744:3 | `if plan.strategy != nil && !strategyCrossedTransport {` | 없음 | 5.1 에서 재실행 안 함 | 블록 744.56-746.4 을 무태그 패키지 시험 어느 것도 실행하지 않음 |
| B56 | if at 756:2 | `if res.Final == journal.StateConfirmed {` | `TestKRDriftDoesNotBlockValidUSAndDispatchDriftCallsNoBroker` · `TestReductionNeverReadsReadinessProvider` (+29) | 5.1 에서 재실행 안 함 | 블록 756.41-758.3 을 시험 31개가 실행, 전부 PASS |
| B57 | if at 759:2 | `if res.Err != nil {` | `TestKRDriftDoesNotBlockValidUSAndDispatchDriftCallsNoBroker` · `TestAbsenceNeedsStableObservationsAndDelta` (+54) | 5.1 에서 재실행 안 함 | 블록 759.20-761.3 을 시험 56개가 실행, 전부 PASS |
