# Branch Test Map: `TestA094AnAckedCancelOrNumberlessPlaceIsOnlyNamed` (post-edit 번호)

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 표 순회(cancel·amend·numberless place) | TestA094AnAckedCancelOrNumberlessPlaceIsOnlyNamed | 컴파일 RED(패키지) | PASS |
| B2 | 상태·읽기 단언 | 같음 | 컴파일 RED | PASS |
| B3 | (새) 기대가 CONFIRMED(취소) | 같음 /cancel | 컴파일 RED | PASS |
| B4 | (새) 알림 0·기록 번호 단언 | 같음 /cancel | 컴파일 RED | PASS |
| B5 | 알림 종목 단언(AMEND·번호 없는 PLACE) | 같음 /amend · /numberless_place | 컴파일 RED | PASS |
