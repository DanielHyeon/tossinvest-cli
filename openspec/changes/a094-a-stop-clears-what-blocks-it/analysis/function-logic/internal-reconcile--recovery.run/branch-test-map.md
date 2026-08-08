# Branch Test Map: `Recovery.Run`

- Source: `internal/reconcile/recovery.go`

> **「진입 실측」은 측정값이다** — 패키지 시험 전체를 `-covermode=set`으로 돌린 프로파일에서 그 분기가 만든 블록의 count다. 어떤 **개별** 시험이 그 분기를 밟는지는 이 실행이 답하지 않는다(시험별 프로파일이 필요하다). 따라서 「Test」 열은 **a094가 요구하는 시험**이며 현존 증명이 아니다.

| Branch | 조건 | 진입 실측 | Test (a094 요구) | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `:214` `if err != nil {` | 아니오 | 기존 — 재시작 순회 무변화 | no | no |
| B2 | `:227` `if err != nil {` | 아니오 | 기존 — 재시작 순회 무변화 | no | no |
| B3 | `:230` `for _, rec := range pending {` | 예 | **a094 4.4** — 순회 무변화 고정 | no | no |
| B4 | `:231` `if rec.State != journal.StateInDoubt {` | 아니오 | **a094 4.1·4.4** | no | no |
| B5 | `:237` `if berr != nil {` | 아니오 | 기존 — 재시작 순회 무변화 | no | no |
| B6 | `:245` `if rerr != nil {` | 예 | **a094 4.3** — 재생이 꺼진 채 남는다 | no | no |
| B7 | `:249` `if settled {` | 예 | 기존 — 재시작 순회 무변화 | no | no |
| B8 | `:254` `if rerr != nil {` | 아니오 | 기존 — 재시작 순회 무변화 | no | no |
| B9 | `:259` `if res.State == journal.StateUnresolvedInDoubt {` | 아니오 | 기존 — 재시작 순회 무변화 | no | no |
| B10 | `:269` `if err != nil {` | 예 | 기존 — 재시작 순회 무변화 | no | no |
| B11 | `:276` `if err != nil {` | 아니오 | 기존 — 재시작 순회 무변화 | no | no |
| B12 | `:291` `if report.Diff.BlocksEntry() {` | 예 | 기존 — 재시작 순회 무변화 | no | no |

**미진입 분기 7개**: B1, B2, B4, B5, B8, B9, B11
**자체 블록 없는 분기 0개**: 없음 — 컴파일러가 별도 블록을 만들지 않는 형태(빈 `default:` 등)이며 미커버와 다르다.
