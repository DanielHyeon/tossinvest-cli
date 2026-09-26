# Function Logic Map: `go_inputs` (Python, `execution_baseline.py`)

## task 7.5.15 — 시한 (2026-09-26)

> 편집 전 `ast.before-7515.json`(revision `1d1e5ca7`) · 편집 후 `ast.after-7515.json` — 분기 · 반환 · raise · 호출 수 불변, 정렬 결과 전부 같음.
> 중간판(`execution_baseline.py` source sha `22ea8c60964f` — **최종이 아니다**, 보수 절 정정) 에서 `:228-279`.

편집 후 `tools/logic-map/execution_baseline.py:228-279` · 분기 25 · 반환 1 · raise 6 · 호출 38 (`ast.after-7515.json`, source sha `22ea8c60964f`)
편집 전 `tools/logic-map/execution_baseline.py:218-267` · 분기 25 · 반환 1 · raise 6 · 호출 38 (`ast.before-7515.json`, revision `1d1e5ca7`, source sha `504ec55c505a`)

| 옛 id | 새 id | 새 줄 | 종류 | 소스(새) | 바뀐 것 |
|---|---|---|---|---|---|
| B1 | B1 | 230 | If | `if tags:` | 같음 |
| B2 | B2 | 236 | If | `if process.returncode:` | 같음 |
| B3 | B3 | 237 | BoolOp | `process.stderr.strip() or 'go package enumeration failed'` | 같음 |
| B4 | B4 | 240 | While | `while offset < len(process.stdout):` | 같음 |
| B5 | B5 | 241 | BoolOp | `offset < len(process.stdout) and process.stdout[offset].isspace()` | 같음 |
| B6 | B6 | 241 | While | `while offset < len(process.stdout) and process.stdout[offset].isspace():` | 같음 |
| B7 | B7 | 243 | If | `if offset >= len(process.stdout):` | 같음 |
| B8 | B8 | 245 | Try | `try:` | 같음 |
| B9 | B9 | 247 | ExceptHandler | `except ValueError as error:` | 같음 |
| B10 | B10 | 249 | If | `if not isinstance(package, dict):` | 같음 |
| B11 | B11 | 252 | If | `if not isinstance(directory, str):` | 같음 |
| B12 | B12 | 253 | If | `if any((field in package for field in fields)):` | 같음 |
| B13 | B13 | 253 | comprehension | `for field in fields` | 같음 |
| B14 | B14 | 257 | Try | `try:` | 같음 |
| B15 | B15 | 259 | ExceptHandler | `except ValueError:` | 같음 |
| B16 | B16 | 261 | For | `for field in fields:` | 같음 |
| B17 | B17 | 263 | If | `if value is None:` | 같음 |
| B18 | B18 | 265 | BoolOp | `not isinstance(value, list) or not all((isinstance(name, str) for name in value))` | 같음 |
| B19 | B19 | 265 | If | `if not isinstance(value, list) or not all((isinstance(name, str) for name in value)):` | 같음 |
| B20 | B20 | 265 | comprehension | `for name in value` | 같음 |
| B21 | B21 | 267 | For | `for name in value:` | 같음 |
| B22 | B22 | 269 | If | `if candidate.is_absolute():` | 같음 |
| B23 | B23 | 270 | Try | `try:` | 같음 |
| B24 | B24 | 272 | ExceptHandler | `except ValueError:` | 같음 |
| B25 | B25 | 276 | If | `if '..' in candidate.parts:` | 같음 |

`subprocess.run` 에 `timeout=60`: 이웃 `check_analysis.go_functions` 의 `go run`. 실측(2026-09-26, `1d1e5ca7`): 따뜻한 캐시 8~13초, 차가운 첫 판(`-tags tossos_testseams`) **49.3초** — 여유가 11초다. 넘으면 판정 줄이 되고 다시 돌리면 캐시가 따뜻하다(막는 쪽으로 틀린다). `TimeoutExpired` 는 `AdoptionError` 가 아니라 `validate` 를 부르는 `resolve_base` 의
`except AdoptionError` 를 지나 `_judged` · `_recording_refusal` 의 `GATE_FAULTS` 경계에서 판정 줄이 된다. 기안 CLI(`main`)는
`(AdoptionError, OSError, ValueError)` 만 받으므로 거기서 시한이 나면 traceback 이다 — 판정 경로가 아니라 이 로트 밖이다.

## 보수 — 최종 파일로 다시 열거 (독립 주장정확성 리뷰 F11, 2026-09-26)

최종 `tools/logic-map/execution_baseline.py:230-282` · 분기 25 · 반환 1 · raise 6 · 호출 38 · source sha `d17ea317f78c` (비교 기준 `ast.after-7515.json`, 그 판 `22ea8c60964f` `:228-279`).
위 절들이 "최종" 이라 적은 sha 는 **중간판**이었다(리뷰 지적 — 표기를 고쳤다). 최종 파일에서 `ast.after-759r.json` 을 다시 뽑아 앞 절의 편집 후 열거와 대조했다 — 분기(종류 · 소스) · 반환 · raise · 호출 수가 **같다**(구조 동일). 바뀐 것은 sha 와 줄 좌표다.
