# Branch Test Map: `Recovery.Run`

- Source: `internal/reconcile/recovery.go`

> **「진입 실측」은 측정값이다** — 패키지 시험 전체를 `-covermode=set`으로 돌린 프로파일에서 그 분기가 만든 블록의 count다. 어떤 **개별** 시험이 그 분기를 밟는지는 이 실행이 답하지 않는다(시험별 프로파일이 필요하다). 따라서 「Test」 열은 **a094가 요구하는 시험**이며 현존 증명이 아니다.

| Branch | 조건 | 진입 실측 | Test (a094 요구) | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `:245` `if err != nil {` | 아니오 | 기존 — 재시작 순회 무변화 | no | no |
| B2 | `:258` `if err != nil {` | 아니오 | 기존 — 재시작 순회 무변화 | no | no |
| B3 | `:261` `for _, rec := range pending {` | 예 | **a094 4.4** — 순회 무변화 고정 | no | no |
| B4 | `:262` `if rec.State != journal.StateInDoubt {` | 아니오 | **a094 4.1·4.4** | no | no |
| B5 | `:268` `if berr != nil {` | 아니오 | 기존 — 재시작 순회 무변화 | no | no |
| B6 | `:276` `if rerr != nil {` | 예 | **a094 4.3** — 재생이 꺼진 채 남는다 | no | no |
| B7 | `:280` `if settled {` | 예 | 기존 — 재시작 순회 무변화 | no | no |
| B8 | `:285` `if rerr != nil {` | 아니오 | 기존 — 재시작 순회 무변화 | no | no |
| B9 | `:290` `if res.State == journal.StateUnresolvedInDoubt {` | 아니오 | 기존 — 재시작 순회 무변화 | no | no |
| B10 | `:302` `if err != nil {` | 예 | 기존 — 재시작 순회 무변화 | no | no |
| B11 | `:309` `if err != nil {` | 아니오 | 기존 — 재시작 순회 무변화 | no | no |
| B12 | `:324` `if report.Diff.BlocksEntry() {` | 예 | 기존 — 재시작 순회 무변화 | no | no |

**미진입 분기 7개**: B1, B2, B4, B5, B8, B9, B11
**자체 블록 없는 분기 0개**: 없음 — 컴파일러가 별도 블록을 만들지 않는 형태(빈 `default:` 등)이며 미커버와 다르다.

**Refresh (2026-09-27, HEAD ddd39a83).** `ast.json` 을 현재 소스에서 `go run ./tools/logic-map` 로 재생성했다(옛 파일은 base `ec29dc72` 소스를 기술). 행의 `:줄` 은 새 AST 의 줄로 옮겼다. 분기 대응은 옛/새 AST 를 (kind, 소스 줄)로 difflib 정렬한 결과다 — 항등(번호 무변). 「진입 실측」 열은 base 에서 잰 값 그대로다(재측정 안 함). 이 함수를 바꾼 이웃 커밋: 1c76a580 (a102).
