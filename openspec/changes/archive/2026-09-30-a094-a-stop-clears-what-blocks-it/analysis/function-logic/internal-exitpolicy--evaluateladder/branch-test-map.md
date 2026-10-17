# Branch Test Map: `EvaluateLadder`

- Source: `internal/exitpolicy/ladder.go`

> **「진입 실측」은 측정값이다** — 패키지 시험 전체를 `-covermode=set`으로 돌린 프로파일에서 그 분기가 만든 블록의 count다. 어떤 **개별** 시험이 그 분기를 밟는지는 이 실행이 답하지 않는다. 따라서 「Test」 열은 **a094가 요구하는 시험**이며 현존 증명이 아니다.

| Branch | 조건 | 진입 실측 | Test (a094 요구) | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `:308` `if err := in.Policy.Validate(); err != nil {` | 예 | 기존 | no | no |
| B2 | `:311` `if in.State.PolicyID != in.Policy.PolicyID {` | 예 | 기존 | no | no |
| B3 | `:318` `if err != nil {` | 아니오 | 기존 | no | no |
| B4 | `:321` `if in.State.PolicyVersion != "" && in.State.PolicyVersion != identity.Version {` | 예 | 기존 | no | no |
| B5 | `:325` `if in.State.PolicyDigest != "" && in.State.PolicyDigest != identity.Digest {` | 예 | 기존 | no | no |
| B6 | `:330` `if err != nil {` | 아니오 | 기존 | no | no |
| B7 | `:334` `if err != nil {` | 아니오 | 기존 | no | no |
| B8 | `:338` `if err != nil {` | 아니오 | 기존 | no | no |
| B9 | `:342` `if err != nil {` | 아니오 | 기존 | no | no |
| B10 | `:345` `if _, err := fraction("taken ratio total", in.State.TakenRatioTotal); err != nil {` | 예 | 기존 | no | no |
| B11 | `:348` `if in.State.ActivatedRung < NoRung \|\| in.State.ActivatedRung >= len(in.Policy.Rungs) {` | 예 | 기존 | no | no |
| B12 | `:355` `if observed.Cmp(probe) > 0 {` | 예 | 기존 | no | no |
| B13 | `:374` `for i, rung := range in.Policy.Rungs {` | 예 | 기존 | no | no |
| B14 | `:376` `if err != nil {` | 아니오 | 기존 | no | no |
| B15 | `:379` `if i > newIndex && returnPct.Cmp(target) >= 0 {` | 예 | 기존 | no | no |
| B16 | `:386` `if newIndex > NoRung {` | 예 | 기존 | no | no |
| B17 | `:388` `if err != nil {` | 아니오 | 기존 | no | no |
| B18 | `:392` `if newIndex == len(in.Policy.Rungs)-1 && in.Policy.RunnerTrailPct != "" {` | 예 | 기존 | no | no |
| B19 | `:394` `if err != nil {` | 아니오 | 기존 | no | no |
| B20 | `:404` `if err != nil {` | 아니오 | 기존 | no | no |
| B21 | `:408` `if err != nil {` | 아니오 | 기존 | no | no |
| B22 | `:420` `if newIndex > in.State.ActivatedRung {` | 예 | 기존 | no | no |
| B23 | `:427` `if in.State.Completed {` | 예 | 기존 | no | no |
| B24 | `:436` `if err != nil {` | 아니오 | 기존 | no | no |
| B25 | `:439` `if observed.Cmp(baseline) < 0 {` | 예 | **a094 4.1** — 손절 조건 성립이 억제와 무관하게 관측된다 | no | no |
| B26 | `:441` `if in.State.PendingAction == ActionLadderStop {` | 예 | **a094 4.1·4.2** — 무장 중 억제는 유지되고, 해제 후 다음 주기에 발의가 난다 | no | no |
| B27 | `:453` `if out.RungPromotedTo == NoRung {` | 예 | 기존 | no | no |
| B28 | `:457` `switch {` | — | 기존 | no | no |
| B29 | `:458` `case newIndex == len(in.Policy.Rungs)-1 && in.Policy.FinalTakeFull:` | 예 | 기존 | no | no |
| B30 | `:461` `case isPositive(rung.PartialRatio):` | 예 | 기존 | no | no |
| B31 | `:466` `default:` | 예 | 기존 | no | no |
| B32 | `:475` `if in.State.PendingAction != ActionNone {` | 예 | 기존 | no | no |

**미진입 분기 11개**: B3, B6, B7, B8, B9, B14, B17, B19, B20, B21, B24
**자체 블록 없는 분기 1개**: B28 — 컴파일러가 별도 블록을 만들지 않는 형태이며 미커버와 다르다.
