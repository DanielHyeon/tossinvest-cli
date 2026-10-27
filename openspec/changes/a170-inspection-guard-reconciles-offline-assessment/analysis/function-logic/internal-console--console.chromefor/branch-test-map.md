# Branch Test Map: `Console.chromeFor`

- Source: `internal/console/chrome.go` (184-247) — 편집 뒤. AST branches 16.
- 측정: 커버리지 시험 바이너리(`go test -c -cover`)로 top-level 시험 607 개를 하나씩 `-test.run '^<Test>$' -test.coverprofile` 실행, 분기 본문 블록 count>0 인 시험만 인용(2026-10-25, a173 close). n 은 그 블록을 지나는 시험 수.

| Branch | Anchor | Scenario | Test | 측정 n | a173 |
|---|---|---|---|---|---|
| B1 | `188:2` | Binary seam 배선 → 설치 stamp 읽기 | `TestStatusStripCellsAreOneLine` | 485 | 무변경 |
| B2 | `195:2` | 엔진 상태 switch | `TestStatusStripCellsAreOneLine` | 485 | 무변경 |
| B3 | `196:2` | !Wired → 미배선 | `TestStatusStripCellsAreOneLine` · `TestEachUnmeasuredCauseIsExplainedOnce` | 465 | 무변경 |
| B4 | `200:2` | Running → 실행 중 | `TestA111ActualPositionsAndPositionManagementRoutesRenderTheSameFreshLine` | 14 | 무변경 |
| B5 | `203:3` | PID>0 → PID 표기 | `TestA111ActualPositionsAndPositionManagementRoutesRenderTheSameFreshLine` | 14 | 무변경 |
| B6 | `206:3` | Stale → 다른 실행 파일 경고 | `TestTheStaleEngineBinaryIsWarnedAbout` | 1 | 무변경 |
| B7 | `209:2` | default → 정지 | `TestA111PositionManagementKeepsJournalFailureTruth` | 17 | 무변경 |
| B8 | `216:2` | 데이터 Mode switch | `TestStatusStripCellsAreOneLine` | 485 | 무변경 |
| B9 | `217:2` | dataCache → 시각(n초 전) | `TestStatusStripCellsAreOneLine` | 143 | 무변경 |
| B10 | `219:2` | dataUnavailable → `읽지 못함`, `DataTitle`=Reason | `TestStatusStripCellsAreOneLine` · `TestEachUnmeasuredCauseIsExplainedOnce` | 483 | a173: Reason 을 본문에서 title 로 |
| B11 | `221:3` | Code 있음 → ` · <코드>` | `TestStatusStripCellsAreOneLine` · `TestEachUnmeasuredCauseIsExplainedOnce` | 483 | a173 추가 |
| B12 | `225:3` | LastAt 있음 → `마지막 성공 …` | `TestTheOrdersCountsCarryTheirOwnProvenanceInTheSameBreath` | 1 | a173: Reason 이어붙이기 대신 단독 문구 |
| B13 | `228:2` | default(요청 시 읽음) → 렌더 시각 | `TestStatusStripCellsAreOneLine` | 291 | 무변경 |
| B14 | `234:2` | Hold 있음 → `갱신 보류`, `DataTitle`=Hold | `TestAKnownRefreshHoldGetsAReasonInsteadOfATone` | 26 | a173: 긴 문구를 title 로 |
| B15 | `239:2` | 검증 실행 있음 | `TestAPendingApprovalIsVisibleFromEveryScreen` | 29 | 무변경 |
| B16 | `240:3` | 승인 대기 → PendingApproval · 링크 | `TestAPendingApprovalIsVisibleFromEveryScreen` | 13 | 무변경 |
