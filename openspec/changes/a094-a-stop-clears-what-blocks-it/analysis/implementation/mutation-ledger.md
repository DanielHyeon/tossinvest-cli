# a094 변이 원장 (2026-09-30)

하네스 `analysis/harness/mutate.py` — 분리 git worktree(공유 트리 무접촉) · 패키지별 무변이 대조군 GREEN 선행 · 치환 원문 정확히 1회 단언 · 변이마다 원복 바이트 대조.

## 1판 — 커밋 `05de14c3` (대조군 11 GREEN)

| id | 변이 | 패키지 · -run | 결과 |
|---|---|---|---|
| M1 | R1 확정 거절 갈래 제거(모호로) | `./internal/execgw/` `TestA094Refusal` | CAUGHT — `TestA094RefusalCodeClassifiesTheAttempt` |
| M2 | R1 두 자리 모순 판정 끄기 | `./internal/execgw/` `TestA094Refusal` | CAUGHT — `TestA094RefusalCodeClassifiesTheAttempt` |
| M3 | R1 대소문자 구별(ToLower 제거) | `./internal/execgw/` `TestA094Refusal` | CAUGHT — `TestA094RefusalCodeClassifiesTheAttempt` |
| M4 | R1 문자열 아닌 code 를 부재로 | `./internal/execgw/` `TestA094Refusal` | CAUGHT — `TestA094RefusalCodeClassifiesTheAttempt` |
| M5 | 해제 판정이 park 도 풂 | `./internal/journal/` `TestA094` | CAUGHT — `TestA094TheReleaseJudgementFollowsTheAttemptState` |
| M6 | 기대 intent 대조 제거 | `./internal/journal/` `TestA094` | CAUGHT — `TestA094ALateReleaseDoesNotClearAnotherProposal` |
| M7 | 분류기가 park 를 비수용으로 | `./internal/journal/` `TestA094` | CAUGHT — `TestA094TheReleaseJudgementFollowsTheAttemptState` |
| M8 | 종결 증거 술어 뒤집기 | `./internal/journal/` `TestA094` | CAUGHT — `TestA094TheReleaseJudgementFollowsTheAttemptState`, `TestA094AConfirmedSellIsReleasedOnlyOnItsClosingRecord`, `TestA094TheClearReleasesAProposalWhoseOrdersClosedFirst` |
| M9 | 청소가 같은 종목 미종결을 안 봄 | `./internal/app/engine/` `TestA094` | CAUGHT — `TestA094AnotherIntentInFlightStopsTheArmReleaseLoop`, `TestA094AnInFlightEngineCancelIsNotCountedButAnInDoubtOneIs` |
| M10 | 매도 취소 접수를 치움으로 | `./internal/app/engine/` `TestA094|TestABreachDisplaces` | **SURVIVED** |
| M11 | 확정 취소된 매도를 다시 취소 | `./internal/app/engine/` `TestA094` | CAUGHT — `TestA094ACancelledSellIsNotClearedUntilItsClosingRecord` |
| M12 | 청소 해제 실패를 무시(치움 완료) | `./internal/app/engine/` `TestA094` | **SURVIVED** |
| M13 | 결과 못 쓴 제출도 해제(4.3c) | `./internal/app/engine/` `TestA094` | CAUGHT — `TestA094AnUnrecordedOutcomeKeepsTheProposalArmed` |
| M14 | 판정 진입 알림 호출 제거 | `./internal/app/engine/` `TestA094` | CAUGHT — `TestA094ABlockedSenderDoesNotHoldAnotherStop`, `TestA094AnEpisodeIsOneRowAcrossRestarts`, `TestA094AParkedTakeProfitIsNotClearedAndIsNamed` |
| M15 | 종결 대기 알림 한계 0 | `./internal/app/engine/` `TestA094` | CAUGHT — `TestA094TheClosingEvidenceWaitIsNamedOnce` |
| M16 | 연속 실패 제외 무시 | `./internal/app/engine/` `TestA094` | CAUGHT — `TestA094AnInFlightEngineCancelIsNotCountedButAnInDoubtOneIs` |
| M17 | 빈 가격을 실패로 | `./internal/app/engine/` `TestA094` | CAUGHT — `TestA094AnEmptyPriceIsNotAClearFailure` |
| M18 | Critical 배선 제거 | `./internal/app/engine/` `TestA094` | CAUGHT — `TestA094TheExitObserverRecordsCriticalsThroughTheNotifier` |
| M19 | 미종결 판정 대상 뒤집기(3.R5a 두 경로) | `./internal/app/engine/` `TestA094AnotherIntentInFlight` | CAUGHT — `TestA094AnotherIntentInFlightStopsTheArmReleaseLoop` |
| M19b | 미종결 판정 대상 뒤집기 — 주문 경로 쪽 | `./internal/execgw/` `TestGatewayRefuses|TestSymbol|InFlight|TestTransportOutcomeTable|TestA094` | CAUGHT — `TestOneInFlightMutationPerSymbol` |
| M20 | 응답 종목 공백 통과(D−5.1 되돌림) | `./internal/execgw/` `TestA094TheReadBack` | CAUGHT — `TestA094TheReadBackNeedsANamedSymbol` |
| M20b | 응답 종목 공백 통과 — 기동 쪽 | `./internal/reconcile/` `TestA094` | **SURVIVED** |
| M21 | 기동 확정이 읽기 판정을 무시 | `./internal/reconcile/` `TestA094` | CAUGHT — `TestA094AnUnconfirmedAckedPlaceStaysAndIsNamed`, `TestA094TheAckedAlertKeyIsStableAcrossRestarts` |
| M22 | 기동 ACKED 확정 끄기 | `./internal/reconcile/` `TestA094` | CAUGHT — `TestA094AnAckedPlaceIsConfirmedByItsRecordedNumber`, `TestA094AnUnconfirmedAckedPlaceStaysAndIsNamed`, `TestA094AnAckedCancelOrNumberlessPlaceIsOnlyNamed` |
| M23 | 따라잡기가 해제하지 않음 | `./internal/app/engine/` `TestA094` | CAUGHT — `TestA094AThawWhoseReleaseFailsIsCaughtUpAtBoot`, `TestA094TheBootCatchUpReleasesOnlyUnacceptedProposals`, `TestA094ABootCatchUpRowFailureIsNamedAndSkipped` |
| M24 | 복구 뒤 따라잡기 호출 제거 | `./cmd/tossctl/` `TestA094` | CAUGHT — `TestA094TheBootCatchUpRunsAfterASuccessfulRecovery` |
| M25 | 해동 audit 부재 검사 제거 | `./internal/app/engine/` `TestA094AThaw|TestA094AnOperatorThaw` | CAUGHT — `TestA094AThawWithoutAnAuditLogChangesNothing` |
| M26 | 해동 뒤 발의 해제 호출 제거 | `./internal/app/engine/` `TestA094AThaw|TestA094AnOperatorThaw` | CAUGHT — `TestA094AnOperatorThawClosesTheParkAndReleasesTheProposal`, `TestA094AThawWhoseReleaseFailsIsCaughtUpAtBoot` |
| M27 | 해동 stale 사전 검사 제거(OperatorResolve 의 from 검사가 뒤에 있음) | `./internal/app/engine/` `TestA094AThaw` | **SURVIVED** |

