# Function Logic Map: `_git` (Python, `execution_baseline.py`)

## task 7.5.15 — 시한 (2026-09-26)

> 편집 전 `ast.before-7515.json`(revision `1d1e5ca7`) · 편집 후 `ast.after-7515.json` — 분기 · 반환 · raise · 호출 수 불변, 정렬 결과 전부 같음.
> 중간판(`execution_baseline.py` source sha `22ea8c60964f` — **최종이 아니다**, 보수 절 정정) 에서 `:36-43`.

편집 후 `tools/logic-map/execution_baseline.py:36-43` · 분기 3 · 반환 1 · raise 1 · 호출 4 (`ast.after-7515.json`, source sha `22ea8c60964f`)
편집 전 `tools/logic-map/execution_baseline.py:30-35` · 분기 3 · 반환 1 · raise 1 · 호출 4 (`ast.before-7515.json`, revision `1d1e5ca7`, source sha `504ec55c505a`)

| 옛 id | 새 id | 새 줄 | 종류 | 소스(새) | 바뀐 것 |
|---|---|---|---|---|---|
| B1 | B1 | 40 | If | `if process.returncode:` | 같음 |
| B2 | B2 | 41 | IfExp | `process.stderr if text else process.stderr.decode('utf-8', 'replace')` | 같음 |
| B3 | B3 | 42 | BoolOp | `error.strip() or 'git command failed'` | 같음 |

`subprocess.run` 에 `timeout=120`: 도우미가 rev-parse 부터 범위 rev-list · diff-tree 까지 싣는다 — 가장 무거운 이웃(`_repairs_after` 의 rev-list · `_self_repair_commits` 의 log)의 값. `TimeoutExpired` 는 `AdoptionError` 가 아니라 `validate` 를 부르는 `resolve_base` 의
`except AdoptionError` 를 지나 `_judged` · `_recording_refusal` 의 `GATE_FAULTS` 경계에서 판정 줄이 된다. 기안 CLI(`main`)는
`(AdoptionError, OSError, ValueError)` 만 받으므로 거기서 시한이 나면 traceback 이다 — 판정 경로가 아니라 이 로트 밖이다.

## 보수 — 최종 파일로 다시 열거 (독립 주장정확성 리뷰 F11, 2026-09-26)

최종 `tools/logic-map/execution_baseline.py:38-45` · 분기 3 · 반환 1 · raise 1 · 호출 4 · source sha `d17ea317f78c` (비교 기준 `ast.after-7515.json`, 그 판 `22ea8c60964f` `:36-43`).
위 절들이 "최종" 이라 적은 sha 는 **중간판**이었다(리뷰 지적 — 표기를 고쳤다). 최종 파일에서 `ast.after-759r.json` 을 다시 뽑아 앞 절의 편집 후 열거와 대조했다 — 분기(종류 · 소스) · 반환 · raise · 호출 수가 **같다**(구조 동일). 바뀐 것은 sha 와 줄 좌표다.
