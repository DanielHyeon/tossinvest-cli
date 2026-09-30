# Branch Test Map: `ReconcileDriver.judgeHoldings`

- Source: `internal/app/engine/adoption.go`

> **「진입 실측」은 측정값이다** — 패키지 시험 전체를 `-covermode=set`으로 돌린 프로파일에서 그 분기가 만든 블록의 count다. 어떤 **개별** 시험이 그 분기를 밟는지는 이 실행이 답하지 않는다. 따라서 「Test」 열은 **a095 3판이 요구하는 시험**이며 현존 증명이 아니다. `[비움 — Qn]`은 사용자 결정이 덮지 않아 Manager 에게 올린 질문이다.

| Branch | 조건 | 진입 실측 | Test (a095 요구) | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `:110` `if stale <= 0 {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | yes | yes |
| B2 | `:119` `for _, holding := range snapshot.Holdings {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | yes | yes |
| B3 | `:121` `if market == "" {` | 아니오 | 기존 — a095는 이 함수를 바꾸지 않는다 | yes | yes |
| B4 | `:125` `if symbol == "" \|\| market == "" \|\| isZeroQuantity(holding.Quantity) {` | 아니오 | 기존 — a095는 이 함수를 바꾸지 않는다 | yes | yes |
| B5 | `:130` `if err != nil {` | 아니오 | 기존 — a095는 이 함수를 바꾸지 않는다 | yes | yes |
| B6 | `:136` `if p.State == journal.PositionClosed \|\| isZeroQuantity(p.Quantity) {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | yes | yes |
| B7 | `:140` `if p.ExitEligible() {` | 예 | **a095 3.1 · 3.2** — 적격이면 수량 증가만 검사하고 `continue`(아래 B8 · B9) | yes | yes |
| B8 | `:143` `if p.Adopted() {` | 예 | **a095 3.1 · 3.4** `TestA095AnAdoptionLookupFailureStaysSilent` · `TestA095TheAdoptedGrowthReplayReportsEveryNewMaximum` | yes | yes |
| B9 | `:145` `} else {` | 예 | **a095 3.2 (10판 새 갈래)** `TestA095AnEngineOpenedPositionThatGrewIsReported` · `TestA095AnAdjustmentHistoryThatNetsToZeroIsNotAnIncrease` | yes | yes |
| B10 | `:152` `if d.blocked(market, symbol) {` | 예 | **a095 2.8** — RECONCILE 무알림(기존 `TestAdoptionIsSilentUnderReconcile`) | yes | yes |
| B11 | `:155` `if !fresh {` | 예 | **a095 2.8** `TestA095AStaleSnapshotStaysSilent` | yes | yes |
| B12 | `:163` `if d.opts.Adoption.Excludes(symbol) {` | 예 | **a095 2.3** `TestA095OperatorChosenStatesStayNormal` | yes | yes |
| B13 | `:171` `if !d.opts.Adoption.Enabled && !d.opts.Adoption.Included(symbol) {` | 예 | **a095 2.4** `TestA095OperatorChosenStatesStayNormal` | yes | yes |
| B14 | `:179` `for _, c := range candidates {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | yes | yes |
| B15 | `:180` `if results[c.position.ID].result != adoptAdopted {` | 예 | **a095 2.2 · 2.6 · 2.12** `TestA095AFailedAdoptionIsCriticalAndDurable` · `TestA095Q2FactsStayNormalInTheirOwnCells` · `TestA095OneCycleCanHoldAFailureAndADeferral` | yes | yes |
| B16 | `:188` `for _, p := range unmanaged {` | 예 | **a095 2.2 · 2.3 · 2.4** — 모인 사유별 등급(위 시험들) | yes | yes |

**미진입 분기 3개**: B3, B4, B5
**자체 블록 없는 분기 0개**: 없음 — 컴파일러가 별도 블록을 만들지 않는 형태(빈 `switch {` 등)이며 미커버와 다르다.
