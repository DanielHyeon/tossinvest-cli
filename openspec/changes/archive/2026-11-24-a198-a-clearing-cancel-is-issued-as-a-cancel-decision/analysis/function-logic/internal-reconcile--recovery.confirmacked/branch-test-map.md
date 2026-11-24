# Branch Test Map: `Recovery.confirmAcked` (post-edit 번호)

시험 이름은 grep 근거(internal/reconcile/a094_acked_boot_test.go · a198_acked_cancel_boot_test.go). RED 는 `analysis/tdd/red-4c.txt`(새 상수 미정의 — 컴파일 RED), GREEN 은 4c.8 선택 시험 원문(`analysis/verify/4c-*.txt`).

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | intent 읽기 실패 | 기존 없음(편집 무관) | n/a | n/a |
| B2 | (새) CANCEL + 기록 번호 → 쓰기 마침 | TestA198AnAckedCancelIsCompletedAtBoot · TestA094AnAckedCancelOrNumberlessPlaceIsOnlyNamed/cancel | 컴파일 RED | PASS |
| B3 | (새) Resume 성공 → ResolveConfirmed | TestA198AnAckedCancelIsCompletedAtBoot | 컴파일 RED | PASS |
| B4 | (새) 확정 쓰기 실패 → 알림·false | TestA198AFailedAckedCancelWriteLeavesTheAckAndNamesIt | 컴파일 RED | PASS |
| B5 | 종류 ≠ PLACE(AMEND·번호 없는 CANCEL) → 알림만 | TestA094AnAckedCancelOrNumberlessPlaceIsOnlyNamed/amend | 컴파일 RED | PASS |
| B6 | 번호 없음(PLACE) | TestA094AnAckedCancelOrNumberlessPlaceIsOnlyNamed/numberless_place | 컴파일 RED | PASS |
| B7 | 읽기 확인 실패 | TestA094AnUnconfirmedAckedPlaceStaysAndIsNamed | n/a(기존) | PASS |
| B8 | Resume 성공(PLACE) | TestA094AnAckedPlaceIsConfirmedByItsRecordedNumber | n/a(기존) | PASS |
| B9 | 확정 쓰기 실패(PLACE) | TestA094AFailedConfirmWriteLeavesTheAckAndNamesIt | n/a(기존) | PASS |
