# Branch Test Map: `ProjectPosition` (편집 전 판)

- Source: `internal/journal/position_projection.go` (65-105), base `884be2a74`.
- 커버 측정: `../../coverage/pre-edit-ProjectPosition-blocks.txt`(선택 시험 8 패턴, `-covermode=set`, PASS).
- 시험 이름은 읽어서 그 분기를 지남을 확인한 것만 적는다. RED/GREEN 열은 구현 단계(2.2 · 2.4)에서 갱신.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | resolveFillOrigin 오류 | 없음(grep) | 해당 없음(기존 분기) | 미실행 — 67.16 hit=0 |
| B2 | 소유 intent 없음 | 없음 — RecordFill 은 소유 확정일 때만 훅을 부름(fills.go:550) | 해당 없음(기존 분기) | 미실행 — 70.12 hit=0 |
| B3 | currentInstance 오류 | 없음(grep) | 해당 없음(기존 분기) | 미실행 — 79.16 hit=0 |
| B4 | position.Apply 오류 | 없음(grep) | 해당 없음(기존 분기) | 미실행 — 94.16 hit=0 |
| B5 | 거절 → 차단(편집 지점) | `TestAnOversellFreezesTheProjectionAndBlocksTheSymbol` · `TestASellWithNothingHeldIsAttributionFailure` · `TestA200CurrentBehaviourAConvergenceCloseThenTheStopFillIsAnAttributionFailure` · `TestA200CurrentBehaviourAStopFillStraddlingTheConvergenceIsAnAttributionFailure` | 구현 2.2 에서 — 본 사고 모양 시험이 RED(차단 0 기대) | 편집 전 101.25 hit=1 |
| (return@104) | 허용 전이 | `TestTheFirstFillOpensTheInstance` | 해당 없음(기존 분기) | 편집 전 104.2 hit=1 |
