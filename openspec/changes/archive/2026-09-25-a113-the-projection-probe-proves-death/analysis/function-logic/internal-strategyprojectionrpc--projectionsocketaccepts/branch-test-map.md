# Branch Test Map: `projectionSocketAccepts`

구현 후 AST 기준(분기 2). 편집 전 B3(owner-write 추정)은 삭제됐다 — 그 삭제를 지키는 것은 B2 행의
두 a113 행(EACCES 는 사망이 아니다)과 뮤테이션 N1(재도입 시 사망, `mutation-ledger.md`).

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 수락 중인 socket | `TestProjectionLivenessClausesEachDecideOnTheirOwn/수락한다` (a108) | no(회귀 핀) | yes |
| B2 | 경로 없음 · 연결 거부만 사망. 쓰기 비트가 깎인 산/죽은 socket(EACCES)은 사망이 아니다 | `TestProjectionLivenessClausesEachDecideOnTheirOwn` 의 `/경로가_없다`·`/아무도_수락하지_않는다`(a108) + `/쓰기_비트가_깎여도_수락_중이면_생존`·`/쓰기_비트가_깎인_죽은_socket_은_묻지_못한다`(a113) | yes — base 에서 a113 두 행이 `projectionSocketAccepts = false, want true` | yes |
