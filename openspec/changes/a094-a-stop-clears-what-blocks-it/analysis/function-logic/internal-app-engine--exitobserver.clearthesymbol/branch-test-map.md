# Branch Test Map: `ExitObserver.clearTheSymbol`

- Source: `internal/app/engine/exitloop.go`

> **「진입 실측」은 측정값이다** — 패키지 시험 전체를 `-covermode=set`으로 돌린 프로파일에서 그 분기가 만든 블록의 count다. 어떤 **개별** 시험이 그 분기를 밟는지는 이 실행이 답하지 않는다(시험별 프로파일이 필요하다). 따라서 「Test」 열은 **a094가 요구하는 시험**이며 현존 증명이 아니다.

| Branch | 조건 | 진입 실측 | Test (a094 요구) | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `:1443` `if err != nil {` | 아니오 | **a094 3.7** — 목록 조회 실패가 손절을 보류시키지 않는다 | no | no |
| B2 | `:1447` `for _, order := range live {` | 예 | **a094 3.1·3.4** — 저널 ∪ 브로커, dedup | no | no |
| B3 | `:1449` `if !buy && !withPending {` | 아니오 | **a094 3.1** — 저널에 없는 매수도 걸린다 | no | no |
| B4 | `:1464` `if err != nil {` | 아니오 | a094 3.3 | no | no |
| B5 | `:1470` `if qerr != nil \|\| perr != nil {` | 아니오 | 기존 | no | no |
| B6 | `:1485` `if err != nil \|\| out.State != journal.StateConfirmed {` | 예 | **a094 3.2·3.3** — 확정 취소 후에만 | no | no |
| B7 | `:1489` `if !clear {` | 예 | **a094 3.3** — 못 치우면 제출 안 함 | no | no |
| B8 | `:1492` `if withPending && m.state.Pending() {` | 예 | 기존 | no | no |
| B9 | `:1493` `if err := o.release(ctx, m, journal.ProposalCancelled); err != nil {` | 아니오 | 기존 | no | no |

**미진입 분기 5개**: B1, B3, B4, B5, B9
**자체 블록 없는 분기 0개**: 없음 — 컴파일러가 별도 블록을 만들지 않는 형태(빈 `default:` 등)이며 미커버와 다르다.

**Refresh (2026-09-27, HEAD ddd39a83).** `ast.json` 을 현재 소스에서 `go run ./tools/logic-map` 로 재생성했다(옛 파일은 base `ec29dc72` 소스를 기술). 행의 `:줄` 은 새 AST 의 줄로 옮겼다. 분기 대응은 옛/새 AST 를 (kind, 소스 줄)로 difflib 정렬한 결과다 — 항등(번호 무변). 「진입 실측」 열은 base 에서 잰 값 그대로다(재측정 안 함). 본문은 base 와 바이트가 같다(위치만 이동).
