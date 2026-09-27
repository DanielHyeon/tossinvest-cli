# Branch Test Map: `ReconcileDriver.judgeHoldings`

- Source: `internal/app/engine/adoption.go`

> **「진입 실측」은 측정값이다** — 패키지 시험 전체를 `-covermode=set`으로 돌린 프로파일에서 그 분기가 만든 블록의 count다. 어떤 **개별** 시험이 그 분기를 밟는지는 이 실행이 답하지 않는다. 따라서 「Test」 열은 **a095 3판이 요구하는 시험**이며 현존 증명이 아니다. `[비움 — Qn]`은 사용자 결정이 덮지 않아 Manager 에게 올린 질문이다.

| Branch | 조건 | 진입 실측 | Test (a095 요구) | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `:78` `if stale <= 0 {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B2 | `:87` `for _, holding := range snapshot.Holdings {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B3 | `:89` `if market == "" {` | 아니오 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B4 | `:93` `if symbol == "" \|\| market == "" \|\| isZeroQuantity(holding.Quantity) {` | 아니오 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B5 | `:98` `if err != nil {` | 아니오 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B6 | `:104` `if p.State == journal.PositionClosed \|\| isZeroQuantity(p.Quantity) {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B7 | `:108` `if p.ExitEligible() {` | 예 | **a095 3.2** — [비움 — Q3] 엔진 개설 포지션의 수량 증가 검사 | no | no |
| B8 | `:109` `if p.Adopted() {` | 예 | **a095 3.1** — 편입된 포지션만 `checkExternalIncrease`에 온다(R2-B2 삭제의 근거) | no | no |
| B9 | `:116` `if d.blocked(market, symbol) {` | 예 | **a095 2.8** — 전이 상태 무알림 유지 | no | no |
| B10 | `:119` `if !fresh {` | 예 | **a095 2.8** — 전이 상태 무알림 유지 | no | no |
| B11 | `:127` `if d.opts.Adoption.Excludes(symbol) {` | 예 | **a095 2.3** — exclude는 normal, 진입 차단 없음 | no | no |
| B12 | `:135` `if !d.opts.Adoption.Enabled && !d.opts.Adoption.Included(symbol) {` | 예 | **a095 2.4** — off∧미지정은 normal(정본 exit-policy `adoption.enabled` false 동등) | no | no |
| B13 | `:143` `for _, c := range candidates {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B14 | `:144` `if !adopted[c.position.ID] {` | 예 | **a095 2.2** — 편입 시도 실패는 운영자가 고른 상태가 아니다 → critical | no | no |
| B15 | `:152` `for _, p := range unmanaged {` | 예 | **a095 2.2 · 2.3 · 2.4** — 모인 사유별 등급 | no | no |

**미진입 분기 3개**: B3, B4, B5
**자체 블록 없는 분기 0개**: 없음 — 컴파일러가 별도 블록을 만들지 않는 형태(빈 `switch {` 등)이며 미커버와 다르다.
