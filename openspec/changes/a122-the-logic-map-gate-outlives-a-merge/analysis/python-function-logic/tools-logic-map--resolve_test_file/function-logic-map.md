# Function Logic Map: `resolve_test_file` (Python, a122 task 7.5.2.3)

## task 7.5.2.3 — 재확인의 입력 집합은 판정의 입력 집합이다 (7.5.2.2 재리뷰, 2026-09-20)

`tools/logic-map/check_analysis.py:1570-1590` · 분기 5 · 반환 3 · raise 0 (편집 전 L1328-1346 · 분기 5 · 반환 3 · raise 0, `ast.before-7.5.2.3.json` = revision `1d12520c`).

인용한 파일을 **고르는** 세 갈래(정규화된 경로 · 패키지 안 · 트리 전체)가 전부 깔때기를 지난다.

| id | 줄 | 종류 | 소스 |
|---|---|---|---|
| B1 | 1583 | If | `if '/' in cited:` |
| B2 | 1585 | IfExp | `qualified if _kind(qualified) == 'reg' else None` |
| B3 | 1587 | If | `if _kind(local) == 'reg':` |
| B4 | 1589 | comprehension | ` for path in _globbed(root, cited) if '.git' not in path.parts` |
| B5 | 1590 | IfExp | `matches[0] if len(matches) == 1 else None` |

## task 7.5.9 — 시험 인용은 추적 파일로만 충족된다 (7.5.2.3 재리뷰 정확성, 2026-09-26)

> 편집 전 `ast.before-759.json`(revision `1d1e5ca7`) · 편집 후 `ast.after-759.json`, `759_flm_rows.py` 로 정렬. 중간판(`check_analysis.py` source sha `ff6142a55df0` — **최종이 아니다**, 보수 절 정정) 에서 `:2226-2250`.

편집 후 `tools/logic-map/check_analysis.py:2191-2215` · 분기 7 · 반환 3 · raise 0 · 호출 3 (`ast.after-759.json`, source sha `194ec1a29941`)
편집 전 `tools/logic-map/check_analysis.py:2157-2177` · 분기 5 · 반환 3 · raise 0 · 호출 4 (`ast.before-759.json`, revision `1d1e5ca7`, source sha `b28398e26f3d`)

| 옛 id | 새 id | 새 줄 | 종류 | 소스(새) | 바뀐 것 |
|---|---|---|---|---|---|
| B1 | B1 | 2208 | If | `if '/' in cited:` | 같음 |
| B2 | — | — | IfExp | `qualified if _kind(qualified) == 'reg' else None` (옛) | **빠짐** |
| B3 | — | — | If | `if _kind(local) == 'reg':` (옛) | **빠짐** |
| B4 | — | — | comprehension | `for path in _globbed(root, cited) if '.git' not in path.parts` (옛) | **빠짐** |
| — | B2 | 2210 | BoolOp | `qualified in files and _kind(qualified) == 'reg'` | **새** |
| — | B3 | 2210 | IfExp | `qualified if qualified in files and _kind(qualified) == 'reg' else None` | **새** |
| — | B4 | 2212 | BoolOp | `local in files and _kind(local) == 'reg'` | **새** |
| — | B5 | 2212 | If | `if local in files and _kind(local) == 'reg':` | **새** |
| — | B6 | 2214 | comprehension | `for path in files if path.name == cited` | **새** |
| B5 | B7 | 2215 | IfExp | `matches[0] if len(matches) == 1 else None` | 번호만 |

**바뀐 것.** 인자 `files`(색인이 읽은 추적 `*_test.go` 전부, `frozenset`)가 늘었고 세 갈래 모두 그 목록 안에서만 고른다.
이름 붙은 갈래(B2·B3)와 패키지 갈래(B4·B5)는 `in files` 를 **먼저** 묻고 그다음 `_kind` 를 묻는다 — 추적 안 된 경로는 원장에도
안 남는다(입력이 아니다). 맨이름 갈래(B6)는 트리 순회 대신 목록에서 이름이 같은 것을 센다.

