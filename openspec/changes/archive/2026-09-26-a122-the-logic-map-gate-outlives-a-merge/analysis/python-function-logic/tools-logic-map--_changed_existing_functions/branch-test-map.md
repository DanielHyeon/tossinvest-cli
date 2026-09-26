# Branch Test Map: `_changed_existing_functions`

## task 7.5.13 · 7.5.26 잔여 (2026-09-27)

짝은 변이로 잡았다(`75_mut.py` 창 `253:262` · `117:118` · 재실행 `257:258` · `261:262`, 사본 `A122_HARNESS_WORK` ext4 · pid 별 · 무변이 대조군 창 양끝 GREEN `Ran 427` / 재실행 `Ran 429`). '—' 는 변이 대신 직접 시험이다.

| 갈래 | 변이 | 잡는 시험 |
|---|---|---|
| 새 B34 · B36 이름은 레코드에서 | MQ2(머리 줄에서 `removeprefix`) | 5 — 편집 · 삭제 · rename 두 방향 · 시험 클래스 `ANameGitQuotesIsJudgedByItsRealName` |
| 새 B36 새 쪽은 레코드의 **마지막** 이름 | MQ6(옛 이름을 씀) | 4 — `test_a_rename_into_a_quoted_name_still_reads_the_new_side` · `test_a_rename_behind_a_forged_blob_does_not_hide_the_edit` 외 |
| B34 `/dev/null` 은 이름이 아니다 | AB5(재조준) | 4 — `test_a_name_git_has_to_quote_is_not_called_a_vanished_body` · `test_a_new_file_with_a_quoted_name_still_requires_nothing` 외 |
| B3 · B4 레코드를 푼다 | (레코드는 가드가 UTF-8 로 확인) | 위 전부 |
| B5~B7 (`section_name`) | 그 번들 | — |
