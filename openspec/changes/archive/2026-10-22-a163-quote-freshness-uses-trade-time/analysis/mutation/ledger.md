# a163 변이 원장 (tasks 4.1)

- HEAD `ff232b3cf` 의 추적 파일 사본(`a163-mutation-1199441`)에서 실행. 원본 worktree 무편집. 생성: `harness.py` (손편집 금지 — 다시 돌릴 것)
- 무변이 대조군: 시험 묶음 26 개 전부 rc=0
- 결과: **32 CAUGHT / 32**

| # | 변이 | 파일 | 판정 | 실패한 시험 |
| --- | --- | --- | --- | --- |
| ① | 보고 사유(시각 없음)가 있으면 판정을 건너뜀 | `internal/app/engine/exitloop.go` | CAUGHT | TestA163AMissingOrMalformedBrokerTimeJudgesTheSame,TestA163AMissingOrMalformedBrokerTimeJudgesTheSame/absent,TestA163AMissingOrMalformedBrokerTimeJudgesTheSame/empty,TestA163AMissingOrMalformedBrokerTimeJudgesTheSame/null,TestA163TheBrokerDataTimeIsReadOnlyWhereTheDesignAllows |
| ② | 보고를 판정 순회 안(판정 직후)으로 이동 | `internal/app/engine/exitloop.go` | CAUGHT | TestA163TheReportComesAfterEveryJudgementAndBeforeTheUnobservedSettlement |
| ③ | 어댑터가 빈 시각을 받은 시각 문자열로 메움 | `internal/official/market_reads.go` | CAUGHT | TestA163AdaptPricesKeepsTheBrokerTimeVerbatim,TestA163PricesCarriesTheBrokerTimeInOneRequest |
| ④a | 어댑터가 FetchedAt 을 브로커 시각으로 | `internal/official/market_reads.go` | CAUGHT | TestA163AdaptPricesKeepsTheBrokerTimeVerbatim |
| ④b | observe 리터럴의 FetchedAt 에 브로커 시각 대입 | `internal/app/engine/exitloop.go` | CAUGHT | TestA163IdentityAndObservedAtComeFromTheReceiptNotTheBrokerTime,TestA163TheBrokerDataTimeIsReadOnlyWhereTheDesignAllows |
| ⑤ | 래치 제거(이미 알린 포지션도 매 주기 알림) | `internal/app/engine/exit_source_age.go` | CAUGHT | TestA163AMissingOrUnparseableTimeIsAlertedOnceAndRearmsWhenATimeArrives |
| ⑥ | 보고 등급을 critical 로(등급표 등재) | `internal/obs/event.go` | CAUGHT | TestA163AMissingOrUnparseableTimeIsAlertedOnceAndRearmsWhenATimeArrives,TestA163TheReportNeverBlocksEntryOrEscalates |
| ⑦ | 어댑터가 빈 · 깨진 timestamp 행을 버림 | `internal/official/market_reads.go` | CAUGHT | TestA163AMissingOrMalformedBrokerTimeJudgesTheSame,TestA163AMissingOrMalformedBrokerTimeJudgesTheSame/+0900,TestA163AMissingOrMalformedBrokerTimeJudgesTheSame/absent,TestA163AMissingOrMalformedBrokerTimeJudgesTheSame/empty,TestA163AMissingOrMalformedBrokerTimeJudgesTheSame/garbage,TestA163AMissingOrMalformedBrokerTimeJudgesTheSame/null |
| ⑧ | 어댑터가 빈 · 깨진 timestamp 에 오류 반환(Prices) | `internal/official/market_reads.go` | CAUGHT | PANIC in -run TestA163AMissingOrMalformedBrokerTimeJudgesTheSame: panic: test timed out after 4m0s |
| ⑨ | observe B7 필터의 effectiveAt 을 브로커 시각 기준으로 | `internal/app/engine/exitloop.go` | CAUGHT | TestA163AStaleBrokerTimeStillLiquidatesInTheSameCycle,TestA163TheBrokerDataTimeIsReadOnlyWhereTheDesignAllows |
| ⑩ | snapshotContext 가 stableObservationID 에 브로커 시각을 넘김 | `internal/app/engine/exitloop.go` | CAUGHT | TestA163TheBrokerDataTimeIsReadOnlyWhereTheDesignAllows,TestA163TheJudgementFunctionsRenderExactlyAsBeforeTheChange |
| ⑪ | 보고기 recover 경계 제거 | `internal/app/engine/exit_source_age.go` | CAUGHT | PANIC in -run TestA163APanickingReportNeitherStopsTheLoopNorChangesTheStops: panic: a163: injected log sink panic |
| ⑫ | engineRuntime 의 SourceAgeLog 배선 삭제 | `cmd/tossctl/engine.go` | CAUGHT | TestA163EngineRuntimeWiresTheSourceAgeLogger |
| ⑬ | JSON 태그에서 omitempty 삭제 | `internal/domain/models.go` | CAUGHT | TestA163TheWTSQuoteJSONIsByteIdenticalAndCarriesNoBrokerTime · TestA163TheOfficialStocksQuoteJSONIsByteIdenticalAndCarriesNoBrokerTime |
| ⑭ | 전이 대신 매 주기 줄 | `internal/app/engine/exit_source_age.go` | CAUGHT | TestA163AnAgeLineIsWrittenOnlyWhenTheBucketChanges |
| ⑮ | held 기준 정리 삭제 | `internal/app/engine/exit_source_age.go` | CAUGHT | TestA163AClosedPositionLeavesTheLatch |
| ⑯ | 주기 상한 삭제 | `internal/app/engine/exit_source_age.go` | CAUGHT | TestA163AtMostEightSourceAlertsLeaveInOneCycle |
| ⑰ | 토글 무시(꺼짐이면 반환 삭제) | `internal/app/engine/exit_source_age.go` | CAUGHT | TestA163TheToggleOffWritesNothingAndJudgesExactlyAsOn |
| ⑱a | 기본값 ON — mergeEngine 이 키 없음을 true 로 | `internal/config/engine.go` | CAUGHT | TestA163TheSourceAgeReportToggleDefaultsOff,TestA163TheSourceAgeReportToggleDefaultsOff/no_key |
| ⑱b | 기본값 ON — 진입 함수 판단을 뒤집음 | `internal/app/engine/exit_source_age.go` | CAUGHT | TestA163TheToggleOffWritesNothingAndJudgesExactlyAsOn |
| ⑲ | merge 를 automation_gate 조기 반환 뒤로 이동 | `internal/config/engine.go` | CAUGHT | TestA163TheSourceAgeReportToggleDefaultsOff,TestA163TheSourceAgeReportToggleDefaultsOff/true_without_automation_gate |
| ⑳ | exitwiring 덮어쓰기 삭제 | `internal/app/engine/exitwiring.go` | CAUGHT | TestA163TheProductionAssemblyReadsTheLoadedToggle,TestA163TheProductionAssemblyReadsTheLoadedToggle/loaded=false,TestA163TheProductionAssemblyReadsTheLoadedToggle/loaded=true |
| ㉑a | audit 행 삭제 | `internal/app/engine/interlock.go` | CAUGHT | TestA163AStartupAuditsTheLoadedReportToggle,TestA163TheReportToggleIsTheLastAuditedSettingAndOnlyItsChangeIsWritten |
| ㉑b | audit 행 값 고정(false) | `internal/app/engine/interlock.go` | CAUGHT | TestA163AStartupAuditsTheLoadedReportToggle,TestA163TheReportToggleIsTheLastAuditedSettingAndOnlyItsChangeIsWritten |
| ㉑c | NewContext 가 로드된 값 대신 false 를 넘김 | `internal/app/engine/engine.go` | CAUGHT | TestA163AStartupAuditsTheLoadedReportToggle |
| ㉒ | 토글 판단을 ObserveOnce 판정 순회 앞으로 — 꺼짐이면 판정 건너뜀 | `internal/app/engine/exitloop.go` | CAUGHT | TestA163AStaleBrokerTimeStillLiquidatesInTheSameCycle,TestA163TheReportToggleIsReadOnlyWhereTheDesignAllows,TestA163TheToggleOffWritesNothingAndJudgesExactlyAsOn |
| ㉓ | 거둔 패닉의 고정 문구 줄 삭제(흔적 없는 패닉) | `internal/app/engine/exit_source_age.go` | CAUGHT | TestA163APanickingReportNeitherStopsTheLoopNorChangesTheStops,TestA163APanickingReportNeitherStopsTheLoopNorChangesTheStops/alerter_panics |
| ㉔ | 고정 문구 줄 래치 삭제(매 주기 같은 줄) | `internal/app/engine/exit_source_age.go` | CAUGHT | TestA163APanickingReportNeitherStopsTheLoopNorChangesTheStops,TestA163APanickingReportNeitherStopsTheLoopNorChangesTheStops/alerter_panics |
| ㉕ | 미래(ahead) 구간도 알림 사유로 | `internal/app/engine/exit_source_age.go` | CAUGHT | TestA163EveryBucketIsClassifiedFromTheReceiptTime |
| ㉖ | 원문 정화 삭제(바이트 자르기 · 제어문자 통과) | `internal/app/engine/exit_source_age.go` | CAUGHT | TestA163TheRawBrokerTimeIsBoundedPrintableAndValidUTF8 |
| ㉗ | 비상 청산(flatten) 경로가 브로커 시각을 읽음 | `internal/flatten/liquidate.go` | CAUGHT | TestA163TheBrokerDataTimeIsReadOnlyWhereTheDesignAllows |
| ㉘ | 거대 원문 해독 상한 삭제(매 주기 끝까지 훑음) | `internal/app/engine/exit_source_age.go` | CAUGHT | TestA163TheRawBrokerTimeIsBoundedPrintableAndValidUTF8 |
