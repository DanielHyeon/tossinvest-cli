# Function Logic Map: `_failed` (Python, a122 task 7.5.2.3)

## task 7.5.2.3 — 재확인의 입력 집합은 판정의 입력 집합이다 (7.5.2.2 재리뷰, 2026-09-20)

`tools/logic-map/check_analysis.py:744-746` · 분기 0 · 반환 1 · raise 0 (새 함수).

새 함수. 실패의 지문(종류 + errno). 권한이 풀린 것도, 없던 파일이 생긴 것도 "달라졌다" 다.

| id | 줄 | 종류 | 소스 |
|---|---|---|---|

## task 7.5.38 — 실패 지문에 이름 (2026-09-27)

> 편집 전 `ast.before-7538.json`(HEAD `5406cac1`) · 편집 후 `ast.after-7538.json`(최종 `check_analysis.py` `463982f42151`) — 분기 0 → 0.

지문 `종류:errno` → `종류:errno:repr(filename)`. 목록의 실패(`UnstatableEntry`)는 목록 **안의 한 항목**의 실패라 이름이 없으면 판정 중 끊긴 항목이 `a` → `b` 로 바뀌어도 같은
지문이었다(RED: 두 디렉터리의 지문 `UnstatableEntry:2` 로 같음 · 원장 재확인 `''`). 파일 실패는 키가 곧 이름이라 바뀌는 판정이 없다.