## 2판 — 커밋 `bfa61fb4` (1판 생존 넷을 죽이는 시험 추가 뒤 · M9 재측정, 대조군 4 GREEN)

| id | 변이 | 패키지 · -run | 결과 |
|---|---|---|---|
| M9 | 청소가 같은 종목 미종결을 안 봄 | `./internal/app/engine/` `TestA094` | CAUGHT — `TestA094AnotherIntentInFlightStopsTheArmReleaseLoop`, `TestA094AnInFlightEngineCancelIsNotCountedButAnInDoubtOneIs`, `TestA094ReplayThe272210LivelockStops` |
| M10 | 매도 취소 접수를 치움으로 | `./internal/app/engine/` `TestA094|TestABreachDisplaces` | CAUGHT — `TestA094AnotherIntentsCancelledSellStillHoldsTheStop` |
| M12 | 청소 해제 실패를 무시(치움 완료) | `./internal/app/engine/` `TestA094` | CAUGHT — `TestA094AParkedTakeProfitIsNotClearedAndIsNamed` |
| M20b | 응답 종목 공백 통과 — 기동 쪽 | `./internal/reconcile/` `TestA094` | **SURVIVED** |
| M27 | 해동 stale 사전 검사 제거(OperatorResolve 의 from 검사가 뒤에 있음) | `./internal/app/engine/` `TestA094AThaw` | CAUGHT — `TestA094AThawOfAnUnparkedAttemptIsStale` |

## 판정

