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


## task 7.5.1 — "못 물었다" 와 "없다" 를 가른 뒤 (2026-09-19, `75_mut.py` 31 변이 · 생존 0)

파서를 엄격하게 다시 썼으므로 위 표의 앵커 12개가 낡았고, 변이 집합을 **현재 소스 기준으로** 다시
썼다. 짝은 여전히 손으로 고르지 않았다 — 각 갈래를 지우는 변이를 걸어 실제로 빨개진 시험을 옮겼다.

| 갈래 (소스) | 변이 | 잡는 시험 |
|---|---|---|
| `if process.returncode: raise …` (못 물었다 = 결함) | V1 | `test_a_committed_record_is_not_read_as_absent_when_git_fails` · `test_a_failure_that_still_printed_is_not_parsed` 외 3 |
| 결함 문장에 git 의 첫 줄 | V2 | `test_a_one_sided_git_failure_cannot_land_a_pre_edit_commit` 외 2 |
| `if header == spec + b" missing"` (**물은 spec 그대로**) | V3 | `test_a_missing_answer_must_echo_the_spec_that_was_asked` |
| `if match is None: raise` (모르는 머리) | V4 | `test_an_unknown_header_is_a_verdict_not_an_absence` 외 1 |
| `if data[position + size] != 0` (NUL 종단) | V5 | `test_a_missing_payload_terminator_is_a_verdict` |
| `if position + size >= len(data)` (넘침 = 잘림) | V6 | `test_a_short_last_payload_is_reported_as_truncated` |
| `for spec in specs: if "\0" in spec` (요청 전체) | V7 | `test_a_nul_in_the_ref_is_refused_by_name` |
| 잘림 문장이 **레코드**를 센다 | V8 | `test_a_truncated_response_counts_records_not_blobs` |
| `if end < 0: raise` | T6 | `test_a_truncated_response_counts_records_not_blobs` · `test_a_truncated_response_is_a_verdict_not_a_partial_answer` |
| 가드 7 양쪽 한 번씩 (호출자 `_landing_refusal`) | V13 | `test_guard_seven_reads_each_side_once_even_when_it_refuses` |

**첫 판 생존 둘과 그 교훈.** V13 은 `all()` 이 첫 번째로 다른 소스에서 멈추므로 **받는** 픽스처에선
소스별로 읽어도 비용이 같아 살아남았다 — 가드 7 이 **거절하는**(끝까지 읽는) V1 모양으로 잰다.
V10(지문을 하한 뒤에)은 첫 시험이 번들을 `record_landing` 의 **걷기 전** 하한 측정에서 써 버려 두
판본이 다 그것을 봤다 — 주입 지점을 `_measure_landing_inputs` 가 스택에 있을 때로 좁혔다. 양성 대조가
둘 다 "도달함" 이라고 말했으므로 계측기가 아니라 시험 공백이었고, 둘 다 **코드가 아니라 시험**을 고쳐
잡았다 ([[mutation-must-reach-the-thing-under-test]]).

## task 7.5.2 — 판정은 한 번 읽은 바이트로 선다 (수리한 트리의 재리뷰, 2026-09-19)

짝은 손으로 고르지 않았다 — 각 갈래를 지우는 변이를 걸어 실제로 빨개진 시험을 옮겼다(`75_mut.py` 52 변이 · 생존 0, 첫 판 생존 W23 은 **안 닿음** → 닿는 시험을 더해 잡음). '—' 는 변이 대신 직접 시험이다.

| 갈래 | 변이 (`75_mut.py`) | 잡는 시험 |
|---|---|---|
| `<oid> submodule` → `None` | W14 | `test_a_submodule_answer_is_not_a_blob` |
| 버전 조언은 rc 129 에서만 | W15 | `test_the_git_version_hint_is_given_only_for_a_usage_error` |
| 파서 오류가 물은 경로를 댄다 | W16 | `test_a_framing_fault_names_the_path_it_was_reading` |
| `>=` 경계(내용은 왔고 NUL 만 없다) | V6 | `test_a_payload_missing_only_its_terminator_is_truncated` |
| 잘림 문장 한 벌이 레코드를 센다 | V8 | `test_a_truncated_response_counts_records_not_blobs` |

## task 7.5.2.2 — 스냅숏은 끝에서 디스크와 다시 대조한다 (7.5.2.1 재리뷰, 2026-09-20)

짝은 손으로 고르지 않았다 — 각 갈래를 지우는 변이를 걸어 실제로 빨개진 시험을 옮겼다(전수 95 · 생존 0: 94 는 한 판에서 CAUGHT(스위트 272 · 무변이 대조군 GREEN) · Z14 는 **도달한 채 살아남아**(`io.BytesIO(None)` 이 조용히 빈 버퍼다) 구조 시험을 더한 뒤 같은 하네스로 다시 재서 CAUGHT(273)). '—' 는 변이 대신 직접 시험이다.

| 갈래 | 변이 (`75_mut.py`) | 잡는 시험 |
|---|---|---|
| B9 git 이 말을 했으면 붙인다(같은 160자 규칙의 셋째 철자였다) | V2 | `test_a_committed_record_is_not_read_as_absent_when_git_fails` |
