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
