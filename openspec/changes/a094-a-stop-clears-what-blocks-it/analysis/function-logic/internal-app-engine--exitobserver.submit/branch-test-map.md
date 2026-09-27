# Branch Test Map: `ExitObserver.submit`

- Source: `internal/app/engine/exitloop.go`

> **「진입 실측」은 측정값이다** — 패키지 시험 전체를 `-covermode=set`으로 돌린 프로파일에서 그 분기가 만든 블록의 count다. 어떤 **개별** 시험이 그 분기를 밟는지는 이 실행이 답하지 않는다(시험별 프로파일이 필요하다). 따라서 「Test」 열은 **a094가 요구하는 시험**이며 현존 증명이 아니다.

| Branch | 조건 | 진입 실측 | Test (a094 요구) | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `:1346` `if err != nil {` | 아니오 | 기존 | no | no |
| B2 | `:1349` `if isZeroQuantity(submitQuantity) {` | 예 | 기존 (a091이 다룸) | no | no |
| B3 | `:1369` `if err != nil {` | 아니오 | 기존 | no | no |
| B4 | `:1378` `if err := o.opts.Journal.AttachExitIntent(ctx, m.position.ID, intentID); err != nil {` | 예 | 기존 | no | no |
| B5 | `:1383` `if err != nil {` | 아니오 | 기존 | no | no |
| B6 | `:1393` `switch {` | — | **a094 2.9** | no | no |
| B7 | `:1394` `case out.State == journal.StateConfirmed:` | 예 | 기존 | no | no |
| B8 | `:1402` `case out.State == journal.StateInDoubt \|\| out.State == journal.StateUnresolvedInDoubt:` | 예 | **a094 2.1** — 409가 더 이상 여기로 오지 않는다 | no | no |
| B9 | `:1407` `case out.Reason == execgw.ReasonSymbolInFlight:` | 아니오 | **a094 6.2** — 272210 라이브락 | no | no |
| B10 | `:1410` `default:` | 예 | **a094 2.9** — 409가 이제 여기로 온다 | no | no |
| B11 | `:1412` `if detail == "" && err != nil {` | 아니오 | 기존 | no | no |

**미진입 분기 5개**: B1, B3, B5, B9, B11
**자체 블록 없는 분기 1개**: B6 — 컴파일러가 별도 블록을 만들지 않는 형태(빈 `default:` 등)이며 미커버와 다르다.

**Refresh (2026-09-27, HEAD ddd39a83).** `ast.json` 을 현재 소스에서 `go run ./tools/logic-map` 로 재생성했다(옛 파일은 base `ec29dc72` 소스를 기술). 행의 `:줄` 은 새 AST 의 줄로 옮겼다. 분기 대응은 옛/새 AST 를 (kind, 소스 줄)로 difflib 정렬한 결과다 — 항등(번호 무변). 「진입 실측」 열은 base 에서 잰 값 그대로다(재측정 안 함). 본문은 base 와 바이트가 같다(위치만 이동).
