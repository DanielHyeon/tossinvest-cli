# Function Logic Map: `_raise_if_inputs_moved` (Python, a122 task 7.5.2)

## task 7.5.2 — 판정은 한 번 읽은 바이트로 선다 (수리한 트리의 재리뷰, 2026-09-19)

`tools/logic-map/check_analysis.py:1732-1738` · 분기 1 · 반환 0 · raise 1 (새 함수).

새 함수. 수락 직전(`compute_landing`)과 선언 경로의 거절 앞(`resolve_landing`)이 **같은** 재확인에 묻는다. 문장은 `INPUTS_MOVED` 한 벌.

| id | 줄 | 종류 | 소스 |
|---|---|---|---|
| B1 | 1737 | If | `if _evidence_fingerprint(root, analysis)[0] != inputs.fingerprint:` |

| raise 줄 | 소스 |
|---|---|
| 1738 | `raise RuntimeError(INPUTS_MOVED)` |
