# Branch Test Map: `Notifier.claimAndDeliver`

- Source: `internal/obs/notifier.go`; **편집 전** 측정 — `analysis/harness/coverage-pre-unit3.json`(연결 워크트리 `b3f14925`, `./internal/obs` 시험 92개를 하나씩).

| Branch | AST anchor | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | if at 263:2 | claim 실패 | `TestAClaimThatFailsAttemptsTheDurableBlock`, `TestAClaimThatFailsBlocksNewEntries` | 편집 전(기준선) | 블록 263.16-276.19을 시험 4개가 실행, PASS |
| B2 | if at 276:3 | 로그 | `TestAClaimThatFailsAttemptsTheDurableBlock`, `TestAClaimThatFailsBlocksNewEntries` | 편집 전(기준선) | 블록 276.19-278.4을 시험 3개가 실행, PASS |
| B3 | if at 279:3 | 래치 | `TestAClaimThatFailsAttemptsTheDurableBlock`, `TestAClaimThatFailsBlocksNewEntries` | 편집 전(기준선) | 블록 279.20-281.4을 시험 3개가 실행, PASS |
| B4 | switch at 284:2 | claim 결과 분기 | `TestA092BothAnnouncersBuildTheSameEvent`, `TestA092EachTransitionIsItsOwnAnnouncement` | 편집 전(기준선) | 블록 284.2-284.27을 시험 44개가 실행, PASS |
| B5 | case at 285:2 | 이미 정착 | `TestConcurrentObservationsOfOneConditionSendOnce`, `TestNotifierIsConcurrencySafe` | 편집 전(기준선) | 블록 285.28-293.27을 시험 6개가 실행, PASS |
| B6 | case at 294:2 | 남의 임차 | `TestAHeldRowIsNotWhispered`, `TestALeaseLineCarriesItsOwnName` | 편집 전(기준선) | 블록 294.34-305.27을 시험 6개가 실행, PASS |
| B7 | if at 310:2 | deliver 가 임차를 잃음 | `TestARowThatVanishedIsNotReportedAsContention`, `TestASenderThatLosesTheLeaseStopsAtOnce` | 편집 전(기준선) | 블록 310.10-315.3을 시험 2개가 실행, PASS |
