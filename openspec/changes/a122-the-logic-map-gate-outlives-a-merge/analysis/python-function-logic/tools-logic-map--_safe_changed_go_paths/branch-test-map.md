# Branch Test Map: `_safe_changed_go_paths` (a122 task 7.5.22 · 7.5.23)

**이 표에는 절대 줄 번호가 없다 (7.5.23).** 7.5.22 가 쓴 줄 번호는 **같은 커밋의 제 수리가 밀어서
전부 +2 만큼 낡은 채로** 커밋됐다. 분기는 `ast*.json` 의 `id` 와 **소스 문구**로 가리킨다 —
줄은 `analysis/harness/7523_coords.py` 가 지문으로 검증하는 `ast*.json` 에만 있다.

시험은 `tools/logic-map/test_check_analysis.py` 의 `AGoFileWithNoTextualDiffIsNotSilentlyEmpty`
(진짜 저장소 · 진짜 git) · `TheGuardAndTheJudgementReadTheSameDiff` · `TheNumstatTableIsNeverInvented`.

## 7.5.22 — git 이 이진으로 다루는 파일

| 분기 | 소스 | 시험 |
|---|---|---|
| `-`/`-` 거절 (참) | `if added == b"-" and deleted == b"-":` | `…a_suppressed_body_is_refused_instead_of_silently_requiring_nothing` · `…the_minus_diff_attribute_suppresses_the_body_too` · `…an_uncommitted_gitattributes_is_enough_to_suppress_the_body` |
| `-`/`-` 거절 (거짓) | 같은 줄 | `…a_mode_only_change_is_not_refused` — **거절의 경계** |
| 거절이 **뒤에** 선다 | 이름 고리 뒤의 둘째 `for added, deleted, paths in records:` | `…the_name_guard_still_speaks_first_when_both_are_wrong`(한 파일) · `…speaks_first_even_when_the_faults_are_in_different_files`(두 파일) |
| rename 양쪽 이름 | `for raw in paths:` | `…the_guard_reads_both_names_of_a_rename`(옛 이름) · `…the_guard_reads_the_new_name_of_a_rename_too`(새 이름) |
| `_numstat_records` 의 rename 갈래 | `if first:` | 위 둘이 양쪽 갈래를 각각 돈다 + `…a_rename_record_carries_both_names` |

## 7.5.23 — git 이 본문을 지우는 파일

| 분기 | 소스 | 시험 |
|---|---|---|
| 판정 diff 의 `--no-textconv` | 판정 `git diff` 의 인자 | `…a_textconv_filter_does_not_empty_the_requirement` |
| 판정 diff 의 `--no-ext-diff` | 같은 자리 | `…an_external_diff_command_does_not_empty_the_requirement` |
| 교차 검사 (참) | `if not had_body:` | `…a_vanished_body_is_refused_rather_than_counted_as_no_change` |
| 두 시야의 **크기** | `if len(bodied) != len(records):` | `…a_judged_diff_that_lists_fewer_files_is_refused` — 그 **하나만** 이 갈래를 못 박는다(7.5.28 정정: 앞 판본은 다른 두 시험을 댔는데 검사를 지워도 둘 다 초록이었다) |
| 건너뛰기 조건의 `and` | `if added in (b"-", b"0") and deleted in (b"-", b"0"):` | `…an_append_only_change_that_lost_its_body_is_still_refused`(`N`/`0` · `0`/`N` 양쪽) |
| 이름이 아니라 **순서**로 짝짓기 | `zip(records, bodied)` | `…a_name_git_has_to_quote_is_not_called_a_vanished_body` |
| 교차 검사 (거짓) | `if added in (b"-", b"0") and deleted in (b"-", b"0"):` | `…a_mode_only_change_is_not_called_a_vanished_body` — **경계** |
| 본문 표시 | 파서 고리의 구역 표시(`bodied.append(False)` · `bodied[-1] = True`) | `AE3`(훅 전에 표시) · `AE4`(마지막 구역만 표시) |
| `_numstat_records` 결함 둘 | `raise RuntimeError("cannot read git diff --numstat record")` · `… rename record` | `…a_record_that_cannot_be_read_is_a_fault_not_an_empty_table` · `…a_truncated_rename_record_is_a_fault_not_an_empty_table` |

