# Branch Test Map: `ExitObserver.record`

- Source: `internal/app/engine/exitloop.go`

> **「진입 실측」은 측정값이다** — 패키지 시험 전체를 `-covermode=set`으로 돌린 프로파일에서 그 분기가 만든 블록의 count다. 어떤 **개별** 시험이 그 분기를 밟는지는 이 실행이 답하지 않는다(시험별 프로파일이 필요하다). 따라서 「Test」 열은 **a094가 요구하는 시험**이며 현존 증명이 아니다.

| Branch | 조건 | 진입 실측 | Test (a094 요구) | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `:1095` `if quote.FetchedAt.IsZero() {` | 예 | 기존 — a094는 이 함수를 바꾸지 않는다 | no | no |
| B2 | `:1097` `} else {` | 예 | 기존 — a094는 이 함수를 바꾸지 않는다 | no | no |
| B3 | `:1117` `if orderable && (snapshot.CancelPendingFirst \|\| isFullExit(proposal)) {` | 예 | **a094 3.1** — 게이트 조건 무변화를 고정한다 | no | no |
| B4 | `:1118` `if m.reJudge && !isProtective(proposal) {` | 예 | 기존 — a094는 이 함수를 바꾸지 않는다 | no | no |
| B5 | `:1140` `} else {` | 예 | **a094 3.1·3.7** | no | no |
| B6 | `:1142` `if err != nil {` | 아니오 | 기존 — a094는 이 함수를 바꾸지 않는다 | no | no |
| B7 | `:1145` `if !cleared {` | 예 | 기존 — a094는 이 함수를 바꾸지 않는다 | no | no |
| B8 | `:1149` `} else {` | 예 | 기존 — a094는 이 함수를 바꾸지 않는다 | no | no |
| B9 | `:1156` `if orderable {` | 예 | 기존 — a094는 이 함수를 바꾸지 않는다 | no | no |
| B10 | `:1158` `if intentID == "" {` | 아니오 | 기존 — a094는 이 함수를 바꾸지 않는다 | no | no |
| B11 | `:1170` `if err != nil {` | 예 | 기존 — a094는 이 함수를 바꾸지 않는다 | no | no |
| B12 | `:1171` `if errors.Is(err, journal.ErrProposalPending) {` | 예 | 기존 — a094는 이 함수를 바꾸지 않는다 | no | no |
| B13 | `:1177` `if errors.Is(err, journal.ErrExitSnapshotQuarantined) {` | 예 | 기존 — a094는 이 함수를 바꾸지 않는다 | no | no |
| B14 | `:1190` `if recorded.ArmedProposal == nil \|\| recorded.ArmOutcome != journal.ExitArmArmed {` | 예 | 기존 — a094는 이 함수를 바꾸지 않는다 | no | no |

**미진입 분기 2개**: B6, B10
**자체 블록 없는 분기 0개**: 없음 — 컴파일러가 별도 블록을 만들지 않는 형태(빈 `default:` 등)이며 미커버와 다르다.
