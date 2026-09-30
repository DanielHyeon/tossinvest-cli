# Branch Test Map: `ExitObserver.ObserveOnce`

AST 분기 8 · return 5 — **편집 뒤 좌표**(difflib 재번호: 번호 불변, FLM 「편집 뒤」). 편집 전 표(좌표 413–470 판)는 git 이력(`ee60c2be`)에 있다.
편집 전 진입 실측: `analysis/harness/observeonce.blocks`(commit `eac13df1`).

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:441` 양보 — 무편집, 포지션 처리 없음 | `TestTheCycleYieldsToFillDetection` (`exitloop_test.go:673`) · `TestA090R6AYieldIsUnchanged` · `TestA090R3bTheThresholdRunsFromTheLastJudgementAcrossYieldsAndBlackouts`(양보 시간도 경과에 듦) | M09(B1 에서 처리) CAUGHT | yes |
| B2 | `:451` 작업 집합 오류 — 무편집(명명 구멍 D11-2) | 없음 — 블록 count 0(시험 원장은 관측 중 살아 있음, a092 번들과 같은 사실) | no | no |
| B3 | `:455` states 빔 — 첫 문장 `settleUnobserved` | `TestAnAccountHoldingNothingIsNotInAnOutage` · `TestA090R12APositionWhoseStateCannotOpenIsCountedAndAlerted`(alone_B3) · `TestA090R13QuarantineReadAndWriteFailuresAreCounted` · `TestA090ABlackoutCyclesMarksDoNotLeakIntoTheNextCycle` | RED-1 · M05 CAUGHT | yes |
| B4 | `:468` 전 종목 미응답 — 무편집 | `TestASustainedOutageBlocksEntriesAndAlertsOnce` · `TestA090R5WhenNoSymbolAnswersOnlyTheAccountLadderMoves` · `TestA090R3bTheThresholdRunsFromTheLastJudgementAcrossYieldsAndBlackouts` | M08(B4 에서 처리) CAUGHT | yes |
| B5 | `:477` 순회 — `noteJudged` · 순회 뒤 `settleUnobserved` | `TestA090R3APositionUnjudgedPastTheThresholdIsRecordedOnceAndTightens` · `TestA090R3cAStopIsSubmittedBeforeTheAlertAndTheNextCycleIsNotDelayed` · `TestA090R16AJudgementThatEndsAtOnceStillCountsAsObserved` | RED-1 · M03 · M04 · M10 CAUGHT | yes |
| B6 | `:479` 응답 없음 — `noteUnobservedCause(no_quote)` | `TestA090R1AZeroQuoteSiblingIsCountedAndLoggedAtNormalGrade` · `TestA090R2AnAbsentSiblingIsCountedLikeAZeroQuote` · `TestA111ValidSiblingIsJudgedWithoutLendingFreshnessToInvalidSymbol`(무변화) | RED-1 · M01 CAUGHT | yes |
| B7 | `:487` 임대 만료 — `noteUnobservedCause(quote_expired)` | `TestA090R7AQuoteThatExpiredMidCycleIsCountedAsExpired` · `TestA111SlowFirstPositionExpiresLaterQuoteWithoutAbandoningStartedProtection`(무변화) | RED-1 · M02 CAUGHT | yes |
| B8 | `:495` 첫 judge 오류 — 무편집 | `TestA090R16AJudgementThatEndsAtOnceStillCountsAsObserved`(selector_stamp_failure — judge 오류여도 관측됨) | no | yes |

변이 원장: `analysis/mutation/round2-27of27.log`(하네스 `a090_mutate.py`).
