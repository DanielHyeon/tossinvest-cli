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