- 1판 29 변이 중 CAUGHT 25 · SURVIVED 4(M10 · M12 · M20b · M27). 2판에서 M10 · M12 · M27 은 새 시험으로 CAUGHT.
  - **M10**(매도 취소 접수를 치움으로): 발의 자신의 매도는 원장 판정(AWAITING_CLOSE)이 두 번째로 막고 있었다 — **다른 intent 의** 매도가 그 판정 밖이었다.
    `TestA094AnotherIntentsCancelledSellStillHoldsTheStop` 이 그 구멍을 잰다.
  - **M12**(청소 해제 거절 무시): 제출은 arming 의 두 번째 발의 거절(armExitProposalTx)이 막았지만 **지연 타이머가 지워져 경보가 영영 안 났다** —
    `TestA094AParkedTakeProfitIsNotClearedAndIsNamed` 에 한계 뒤 지연 경보 1 단언을 더했다(침묵 금지).
  - **M27**(해동 stale 사전 검사 제거): OperatorResolve 의 from 검사가 전이를 막지만 audit 줄이 거절된 해소에 남았다 — stale 거절 시 audit 줄 0 단언.
- **M20b 생존 — 동등 변이(메시지 전용)**: `judgePlacedOrder` 의 「응답 종목 공백」 검사는 뒤의 `want == "" || facts.Symbol != want` 가 이미 덮는다(빈 응답 종목은 비지
  않은 발주 종목과 늘 다르다). 기동 경로 시험(reconcile)은 ACKED 잔존 · 알림으로 행동을 재므로 생존하고, 발주 직후 경로 시험(execgw, M20)은 오류 문구의
  "symbol" 을 재서 CAUGHT 다. 안전은 우연이 아니라 불일치 비교가 진다 — 명시 검사는 운영자가 읽을 사유 문구를 위해 남긴다.
- **M9 인과 확정(6.2)**: 청소가 같은 종목 미종결을 안 보게 하면 `TestA094ReplayThe272210LivelockStops` 가 PROPOSAL_CANCELLED 반복으로 실패한다 — 272210 의
  라이브락은 D−4.7 셋째 기전이다.

## 3판 — 커밋 `cb36caf4` (다각 리뷰 수리 뒤 · 38 변이, 대조군 12 GREEN)

