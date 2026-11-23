# a198 변이 원장 (tasks 4.1, 2026-11-23)

- 하네스: `a198_mutate.py`(이 디렉터리). 사본 = `git worktree add --detach ~/tossos-work/a198-mut-<pid> <sha>`, 시작 sha 단언, 무변이 대조군 GREEN 뒤 변이마다 적용 확인(표식·앵커 1회) → 닿음(변이 판 coverprofile 의 변이 줄 블록 계수) → 시험 → 원복(원문 바이트 비교 + `git diff --quiet`) → 끝 대조군 GREEN → 사본 제거. 한 번에 한 판. `GOFLAGS` 는 사본에서 끔(-trimpath 가 runtime.Caller 경로 시험을 깨뜨림 — 4.2 관찰).
- 대상 sha `09bb9e8c1`(GREEN 커밋). 1판 시험 = `go test ./internal/app/engine -run A198 -count=1`(ledger-r1.jsonl). 2판 = M7 만, `-run 'A198|A094'`(ledger-r2-m7.jsonl).
- 1판 하네스 결함(정정): 닿음 줄을 「변이 새 글의 첫 줄을 변이 파일에서 처음 찾은 자리」로 계산해 M7 이 exitloop.go:364(엉뚱한 `}`)를 재고 0 을 냈다. 2판은 원문 앵커 위치로 계산(1728, 계수 1). 다른 변이의 줄은 실제 편집 자리와 일치(아래 표).

| 변이 | 적용 | 닿음(줄·계수) | 결과 | 죽인 시험 |
|---|---|---|---|---|
| M1 결함 복원(RiskGuardian 매도·PLACE 결정) | ✓ | exitloop.go:1694 · 1 | **CAUGHT** | A198 AWorkingEngineBuy · ArmedSell · ARefusedCancel · …(7 중 다수) |
| M2 결정 방향을 매도로 고정 | ✓ | :1699 · 1 | **CAUGHT** | AWorkingEngineBuy · ARefusedCancel · AnInDoubt · ProductionAssembly |
| M3 결정에 멱등 키가 붙는 발행(KindPlace) | ✓ | :1696 · 1 | **CAUGHT** | AWorkingEngineBuy · ArmedSell · ARefusedCancel · … |
| M4 비확정 취소 실패 기록 삭제 | ✓ | :1720 · 1 | **CAUGHT** | ARefusedCancel |
| M4b 발행 실패 기록 삭제 | ✓ | :1704 · 1 | **CAUGHT** | AnUnrecordableCancelDecision |
| M5 실패 기록의 중복 억제 삭제(매 주기 한 줄) | ✓ | a198_clear_cancel_log.go:31 · 1 | **CAUGHT** | ARefusedCancel · AnUnrecordableCancelDecision |
| M6 실패 줄에 원문 싣기(LiveOrder 전체 — 계좌·수량·가격) | ✓ | a198_clear_cancel_log.go:60 · 1 | **CAUGHT** | ARefusedCancel |
| M7 매도 취소 접수를 치움으로 판정 | ✓ | :1728 · 1 | A198 만: **생존** · A198+A094: **CAUGHT** | TestA094ACancelledSellStillHoldsTheStopAfterTheProposalIsReleased · TestA094AnotherIntentsCancelledSellStillHoldsTheStop(가짜 제출자 기존 시험) |

- M7 생존 이유(A198 만): 실제 게이트웨이 시험 3(ArmedSell)에서는 그 매도가 무장된 발의 자신의 주문이라, B12 를 지워도 뒤의 `ReleaseClearedExitProposal`(B14→B17)이 「그 intent 의 접수 확정 주문에 종결 증거 없음」으로 해제를 거절해 같은 주기 손절이 여전히 막힌다 — 다른 가드가 대신 지킴. B12 가 유일한 벽이 되는 모양(다른 intent 의 매도 · 발의 해제 뒤의 매도)은 기존 A094 시험 2 개가 가짜 제출자로 죽인다. 실제 게이트웨이로 그 모양을 재는 시험은 이 change 에 없음(보고).
- 끝 대조군: 두 판 모두 rc 0, 사본 제거 확인.

