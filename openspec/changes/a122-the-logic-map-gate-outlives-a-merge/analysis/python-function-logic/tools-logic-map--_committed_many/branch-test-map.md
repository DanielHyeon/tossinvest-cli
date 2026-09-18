# Branch Test Map: `_committed_many`

번호가 아니라 **소스 한 줄**로 적는다 — [[positional-branch-ids-break-hand-renumbering]].
짝은 손으로 고르지 않았다: 각 갈래를 지우는 변이(`analysis/harness/75_mut.py` T1~U3)를 걸어
**실제로 빨개진** 시험을 옮겼다. 하네스는 이름을 셋까지만 찍는다 — 넷 이상이면 "외 N".
2026-09-18 재실행: **22 변이 전부 CAUGHT, 생존 0.**

| 갈래 (소스) | 변이 | 잡는 시험 | 덮이나 |
|---|---|---|---|
| `["git", "cat-file", "--batch", "-Z"]` (출력까지 NUL) | T1 | 78개 — `test_a_blob_holding_nul_bytes_is_read_whole` 외 77 | yes |
| 입력이 NUL 로 끊긴다 | T2 | `test_a_newline_in_a_path_still_names_one_file` · `test_a_truncated_response_is_a_verdict_not_a_partial_answer` | yes |
| `if fields[1] == b"blob":` (blob 만 내용) | T3 | `test_a_tree_is_not_a_blob` | yes |
| `position += size + 1` 이 **blob 밖**에 있다 | T4 | `test_a_tree_is_not_a_blob` | yes — 트리 **뒤**의 자리를 같이 물어야 닿는다 |
| `if process.returncode: return found` | T5 · T17 | `test_a_failure_that_still_printed_is_not_parsed` · `test_git_failing_leaves_every_path_unanswered` | yes — 빈 출력으로는 안 갈린다(부분 출력이 필요) |
| `if end < 0: raise` (**부분 답을 안 쓴다**) | T6 · T6b · T17 | `test_a_truncated_response_is_a_verdict_not_a_partial_answer` | yes |
| `if position != len(data): raise` | U1 | `test_bytes_left_unread_are_a_verdict` | yes |
| `if "\0" in relative: raise` | U2 | `test_a_nul_in_a_path_is_refused_before_it_desyncs_the_batch` | yes |
| `.encode("utf-8", "surrogateescape")` | U3 | `test_a_path_that_is_not_utf8_still_reaches_git` | yes |
| `size = int(fields[2])` (선언된 크기) | T7 | `test_a_blob_holding_nul_bytes_is_read_whole` · `test_a_tree_is_not_a_blob` | yes — 내용에 NUL 이 있어야 닿는다 |
| `found = {relative: None …}` (사전 채움) | T18 | 60개 | yes |
| `if not wanted:` (빈 요청은 git 을 안 부른다) | — | `test_nothing_asked_means_no_git_at_all` (spawn 0 단언) | yes — 행동 단언 |
| 배치가 **한 번**인가 — `_pinning_at` | T8 · T10 | `test_judging_one_commit_costs_one_git_process` · `test_a_commits_blob_is_read_by_one_function` | yes |
| 배치가 **한 번**인가 — `_unheld_bundles` | T9 · T11 | `test_the_unheld_check_also_costs_one_git_process` · `test_the_signal_survives_archiving_the_change` | yes |
| 목록을 **한 번** 재서 넘긴다 (계산 · 선언) | T12 · T13 | 34개 · 21개 | yes |
| `committed is None and before` (아카이브 대안은 대체가 아니다) | T14 | `test_an_archived_bundle_held_at_the_candidate_is_not_called_unheld` | yes — 옮긴 **뒤** 커밋을 봐야 닿는다 |
| `anchor = root.resolve()` (검사와 계산이 같은 닻) | T15 | 59개 | yes |
| 철자가 하나인가 (`_committed_bytes` → `_committed_many`) | T16 | `test_a_commits_blob_is_read_by_one_function` | yes — **구조** 시험뿐이다(행동이 같다) |

## 정정 — 첫 판의 "T6 은 동등 변이다" 는 **거짓이었다** (2026-09-18 독립 리뷰 P0)

첫 판은 `if end < 0: break` 를 두고 "사전 채움이 `break` 와 `continue` 를 같게 만든다,
그 사전 채움은 T18 이 잡으므로 T6 은 동등 변이"라고 적었다. **두 명제가 다르다.**

- 반례(재현으로 확인): 마지막 머리가 멀쩡하고 **내용만 잘린** 응답에서 원본은 `None`,
  변이는 `b""` 를 낸다. `None` 과 `b""` 는 `_unheld_bundles` 의 아카이브 대체 갈래를
  여닫으므로 **판정이 갈린다**(`[]` → `['pkg--fn']`).
- T18 이 잡는 것은 `missing`/`continue` 갈래(도달 153회)이고 `end < 0` 갈래는 **182개
  시험에서 0회 · 116개 코퍼스에서 0회** 도달이었다. 살아남은 이유는 동등이 아니라 **미도달**
  이다 ([[mutation-must-reach-the-thing-under-test]]).
- 같은 로트의 T14 에는 정확한 규율을 썼다(`assertTrue(_pre_archive_path(ast_path))` —
  픽스처가 그 갈래에 닿는지 **먼저 단언**). T6 만 그 규율을 안 썼고, 살아남은 것도 그것이다.

수리: `end < 0` 을 **판정**으로 바꾸고(부분 답을 안 쓴다) 닿는 픽스처를 세웠다. 이제
`break` 로 되돌리는 변이(T6)와 `end = len(data)` 변이(T6b) 둘 다 잡힌다.

하네스에도 **양성 대조**를 넣었다 — 변이가 **돌았는지**까지 재지 않으면 눈먼 계측기와
진짜 음성이 같게 기록된다. T6 이 그 결과였다.