| id | 변이 | 패키지 · -run | 결과 |
|---|---|---|---|
| M1 | R1 확정 거절 갈래 제거(모호로) | `./internal/execgw/` `TestA094Refusal` | CAUGHT — `TestA094RefusalCodeClassifiesTheAttempt` |
| M2 | R1 두 자리 모순 판정 끄기 | `./internal/execgw/` `TestA094Refusal` | CAUGHT — `TestA094RefusalCodeClassifiesTheAttempt` |
| M3 | R1 대소문자 구별(ToLower 제거) | `./internal/execgw/` `TestA094Refusal` | CAUGHT — `TestA094RefusalCodeClassifiesTheAttempt` |
| M4 | R1 문자열 아닌 code 를 부재로(모순 강제 해제) | `./internal/execgw/` `TestA094Refusal` | CAUGHT — `TestA094RefusalCodeClassifiesTheAttempt` |
| M5 | 해제 판정이 park 도 풂 | `./internal/journal/` `TestA094` | CAUGHT — `TestA094TheReleaseJudgementFollowsTheAttemptState` |
| M6 | 기대 intent 대조 제거 | `./internal/journal/` `TestA094` | CAUGHT — `TestA094ALateReleaseDoesNotClearAnotherProposal` |
| M7 | 분류기가 park 를 비수용으로 | `./internal/journal/` `TestA094` | CAUGHT — `TestA094TheReleaseJudgementFollowsTheAttemptState` |
| M8 | 종결 증거 술어 뒤집기 | `./internal/journal/` `TestA094` | CAUGHT — `TestA094TheReleaseJudgementFollowsTheAttemptState`, `TestA094AConfirmedSellIsReleasedOnlyOnItsClosingRecord`, `TestA094TheClearReleasesAProposalWhoseOrdersClosedFirst` |
| M9 | 청소가 같은 종목 미종결을 안 봄 | `./internal/app/engine/` `TestA094` | CAUGHT — `TestA094AnotherIntentInFlightStopsTheArmReleaseLoop`, `TestA094AnInFlightEngineCancelIsNotCountedButAnInDoubtOneIs`, `TestA094ReplayThe272210LivelockStops` |
| M10 | 매도 취소 접수를 치움으로 | `./internal/app/engine/` `TestA094|TestABreachDisplaces` | CAUGHT — `TestA094AnotherIntentsCancelledSellStillHoldsTheStop` |
| M11 | 확정 취소된 매도를 다시 취소 | `./internal/app/engine/` `TestA094` | CAUGHT — `TestA094ACancelledSellIsNotClearedUntilItsClosingRecord` |
| M12 | 청소 해제 실패를 무시(치움 완료) | `./internal/app/engine/` `TestA094` | CAUGHT — `TestA094AParkedTakeProfitIsNotClearedAndIsNamed` |
| M13 | 결과 못 쓴 제출도 해제(4.3c) | `./internal/app/engine/` `TestA094` | CAUGHT — `TestA094AnUnrecordedOutcomeKeepsTheProposalArmed` |
| M14 | 판정 진입 알림 호출 제거 | `./internal/app/engine/` `TestA094` | CAUGHT — `TestA094ABlockedSenderDoesNotHoldAnotherStop`, `TestA094AnEpisodeIsOneRowAcrossRestarts`, `TestA094AParkedTakeProfitIsNotClearedAndIsNamed` |
| M15 | 종결 대기 알림 한계 0 | `./internal/app/engine/` `TestA094` | CAUGHT — `TestA094TheClosingEvidenceWaitIsNamedOnce` |
| M16 | 연속 실패 제외 무시 | `./internal/app/engine/` `TestA094` | CAUGHT — `TestA094AnInFlightEngineCancelIsNotCountedButAnInDoubtOneIs` |
| M17 | 빈 가격을 실패로 | `./internal/app/engine/` `TestA094` | CAUGHT — `TestA094AnEmptyPriceIsNotAClearFailure` |
| M18 | Critical 배선 제거 | `./internal/app/engine/` `TestA094` | CAUGHT — `TestA094TheExitObserverRecordsCriticalsThroughTheNotifier` |
| M19 | 미종결 판정 대상 뒤집기(3.R5a 두 경로) | `./internal/app/engine/` `TestA094AnotherIntentInFlight` | CAUGHT — `TestA094AnotherIntentInFlightStopsTheArmReleaseLoop` |
| M19b | 미종결 판정 대상 뒤집기 — 주문 경로 쪽 | `./internal/execgw/` `TestGatewayRefuses|TestSymbol|InFlight|TestTransportOutcomeTable|TestA094` | CAUGHT — `TestOneInFlightMutationPerSymbol` |
| M20 | 응답 종목 공백 통과(D−5.1 되돌림) | `./internal/execgw/` `TestA094TheReadBack` | CAUGHT — `TestA094TheReadBackNeedsANamedSymbol` |
| M20b | 응답 종목 공백 통과 — 기동 쪽 | `./internal/reconcile/` `TestA094` | **SURVIVED** |
| M21 | 기동 확정이 읽기 판정을 무시 | `./internal/reconcile/` `TestA094` | CAUGHT — `TestA094AnUnconfirmedAckedPlaceStaysAndIsNamed`, `TestA094TheAckedAlertKeyIsStableAcrossRestarts`, `TestA094TheAckedAlertWithholdsTheBrokerBody` |
| M22 | 기동 ACKED 확정 끄기 | `./internal/reconcile/` `TestA094` | CAUGHT — `TestA094AnAckedPlaceIsConfirmedByItsRecordedNumber`, `TestA094AnUnconfirmedAckedPlaceStaysAndIsNamed`, `TestA094AnAckedCancelOrNumberlessPlaceIsOnlyNamed` |
| M23 | 따라잡기가 해제하지 않음 | `./internal/app/engine/` `TestA094` | CAUGHT — `TestA094AThawWhoseReleaseFailsIsCaughtUpAtBoot`, `TestA094TheBootCatchUpReleasesOnlyUnacceptedProposals`, `TestA094ABootCatchUpRowFailureIsNamedAndSkipped` |
| M24 | 복구 뒤 따라잡기 호출 제거 | `./cmd/tossctl/` `TestA094` | CAUGHT — `TestA094TheBootCatchUpRunsAfterASuccessfulRecovery` |
| M25 | 해동 audit 부재 검사 제거 | `./internal/app/engine/` `TestA094AThaw|TestA094AnOperatorThaw` | CAUGHT — `TestA094AThawWithoutAnAuditLogChangesNothing` |
| M26 | 해동 뒤 발의 해제 호출 제거 | `./internal/app/engine/` `TestA094AThaw|TestA094AnOperatorThaw` | CAUGHT — `TestA094AnOperatorThawClosesTheParkAndReleasesTheProposal`, `TestA094AThawWhoseReleaseFailsIsCaughtUpAtBoot` |
| M27 | 해동 stale 사전 검사 제거(OperatorResolve 의 from 검사가 뒤에 있음) | `./internal/app/engine/` `TestA094AThaw` | CAUGHT — `TestA094AThawOfAnUnparkedAttemptIsStale` |
| M28 | 3.E5 배제의 대상 주문 결속 제거 | `./internal/app/engine/` `TestA094AnInFlightEngineCancel` | **SURVIVED** |
| M29 | 3.E5 모든 대상 주문에 취소 요구 제거 | `./internal/app/engine/` `TestA094AnInFlightEngineCancel` | **SURVIVED** |
| M30 | 판정 진입의 비수용 발의 해제 제거 | `./internal/app/engine/` `TestA094` | CAUGHT — `TestA094AnUnacceptedProposalLeftArmedIsReleasedNextObservation` |
| M31 | 연속 실패 경보 래치를 기록 전에 | `./internal/app/engine/` `TestA094` | CAUGHT — `TestA094AFailedStreakRecordIsRetried` |
| M32 | 다른 intent 매도의 종결 대기 알림 제거 | `./internal/app/engine/` `TestA094` | CAUGHT — `TestA094AnotherIntentsCloseWaitIsNamedToo` |
| M33 | intent 없는 발의 알림 제거 | `./internal/app/engine/` `TestA094` | CAUGHT — `TestA094AnIntentlessArmedProposalIsNamed` |
| M34 | ACKED 알림에 브로커 본문 노출 | `./internal/reconcile/` `TestA094` | CAUGHT — `TestA094TheAckedAlertWithholdsTheBrokerBody` |
| M35 | 따라잡기 목록 실패 critical 제거 | `./internal/app/engine/` `TestA094` | CAUGHT — `TestA094ABootCatchUpListFailureIsNamed` |
| M36 | 해동 해제 실패 critical 제거 | `./internal/app/engine/` `TestA094AThaw` | CAUGHT — `TestA094AThawWhoseReleaseFailsIsCaughtUpAtBoot` |

