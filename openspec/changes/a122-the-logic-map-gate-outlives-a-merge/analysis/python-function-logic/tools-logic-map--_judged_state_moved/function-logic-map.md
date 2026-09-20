# Function Logic Map: `_judged_state_moved` (Python, a122 task 7.5.2.2)

## task 7.5.2.2 — 스냅숏은 끝에서 디스크와 다시 대조한다 (7.5.2.1 재리뷰, 2026-09-20)

`tools/logic-map/check_analysis.py:1881-1897` · 분기 2 · 반환 3 · raise 0 (새 함수).

새 함수. 판정한 `head` 와 증거가 지금도 거기 있는가 — `HEAD` 를 다시 풀어 비교하고, 같으면 증거를 다시 읽어 `Evidence` 전체를 비교한다. 이 읽기의 바이트는 판정에 안 들어간다. `check` 의 출구 · 기록 명령의 쓰기 직전이 같은 함수 · 같은 문장(`JUDGED_STATE_MOVED`)을 쓴다. 7.5.2.1 의 "실행 중 커밋은 판정을 안 멈춘다" 를 되돌린다(사람 결정).

| id | 줄 | 종류 | 소스 |
|---|---|---|---|
| B1 | 1893 | If | `if now != head:` |
| B2 | 1895 | If | `if _read_evidence(evidence.directory) != evidence:` |
