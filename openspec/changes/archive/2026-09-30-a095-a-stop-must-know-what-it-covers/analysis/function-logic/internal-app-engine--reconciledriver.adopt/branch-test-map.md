# Branch Test Map: `ReconcileDriver.adopt`

- Source: `internal/app/engine/adoption.go`

> **「진입 실측」은 측정값이다** — 패키지 시험 전체를 `-covermode=set`으로 돌린 프로파일에서 그 분기가 만든 블록의 count다. 어떤 **개별** 시험이 그 분기를 밟는지는 이 실행이 답하지 않는다. 따라서 「Test」 열은 **a095 3판이 요구하는 시험**이며 현존 증명이 아니다. `[비움 — Qn]`은 사용자 결정이 덮지 않아 Manager 에게 올린 질문이다.

| Branch | 조건 | 진입 실측 | Test (a095 요구) | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `:213` `if len(candidates) == 0 {` | 예 | 분기 무변화 — 결과 형태는 `TestA095Q2FactsStayNormalInTheirOwnCells` · `TestA095OneCycleCanHoldAFailureAndADeferral`가 잰다 | yes | yes |
| B2 | `:218` `if err != nil {` | 예 | 분기 무변화 — 결과 형태는 `TestA095Q2FactsStayNormalInTheirOwnCells` · `TestA095OneCycleCanHoldAFailureAndADeferral`가 잰다 | yes | yes |
| B3 | `:220` `if cycle.Err == nil {` | 예 | 분기 무변화 — 결과 형태는 `TestA095Q2FactsStayNormalInTheirOwnCells` · `TestA095OneCycleCanHoldAFailureAndADeferral`가 잰다 | yes | yes |
| B4 | `:227` `if bound <= 0 {` | 예 | 분기 무변화 — 결과 형태는 `TestA095Q2FactsStayNormalInTheirOwnCells` · `TestA095OneCycleCanHoldAFailureAndADeferral`가 잰다 | yes | yes |
| B5 | `:230` `for _, c := range candidates {` | 예 | 분기 무변화 — 결과 형태는 `TestA095Q2FactsStayNormalInTheirOwnCells` · `TestA095OneCycleCanHoldAFailureAndADeferral`가 잰다 | yes | yes |
| B6 | `:233` `if !ok {` | 예 | 분기 무변화 — 결과 형태는 `TestA095Q2FactsStayNormalInTheirOwnCells` · `TestA095OneCycleCanHoldAFailureAndADeferral`가 잰다 | yes | yes |
| B7 | `:239` `if age := d.clk.Now().Sub(readAt); age > bound {` | 예 | 분기 무변화 — 결과 형태는 `TestA095Q2FactsStayNormalInTheirOwnCells` · `TestA095OneCycleCanHoldAFailureAndADeferral`가 잰다 | yes | yes |
| B8 | `:248` `if d.adoptOne(ctx, c, observed) {` | 예 | 분기 무변화 — 결과 형태는 `TestA095Q2FactsStayNormalInTheirOwnCells` · `TestA095OneCycleCanHoldAFailureAndADeferral`가 잰다 | yes | yes |

**미진입 분기 0개**: 없음
**자체 블록 없는 분기 0개**: 없음 — 컴파일러가 별도 블록을 만들지 않는 형태(빈 `switch {` 등)이며 미커버와 다르다.
