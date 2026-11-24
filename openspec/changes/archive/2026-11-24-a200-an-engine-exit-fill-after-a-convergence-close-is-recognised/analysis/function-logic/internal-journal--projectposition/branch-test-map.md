# Branch Test Map: `ProjectPosition` (편집 뒤 판)

- Source: `internal/journal/position_projection.go` (65-110), 작업 트리. 재번호: B1~B5 유지, 새 B6.
- 커버 측정: 편집 전 `../../coverage/pre-edit-ProjectPosition-blocks.txt`, 편집 뒤 `../../coverage/after-ProjectPosition-blocks.txt`.
- 시험 이름은 읽어서 그 분기를 지남을 확인한 것만 적는다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | resolveFillOrigin 오류 | 없음(grep) | 해당 없음(기존 분기) | 미실행 — 67.16 hit=0(편집 전도 0) |
| B2 | 소유 intent 없음 | 없음 — RecordFill 은 소유 확정일 때만 훅을 부름(fills.go:550) | 해당 없음(기존 분기) | 미실행 — 70.12 hit=0(편집 전도 0) |
| B3 | currentInstance 오류 | 없음(grep) | 해당 없음(기존 분기) | 미실행 — 79.16 hit=0(편집 전도 0) |
| B4 | position.Apply 오류 | 없음(grep) | 해당 없음(기존 분기) | 미실행 — 94.16 hit=0(편집 전도 0) |
| B5 | 거절 → 인정 아님 → 차단 | `TestAnOversellFreezesTheProjectionAndBlocksTheSymbol` · `TestASellWithNothingHeldIsAttributionFailure` · `TestA200AFillFromAnotherIntentStaysAnAttributionFailure` · `TestA200AnAmendedStopStaysAnAttributionFailure` · `TestA200AManualKindCloseStaysAnAttributionFailure` · `TestA200CurrentBehaviourAStopFillStraddlingTheConvergenceIsAnAttributionFailure` | 편집 전에도 통과(현재 동작 고정 — `analysis/tdd/red.txt`) | yes — 101.25 hit=1 |
| B6 | 인정 → 표지 · nil | `TestA200AStopFillAfterAConvergenceCloseIsRecognised` · `TestA200AnAdoptedStopFillIsRecognisedWithoutAnOutcome` · `TestA200PartialStopFillsAfterAConvergenceCloseAreEachRecognised` · `TestA200APartialTakeProfitBelowTheDecreaseIsRecognised` · `TestA200W4AManagedStopFillLeavesNoBlock` | yes — 넷 모두 RED 에서 실패(`analysis/tdd/red.txt`) | yes — 104.116 hit=1 |
| (return@109) | 허용 전이 | `TestTheFirstFillOpensTheInstance` | 해당 없음(기존 분기) | yes — 109.2 hit=1 |