**깃발 둘은 시험이 **간접으로** 못 박는다.** `--no-textconv`/`--no-ext-diff` 를 지우면 그 저장소의
본문이 사라지고 **교차 검사가 런타임에 거절**하므로, 위 두 픽스처 시험이 "올바른 판정" 대신 예외를 받는다.
실측: 깃발을 지우면 `RuntimeError: git reported content changes but emitted no diff body for x.go`.
7.5.22 까지 그 둘은 **지워도 313 시험이 전부 초록**인 공짜 삭제였다.

## 7.5.27 — 워킹트리를 다시 쓰는 알려진 문

시험은 `TheWorktreeIsNotRewrittenUnderTheGate`(워킹트리 대상, 진짜 git).

| 갈래 | 소스 | 시험 |
|---|---|---|
| 가드도 고정을 받는다 | `_safe_changed_go_paths(root, base, target, pins)` · argv 의 `*(pins or [])` | `…a_clean_filter_does_not_hide_a_worktree_edit`(한쪽만 고정하면 두 시야가 갈려 거절) + 구조 `…both_git_calls_see_the_same_files` |
| fsmonitor 고정 | `_git_view_pins` 의 `core.fsmonitor=false` | `…a_lying_fsmonitor_does_not_hide_a_worktree_edit` |
| 필터 `clean` · `process` · `required` | `_git_view_pins` 의 드라이버 고리 | **계약** `…every_configured_driver_gets_all_three_pins` 가 셋 각각을 못 박는다 — 행동 시험 `…clean_filter…` · `…process_filter…` · `…required_filter_with_a_dotted_name…` 은 `clean=` 을 빼도 **초록**이다(git 의 우연, 7.5.28 정정) · 거절 갈래는 `…driver_name_that_cannot_be_pinned_is_refused` |
| 인덱스 플래그 (참) | 판정 끝의 `if not target:` → `_hidden_by_index_flags` | `…an_index_flag_that_hides_an_edit_is_refused_by_name`(subTest 둘) |
| 인덱스 플래그 (경계) | 해시 대조 · 파일 없음 건너뛰기 | `…an_index_flag_that_hides_nothing_is_not_refused` |
| 커밋 대상은 안 묻는다 | `if not target:` 의 거짓 갈래 | `…a_commit_target_does_not_read_the_worktree` |

## 7.5.28 — 줄 경계 · stat 캐시 · `ident` · 머리 줄의 탭

| 갈래 | 소스 | 시험 |
|---|---|---|
| 판정 diff 를 `\n` 에서만 자름 | `process.stdout.decode("utf-8", "strict").split(…)` | `…a_unicode_line_separator_in_a_path_does_not_split_the_diff`(subTest 셋) · `…a_rename_to_a_name_with_a_line_separator…` |
| 머리 줄 끝의 탭 | `_header_name` | `…a_name_with_a_space_is_not_refused` |
| stat 캐시 고정 | `_git_view_pins` 의 `core.checkStat=default` · `core.trustctime=true` | `…a_minimal_stat_check_does_not_hide_a_same_size_edit` — **`checkStat` 만** (7.5.25 정정 F1: `trustctime` 은 아무 시험도 못 박았다 · `AG2` 가 둘을 한꺼번에 뺐다) |
| `ident` 거절 | `_ident_go_paths` · 판정 끝의 `rewritten` | `…the_ident_attribute_is_refused_by_name`(`AG3`·`AG5`) · 경계 `…an_explicitly_unset_ident_is_not_refused`(`AG4`) |
| 드라이버 이름은 `=` 만 거절 | `if "=" in driver:` | `…a_driver_name_with_a_space_is_pinned_not_refused` · `…cannot_be_pinned_is_refused` |
| 플래그 해시의 고정 · `100755` | `["git", *pins, "hash-object", …]` · 모드 조건 | `…a_flagged_file_behind_a_filter…` · `…an_executable_flagged_file…` |

## 픽스처가 허구가 아님을 먼저 못 박는다

양성 대조군 둘 — `…git_really_suppresses_the_body_for_a_binary_marked_go_file`(`Binary files` 가 나오고
`@@` 가 없고 numstat 이 `-\t-\t`)과 `…git_really_empties_the_body_while_numstat_looks_normal`
(numstat 은 `1\t1\t` 인데 훅이 0 이고 `--no-textconv` 를 주면 돌아온다). 이것이 틀리면 나머지는
아무것도 재지 않는다 ([[mutation-must-reach-the-thing-under-test]]).

## RED

