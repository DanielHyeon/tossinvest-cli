# Branch Test Map: `EvaluateRatchet`

- Source: `internal/exitpolicy/ratchet.go`

> **「진입 실측」은 측정값이다** — 패키지 시험 전체를 `-covermode=set`으로 돌린 프로파일에서 그 분기가 만든 블록의 count다. 어떤 **개별** 시험이 그 분기를 밟는지는 이 실행이 답하지 않는다. 따라서 「Test」 열은 **a094가 요구하는 시험**이며 현존 증명이 아니다.

| Branch | 조건 | 진입 실측 | Test (a094 요구) | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `:337` `if in.Config != nil {` | 예 | 기존 | no | no |
| B2 | `:340` `if err := cfg.Validate(); err != nil {` | 예 | 기존 | no | no |
| B3 | `:344` `if err != nil {` | 예 | 기존 | no | no |
| B4 | `:348` `if err != nil {` | 예 | 기존 | no | no |
| B5 | `:352` `if err != nil {` | 예 | 기존 | no | no |
| B6 | `:356` `if err != nil {` | 예 | 기존 | no | no |
| B7 | `:360` `if err != nil {` | 예 | 기존 | no | no |
| B8 | `:364` `if err != nil {` | 예 | 기존 | no | no |
| B9 | `:371` `if observed.Cmp(probe) > 0 {` | 예 | 기존 | no | no |
| B10 | `:378` `if err != nil {` | 아니오 | 기존 | no | no |
| B11 | `:394` `if levelCandidate != "" {` | 예 | 기존 | no | no |
| B12 | `:400` `if level.Rank() >= LevelBreakeven.Rank() {` | 예 | 기존 | no | no |
| B13 | `:405` `if err != nil {` | 아니오 | 기존 | no | no |
| B14 | `:409` `if err != nil {` | 아니오 | 기존 | no | no |
| B15 | `:419` `if err != nil {` | 아니오 | 기존 | no | no |
| B16 | `:422` `if observed.Cmp(baseline) < 0 {` | 예 | **a094 4.1** | no | no |
| B17 | `:423` `if in.PendingAction == ActionBaselineBreach {` | 예 | **a094 4.1·4.2** — RATCHET에서도 해제 후 발의가 난다 | no | no |
| B18 | `:438` `if wantsPartial {` | 예 | 기존 | no | no |
| B19 | `:439` `switch {` | — | 기존 | no | no |
| B20 | `:440` `case taken.Sign() > 0:` | 예 | 기존 | no | no |
| B21 | `:442` `case in.PendingAction != ActionNone:` | 예 | 기존 | no | no |
| B22 | `:444` `default:` | 예 | 기존 | no | no |

**미진입 분기 4개**: B10, B13, B14, B15
**자체 블록 없는 분기 1개**: B19 — 컴파일러가 별도 블록을 만들지 않는 형태이며 미커버와 다르다.
