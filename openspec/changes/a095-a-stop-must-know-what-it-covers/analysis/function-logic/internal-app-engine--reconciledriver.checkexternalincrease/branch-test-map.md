# Branch Test Map: `ReconcileDriver.checkExternalIncrease`

- Source: `internal/app/engine/adoption.go`

> **「진입 실측」은 측정값이다** — 네 패키지의 시험 전체를 `-covermode=set`으로 돌린 프로파일에서 그 분기가 만든 블록의 count다. 어떤 **개별** 시험이 그 분기를 밟는지는 이 실행이 답하지 않는다(시험별 프로파일이 필요하다). 따라서 「Test」 열은 **a095가 요구하는 시험**이며 현존 증명이 아니다.

| Branch | 조건 | 진입 실측 | Test (a095 요구) | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `:442` `if d.grown[p.ID] {` | 아니오 | **a095 2.4** — 한 프로세스에서 한 번만 우는 것을 고정하고, 그 한 번이 durable해야 한다 | no | no |
| B2 | `:446` `if err != nil {` | 아니오 | **a095 2.3** — 편입 기록이 없는 포지션에서 조용히 반환하는 것이 실측으로 미진입이다 | no | no |
| B3 | `:450` `if err != nil \|\| cmp <= 0 {` | 예 | **a095 2.1** — 수량 증가가 알림으로 이어진다 | no | no |

**미진입 분기 2개**: B1, B2
**자체 블록 없는 분기 0개**: 없음 — 컴파일러가 별도 블록을 만들지 않는 형태(빈 `switch {` 등)이며 미커버와 다르다.
