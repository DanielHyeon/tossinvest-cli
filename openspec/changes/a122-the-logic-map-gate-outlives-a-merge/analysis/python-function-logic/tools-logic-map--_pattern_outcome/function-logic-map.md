# Function Logic Map: `_pattern_outcome` (Python, a122 task 7.5.2.3)

## task 7.5.2.3 — 재확인의 입력 집합은 판정의 입력 집합이다 (7.5.2.2 재리뷰, 2026-09-20)

`tools/logic-map/check_analysis.py:824-828` · 분기 1 · 반환 1 · raise 0 (새 함수).

새 함수. 트리 순회의 지문. 시험 함수 색인은 `*_test.go` **집합**이 판정의 입력이다(a112 실측 962 파일) — 파일 하나가 생겨도 인용 판정이 낡는다.

| id | 줄 | 종류 | 소스 |
|---|---|---|---|
| B1 | 827 | comprehension | ` for path in matches` |
