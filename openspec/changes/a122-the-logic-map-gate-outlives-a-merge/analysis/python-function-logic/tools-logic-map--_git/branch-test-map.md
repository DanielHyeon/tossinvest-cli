# Branch Test Map: `_git` (Python, `execution_baseline.py`)

## task 7.5.15 (2026-09-26)

짝은 손으로 고르지 않았다 — 갈래마다 변이를 걸어 실제로 빨개진 시험을 옮겼다(`analysis/harness/75_mut.py` 창 `228:246` · `97:98` · `99:100` · `101:102` · `102:103` · `112:113` · `239:240` · `244:246`(재실행), `7515_mut.py`; 사본 `A122_HARNESS_WORK` · pid 별 · 무변이 대조군 창 양끝 GREEN `Ran 407`, 구조 시험을 더한 뒤 창 `102:103` · `244:246` 은 `Ran 408`). '—' 는 변이 대신 직접 시험이다.

| 갈래 | 변이 (`7515_mut.py`) | 잡는 시험 |
|---|---|---|
| `timeout=` | EB1 | `test_every_child_process_in_the_gate_modules_names_a_timeout`(구조) **하나뿐** — 행동 시험(`…verdict_starts…`)의 픽스처는 이관 경로를 안 탄다 |
| 그 밖 분기 | (이 로트가 안 바꿈) | 이 표 밖 — 기존 이관 시험들 |