7.5.22(수리 전 `197a0355`): `AGoFileWithNoTextualDiffIsNotSilentlyEmpty` **Ran 7 · 실패 4 · 통과 3**.
통과 셋은 통과해야 하는 셋이다 — 양성 대조군 · mode-only 경계 · 순서(가릴 둘째 가드가 아직 없다).
*(7.5.23 정정: 이 줄의 앞 판본은 `Ran 6 · 통과 2` 라고 적었다. 순서 시험을 더하기 **전**에 재고
다시 안 쟀다 — 같은 커밋의 `review.md`·`tasks.md` 와도 어긋났다.)*

## 안 닫은 것 — 사유

- `_numstat_records` 의 결함 둘은 **지어낸 바이트**로만 시험한다. 진짜 git 으로 그 모양을 못 만들었다
  (충돌 중 워킹트리도 평범한 3칸이었다). 그래도 되는 이유는 거기서 재는 것이 *git 의 행동*이 아니라
  *순수 함수의 결함 처리*이기 때문이다 — git 이 그 모양을 낸다고 주장하지 않는다.
- 이진 rename 에서 거절 문장이 **어느 쪽 이름**을 대는지(`paths[-1]`)에 시험이 없다. 진짜 git 은
  `-\t-\t\0old\0new\0` 를 내고 가드는 새 이름을 댄다 — 행동은 쟀지만 못 박지는 않았다.
- 가드의 `subprocess.run` 에 `timeout=` 이 없다. 같은 결함이 일곱 자리에 걸쳐 있어
  **값 단위로** 한 번에 닫는다 → task 7.5.15 ([[correction-unit-must-be-the-value]]).

## 7.5.25 — 스냅숏 환경

| 갈래 | 변이 | 잡는 시험 |
|---|---|---|
| 워킹트리 대상인데 환경이 없으면 거절 | `AH11_the_guard_takes_no_snapshot` | `test_a_worktree_guard_without_the_snapshot_is_refused` |
| 가드도 스냅숏 환경을 받는다 | `AH2_the_guard_skips_the_snapshot` | 27 시험 (7.5.31 창 `129:171`; 7.5.25 는 21 이라 적었는데 같은 커밋의 뒤 시험으로 22 가 됐다 — 정정) |
| `--find-renames` (구조) | `AD6_the_guard_ignores_renames` | `test_both_git_calls_see_the_same_files` |

7.5.27 · 7.5.28 의 `pins` 행(필터 고정 · stat 고정)은 7.5.25 에서 겨눌 코드가 없어졌다 — 그 행은 기록으로 둔다.

## task 7.5.34 (창 `174:187` 재실행 · `143:190`)

| 갈래 | 변이 | 잡는 시험 |
|---|---|---|
| diff 가 격리한 환경에서 돈다 | `AJ48_the_guard_runs_outside_the_comparison` | 154 시험 |
| `--find-renames` | `AD6_the_guard_ignores_renames` | `test_both_git_calls_see_the_same_files` |

편집 전 B1·B2(스냅숏 없는 워킹트리 비교 거절)와 그 변이 AH11 은 없어졌다 — 서명이 `Comparison` 을 받아 그 입력을 만들 수 없다.
pathspec 이 없다는 것은 `test_both_git_calls_see_the_same_files`(구조)와 `test_pathspec_variables_do_not_change_the_judgement`(행동)가 본다.

## task 7.5.15 (2026-09-26)

짝은 손으로 고르지 않았다 — 갈래마다 변이를 걸어 실제로 빨개진 시험을 옮겼다(`analysis/harness/75_mut.py` 창 `228:246` · `97:98` · `99:100` · `101:102` · `102:103` · `112:113` · `239:240` · `244:246`(재실행), `7515_mut.py`; 사본 `A122_HARNESS_WORK` · pid 별 · 무변이 대조군 창 양끝 GREEN `Ran 407`, 구조 시험을 더한 뒤 창 `102:103` · `244:246` 은 `Ran 408`). '—' 는 변이 대신 직접 시험이다.

| 갈래 | 변이 | 잡는 시험 |
|---|---|---|
| `git diff --numstat` 의 `timeout=30` | AP1 | `test_every_child_process_in_the_gate_modules_names_a_timeout`(구조) · `test_every_child_the_verdict_starts_is_given_a_timeout`(행동 — 판정 · 기록 한 판이 실제로 띄운 자식 전부) |

위 "안 닫은 것" 의 `timeout=` 줄은 이것으로 닫혔다.
