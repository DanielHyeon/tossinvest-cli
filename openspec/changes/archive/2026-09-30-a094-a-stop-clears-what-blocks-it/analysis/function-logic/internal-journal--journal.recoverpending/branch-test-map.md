# Branch Test Map: `Journal.RecoverPending`

- Source: `internal/journal/recovery.go`

> **「진입 실측」은 측정값이다** — 패키지 시험 전체를 `-covermode=set`으로 돌린 프로파일에서 그 분기가 만든 블록의 count다. 어떤 **개별** 시험이 그 분기를 밟는지는 이 실행이 답하지 않는다. 따라서 「Test」 열은 **a094가 요구하는 시험**이며 현존 증명이 아니다.

| Branch | 조건 | 진입 실측 | Test (a094 요구) | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `:88` `if err != nil {` | 아니오 | 기존 | no | no |
| B2 | `:93` `for _, rec := range pending {` | 예 | 기존 | no | no |
| B3 | `:94` `switch rec.State {` | — | 기존 | no | no |
| B4 | `:95` `case StateRecorded:` | 예 | **a094 4.4a** — 세션 중 경로가 이 분기를 밟지 않는다 | no | no |
| B5 | `:97` `if err := handle.Settle(ctx, StateNotDispatched, ReasonRestartNotDispatched,` | — | 기존 | no | no |
| B6 | `:103` `case StateDispatchStarted:` | 예 | **a094 4.4a** — 같음 | no | no |
| B7 | `:105` `if err := handle.MarkInDoubt(ctx, ReasonRestartInDoubt,` | — | 기존 | no | no |
| B8 | `:111` `if err != nil {` | 아니오 | 기존 | no | no |
| B9 | `:116` `case StateAcked, StateInDoubt:` | 예 | **a094 4.4** — 재시작 순회 무변화 | no | no |
| B10 | `:118` `if err != nil {` | 아니오 | 기존 | no | no |

**미진입 분기 3개**: B1, B8, B10
**자체 블록 없는 분기 3개**: B3, B5, B7 — 컴파일러가 별도 블록을 만들지 않는 형태이며 미커버와 다르다.
