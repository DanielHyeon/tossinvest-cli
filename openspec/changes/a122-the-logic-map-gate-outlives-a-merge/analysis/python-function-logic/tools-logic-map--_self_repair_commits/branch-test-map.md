# Branch Test Map: `_self_repair_commits`

번호가 아니라 **소스 한 줄**로 적는다 — [[positional-branch-ids-break-hand-renumbering]].
짝은 손으로 고르지 않았다: 갈래마다 변이(`726_mut.py` S1~S11)를 걸어 **실제로 빨개진** 시험을
옮겼다. 무변이 대조군이 GREEN 임을 확인한 뒤에만 돈다.

| 갈래 (소스) | 변이 | 잡는 시험 |
|---|---|---|
| 규칙 자체 (`_landing_refusal` 의 `later`) | S1 (가드 삭제) | **7** — `test_a_review_fix_to_a_pinned_file_after_the_record_is_refused` · `test_any_go_file_counts_not_only_the_pinned_one` · `test_merge_commits_are_not_read` 외 4 |
| `any(name.endswith(".go") …)` = **ANYGO** | S2 (고정 소스만) | `test_any_go_file_counts_not_only_the_pinned_one` |
| `--no-merges` ×2 | S3 | **없다 — 동등 변이(아래)** |
| `not _is_ancestor(root, commit, candidate)` | S4 (뒤집기) | **17** — 착지 관련 시험 거의 전부 |
| 자기 자신은 안 센다 | S5 | `test_one_changed_pinned_source_is_enough` · `test_the_repair_commit_itself_is_the_landing_when_it_refreshes_the_evidence` |
| 가드의 **자리**(맨 뒤) | S6 (맨 앞으로) | `test_a_candidate_its_evidence_does_not_describe_keeps_that_sentence` |
| `reversed(hashes)` (가장 오래된 것을 이름으로) | S7 (`later[-1]`) | `test_the_oldest_repair_is_the_one_named` |
| `if change_dir.startswith(ARCHIVE_PREFIX)` | S8 (옮기기 전 경로 무시) | `test_the_signal_survives_archiving_the_change` |
| `LANDING_RECOVERY` 를 두 자리에서 인용 | S9 | `test_a_record_on_disk_is_named_as_not_committed` · `test_the_refusal_says_how_to_move_the_record` |
| 계산 경로가 같은 집합을 넘긴다 | S10 | `test_the_recorder_will_not_write_a_landing_its_own_later_work_outruns` |
| 선언 경로가 같은 집합을 넘긴다 | S11 | `test_any_go_file_counts_not_only_the_pinned_one` |
| `--full-history` (역사를 단순화하지 않는다) | S12 | `test_a_fix_the_merge_hides_is_still_read` · `test_a_git_failure_becomes_a_verdict_not_a_silent_pass` |
| 첫 조회 실패 → 판정 | S13 | `test_a_git_failure_becomes_a_verdict_not_a_silent_pass` |
| 둘째 조회 실패 → 판정 | S14 | 같은 시험 (둘째 주입) |
| `rev-list` 실패 → 판정 | S17 | 같은 시험 (셋째 주입) |
| `-c core.quotePath=false` | S15 | `test_a_non_ascii_go_name_is_read` |
| `-c diff.renames=false` | S16 | `test_a_rename_out_of_go_is_read_whatever_the_machine_configures` |
| 계산값 대조 거절도 복구 경로를 말한다 | S9b | `test_a_later_commit_that_also_matches_is_refused` |

**18개 중 17 CAUGHT · 1 SURVIVED(S3).** 아래 여섯 줄(S9b · S12~S17)은 2026-09-16 적대 리뷰가
연 갈래다 — 그 전에는 재는 시험이 **0** 이었다.

## S3 는 **동등 변이**다 — 왜인지 재서 적는다

[[surviving-mutant-may-mean-accidental-safety]]: "동등 변이"로 넘기지 않는다. `--no-merges` 를
양쪽에서 빼도 판정이 안 바뀌는 이유는 **git 의 기본값**이다 — `git log` 는 `--diff-merges` 를
명시해야만 병합 커밋의 diff 를 낸다. git 2.43 에서 직접 쟀다: 기본 · `log.diffMerges=first-parent` ·
`diff.merges=first-parent` 셋 다 병합 커밋의 이름 목록이 **비었고**, `--diff-merges=first-parent`
를 명시했을 때만 나왔다. 그러므로 `--no-merges` 없이도 병합은 깃발이 안 서고, 이 한계는 저장소
설정의 함수가 아니다.

가드를 지운 것이 아니라 **한계를 못 박았다**: `test_a_fix_made_inside_the_merge_commit_itself_is_the_known_limit`
이 병합 커밋 자신이 Go 와 change 디렉터리를 같이 고친 모양에서 깃발이 0 임을 단언한다. 누가
`--diff-merges` 를 더해 한계를 없애면 그 시험이 빨개진다.

## 가지치기 — 선형 픽스처로는 못 잰다

