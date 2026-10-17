# Branch Test Map: `_tracked_outcome` (Python, a122 task 7.5.9)

## task 7.5.9 · 7.5.10 (2026-09-26)

짝은 손으로 고르지 않았다 — 갈래마다 변이를 걸어 실제로 빨개진 시험을 옮겼다(`analysis/harness/75_mut.py` 창 `228:246` · `97:98` · `99:100` · `101:102` · `102:103` · `112:113` · `239:240` · `244:246`(재실행), `7515_mut.py`; 사본 `A122_HARNESS_WORK` · pid 별 · 무변이 대조군 창 양끝 GREEN `Ran 407`, 구조 시험을 더한 뒤 창 `102:103` · `244:246` 은 `Ran 408`). '—' 는 변이 대신 직접 시험이다.

| 갈래 | 변이 | 잡는 시험 |
|---|---|---|
| B1 · B2 자식을 못 띄움 · 시한 → 실패 값 | — | 행동 시험 없음(자식 프로세스를 못 띄우게 하는 픽스처를 안 만들었다). 구조: `_tracked` 가 `BaseException` 을 올린다(AM9) |
| B3 rc≠0 → `RuntimeError("cannot list the tracked files: …")` | AM5 | `test_a_tracked_file_list_that_cannot_be_read_is_a_fault_not_an_empty_index` |
| B4 목록 | AM6 | 위 `test_index` 표 |
| 지문 = git 바이트의 해시 (7.5.10) | AN3 — **첫 판 생존**: 시험이 저장소 **둘**을 견줘 뿌리가 달랐다. 한 저장소의 두 상태로 고쳐 재실행 CAUGHT | `test_the_tracked_list_fingerprint_is_the_bytes_git_said` |

## 보수 (독립 리뷰 둘, 2026-09-26)

짝은 변이로 잡았다(`75_mut.py` 창 `246:253` · `233:234`, 사본 `A122_HARNESS_WORK` · pid 별 · 무변이 대조군 창 양끝 GREEN `Ran 416`; 사본은 최종 파일과 주석 한 덩이만 다르고 AST 가 같다 — 실측).

| 갈래 | 변이 | 잡는 시험 |
|---|---|---|
| B1 · B2 자식을 못 띄움 · 시한 → 실패 값 | **MZ1**(빈 목록 + 빈 해시 — 리뷰어가 생존을 실측) | `test_a_tracked_listing_that_cannot_start_or_times_out_is_a_fault` |
| `SNAPSHOT_PINS` | MR3(핀 삭제) | `test_the_tracked_listing_does_not_start_the_repositorys_fsmonitor` |
| B4 추적 목록(스테이징 포함) | AM6 재조준 | 16 |
