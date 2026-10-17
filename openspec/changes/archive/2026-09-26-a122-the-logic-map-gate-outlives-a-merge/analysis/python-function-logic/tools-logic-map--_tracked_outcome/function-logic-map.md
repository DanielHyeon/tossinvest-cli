# Function Logic Map: `_tracked_outcome` (Python, a122 task 7.5.9 — 새 함수)

## task 7.5.9 — 새 함수 (2026-09-26)

편집 후 `tools/logic-map/check_analysis.py:1411-1430` · 분기 4 · 반환 3 · raise 0 · 호출 8 (`ast.after-759.json`, source sha `194ec1a29941`)

| id | 줄 | 종류 | 소스 |
|---|---|---|---|
| B1 | 1422 | Try | `try:` |
| B2 | 1424 | ExceptHandler | `except (OSError, subprocess.SubprocessError) as exc:` |
| B3 | 1426 | If | `if process.returncode:` |
| B4 | 1429 | comprehension | `for raw in process.stdout.split(b'\x00') if raw` |

`git ls-files -z` 의 대답을 `(지문, 경로들 또는 실패)` 로. 지문은 git 이 낸 **바이트 그대로**의 해시(task 7.5.10 — `-z` 는 NUL 로 끝나는
경로의 이음이고 경로에 NUL 은 못 든다 → 출력이 같으면 목록이 같다). 실패는 셋 다 값으로 돌려준다: 자식을 못 띄움(`OSError`) · 시한
(`TimeoutExpired`, 이웃 `_worktree_entries` 의 `ls-files` 와 같은 30) · rc≠0(`RuntimeError("cannot list the tracked files: …")`).
중간판(`check_analysis.py` source sha `ff6142a55df0` — **최종이 아니다**, 보수 절 정정) 에서 `:1446-1465`.

## 보수 — `ls-files` 에 `SNAPSHOT_PINS` · 결함 갈래의 행동 시험 (독립 리뷰 F9 · P1-3, 2026-09-26)

> 편집 전 `ast.before-759r.json`(`22c1e10e9d94`) · 편집 후 `ast.after-759r.json`(최종 `tools/logic-map/check_analysis.py:1481-1503` · 분기 4 · 반환 3 · raise 0 · 호출 8 · source sha `ff590b79db6e`) — 분기 4 → 4, 반환 · raise · 호출 수 불변.

편집 후 `tools/logic-map/check_analysis.py:1481-1503` · 분기 4 · 반환 3 · raise 0 · 호출 8 (`ast.after-759r.json`, source sha `ff590b79db6e`)
편집 전 `tools/logic-map/check_analysis.py:1446-1465` · 분기 4 · 반환 3 · raise 0 · 호출 8 (`ast.before-759r.json`, revision `worktree`, source sha `22c1e10e9d94`)

| 옛 id | 새 id | 새 줄 | 종류 | 소스(새) | 바뀐 것 |
|---|---|---|---|---|---|
| B1 | B1 | 1492 | Try | `try:` | 같음 |
| B2 | B2 | 1497 | ExceptHandler | `except (OSError, subprocess.SubprocessError) as exc:` | 같음 |
| B3 | B3 | 1499 | If | `if process.returncode:` | 같음 |
| B4 | B4 | 1502 | comprehension | `for raw in process.stdout.split(b'\x00') if raw` | 같음 |

`git ls-files -z` 앞에 `*SNAPSHOT_PINS`(`-c core.fsmonitor=false`). 없을 때 저장소가 `core.fsmonitor` 에 적은 프로그램이 판정 중 **떴다**(리뷰어 실측 · 이 보수의
RED `test_the_tracked_listing_does_not_start_the_repositorys_fsmonitor` 가 표식 파일로 재현). 모듈 머리의 `SNAPSHOT_PINS` 주석("인덱스를 읽는 git 호출이 모두
받는 고정")은 7.5.9 첫 판에서 거짓이었다. 이 보수로 추적 목록은 고정되지만 주석은 여전히 **전부**를 말할 수 없다 — `_recording_refusal` 의 `git diff --quiet` 가
이 로트 **전부터** 핀이 없다(열린 7.5.40). 주석을 그 범위로 좁혔다.
B1 · B2(자식을 못 띄움 · 시한)는 행동이 맞았지만 시험이 0 이었다 — 변이 MZ1(그 갈래가 빈 목록 + 빈 해시)이 408 을 초록으로 통과했다(리뷰어 실측). 시험
`test_a_tracked_listing_that_cannot_start_or_times_out_is_a_fault`(두 subTest — `TimeoutExpired` · `FileNotFoundError`)가 MZ1 을 잡는다. 그 시험은 수리 전에도
초록이다 — 행동이 이미 맞았으므로 RED 가 아니라 **변이**가 영수증이다.
