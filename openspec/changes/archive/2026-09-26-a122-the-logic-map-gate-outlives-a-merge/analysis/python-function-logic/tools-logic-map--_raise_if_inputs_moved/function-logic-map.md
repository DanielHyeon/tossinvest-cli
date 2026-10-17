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

## task 7.5.2.1 — 판정이 읽는 입력을 전부 세고 하나씩 묶는다 (7.5.2 재리뷰, 2026-09-20)

`tools/logic-map/check_analysis.py:1797-1814` · 분기 4 · 반환 0 · raise 3 (편집 전 L1732-1738 · 분기 1 · 반환 0 · raise 1, `ast.before-7.5.2.1.json` = revision `fc35eb2d`).

`(root, inputs)` — 증거를 다시 읽어 **먼저** `Evidence` 전체(존재 · 목록 · 바이트)를 비교하고, 같을 때만 고정 목록을 고른다. 고르기의 `ValueError` 도 움직임이다 — 7.5.2 는 고정 목록부터 골라서 "움직였다" 대신 `escapes repository` 를 냈다(레드팀). 해시 대신 바이트를 그대로 비교한다(재리뷰 단순화). 역사는 비교하지 않는다.

| id | 줄 | 종류 | 소스 |
|---|---|---|---|
| B1 | 1807 | If | `if now != inputs.evidence:` |
| B2 | 1809 | Try | `try:` |
| B3 | 1811 | ExceptHandler | `except ValueError as exc:` |
| B4 | 1813 | If | `if bundles != inputs.bundles:` |

| raise 줄 | 소스 |
|---|---|
| 1808 | `raise RuntimeError(INPUTS_MOVED)` |
| 1812 | `raise RuntimeError(INPUTS_MOVED) from exc` |
| 1814 | `raise RuntimeError(INPUTS_MOVED)` |
