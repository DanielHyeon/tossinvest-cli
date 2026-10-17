# Function Logic Map: `_decoded` (Python, a122 task 7.5.2)

## task 7.5.2 — 판정은 한 번 읽은 바이트로 선다 (수리한 트리의 재리뷰, 2026-09-19)

`tools/logic-map/check_analysis.py:1346-1352` · 분기 0 · 반환 1 · raise 0 (새 함수).

새 함수. `Path.read_text(encoding="utf-8")` 와 같은 글자(줄바꿈 변환 `newline=None` 까지) — `io.TextIOWrapper` 가 `read_text` 가 여는 것과 같은 래퍼다. 시험이 네 입력(CRLF · 홀로 CR · 한글 · 빈 것)과 못 푸는 바이트에서 `read_text` 와 같음을 단언한다(변이 W22).

| id | 줄 | 종류 | 소스 |
|---|---|---|---|
