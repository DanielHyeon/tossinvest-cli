# CodeGraphContext supporting context — a208

- 도구: `codegraphcontext analyze callers <심볼>`(색인 갱신 없이 읽기만, 2026-11-28). 색인이 이 기계의 여러 worktree(`~/tossos-work/a0xx` ·
  `~/tossos-gate/*`)를 함께 담고 있어 같은 호출자가 worktree 마다 한 줄씩 나옴 — 이름만 읽음.

| 심볼 | CGC 호출자(이름) | CodeGraph 와 |
|---|---|---|
| `EvaluateLadder` | `EvaluateLadderSnapshot` | 같음 |
| `judgeLadder` | `judge` | 같음 |
| `adoptOne` | `adopt` | 같음 |
| `SaveEngineAdoption` | `Save`(cmd/tossctl `consoleAdoptionSettings.Save`) | **다름** — CodeGraph 는 호출자 없음 |
| `handleSettingsSave` | 결과 없음 | CodeGraph 는 route 로 찾음 |

누락 후보: 없음(설정 저장 seam 의 호출자를 CGC 가 하나 더 찾음 — 아래 조정).
GBrain 의미 검색: **안 돌림**(그 채널은 안 봤음).
