# Function Logic Map: `_recording_moved` (Python, a122 task 7.5.2.3)

## task 7.5.2.3 — 재확인의 입력 집합은 판정의 입력 집합이다 (7.5.2.2 재리뷰, 2026-09-20)

`tools/logic-map/check_analysis.py:2447-2467` · 분기 2 · 반환 2 · raise 0 (새 함수).

새 함수. 쓰기 직전에 **거절 집합 전체**를 다시 묻는다. 순회는 133~219초이고 그 사이의 Go 편집은 역사에도 원장에도 없다 — 7.5.2.2 는 그 상태에서 기록을 **영구히** 만들었다. 순서가 곧 사유다: 역사 → 거절 → 원장.

| id | 줄 | 종류 | 소스 |
|---|---|---|---|
| B1 | 2462 | If | `if moved:` |
| B2 | 2467 | BoolOp | `refusal or _judged_state_moved(root, head, book)` |

## task 7.5.16 — 말하는 순서: 역사 → 원장 → 거절 (7.5.2.3 재리뷰 정확성, 2026-09-26)

> 편집 전 `ast.before-7516.json`(revision `1d1e5ca7`) · 편집 후 `ast.after-7516.json`. 중간판(`check_analysis.py` source sha `ff6142a55df0` — **최종이 아니다**, 보수 절 정정) 에서 `:3138-3164`.

편집 후 `tools/logic-map/check_analysis.py:3138-3164` · 분기 2 · 반환 2 · raise 0 · 호출 4 (`ast.after-7516.json`, source sha `ff6142a55df0`)
편집 전 `tools/logic-map/check_analysis.py:3063-3083` · 분기 2 · 반환 2 · raise 0 · 호출 4 (`ast.before-7516.json`, revision `1d1e5ca7`, source sha `b28398e26f3d`)

| 옛 id | 새 id | 새 줄 | 종류 | 소스(새) | 바뀐 것 |
|---|---|---|---|---|---|
| B1 | B1 | 3159 | If | `if moved:` | 같음 |
| B2 | — | — | BoolOp | `refusal or _judged_state_moved(root, head, book)` (옛) | **빠짐** |
| — | B2 | 3164 | BoolOp | `_judged_state_moved(root, head, book) or refusal` | **새** |

**결함.** 편집 전 B2 는 `refusal or _judged_state_moved(...)` — 거절이 원장보다 먼저 말했다. 걷는 동안 원장에 든 추적 파일이 바뀌면(기록
경로의 원장에서 증거 밖의 유일한 읽기는 `base-commit.txt` — 증거의 움직임은 `compute_landing` 의 `_raise_if_inputs_moved` 가 먼저 잡는다)
그 편집이 트리를 더럽히고 거절도 그 새 바이트로 사유를 만든다. 영수증 `analysis/harness/7516_order.py`(같은 모양을 세 리비전의 코드로):
`5a54f78d` → "the working tree has uncommitted changes to tracked files — commit them first"(리뷰가 적은 모양), `1d1e5ca7` → 6.4(b) 의
"cannot resolve the comparison base: … is not the one committed in HEAD"(6.4 가 거절 집합에서 base 대조를 더러운 트리 앞에 세웠다),
편집 뒤 → "`…/base-commit.txt` changed while this change was being judged — … run it again". 셋 다 기록을 안 썼다(`written=False`) —
결함은 **틀린 사유**이지 잘못된 기록이 아니다.

**수리.** B2 를 `_judged_state_moved(...) or refusal` 로. 거절은 여전히 **먼저 계산한다**(B1 뒤의 `_recording_refusal` 호출은 그대로) —
그 읽기(증거 · `base-commit.txt`)가 원장에 들어가야 마지막 대조가 본다. 바뀐 것은 말하는 순서뿐이다. 역사 물음(B1)은 여전히 맨 앞이다.

**이 수리가 B1 을 못에서 뽑았다** ([[a-new-guard-unpins-the-guards-behind-it]]). `_judged_state_moved` 도 역사를 먼저 묻고 이제 거절보다
먼저 말하므로, B1 을 지운 변이 AA8 이 같은 "HEAD moved" 문장을 내며 **살아남았다**(도달 계측 "못 쟀다"). B1 이 남아서 하는 일은 역사가
움직였을 때 거절을 **계산하지 않는 것**이고 출력으로는 안 갈린다. 지우지 않고(움직인 역사 위에서 증거 · base 를 다시 읽고 하한을 푸는
일을 막는다) 순서를 AST 로 못 박았다 — `test_the_history_is_asked_before_the_refusal_is_computed`, 재실행 AA8 CAUGHT.

## 보수 — 최종 파일로 다시 열거 (독립 주장정확성 리뷰 F11, 2026-09-26)

최종 `tools/logic-map/check_analysis.py:3188-3214` · 분기 2 · 반환 2 · raise 0 · 호출 4 · source sha `ff590b79db6e` (비교 기준 `ast.after-7516.json`, 그 판 `ff6142a55df0` `:3138-3164`).
위 절들이 "최종" 이라 적은 sha 는 **중간판**이었다(리뷰 지적 — 표기를 고쳤다). 최종 파일에서 `ast.after-759r.json` 을 다시 뽑아 앞 절의 편집 후 열거와 대조했다 — 분기(종류 · 소스) · 반환 · raise · 호출 수가 **같다**(구조 동일). 바뀐 것은 sha 와 줄 좌표다. 줄이 실제로 밀렸다(`:3138` → `:3188` — 보수가 앞에 새 깔때기 둘 · 해소기 수리 · 주석을 더했다).
