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
| 교차 검사 (참) | `if not bodied.intersection(names):` | `…a_vanished_body_is_refused_rather_than_counted_as_no_change` |
| 교차 검사 (거짓) | `if added in (b"-", b"0") and deleted in (b"-", b"0"):` | `…a_mode_only_change_is_not_called_a_vanished_body` — **경계** |
| 본문 표시 | 파서 고리의 `bodied.update(...)` 두 자리 | 위 셋이 같이 돈다(`AD5` 가 한 자리만 남기면 빨개진다) |
| `_numstat_records` 결함 둘 | `raise RuntimeError("cannot read git diff --numstat record")` · `… rename record` | `…a_record_that_cannot_be_read_is_a_fault_not_an_empty_table` · `…a_truncated_rename_record_is_a_fault_not_an_empty_table` |

**깃발 둘은 시험이 **간접으로** 못 박는다.** `--no-textconv`/`--no-ext-diff` 를 지우면 그 저장소의
본문이 사라지고 **교차 검사가 런타임에 거절**하므로, 위 두 픽스처 시험이 "올바른 판정" 대신 예외를 받는다.
실측: 깃발을 지우면 `RuntimeError: git reported content changes but emitted no diff body for x.go`.
7.5.22 까지 그 둘은 **지워도 313 시험이 전부 초록**인 공짜 삭제였다.

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
