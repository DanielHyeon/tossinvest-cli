# a163 RED → GREEN 기록

| 파일 | 무엇 |
| --- | --- |
| `red-1-compile.txt` | RED 1단계 — 새 시험만 더한 트리(HEAD `78efe1951` + 시험). 새 심볼이 없어 컴파일 실패(official · config · engine), cmd/tossctl 구조 시험은 행동 실패 |
| `red-2-behaviour.txt` | RED 2단계 — 타입 표면만 더한 골격(필드 · 상수 · 끝 인자 · 빈 `reportSourceAge`, 동작 0). 19 시험이 **의도한 이유**로 실패 |
| `green.txt` | GREEN — 같은 시험 45 PASS · 0 FAIL |

## task → 시험 → RED 이유

| task | 시험 | RED 이유(red-2) |
| --- | --- | --- |
| 2.5 | `TestA163AdaptPricesKeepsTheBrokerTimeVerbatim` · `TestA163PricesCarriesTheBrokerTimeInOneRequest` | `SourceTimestamp = ""` — 어댑터가 원문을 버림 |
| 2.6 | `TestA163AMissingOrUnparseableTimeIsAlertedOnceAndRearmsWhenATimeArrives` · `TestA163AHeldPositionThatWasNotJudgedKeepsItsLatch` · `TestA163AClosedPositionLeavesTheLatch` · `TestA163AtMostEightSourceAlertsLeaveInOneCycle` | 알림 0 · 래치 맵 비어 있음 · 주기별 `[0 0 0]` |
| 2.7 | `TestA163TheReportComesAfterEveryJudgementAndBeforeTheUnobservedSettlement` | 보고 없음(`at = -1`) |
| 2.8 | `TestA163NoReportLineOrAlertCarriesTheAccount` | 줄 0(필드 목록을 잴 대상이 없음) |
| 2.9 | `TestA163TheReportNeverBlocksEntryOrEscalates` | 골격에서도 통과(보고가 없으면 막을 것도 없음) — 이빨은 변이 ⑥ |
| 2.10 | `TestA163APanickingReportNeitherStopsTheLoopNorChangesTheStops` | 골격에서도 통과(보고가 패닉할 코드가 없음) — 이빨은 변이 ⑪ |
| 2.11(a) | `TestA163EngineRuntimeWiresTheSourceAgeLogger` | 편집 전부터 실패(red-1) — 배선 키 없음 |
| 2.11(b) | `TestA163TheAssembledObserverWritesItsLinesToTheDedicatedLogger` | 전용 로거 줄 0 |
| 2.12 | `TestA163AnAgeLineIsWrittenOnlyWhenTheBucketChanges` · `TestA163EveryBucketIsClassifiedFromTheReceiptTime` | 줄 0 · 구간 열 빔 |
| 2.13 | `TestA163TheSourceAgeReportToggleDefaultsOff` · `TestA163AnOffToggleLeavesConfigShowByteIdentical` · `TestA163ThePublishedSchemaDeclaresTheToggleOffByDefault` | `explicit true` · `true without automation_gate` 가 false · 켬이 안 보임 · 스키마 속성 없음 |
| 2.14 | `TestA163TheToggleOffWritesNothingAndJudgesExactlyAsOn` | ON 관측자가 줄 · 알림 0 |
| 2.15 | `TestA163TheProductionAssemblyReadsTheLoadedToggle` | 호출자 값이 이김 |
| 2.16 | `TestA163TheReportToggleIsTheLastAuditedSettingAndOnlyItsChangeIsWritten` · `TestA163AStartupAuditsTheLoadedReportToggle` | 끝 행이 `per_trade_risk_budget` · 토글이 false 로 로드됨 |

시험 고침(실패 이유가 틀렸던 것, GREEN 전): `TestA163ThePublishedSchemaDeclaresTheToggleOffByDefault` 는 첫 판이 다른 속성의 문자열 default 를
`*bool` 로 풀다 실패 — `json.RawMessage` 로 고쳐 「속성 없음」 으로 실패하게 함. 줄을 position_id 로 고르던 두 시험은 로거가 그 값을 가림
(`pos-[redacted]`)이라 종목으로 고르게 고침(GREEN 첫 판에서 발견 — 기능 결함 아님, a090 줄도 같음).