## 4b 보강 변이 (2026-11-24, tasks 4b.8)

- 하네스 `a198_mutate_4b.py`(같은 절차). 대상 sha `e92b7621b`(보강 GREEN). 시험 = `go test ./internal/app/engine ./internal/execgw -run 'A198|A094|A165' -count=1`. 원장 `ledger-4b.jsonl`. 대조군 rc 0 → 13 변이 → 끝 대조군 rc 0 → 사본 제거.

| 변이 | 닿음(줄·계수) | 결과 | 죽인 시험 |
|---|---|---|---|
| V1 D7 을 PLACE·AMEND 에도(kind 검사 삭제) | cancel_refusal.go:45 · 1 | CAUGHT | TestA198OtherAnswersKeepTheirOutcome · TestA165NilPolicyAndNonCancelKeepTheOriginalOutcome · TestA165PolicyDoesNotChangeNilOrNonCancelAmbiguousOutcomes |
| V2 부분문자열(접두) 일치로 완화 | a198_settled_cancel_refusal.go:46 · 1 | CAUGHT | TestA198OtherAnswersKeepTheirOutcome |
| V3 status 409 전체를 확정 거절로 | :38 · 1 | CAUGHT | AnAlreadyModified · OtherAnswers · AlreadyProcessingKeepsTheA165Path · A165 엔진 시험 다수 |
| V4 already-modified 를 D7 에 | :25(맵 리터럴 — 계수 0, 패키지 초기화라 블록 계수 밖) | CAUGHT | TestA198AnAlreadyModifiedCancelStaysInDoubt · OtherAnswers |
| V5 D7 을 flatten 정책에도 | cancel_refusal.go:53 · 1 | CAUGHT | TestA198OtherAnswersKeepTheirOutcome(flatten 정책 케이스) |
| V6 D7 삭제 | :53 · 1 | CAUGHT | AnAlreadyFilledBuy… · AnAlreadyCanceledAfterAlreadyProcessing · EngineSettledOrderRefusalsAreDefinitive |
| V7 D8 삭제 | exitloop.go:1678 · 1 | CAUGHT | AnAcceptedBuyCancelIsNotSentAgain… · OnlyTheRefusedBuyIsCancelledAgain |
| **V8 D8 을 매도에도 「치움」(초과 매도 방향)** | :1678 · 1 | **CAUGHT** — 단 A094(가짜 제출자) 시험 1개만 죽임 | TestA094ACancelledSellStillHoldsTheStopAfterTheProposalIsReleased |
| N10 발행 실패에도 cleared 유지 | :1708 · 1 | CAUGHT | OnlyTheCancelDecisionFailingStillWithholdsTheStop |
| N2 최대 수량 ×2 | :1703 · 1 | CAUGHT | TheCancelDecisionCeilingIsTheOrderQuantity |
| N6 접수 확정 뒤 표시 정리 삭제 | :1727(삭제 앞 줄 — `}` 라 계수 0) | CAUGHT | AConfirmedCancelClearsTheFailureMark |
| N9 해석 실패 줄 삭제 | :1690 · 1 | CAUGHT | AnUnreadableWorkingOrderIsLoggedOnce |
| N11 표시를 쓰기 앞에 | a198_clear_cancel_log.go:47 · 1 | CAUGHT | AFailureLineWhoseWritePanickedIsWrittenNextCycle |