### 3판 판정

- CAUGHT 35 · SURVIVED 3(M20b 동등 — 1판 판정 유지 · **M28 · M29**).
- **M28 · M29(3.E5 결속)**: 두 검사(미종결 취소가 치울 대상을 겨눔 · 치울 대상 전부에 취소가 있음)가 서로를 가렸다 — 「다른 주문을 겨눈 취소」 행 하나는
  둘 중 하나만 있어도 셈. 결속이 풀리면 무엇이 통과하는지를 따로 세운 두 행을 더했다: 「대상 취소 + 무관한 미종결 취소」(M28 이 통과시키던 것 — 무관한 취소
  때문에 이른 경보가 꺼짐) · 「취소 없는 둘째 작업 매수」(M29 가 통과시키던 것 — 치우지 못한 주문이 남았는데 안 셈).
- **3b — M28 · M29 재측정**: 커밋 `cb36caf4` 의 분리 worktree 에 보강 시험 파일만 덮어 두 변이를 다시 돌렸다(Go 착지 동결 중 — 시험 파일 커밋은 동결 해제 뒤).
  결과: **M28 CAUGHT**(`…/a_bound_cancel_plus_an_unrelated_in-flight_cancel`) · **M29 CAUGHT**(`…/a_second_working_buy_with_no_cancel`). 무변이 대조군 GREEN.
- 최종: 38 변이 CAUGHT 37 · 동등 1(M20b).

## codex i1 수리 되돌림 변이 (2026-10-01, HEAD 위 분리 worktree + 수리 파일)

| id | 변이 | 결과 |
|---|---|---|
| F1-revert | 청소에서 `withPending=false` 건너뜀을 확정 취소 검사 **앞**으로 되돌림 | CAUGHT — `TestA094ACancelledSellStillHoldsTheStopAfterTheProposalIsReleased` |
| F2-revert | 판정 진입이 해제해도 그 주기 판정을 계속 | CAUGHT — `TestA094AReleasedLadderRungIsProposedAgain` |

무변이 대조군(`-run TestA094`) GREEN.
