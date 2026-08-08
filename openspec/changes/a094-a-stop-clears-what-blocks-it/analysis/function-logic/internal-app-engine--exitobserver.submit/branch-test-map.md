# Branch Test Map: `ExitObserver.submit`

- Source: `internal/app/engine/exitloop.go`

> **「진입 실측」은 측정값이다** — 패키지 시험 전체를 `-covermode=set`으로 돌린 프로파일에서 그 분기가 만든 블록의 count다. 어떤 **개별** 시험이 그 분기를 밟는지는 이 실행이 답하지 않는다(시험별 프로파일이 필요하다). 따라서 「Test」 열은 **a094가 요구하는 시험**이며 현존 증명이 아니다.

| Branch | 조건 | 진입 실측 | Test (a094 요구) | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `:1240` `if err != nil {` | 아니오 | 기존 | no | no |
| B2 | `:1243` `if isZeroQuantity(submitQuantity) {` | 예 | 기존 (a091이 다룸) | no | no |
| B3 | `:1263` `if err != nil {` | 아니오 | 기존 | no | no |
| B4 | `:1272` `if err := o.opts.Journal.AttachExitIntent(ctx, m.position.ID, intentID); err != nil {` | 예 | 기존 | no | no |
| B5 | `:1277` `if err != nil {` | 아니오 | 기존 | no | no |
| B6 | `:1287` `switch {` | — | **a094 2.9** | no | no |
| B7 | `:1288` `case out.State == journal.StateConfirmed:` | 예 | 기존 | no | no |
| B8 | `:1296` `case out.State == journal.StateInDoubt \|\| out.State == journal.StateUnresolvedInDoubt:` | 예 | **a094 2.1** — 409가 더 이상 여기로 오지 않는다 | no | no |
| B9 | `:1301` `case out.Reason == execgw.ReasonSymbolInFlight:` | 아니오 | **a094 6.2** — 272210 라이브락 | no | no |
| B10 | `:1304` `default:` | 예 | **a094 2.9** — 409가 이제 여기로 온다 | no | no |
| B11 | `:1306` `if detail == "" && err != nil {` | 아니오 | 기존 | no | no |

**미진입 분기 5개**: B1, B3, B5, B9, B11
**자체 블록 없는 분기 1개**: B6 — 컴파일러가 별도 블록을 만들지 않는 형태(빈 `default:` 등)이며 미커버와 다르다.
