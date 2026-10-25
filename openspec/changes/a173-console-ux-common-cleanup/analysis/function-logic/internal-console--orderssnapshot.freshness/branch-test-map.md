# Branch Test Map: `ordersSnapshot.freshness`

- Source: `internal/console/chrome.go` (287-305) — 편집 뒤. AST branches 6.
- 측정: 커버리지 시험 바이너리(`go test -c -cover`)로 top-level 시험 607 개를 하나씩 `-test.run '^<Test>$' -test.coverprofile` 실행, 분기 본문 블록 count>0 인 시험만 인용(2026-10-25, a173 close). n 은 그 블록을 지나는 시험 수.

| Branch | Anchor | Scenario | Test | 측정 n | a173 |
|---|---|---|---|---|---|
| B1 | `288:2` | 상태 switch | `TestStatusStripCellsAreOneLine` | 51 | 무변경 |
| B2 | `289:2` | !Wired → unavailable · seam_unwired | `TestStatusStripCellsAreOneLine` · `TestEachUnmeasuredCauseIsExplainedOnce` | 23 | a173: Code 추가 |
| B3 | `291:2` | Error → unavailable · broker_read_failed | `TestAWholeSeamFailureIsUnmeasuredAndNotAnEmptyAccount` · `TestEachUnmeasuredCauseIsExplainedOnce` | 6 | a173: Code 추가 |
| B4 | `293:3` | Error 이면서 Present → LastAt | `TestTheOrdersCountsCarryTheirOwnProvenanceInTheSameBreath` | 1 | 무변경 |
| B5 | `297:2` | !Present → unavailable · never_fetched | `TestTheOrdersRefreshYieldsToALiveVerification` | 1 | a173: Code 추가 |
| B6 | `301:2` | Held → Hold 문구 | 측정상 없음(기존 공백) | 0 | 무변경 — 측정 시험 0(기존 공백, a173 무관) |
