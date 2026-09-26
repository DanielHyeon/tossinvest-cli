# Function Logic Map: `_reads_moved` (Python, a122 task 7.5.2.3)

## task 7.5.2.3 — 재확인의 입력 집합은 판정의 입력 집합이다 (7.5.2.2 재리뷰, 2026-09-20)

`tools/logic-map/check_analysis.py:887-901` · 분기 5 · 반환 3 · raise 0 (새 함수).

새 함수. 원장에 적힌 것을 **전부** 다시 읽어 지문을 견준다. 이 읽기의 바이트는 판정에 안 들어간다. 7.5.2.2 는 이 집합을 손으로 골라(`HEAD` + `Evidence`) 판정이 읽는 1,739 경로 중 149 만 봤다(a112 실측).

| id | 줄 | 종류 | 소스 |
|---|---|---|---|
| B1 | 892 | If | `if book.diverged:` |
| B2 | 895 | Try | `try:` |
| B3 | 896 | For | `for (kind, key), before in sorted(book.seen.items()):` |
| B4 | 897 | If | `if _PROBE_NOW[kind](key) != before:` |
| B5 | 898 | IfExp | `key.split(chr(10))[1] if kind == 'glob' else _shown(root, key)` |

## task 7.5.9 — 시험 인용은 추적 파일로만 충족된다 (7.5.2.3 재리뷰 정확성, 2026-09-26)

> 편집 전 `ast.before-759.json` · 편집 후 `ast.after-759.json`. 중간판(`check_analysis.py` source sha `ff6142a55df0` — **최종이 아니다**, 보수 절 정정) 에서 `:1528-1542`.

편집 후 `tools/logic-map/check_analysis.py:1493-1507` · 분기 5 · 반환 3 · raise 0 · 호출 5 (`ast.after-759.json`, source sha `194ec1a29941`)
편집 전 `tools/logic-map/check_analysis.py:1474-1488` · 분기 5 · 반환 3 · raise 0 · 호출 7 (`ast.before-759.json`, revision `1d1e5ca7`, source sha `b28398e26f3d`)

| 옛 id | 새 id | 새 줄 | 종류 | 소스(새) | 바뀐 것 |
|---|---|---|---|---|---|
| B1 | B1 | 1498 | If | `if book.diverged:` | 같음 |
| B2 | B2 | 1501 | Try | `try:` | 같음 |
| B3 | B3 | 1502 | For | `for (kind, key), before in sorted(book.seen.items()):` | 같음 |
| B4 | B4 | 1503 | If | `if _PROBE_NOW[kind](key) != before:` | 같음 |
| B5 | — | — | IfExp | `key.split(chr(10))[1] if kind == 'glob' else _shown(root, key)` (옛) | **빠짐** |
| — | B5 | 1504 | IfExp | `'the tracked file list' if kind == 'tracked' else _shown(root, key)` | **새** |

원장 종류 `glob`(트리 순회)이 없어지고 `tracked`(추적 목록)가 들어왔다. 순회는 키에 뿌리와 패턴을 담아 패턴을 이름으로 댔고, 추적
목록은 키가 뿌리 하나라 **"the tracked file list"** 로 댄다(`_shown(root, root)` 은 `.` 이다). `_PROBE_NOW` 표도 `glob` → `tracked`.

## 보수 — 최종 파일로 다시 열거 (독립 주장정확성 리뷰 F11, 2026-09-26)

최종 `tools/logic-map/check_analysis.py:1567-1581` · 분기 5 · 반환 3 · raise 0 · 호출 5 · source sha `ff590b79db6e` (비교 기준 `ast.after-759.json`, 그 판 `194ec1a29941` `:1493-1507`).
위 절들이 "최종" 이라 적은 sha 는 **중간판**이었다(리뷰 지적 — 표기를 고쳤다). 최종 파일에서 `ast.after-759r.json` 을 다시 뽑아 앞 절의 편집 후 열거와 대조했다 — 분기(종류 · 소스) · 반환 · raise · 호출 수가 **같다**(구조 동일). 바뀐 것은 sha 와 줄 좌표다.
