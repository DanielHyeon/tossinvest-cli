# Function Logic Map: `_names_outcome` (Python, a122 task 7.5.9 보수 — 새 함수)

## 새 함수 (2026-09-26)

편집 후 `tools/logic-map/check_analysis.py:1449-1459` · 분기 4 · 반환 2 · raise 0 · 호출 9 (`ast.after-759r.json`, source sha `ff590b79db6e`)

| id | 줄 | 종류 | 소스 |
|---|---|---|---|
| B1 | 1454 | Try | `try:` |
| B2 | 1456 | ExceptHandler | `except OSError as exc:` |
| B3 | 1458 | comprehension | `for raw in (name.encode('utf-8') for name in names)` |
| B4 | 1458 | comprehension | `for name in names` |

이름만 읽는 목록의 지문 — `os.listdir`(stat 0)을 정렬하고 7.5.10 과 같은 단사 인코딩(이름 길이 8바이트 · 이름)으로 해시. 실패는 `(종류:errno, 예외)`.
