# Branch Test Map: `Recovery.confirmAcked` (pre-edit 번호)

시험 이름은 grep 근거(internal/reconcile/a094_acked_boot_test.go). RED/GREEN 은 4c.6·4c.8 원문에서 채움.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | intent 읽기 실패 | 기존 없음(편집 무관) | n/a | n/a |
| B2 | 종류 ≠ PLACE → 알림만 | TestA094AnAckedCancelOrNumberlessPlaceIsOnlyNamed | 4c.6 | 4c.8 |
| B3 | 번호 없음 | 4c 에서 번호 없는 PLACE 행 추가 | 4c.6 | 4c.8 |
| B4 | 읽기 확인 실패 | TestA094AnUnconfirmedAckedPlaceStaysAndIsNamed | n/a(기존) | 4c.8 |
| B5 | Resume 성공 | TestA094AnAckedPlaceIsConfirmedByItsRecordedNumber | n/a(기존) | 4c.8 |
| B6 | 확정 쓰기 실패 | TestA094AFailedConfirmWriteLeavesTheAckAndNamesIt | n/a(기존) | 4c.8 |
