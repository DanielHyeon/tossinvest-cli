# Branch Test Map: `_committed_many`

번호가 아니라 **소스 한 줄**로 적는다 — [[positional-branch-ids-break-hand-renumbering]].
짝은 손으로 고르지 않았다: 각 갈래를 지우는 변이(`75_mut.py` T1~T18)를 걸어 **실제로 빨개진**
시험을 옮겼다. 하네스는 이름을 셋까지만 찍는다 — 넷 이상이면 "외 N".

| 갈래 (소스) | 변이 | 잡는 시험 | 덮이나 |
|---|---|---|---|
| `["git", "cat-file", "--batch", "-Z"]` (출력까지 NUL) | T1 | 56개 — `test_a_blob_holding_nul_bytes_is_read_whole` 외 55 | yes |
| 입력이 NUL 로 끊긴다 (`f"{ref}:{relative}\0"`) | T2 | `test_a_newline_in_a_path_still_names_one_file` | yes — 줄 단위로 바꾸면 경로가 쪼개진다 |
| `if fields[1] == b"blob":` (blob 만 내용) | T3 | `test_a_tree_is_not_a_blob` | yes |
| `position += size + 1` 이 **blob 밖**에 있다 | T4 | `test_a_tree_is_not_a_blob` | yes — 트리 **뒤**의 자리를 같이 물어야 닿는다 |
| `if process.returncode: return found` | T5 · T17 | `test_a_failure_that_still_printed_is_not_parsed` | yes — 빈 출력으로는 안 갈린다(부분 출력이 필요) |
| `if end < 0: break` | T6 | **없다 — 동등 변이** | 사전 채움이 `break` 와 `continue` 를 같게 만든다. 그 사전 채움은 T18 이 **59개**로 잡는다 |
| `size = int(fields[2])` (선언된 크기) | T7 | `test_a_blob_holding_nul_bytes_is_read_whole` · `test_a_tree_is_not_a_blob` | yes — 내용에 NUL 이 있어야 닿는다 |
| `found = {relative: None for relative in wanted}` (사전 채움) | T18 | 59개 | yes |
| `if not wanted:` (빈 요청은 git 을 안 부른다) | — | `test_nothing_asked_means_no_git_at_all` (spawn 0 단언) | yes — 행동 단언 |
| 배치가 **한 번**인가 — `_pinning_at` | T8 · T10 | `test_judging_one_commit_costs_one_git_process` · `test_a_commits_blob_is_read_by_one_function` | yes |
| 배치가 **한 번**인가 — `_unheld_bundles` | T9 · T11 | `test_the_unheld_check_also_costs_one_git_process` · `test_the_signal_survives_archiving_the_change` | yes |
| 목록을 **한 번** 재서 넘긴다 (계산 경로) | T12 | 34개 | yes |
| 목록을 **한 번** 재서 넘긴다 (선언 경로) | T13 | 21개 | yes |
| `committed is None and before` (아카이브 대안은 대체가 아니다) | T14 | `test_an_archived_bundle_held_at_the_candidate_is_not_called_unheld` | yes — 옮긴 **뒤** 커밋을 봐야 닿는다 |
| `anchor = root.resolve()` (검사와 계산이 같은 닻) | T15 | 59개 | yes |
| 철자가 하나인가 (`_committed_bytes` → `_committed_many`) | T16 | `test_a_commits_blob_is_read_by_one_function` | yes — **구조** 시험뿐이다(행동이 같다) |

**T6 을 "동등 변이"로 넘기지 않았다** ([[surviving-mutant-may-mean-accidental-safety]]).
무엇이 덮고 있는지를 가설이 아니라 변이(T18)로 확인했고, 덮는 것 자체가 못 박혀 있다.
