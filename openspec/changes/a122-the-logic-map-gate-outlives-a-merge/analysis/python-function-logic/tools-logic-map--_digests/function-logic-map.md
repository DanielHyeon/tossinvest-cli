# Function Logic Map: `_digests` (Python, a122 task 7.5.2)

## task 7.5.2 — 판정은 한 번 읽은 바이트로 선다 (수리한 트리의 재리뷰, 2026-09-19)

`tools/logic-map/check_analysis.py:613-618` · 분기 2 · 반환 1 · raise 0 (새 함수).

새 함수. (경로, 바이트 해시). 못 읽은 자리는 빈 해시.

| id | 줄 | 종류 | 소스 |
|---|---|---|---|
| B1 | 615 | comprehension | ` for path, raw in reads` |
| B2 | 616 | IfExp | `'' if raw is None else hashlib.sha256(raw).hexdigest()` |