- 계수 0 인 두 줄(V4 맵 리터럴·N6 삭제 앞 `}`)은 닿음 측정 방식의 한계 — 두 변이 모두 지목한 시험이 죽였으므로 닿음은 결과로 확인됨.
- V8(초과 매도 방향)을 실제 게이트웨이 시험은 죽이지 못한다: A198 의 무장된 매도 시험에서는 뒤의 `ReleaseClearedExitProposal` 이 대신 막는다(1판 M7 과 같은 구조). 죽이는 것은 A094 의 가짜 제출자 시험 1개 — 범위 안(선택 시험 묶음) 시험이다.

## 구현 리뷰 r2 추가 변이 (2026-11-24)

- 같은 하네스(`A198_ONLY="R13,R3"`), 대상 sha `381f229de`. 원장 `ledger-r2.jsonl`. 대조군 rc 0 → 끝 대조군 rc 0 → 사본 제거. (첫 시도는 필터 편집이 적용되지 않은 채 전체 목록으로 돌기 시작해 중단했고, 남은 사본 worktree 는 `worktree remove --force` 로 지운 뒤 다시 돌렸다.)

| 변이 | 닿음 | 결과 | 죽인 시험 |
|---|---|---|---|
| R13 조회 오류여도 치움 유지 | exitloop.go:1676 · 0(그 갈래를 어느 시험도 실행하지 않음) | **생존** | 없음 — 조회 오류 주입 수단 없음(design §6 「시험 공백」) |
| R3 해석 실패 사유 값 바꾸기 | :1693 · 1 | CAUGHT | TestA198AnUnreadableWorkingOrderIsLoggedOnce(사유 값 단언 추가 뒤) |

## 보강 2 변이 (2026-11-24, tasks 4c.8 — D9·D10)

- 하네스 `a198_mutate_4c.py`(같은 절차 — 사본 worktree · 시작 sha 단언 · 무변이 대조군 GREEN · 변이마다 닿음(coverprofile) · 원복 바이트 비교 + `git diff --quiet` · 끝 대조군 · 사본 제거, 한 변이가 여러 자리를 고칠 수 있음). 시험 = `go test ./internal/app/engine ./internal/journal ./internal/execgw ./internal/reconcile -run 'A198|A193|A094' -count=1`(GOFLAGS 없음).
- 1판: 대상 sha `e4f00a720`(GREEN), 원장 `ledger-4c.jsonl` — 대조군 rc 0(engine 139.5s · journal 29.6s · execgw 28.5s · reconcile 12.8s) → 18 변이 → 끝 대조군 rc 0 → 사본 제거 확인. W1 생존 → 시험 추가(`TestA198ACancelIsNotClosedOutsideItsTargetsOwnScope/closing_state_without_the_terminal_flag`) 뒤 2판 W1 만 재측정(`ledger-4c-w1.jsonl`).
- 닿음 0 인 줄(W3·W3b·W11·W17)은 SQL 원문 문자열·삭제된 줄·`case` 머리라 블록 계수 밖 — 지목한 시험이 죽였으므로 결과로 닿음 확인(4b 와 같은 한계).

