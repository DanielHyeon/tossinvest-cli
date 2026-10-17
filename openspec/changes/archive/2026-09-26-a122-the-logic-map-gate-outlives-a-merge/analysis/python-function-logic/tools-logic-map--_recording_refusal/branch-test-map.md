# Branch Test Map: `_recording_refusal`

번호가 아니라 **소스 한 줄**로 적는다 — [[positional-branch-ids-break-hand-renumbering]].
짝은 손으로 고르지 않았다: 각 갈래를 부수는 변이(`77_mut.py` R1~R22, 최종 코드 sha256 `1dfa432b0081` 사본)를 걸어
**실제로 빨개진** 시험을 옮겼다. 하네스는 이름을 셋까지만 찍는다 — 넷 이상이면 "외 N".

| 갈래 (소스) | 변이 | 잡는 시험 |
|---|---|---|
| `if landing_file.is_symlink()` | R1 | `test_a_dangling_symlink_record_is_refused_not_followed` · `test_the_advice_never_names_a_command_that_would_refuse` |
| `if _landing_record(change_dir, root) is not None` | R2 | `test_a_record_on_disk_is_named_as_not_committed` · `test_an_undecodable_committed_record_is_reported_not_raised` |
| `if landing_file.exists()` | R3 · R4(문장 합침) | `test_a_record_on_disk_is_named_as_not_committed` · `test_it_refuses_to_overwrite_an_existing_record` · `test_the_advice_never_names_a_command_that_would_refuse` — R4 는 첫째 하나만 잡는다(명령과 조언이 같은 판정이라 둘이 같이 바뀐다) |
| `if (change_dir / 'analysis' / 'function-logic-reference.txt').exists()` | R5 · R10(dirty 를 이 앞으로) | `test_it_refuses_a_change_that_borrows_its_evidence` · `test_a_refusal_that_outlives_a_commit_is_named_before_one_that_does_not` |
| `try:` / `except (OSError, RuntimeError, ValueError, json.JSONDecodeError) as exc:` | R6 | `test_a_base_it_cannot_resolve_is_named_as_the_base` — **첫 판 SURVIVED**. 바깥 `GATE_FAULTS` 가 rc 1 을 대신 만들어 가려졌다 |
| `if facts.get('execution_baseline_adoption')` | R7 · R14'(문장 갈림) | `test_the_recorder_refuses_the_adoption_path_too` |
| `if why:` (걷기 전 하한 사유) | R9 · R22(dirty 를 이 앞으로) | `test_the_advice_never_names_a_command_that_would_refuse` · `test_a_refusal_that_outlives_a_commit_is_named_before_one_that_does_not` — 기록 명령 쪽은 `compute_landing` 이 같은 사유로 대신 거절하므로 **조언 쪽 시험만** 잡는다 |
| `if dirty.returncode:` | R8 | `test_it_refuses_while_tracked_files_are_modified` · `test_the_advice_never_names_a_command_that_would_refuse` |
| `return ('', base)` (멈추지 않음) | R11(base 대신 빈칸) | 9 — `test_a_candidate_older_than_its_evidence_is_not_the_landing` · `test_a_record_on_disk_is_named_as_not_committed` · `test_it_does_not_record_the_base_even_when_the_base_matches` 외 6 |

## 호출자 쪽 배관

| 자리 | 변이 | 잡는 시험 |
|---|---|---|
| `record_landing`: `if refusal: return 1, …` | R12 | 10 — `test_a_base_it_cannot_resolve_is_named_as_the_base` · `test_a_dangling_symlink_record_is_refused_not_followed` · `test_a_record_on_disk_is_named_as_not_committed` 외 7 |
| `record_landing`: 판정 호출이 `GATE_FAULTS` 경계 **안** | R21 | `test_record_landing_names_a_bundle_that_escapes_the_repository` (7.4) — GREEN 첫 판을 실제로 잡은 시험 |
| `main`: `if refusal:` | R13 | `test_the_advice_never_names_a_command_that_would_refuse` · `test_a_fault_while_asking_the_recorder_does_not_become_advice` · `test_a_refusal_that_outlives_a_commit_is_named_before_one_that_does_not` |
| `main`: `except GATE_FAULTS` → 권하지 않음 | R14 | `test_a_fault_while_asking_the_recorder_does_not_become_advice` |
| `main`: 해소기가 찾은 디렉터리를 넘김 | R15(활성 경로) | `test_the_advice_never_names_a_command_that_would_refuse` (staged 아카이브 모양) |

## task 7.5.2.1 — 판정이 읽는 입력을 전부 세고 하나씩 묶는다 (7.5.2 재리뷰, 2026-09-20)

짝은 손으로 고르지 않았다 — 각 갈래를 지우는 변이를 걸어 실제로 빨개진 시험을 옮겼다(`75_mut.py` 76 변이 · 생존 0 · 무변이 대조군 초록). '—' 는 변이 대신 직접 시험이다.

| 갈래 | 변이 (`75_mut.py`) | 잡는 시험 |
|---|---|---|
| `head` 와 비교 | Y7 | `test_every_history_read_uses_the_one_resolved_commit` |
| 0 · 1 말고는 결함 | Y12 | `test_a_cleanliness_check_git_cannot_answer_is_not_called_dirty` |

## task 7.5.14 (2026-09-27)

짝은 변이로 잡았다(`75_mut.py` 창 `264:267`, 사본 ext4 · pid 별 · 무변이 대조군 창 양끝 GREEN `Ran 439`).

| 갈래 | 변이 | 잡는 시험 |
|---|---|---|
| 새 B3 디스크의 기록은 깔때기로 | MU2(`exists()` 로 되돌림) | `test_no_read_primitive_lives_outside_the_funnels`(구조 **하나뿐** — 행동은 같다. 첫 판 표의 1 은 second_site 시험을 더하기 **전**의 판에서 잰 값이었다 — 그 시험이 들어간 파일에서는 2(적대 리뷰 실측: second_site 가 위반 목록 전체를 단언해 같이 빨개졌다). 그 시험을 독립시킨 최종 파일에서 재실측 1) |
| 새 B4~B6 빌림 표지는 판정과 같게 | MU1(`exists()` 로 되돌림) | ~~3~~ **5**(첫 판 3 은 second_site 추가 전 판의 값 — 추가 뒤 4(리뷰어); 수리 뒤 최종 파일 재실측 5, `Ran 445`) — 권한 없는 · FIFO · **심링크 고리** · **폴더** 표지 · 구조 시험 |
| 읽히는 표지는 빌림 | — | `test_a_readable_mark_is_still_a_borrow`(양성 대조, 편집 전에도 초록) |
