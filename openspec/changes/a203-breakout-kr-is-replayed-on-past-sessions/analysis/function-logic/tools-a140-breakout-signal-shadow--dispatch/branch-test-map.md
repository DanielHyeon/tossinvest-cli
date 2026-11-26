# Branch Test Map: `dispatch`

편집 뒤(현재 AST, 56-70) 갈래. 편집 전 표(B1~B5)는 커밋 `f71ad60ba` 의 이 파일.
GREEN 원문: `../../verify/go-test-a140.txt`. RED 원문: `../../verify/red-dispatch-before-edit.txt` — 시험을 편집 뒤에 썼으므로 RED 는 **사후**에 dispatch 를 편집 전 본문으로 되돌려 잰 것(rc 1, 첫 단언 B1 에서 멈춤 — B5 · B6 의 RED 는 따로 관측하지 않음).

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 인자 없음 → 오류(세 이름을 말함) | `TestA203DispatchRoutesReplay` | yes(사후) — 편집 전 본문에서 「no args: a subcommand is required: "run" or "report"」 FAIL | pass |
| B2 | switch 60 `args[0]` — 이름별 갈래 선택 | 아래 B3~B6 가 함께 잼 | — | pass |
| B3 | `run` | `TestA166Breakout*`(runCommand 직접) | no | pass |
| B4 | `report` → 빈 원장 보고 | `TestA203DispatchRoutesReplay` | no | pass |
| B5 | `replay` → replayCommand(단계 없음 오류) | `TestA203DispatchRoutesReplay` | 따로 관측 안 함(같은 시험이 B1 에서 먼저 멈춤) | pass |
| B6 | 모르는 이름 → 오류(세 이름) | `TestA203DispatchRoutesReplay` | 따로 관측 안 함(같은 이유) | pass |
