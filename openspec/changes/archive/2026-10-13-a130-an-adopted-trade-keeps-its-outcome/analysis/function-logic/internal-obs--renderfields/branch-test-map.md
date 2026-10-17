# Branch Test Map: `renderFields`

- Source: `internal/obs/notifier.go` (1029-1046) — `727a841e` 기준(a133 `35936cc8` 편집 뒤, 그 뒤로 이 파일 무변경).
- 출처: a133 이 편집한 함수가 a130 창 안에 끼어서 요구됨 — 겹친 change. 시험은 a133 · 기존 시험을 인용함.
- 측정(2026-10-13, HEAD `272dba00`): 시험마다 따로
  `go test -count=1 -run '^<Test>$' -covermode=count -coverpkg=./internal/obs,./internal/redact -coverprofile=<f> ./<pkg>`
  (전부 rc 0). 칸의 `a/b` = 분기 문장을 담은 블록의 실행 횟수 / 분기 본문(같은 줄, 조건 뒤에서 시작하는 블록)의 실행 횟수.
  `if` 는 `a > b` 이면 거짓 갈래도 지남.

| Branch | Anchor | Scenario | Test | 측정 |
|---|---|---|---|---|
| B1 | range at 1031:2 | 키 모으기 순회 | `TestA133CriticalFieldsLeaveNoAccount` · `TestCriticalAlertIsDurableBeforeItIsSent` | 1/10 · 1/2 |
| B2 | if at 1034:3 | 계좌 키는 건너뛰고(참) 나머지 키는 남음(거짓) | `TestA133CriticalFieldsLeaveNoAccount` · `TestA133BestEffortPathLeavesNoAccount` · `TestA133NormalRelayLeavesNoAccount` | 10/5 · 10/5 · 10/5 — 두 갈래 다 |
| B3 | range at 1042:2 | 정렬된 키마다 `k: v` 줄 | `TestA133CriticalFieldsLeaveNoAccount` · `TestCriticalAlertIsDurableBeforeItIsSent` | 1/5 · 1/2 |

- B2 참 갈래는 a133 시험만 지남 — `TestCriticalAlertIsDurableBeforeItIsSent` 는 2/0(거짓 갈래만)이라 B2 의 근거로 쓰지 않음.
- 미덮음 분기: 없음.
