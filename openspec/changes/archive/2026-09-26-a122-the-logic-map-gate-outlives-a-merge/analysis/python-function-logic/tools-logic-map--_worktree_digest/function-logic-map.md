# Function Logic Map: `_worktree_digest` (Python, a122 task 7.5.2.3)

## task 7.5.2.3 — 재확인의 입력 집합은 판정의 입력 집합이다 (7.5.2.2 재리뷰, 2026-09-20)

`tools/logic-map/check_analysis.py:855-860` · 분기 2 · 반환 2 · raise 0 (새 함수).

새 함수. 워킹트리 파일의 sha256 — `is_file()` + `read_bytes()` 두 syscall 이 깔때기 하나가 된다.

| id | 줄 | 종류 | 소스 |
|---|---|---|---|
| B1 | 857 | Try | `try:` |
| B2 | 859 | ExceptHandler | `except OSError:` |
