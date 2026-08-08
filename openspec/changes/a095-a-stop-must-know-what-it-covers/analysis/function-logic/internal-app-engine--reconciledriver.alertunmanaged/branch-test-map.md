# Branch Test Map: `ReconcileDriver.alertUnmanaged`

- Source: `internal/app/engine/adoption.go`

> **「진입 실측」은 측정값이다** — 네 패키지의 시험 전체를 `-covermode=set`으로 돌린 프로파일에서 그 분기가 만든 블록의 count다. 어떤 **개별** 시험이 그 분기를 밟는지는 이 실행이 답하지 않는다(시험별 프로파일이 필요하다). 따라서 「Test」 열은 **a095가 요구하는 시험**이며 현존 증명이 아니다.

| Branch | 조건 | 진입 실측 | Test (a095 요구) | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `:393` `if d.unmanaged[p.ID] {` | 예 | **a095 2.4** — 프로세스당 1회 억제가 durable 알림과 어떻게 맞물리는지 | no | no |
| B2 | `:404` `switch {` | — | 기존 — 사유 분기는 바꾸지 않는다 | no | no |
| B3 | `:405` `case d.opts.Adoption.Rejected != "":` | 예 | 기존 — 사유 분기는 바꾸지 않는다 | no | no |
| B4 | `:407` `case d.opts.Adoption.Excludes(p.Symbol):` | 예 | 기존 — 사유 분기는 바꾸지 않는다 | no | no |
| B5 | `:409` `case d.opts.Adoption.Enabled:` | 예 | 기존 — 사유 분기는 바꾸지 않는다 | no | no |
| B6 | `:412` `case d.opts.Adoption.Included(p.Symbol):` | 예 | 기존 — 사유 분기는 바꾸지 않는다 | no | no |

**미진입 분기 0개**: 없음
**자체 블록 없는 분기 1개**: B2 — 컴파일러가 별도 블록을 만들지 않는 형태(빈 `switch {` 등)이며 미커버와 다르다.
