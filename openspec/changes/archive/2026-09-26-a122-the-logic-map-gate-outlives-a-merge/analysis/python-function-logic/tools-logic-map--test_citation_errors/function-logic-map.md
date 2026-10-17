# Function Logic Map: `test_citation_errors` (Python, a122 task 7.5.2.3)

## task 7.5.2.3 — 재확인의 입력 집합은 판정의 입력 집합이다 (7.5.2.2 재리뷰, 2026-09-20)

`tools/logic-map/check_analysis.py:1593-1636` · 분기 8 · 반환 1 · raise 0 (편집 전 L1349-1387 · 분기 6 · 반환 1 · raise 0, `ast.before-7.5.2.3.json` = revision `1d12520c`).

줄 수를 세는 읽기도 깔때기로. 고를 때는 정규 파일이었는데 읽을 때 사라졌으면 그 줄에 대해 **아무 주장도 하지 않는다** — 그 변화는 원장에 남아 끝의 재확인이 댄다.

| id | 줄 | 종류 | 소스 |
|---|---|---|---|
| B1 | 1597 | For | `for name in sorted(set(CITED_TEST.findall(text))):` |
| B2 | 1598 | If | `if name not in index:` |
| B3 | 1619 | For | `for line in text.splitlines():` |
| B4 | 1620 | For | `for basename, raw in CITED_TEST_LINE.findall(line):` |
| B5 | 1622 | If | `if path is None:` |
| B6 | 1625 | Try | `try:` |
| B7 | 1627 | ExceptHandler | `except OSError:` |
| B8 | 1631 | If | `if number > total:` |

## task 7.5.9 — 시험 인용은 추적 파일로만 충족된다 (7.5.2.3 재리뷰 정확성, 2026-09-26)

> 편집 전 `ast.before-759.json` · 편집 후 `ast.after-759.json` — 분기 **8 → 8**, 정렬 결과 전부 같음. 중간판(`check_analysis.py` source sha `ff6142a55df0` — **최종이 아니다**, 보수 절 정정) 에서 `:2253-2298`.

편집 후 `tools/logic-map/check_analysis.py:2218-2263` · 분기 8 · 반환 1 · raise 0 · 호출 13 (`ast.after-759.json`, source sha `194ec1a29941`)
편집 전 `tools/logic-map/check_analysis.py:2180-2223` · 분기 8 · 반환 1 · raise 0 · 호출 13 (`ast.before-759.json`, revision `1d1e5ca7`, source sha `b28398e26f3d`)

| 옛 id | 새 id | 새 줄 | 종류 | 소스(새) | 바뀐 것 |
|---|---|---|---|---|---|
| B1 | B1 | 2222 | For | `for name in sorted(set(CITED_TEST.findall(text))):` | 같음 |
| B2 | B2 | 2223 | If | `if name not in index:` | 같음 |
| B3 | B3 | 2246 | For | `for line in text.splitlines():` | 같음 |
| B4 | B4 | 2247 | For | `for basename, raw in CITED_TEST_LINE.findall(line):` | 같음 |
| B5 | B5 | 2249 | If | `if path is None:` | 같음 |
| B6 | B6 | 2252 | Try | `try:` | 같음 |
| B7 | B7 | 2254 | ExceptHandler | `except OSError:` | 같음 |
| B8 | B8 | 2258 | If | `if number > total:` | 같음 |

바뀐 것은 둘이고 분기는 안 바뀌었다: 거절 문장 `anywhere in the tree` → `in any tracked file`(저자의 디스크에만 있는 시험이 여기서
"없다" 이므로 "트리 어디에도 없다" 는 틀린 곳을 가리킨다), `resolve_test_file` 에 `index.files` 를 넘긴다. 문장을 단언하는 기존 시험 0
(`rg 'anywhere in the tree|not a Go test function'`).

## 보수 — 최종 파일로 다시 열거 (독립 주장정확성 리뷰 F11, 2026-09-26)

최종 `tools/logic-map/check_analysis.py:2301-2346` · 분기 8 · 반환 1 · raise 0 · 호출 13 · source sha `ff590b79db6e` (비교 기준 `ast.after-759.json`, 그 판 `194ec1a29941` `:2218-2263`).
위 절들이 "최종" 이라 적은 sha 는 **중간판**이었다(리뷰 지적 — 표기를 고쳤다). 최종 파일에서 `ast.after-759r.json` 을 다시 뽑아 앞 절의 편집 후 열거와 대조했다 — 분기(종류 · 소스) · 반환 · raise · 호출 수가 **같다**(구조 동일). 바뀐 것은 sha 와 줄 좌표다.