| 변이 | 자리 · 닿음 | 결과 | 죽인 시험 |
|---|---|---|---|
| W1 종결 아닌 대상에도 닫기(terminal 조건 삭제) | a198_cancel_target_close.go:104 · 1 | 1판 **생존** → 2판(sha `fcd5fa77b`, 대조군·끝 대조군 rc 0) **CAUGHT** | 1판 없음 — 판정 표가 state 로 다시 거르고, 운영 쓰기는 state 와 terminal 을 함께 씀(filldetect/ledger.go:60-61). terminal 0 인 종결 state 행(호출자가 넘긴 관측) 시험 추가 뒤 ACancelIsNotClosedOutsideItsTargetsOwnScope |
| W2 범위 결속 삭제(거래일) | :104 · 1 | CAUGHT | AnInDoubtCancelStaysWhileItsTargetIsNotClosed · ACancelIsNotClosedOutsideItsTargetsOwnScope |
| W3 소유 범위 삭제(계좌) | :59 · 0 | CAUGHT | ACancelIsNotClosedOutsideItsTargetsOwnScope |
| W3b 소유 범위 삭제(방향) | :60 · 0 | CAUGHT | ACancelIsNotClosedOutsideItsTargetsOwnScope |
| W4 소유 유일 조건 삭제 | :98 · 1 | CAUGHT | AnInDoubtCancelWithTwoOwnersIsNotClosed · ACancelIsNotClosedOutsideItsTargetsOwnScope |
| W5 판정 표 뒤집기(취소로 끝남 → 확정 거절) | :43 · 1 | CAUGHT | TheSettlerClosesAnEngineCancelWithoutBrokerOrAlert · TheCancelIsClosedBeforeThePlaceResolution · AnInDoubtBuyCancelClosesOnItsClosingRecordAndTheStopFollows … |
| W6 fail_closed 행도 닫기 | :119 · 1 | CAUGHT | AnInDoubtCancelStaysWhileItsTargetIsNotClosed · ACancelIsNotClosedOutsideItsTargetsOwnScope |
| W7 REPLACED 를 닫기 | :42 · 1 | CAUGHT | AnInDoubtCancelStaysWhileItsTargetIsNotClosed · CancelVerdictCoversEveryBrokerState · ACancelIsNotClosedOutsideItsTargetsOwnScope |
| W8 세션에서 ACKED 도 닫기 | in_doubt_settle.go:182 · 1 | CAUGHT | TheSettlerLeavesFlattenAndAckedCancelsAlone |
| W9 D9 를 AMEND 에도 | acked_boot.go:56 · 1 | CAUGHT | TestA094AnAckedCancelOrNumberlessPlaceIsOnlyNamed(amend 행) |
| W10 D10 이 PLACE 를 건드리기 | in_doubt_settle.go:180 · 1 | CAUGHT | A193AnInDoubtPlaceFoundInTheListIsConfirmedWhileTheEngineRuns · A193AnInDoubtSellWithoutBaselineParks… · A193AnInDoubtBuyProvenAbsentIsFailedConfirmed … |
| W11 닫기를 발주 해소 뒤로 | in_doubt_settle.go:193 · 0 | CAUGHT | TheCancelIsClosedBeforeThePlaceResolution |
| W12 source 한정 삭제 | a198_cancel_target_close.go:87 · 1 | CAUGHT | TheSettlerLeavesFlattenAndAckedCancelsAlone · ACancelIsNotClosedOutsideItsTargetsOwnScope |
| W13 D9 쓰기 실패를 true 로 | acked_boot.go:63 · 1 | CAUGHT | A198AFailedAckedCancelWriteLeavesTheAckAndNamesIt |
| W14 D9 삭제 | acked_boot.go:56 · 1 | CAUGHT | A094AnAckedCancelOrNumberlessPlaceIsOnlyNamed · A198AnAckedCancelIsCompletedAtBoot · A198AFailedAckedCancelWriteLeavesTheAckAndNamesIt |
| W15 스윕이 해소기(브로커 경로)를 부르기 | a198_in_doubt_cancel_close.go:19 · 1 | CAUGHT | A193OnlyInDoubtPlaceAttemptsAreTouched · TheSettlerClosesAnEngineCancelWithoutBrokerOrAlert · TheCancelIsClosedBeforeThePlaceResolution |
| W16 CANCEL 에 처음 봄 알림 | in_doubt_settle.go:193 · 1 | CAUGHT | A193OnlyInDoubtPlaceAttemptsAreTouched · TheSettlerClosesAnEngineCancelWithoutBrokerOrAlert |
| W17 D10 삭제(취소를 고르지 않음) | in_doubt_settle.go:182 · 0 | CAUGHT | TheSettlerClosesAnEngineCancelWithoutBrokerOrAlert · TheCancelIsClosedBeforeThePlaceResolution · AnInDoubtBuyCancelClosesOnItsClosingRecordAndTheStopFollows … |
