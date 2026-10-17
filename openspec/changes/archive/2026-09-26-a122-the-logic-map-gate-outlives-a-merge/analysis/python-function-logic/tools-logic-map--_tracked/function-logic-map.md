# Function Logic Map: `_tracked` (Python, a122 task 7.5.9 — 새 함수)

## task 7.5.9 — 새 함수 (2026-09-26)

편집 후 `tools/logic-map/check_analysis.py:1433-1441` · 분기 1 · 반환 1 · raise 1 · 호출 4 (`ast.after-759.json`, source sha `194ec1a29941`)

| id | 줄 | 종류 | 소스 |
|---|---|---|---|
| B1 | 1439 | If | `if isinstance(value, BaseException):` |

추적 목록을 묻는 **유일한 자리**. 결과(실패 포함)를 원장 `tracked` 에 적고 실패면 올린다 — 빈 목록으로 두면 모든 인용이 "없는 시험" 이
되어 이유가 사라진다. 중간판(`check_analysis.py` source sha `ff6142a55df0` — **최종이 아니다**, 보수 절 정정) 에서 `:1468-1476`.

## 보수 — 최종 파일로 다시 열거 (독립 주장정확성 리뷰 F11, 2026-09-26)

최종 `tools/logic-map/check_analysis.py:1506-1514` · 분기 1 · 반환 1 · raise 1 · 호출 4 · source sha `ff590b79db6e` (비교 기준 `ast.after-759.json`, 그 판 `194ec1a29941` `:1433-1441`).
위 절들이 "최종" 이라 적은 sha 는 **중간판**이었다(리뷰 지적 — 표기를 고쳤다). 최종 파일에서 `ast.after-759r.json` 을 다시 뽑아 앞 절의 편집 후 열거와 대조했다 — 분기(종류 · 소스) · 반환 · raise · 호출 수가 **같다**(구조 동일). 바뀐 것은 sha 와 줄 좌표다.