`git log <경로>` 의 기본 단순화는 병합이 그 경로에 대해 한 부모와 같으면 반대편 가지를 통째로
버린다. 첫 판은 그것을 안 막았고, **선형 픽스처만 있어서** 시험도 못 잡았다
([[linear-fixtures-never-exercise-dag-guards]]). `test_a_fix_the_merge_hides_is_still_read` 는
가지치기가 **실제로 일어나는지**를 먼저 단언한 뒤(`assertNotIn(hidden, simplified)`) 규칙을 잰다 —
안 하면 가지치기가 안 일어나는 픽스처 위에서 초록인 시험이 된다.

## 안 덮는 것 — 알려진 한계

| 모양 | 시험 | 왜 두나 |
|---|---|---|
| Go 수리와 문서를 **다른 커밋**으로 쪼갬 | `test_splitting_the_fix_from_its_notes_evades_the_guard` (초록으로 못 박음) | 망각 가드이지 위조 가드가 아니다. 내용으로 가르는 규칙은 이웃 때문에 68건 중 61/55 를 거절한다 |
| 병합 커밋 자신의 수리 | `test_a_fix_made_inside_the_merge_commit_itself_is_the_known_limit` | 7.2.4 H7 과 같은 부류 |
| 이웃이 나중에 같은 파일을 고침 | `test_a_neighbours_later_go_edit_is_not_this_changes_own_repair` | 거절하면 안 되는 자리 — 내용으로는 자기 수리와 안 갈린다 |

---

**여기 인용된 `722_mut.py` · `726_mut.py` · `76_mut.py` 는 저장소에 없다** (2026-09-18 확인).
그 세션들의 스크래치패드에 살다가 사라져서 **인용만 남고 재현이 안 된다** —
[[borrowed-flm-evidence-goes-stale]] 와 같은 모양이다. 7.5 부터는 하네스를
`analysis/harness/` 에 커밋한다(`README.md` 참조). 위 표의 짝을 다시 확인해야 하면
`75_mut.py` 를 본떠 그 task 의 변이를 다시 쓸 것 — **짝은 손으로 고르지 않는다.**

## task 7.5.2.1 — 판정이 읽는 입력을 전부 세고 하나씩 묶는다 (7.5.2 재리뷰, 2026-09-20)

짝은 손으로 고르지 않았다 — 각 갈래를 지우는 변이를 걸어 실제로 빨개진 시험을 옮겼다(`75_mut.py` 76 변이 · 생존 0 · 무변이 대조군 초록). '—' 는 변이 대신 직접 시험이다.

| 갈래 | 변이 (`75_mut.py`) | 잡는 시험 |
|---|---|---|
| `head` 에서 걷기 | Y4 | `test_every_history_read_uses_the_one_resolved_commit · test_a_branch_switch_mid_run_judges_the_history_it_started_on` |

## task 7.5.2.2 — 스냅숏은 끝에서 디스크와 다시 대조한다 (7.5.2.1 재리뷰, 2026-09-20)

짝은 손으로 고르지 않았다 — 각 갈래를 지우는 변이를 걸어 실제로 빨개진 시험을 옮겼다(전수 95 · 생존 0: 94 는 한 판에서 CAUGHT(스위트 272 · 무변이 대조군 GREEN) · Z14 는 **도달한 채 살아남아**(`io.BytesIO(None)` 이 조용히 빈 버퍼다) 구조 시험을 더한 뒤 같은 하네스로 다시 재서 CAUGHT(273)). '—' 는 변이 대신 직접 시험이다.

| 갈래 | 변이 (`75_mut.py`) | 잡는 시험 |
|---|---|---|
| 분기 변화 0 — git 의 말 첫 줄이 `_first_line` 한 벌로 | V2 | `test_a_one_sided_git_failure_cannot_land_a_pre_edit_commit` |

## task 7.5.29 (2026-09-27)

짝은 변이로 잡았다(`75_mut.py` 창 `253:262` · `117:118` · 재실행 `257:258` · `261:262`, 사본 `A122_HARNESS_WORK` ext4 · pid 별 · 무변이 대조군 창 양끝 GREEN `Ran 427` / 재실행 `Ran 429`). '—' 는 변이 대신 직접 시험이다.

| 갈래 | 변이 | 잡는 시험 |
|---|---|---|
| `-z` 로 받는다 | MS1(`-z` 뗌) | 104 — 자기 수리를 거치는 시험 전부 |
| 새 B12 이름이 `.go` 로 **끝난다** | MS2(`.go` 가 들어 있으면) | `test_a_non_go_name_with_a_line_separator_raises_no_flag` |
| 새 B9~B11 모양이 다르면 결함 | MS3 — **첫 판 생존**(정상 git 은 늘 그 모양). git 대답을 바꿔 치우는 시험을 더해 재실행 CAUGHT | `test_a_listing_of_an_unexpected_shape_is_a_fault` |
| 새 B8 커밋 경계는 NUL 둘 | — | `test_two_repairs_are_each_read`(커밋 둘 — 하나는 Go, 하나는 아님) |
| 인용되는 Go 이름도 깃발 | — | `test_a_go_name_git_quotes_raises_the_flag`(편집 전 빨강) |

## task 7.5.42 (마감 수리, 2026-09-27)

| 갈래 | 변이 | 잡는 시험 |
|---|---|---|
| `log.showRoot=true` | MX8(고정 삭제) | `test_the_root_commit_is_read_whatever_log_show_root_says` |
