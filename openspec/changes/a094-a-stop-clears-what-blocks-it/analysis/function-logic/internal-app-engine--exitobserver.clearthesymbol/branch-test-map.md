# Branch Test Map: `ExitObserver.clearTheSymbol`

- Source: `internal/app/engine/exitloop.go`

> **「진입 실측」은 측정값이다** — 패키지 시험 전체를 `-covermode=set`으로 돌린 프로파일에서 그 분기가 만든 블록의 count다. 어떤 **개별** 시험이 그 분기를 밟는지는 이 실행이 답하지 않는다(시험별 프로파일이 필요하다). 따라서 「Test」 열은 **a094가 요구하는 시험**이며 현존 증명이 아니다.

| Branch | 조건 | 진입 실측 | Test (a094 요구) | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `:1337` `if err != nil {` | 아니오 | **a094 3.7** — 목록 조회 실패가 손절을 보류시키지 않는다 | no | no |
| B2 | `:1341` `for _, order := range live {` | 예 | **a094 3.1·3.4** — 저널 ∪ 브로커, dedup | no | no |
| B3 | `:1343` `if !buy && !withPending {` | 아니오 | **a094 3.1** — 저널에 없는 매수도 걸린다 | no | no |
| B4 | `:1358` `if err != nil {` | 아니오 | a094 3.3 | no | no |
| B5 | `:1364` `if qerr != nil \|\| perr != nil {` | 아니오 | 기존 | no | no |
| B6 | `:1379` `if err != nil \|\| out.State != journal.StateConfirmed {` | 예 | **a094 3.2·3.3** — 확정 취소 후에만 | no | no |
| B7 | `:1383` `if !clear {` | 예 | **a094 3.3** — 못 치우면 제출 안 함 | no | no |
| B8 | `:1386` `if withPending && m.state.Pending() {` | 예 | 기존 | no | no |
| B9 | `:1387` `if err := o.release(ctx, m, journal.ProposalCancelled); err != nil {` | 아니오 | 기존 | no | no |

**미진입 분기 5개**: B1, B3, B4, B5, B9
**자체 블록 없는 분기 0개**: 없음 — 컴파일러가 별도 블록을 만들지 않는 형태(빈 `default:` 등)이며 미커버와 다르다.