**결함 두 모양(RED, 편집 전 코드).** (1) 추적 안 된 사본이 있으면 맨이름이 둘을 찾아 `None` → 호출자가 `continue` → 대조군에서
나오던 `past the end` 거절이 **사라졌다**. tasks.md 는 이것을 "없던 거절을 만든다" 로 적었다 — **틀렸다**: 해소 실패는 거절이 아니라
건너뛰기다(`test_an_untracked_copy_does_not_hide_a_bare_name_coordinate`). (2) 패키지 안의 추적 안 된 긴 파일이 좌표를 받아 추적된
짧은 파일의 거절을 가렸다(`test_an_untracked_file_in_the_package_does_not_answer_for_the_tracked_one`). 이름 붙은 갈래는 반대 방향이다 —
추적 안 된 파일의 줄 수로 **거절을 지어냈다**. 편집 뒤에는 머지 뒤와 같이 해소되지 않는다(해소 못 한 좌표는 오늘 오류가 아니다 —
`test_a_qualified_path_that_does_not_exist_is_not_silently_skipped` 가 그 정책을 못 박는다).

## 보수 — 이름 붙은 좌표의 경로를 풀어서 대조한다 (독립 적대 리뷰 P1-1 — 7.5.9 첫 판의 회귀, 2026-09-26)

> 편집 전 `ast.before-759r.json`(워킹트리 판, source sha `22c1e10e9d94`) · 편집 후 `ast.after-759r.json`(최종 `tools/logic-map/check_analysis.py:2267-2298` · 분기 9 · 반환 4 · raise 0 · 호출 7 · source sha `ff590b79db6e`), `759_flm_rows.py` 로 정렬.

편집 후 `tools/logic-map/check_analysis.py:2267-2298` · 분기 9 · 반환 4 · raise 0 · 호출 7 (`ast.after-759r.json`, source sha `ff590b79db6e`)
편집 전 `tools/logic-map/check_analysis.py:2226-2250` · 분기 7 · 반환 3 · raise 0 · 호출 3 (`ast.before-759r.json`, revision `worktree`, source sha `22c1e10e9d94`)

| 옛 id | 새 id | 새 줄 | 종류 | 소스(새) | 바뀐 것 |
|---|---|---|---|---|---|
| B1 | B1 | 2284 | If | `if '/' in cited:` | 같음 |
| — | B2 | 2289 | Try | `try:` | **새** |
| — | B3 | 2291 | ExceptHandler | `except ValueError:` | **새** |
| B2 | B4 | 2293 | BoolOp | `qualified in files and _kind(qualified) == 'reg'` | 번호만 |
| B3 | B5 | 2293 | IfExp | `qualified if qualified in files and _kind(qualified) == 'reg' else None` | 번호만 |
| B4 | B6 | 2295 | BoolOp | `local in files and _kind(local) == 'reg'` | 번호만 |
| B5 | B7 | 2295 | If | `if local in files and _kind(local) == 'reg':` | 번호만 |
| B6 | B8 | 2297 | comprehension | `for path in files if path.name == cited` | 번호만 |
| B7 | B9 | 2298 | IfExp | `matches[0] if len(matches) == 1 else None` | 번호만 |

**결함.** 7.5.9 첫 판의 B2·B3 은 `root / cited` 를 그대로 추적 목록과 대조했다. `pathlib` 은 `//` 와 `./` 는 접지만 `..` 도 심링크도 안 푼다 —
`internal/../internal/near_test.go:4000` · `alias/near_test.go:4000`(추적 심링크 `alias -> internal`)이 목록 밖 → `None` → 호출자 `continue` 로 **조용히 통과**했다.
편집 전(1d1e5ca7)은 디스크의 정규 파일로 골라 `past the end of a 1-line file` 로 거절했다(리뷰어 실측). 이 보수의 RED(수리 전 워킹트리 `22c1e10e`):
`..` subTest 와 심링크 시험이 `[]` 로 빨갛다. `//` · `./` subTest 는 수리 전에도 초록이다(`pathlib` 이 접는다 — 못 박기용).

**수리(새 B2 · B3).** `os.path.realpath(root / cited)` 를 `realpath(root)` 에 대한 상대경로로 바꿔 `root / 그것` 을 목록과 대조한다. 저장소 밖이면(`ValueError`)
`None` — 해소 못 함(오류 아님, 기존 정책). `realpath` 의 lstat 들은 원장 밖이다(`normalized_source` 와 같은 부류 — 열린 7.5.12).
**거부하는 정상 입력**: 오늘 코퍼스의 `..` · `//` 좌표 0 건(독립 적대 리뷰 측정 — 이 세션이 다시 재지 않았다). 이 수리는 **더 거절하는** 쪽이다(편집 전 동작으로 되돌림).
