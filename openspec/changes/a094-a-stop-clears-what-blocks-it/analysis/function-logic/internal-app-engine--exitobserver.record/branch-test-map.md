# Branch Test Map: `ExitObserver.record`

- Source: `internal/app/engine/exitloop.go`

> **「진입 실측」은 측정값이다** — 패키지 시험 전체를 `-covermode=set`으로 돌린 프로파일에서 그 분기가 만든 블록의 count다. 어떤 **개별** 시험이 그 분기를 밟는지는 이 실행이 답하지 않는다(시험별 프로파일이 필요하다). 따라서 「Test」 열은 **a094가 요구하는 시험**이며 현존 증명이 아니다.

| Branch | 조건 | 진입 실측 | Test (a094 요구) | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `:1180` `if !o.quoteUsable(quote) {` | — (a111 이 넣은 분기, 진입 실측 없음) | 기존 — `TestA111LeaseIsRecheckedAtTheRecordOrRefreshBoundary` (a111 BTM B1, 같은 소스 sha 522d5d81) | no | no |
| B2 | `:1199` `if judgement.ObservationSource == "" {` | — (a111 이 넣은 분기, 진입 실측 없음) | 기존 — `TestA111LeaseIsRecheckedAtTheRecordOrRefreshBoundary` (a111 BTM B2, 같은 소스 sha 522d5d81) | no | no |
| B3 | `:1200` `if quote.FetchedAt.IsZero() {` | 예 | 기존 — a094는 이 함수를 바꾸지 않는다 | no | no |
| B4 | `:1202` `} else {` | 예 | 기존 — a094는 이 함수를 바꾸지 않는다 | no | no |
| B5 | `:1223` `if orderable && (snapshot.CancelPendingFirst \|\| isFullExit(proposal)) {` | 예 | **a094 3.1** — 게이트 조건 무변화를 고정한다 | no | no |
| B6 | `:1224` `if m.reJudge && !isProtective(proposal) {` | 예 | 기존 — a094는 이 함수를 바꾸지 않는다 | no | no |
| B7 | `:1246` `} else {` | 예 | **a094 3.1·3.7** | no | no |
| B8 | `:1248` `if err != nil {` | 아니오 | 기존 — a094는 이 함수를 바꾸지 않는다 | no | no |
| B9 | `:1251` `if !cleared {` | 예 | 기존 — a094는 이 함수를 바꾸지 않는다 | no | no |
| B10 | `:1255` `} else {` | 예 | 기존 — a094는 이 함수를 바꾸지 않는다 | no | no |
| B11 | `:1262` `if orderable {` | 예 | 기존 — a094는 이 함수를 바꾸지 않는다 | no | no |
| B12 | `:1264` `if intentID == "" {` | 아니오 | 기존 — a094는 이 함수를 바꾸지 않는다 | no | no |
| B13 | `:1276` `if err != nil {` | 예 | 기존 — a094는 이 함수를 바꾸지 않는다 | no | no |
| B14 | `:1277` `if errors.Is(err, journal.ErrProposalPending) {` | 예 | 기존 — a094는 이 함수를 바꾸지 않는다 | no | no |
| B15 | `:1283` `if errors.Is(err, journal.ErrExitSnapshotQuarantined) {` | 예 | 기존 — a094는 이 함수를 바꾸지 않는다 | no | no |
| B16 | `:1296` `if recorded.ArmedProposal == nil \|\| recorded.ArmOutcome != journal.ExitArmArmed {` | 예 | 기존 — a094는 이 함수를 바꾸지 않는다 | no | no |

**미진입 분기 2개**: B8, B12 (base 번호 B6, B10 — 4판 재번호; 진입 실측은 base 값이다. a111 이 더한 B1·B2 는 이 실측의 대상이 아니었다)
**자체 블록 없는 분기 0개**: 없음 — 컴파일러가 별도 블록을 만들지 않는 형태(빈 `default:` 등)이며 미커버와 다르다.

**Refresh (2026-09-27, HEAD ddd39a83).** `ast.json` 을 현재 소스에서 `go run ./tools/logic-map` 로 재생성했다(옛 파일은 base `ec29dc72` 소스를 기술). 행의 `:줄` 은 새 AST 의 줄로 옮겼다. 분기 대응은 옛/새 AST 를 (kind, 소스 줄)로 difflib 정렬한 결과다 — B1->B3, B2->B4, B3->B5, B4->B6, B5->B7, B6->B8, B7->B9, B8->B10, B9->B11, B10->B12, B11->B13, B12->B14, B13->B15, B14->B16; 새 분기 B1, B2. 「진입 실측」 열은 base 에서 잰 값 그대로다(재측정 안 함). 이 함수를 바꾼 이웃 커밋: 882a0b49 (a111).
