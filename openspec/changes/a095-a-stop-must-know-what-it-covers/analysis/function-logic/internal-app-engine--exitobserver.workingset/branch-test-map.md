# Branch Test Map: `ExitObserver.workingSet`

- Source: `internal/app/engine/exitloop.go`

> **「진입 실측」은 측정값이다** — 패키지 시험 전체를 `-covermode=set`으로 돌린 프로파일에서 그 분기가 만든 블록의 count다. 어떤 **개별** 시험이 그 분기를 밟는지는 이 실행이 답하지 않는다. 따라서 「Test」 열은 **a095 3판이 요구하는 시험**이며 현존 증명이 아니다. `[비움 — Qn]`은 사용자 결정이 덮지 않아 Manager 에게 올린 질문이다.

| Branch | 조건 | 진입 실측 | Test (a095 요구) | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `:495` `if err != nil {` | 아니오 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B2 | `:499` `if err != nil {` | 아니오 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B3 | `:503` `for _, result := range stateResults {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B4 | `:508` `for _, p := range positions {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B5 | `:509` `if p.State == journal.PositionClosed \|\| isZeroQuantity(p.Quantity) {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B6 | `:512` `if !p.ExitEligible() {` | 예 | **a095 2.1** — 이 자리의 사실은 normal · `TestAPositionWithNoEntryDecisionIsSkippedAndAlertedOnce`가 이미 normal을 고정한다(`exitloop_test.go:508`) | no | no |
| B7 | `:525` `if !ok {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B8 | `:527` `if err != nil {` | 아니오 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B9 | `:528` `if cycle.Err == nil {` | 아니오 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B10 | `:533` `if opened.PositionID == "" {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B11 | `:542` `if result.Corruption != nil {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B12 | `:545` `if qerr != nil {` | 아니오 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B13 | `:546` `if cycle.Err == nil {` | 아니오 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B14 | `:556` `if q, active, qerr := o.opts.Journal.ActiveExitSnapshotQuarantine(ctx, p.ID, p.InstanceSeq); qerr != nil {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B15 | `:561` `} else if active && !q.NeedsReJudgement() {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B16 | `:557` `if cycle.Err == nil {` | 아니오 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B17 | `:561` `} else if active && !q.NeedsReJudgement() {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B18 | `:565` `} else if active {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B19 | `:565` `} else if active {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B20 | `:589` `if identityErr != nil {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B21 | `:592` `if qerr != nil {` | 아니오 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B22 | `:593` `if cycle.Err == nil {` | 아니오 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |

**미진입 분기 9개**: B1, B2, B8, B9, B12, B13, B16, B21, B22
**자체 블록 없는 분기 0개**: 없음 — 컴파일러가 별도 블록을 만들지 않는 형태(빈 `switch {` 등)이며 미커버와 다르다.
