# Branch Test Map: `Tracker.persist` (reconcile)

- Source: `internal/reconcile/mismatch.go` (878-954). 커버 측정: `../../coverage/after-branch-counts.txt`(cover mode set — 1 = 실행됨). 편집 전 측정 `../../coverage/base-branch-counts.txt`.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 기존 분기 | 기존 시험(패키지 GREEN) | 해당 없음(기존 분기) | yes — cover cond=1 body=1 |
| B2 | 기존 분기 | 기존 시험(패키지 GREEN) | 해당 없음(기존 분기) | yes — cover cond=1 body=1 |
| B3 | 기존 분기 | 기존 시험(패키지 GREEN) | 해당 없음(기존 분기) | yes — cover cond=1 body=1 |
| B4 | 삽입 쓰기 오류 | `TestA192AnUnprovenInsertIsNeverBound` | 해당 없음(기존 분기 — 결속 없음 쪽은 base 에서도 통과) | yes — cover cond=1 body=1 |
| B5 | 기존 분기 | 기존 시험(패키지 GREEN) | 해당 없음(기존 분기) | yes — cover cond=1 body=1 |
| B6 | 기존 분기 | 기존 시험(패키지 GREEN) | 해당 없음(기존 분기) | yes — cover cond=1 body=1 |
| B7 | 종결 증명 해제 갈래 | `TestA192MissingOrderBlockArmsThenReleasesAfterTheTerminalProof` | yes — `../../red/red-1-reconcile.txt`(단언 실패, 컴파일 실패 아님) | yes — cover cond=1 body=1 |
| B8 | 해제 기록 실패 | `TestA192ReleaseWriteFailureKeepsTheGate` | 관측 안 함 — GREEN 뒤 커버 측정에서 0 을 보고 시험 추가. 변이 대신: 이 갈래는 오류 전달뿐이라 변이 대상 판정 없음, 시험이 오류·게이트 유지를 단언 | yes — cover cond=1 body=1 |
| B9 | 재판정 해제/유지 | `TestA192AFillCommittedBetweenVerdictAndReleaseKeepsTheBlock` | yes — red-1(`release attempts = 0`) | yes — cover cond=1 body=1 |
| B10 | 기존 분기 | 기존 시험(패키지 GREEN) | 해당 없음(기존 분기) | yes — cover cond=1 body=1 |
| B11 | 기존 분기 | 기존 시험(패키지 GREEN) | 해당 없음(기존 분기) | yes — cover cond=1 body=1 |
