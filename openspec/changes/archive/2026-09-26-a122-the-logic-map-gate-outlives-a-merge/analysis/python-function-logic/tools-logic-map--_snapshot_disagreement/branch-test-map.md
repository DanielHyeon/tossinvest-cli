# Branch Test Map: `_snapshot_disagreement` (Python, a122 task 7.5.31)

변이 로그(`75_mut.py`, 창 `129:171`, 대조군 양끝 GREEN `Ran 362`)에서 옮겼다. 변이마다 값 하나.

| 갈래 | 변이 | 잡는 시험 |
|---|---|---|
| 호출 자체(몸통 B40) | `AI6_the_store_is_not_cross_checked` | `test_an_object_store_that_lies_about_the_edit_is_refused` |
| B1 `ls-tree` 실패 | `AI11_a_failed_ls_tree_is_read` | `test_the_store_cross_check_reads_each_rule_directly` |
| B11 같은 oid 는 넘어간다 | `AI12_an_unchanged_path_is_checked` | 77 시험 |
| B13 규칙 1(레코드에 없음) | `AI8_a_missing_record_is_fine` | `…object_store_that_lies…` · `…reads_each_rule_directly` |
| B13 규칙 3(`0`/`0`) | `AI7_same_content_is_a_change` | `…reads_each_rule_directly` **하나** — 위조 저장소 시험에서는 git 이 그 경로를 레코드로 **안 냈다**(규칙 1 이 잡는다) |
| B15 규칙 2(삭제) | `AI9_a_hidden_deletion_is_fine` | `…reads_each_rule_directly` |

**재지 않은 갈래**: B4 · B5(레코드 모양) — git 이 내지 않는 모양이라 행동 시험이 못 닿고, 변이를 넣지 않았다. B6 의 종류
거름은 지웠다(위 FLM).
