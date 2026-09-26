# Function Logic Map: `_pattern_outcome` (Python, a122 task 7.5.2.3)

## task 7.5.2.3 — 재확인의 입력 집합은 판정의 입력 집합이다 (7.5.2.2 재리뷰, 2026-09-20)

`tools/logic-map/check_analysis.py:824-828` · 분기 1 · 반환 1 · raise 0 (새 함수).

새 함수. 트리 순회의 지문. 시험 함수 색인은 `*_test.go` **집합**이 판정의 입력이다(a112 실측 962 파일) — 파일 하나가 생겨도 인용 판정이 낡는다.

| id | 줄 | 종류 | 소스 |
|---|---|---|---|
| B1 | 827 | comprehension | ` for path in matches` |

## task 7.5.9 — 지웠다 (2026-09-26)

`ast.before-759.json`(revision `1d1e5ca7`)이 지우기 직전의 열거다. 트리 순회의 지문였다. 시험 색인과 인용 해소가 추적 목록(`_tracked` ·
`_tracked_outcome`)으로 옮겨 호출자가 0 이 됐다. 이 함수가 들고 있던 결함 둘은 함수와 같이 사라졌다 — 지문이 경로를 `\n` 으로 이어
충돌했고(7.5.10, 영수증 `analysis/harness/7510_collide.py` 가 `1d1e5ca7` 에서 `COLLIDE` 를 찍는다), `rglob` 이 못 읽는 하위 트리의
`OSError` 를 삼켰다(7.5.11, 3.12.3 영수증 `7511_isdir.py`).
