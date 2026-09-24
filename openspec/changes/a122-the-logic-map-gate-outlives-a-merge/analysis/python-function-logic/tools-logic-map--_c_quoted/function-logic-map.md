# Function Logic Map: `_c_quoted` (Python, a122 task 7.5.31 — 새 함수)

`ast.after-7.5.31.json` — 분기 0. 대체 저장소 목록(`GIT_ALTERNATE_OBJECT_DIRECTORIES`)의 한 항목을 C 인용으로 적는다 —
git 은 `"` 로 시작하는 항목을 인용을 풀어 읽는다(7.5.25 적대 재리뷰 F10, git 2.43 실측). 7.5.25 는 경로에 `:`·`"` 가
있으면 거절했다(헛거절). 변이 `AI2`(인용 안 함) · `AI3`(`"` 를 안 이스케이프)을 `test_a_repository_path_with_a_colon_or_a_quote_is_judged`
가 잡는다.
