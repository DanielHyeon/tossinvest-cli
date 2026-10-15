# a138 검증 영수증

워킹트리 = base `ab5edc48` + a138 편집(GREEN 커밋 직전), 2026-10-15 KST. `rtk proxy` 로 잘리지 않은 출력.

| 명령 | 대상 | 결과 |
|---|---|---|
| `gofmt -l internal cmd` | 전체 | 출력 0 |
| `go vet ./...` · `go vet -tags tossos_testseams ./...` | 전체 | exit 0 · exit 0 |
| `go test -count=1 -timeout 40m ./internal/journal` | journal 전체 | ok 627.9s |
| `go test -count=1 -timeout 40m ./internal/app/engine` | engine 전체 | ok 415.3s |
| `go test -count=1 ./internal/execgw` | execgw 전체 | ok 163.9s |
| `go test -count=1 ./cmd/tossctl ./internal/console ./internal/risk ./internal/config ./internal/measure` | 다섯 | 전부 ok — console 은 census 시험 둘(`TestTheSubmitContractIsUnchanged` · `TestTheSettingsScreensAskForNoTyping`)에 새 칸을 등록한 뒤 ok 40.8s |
| `go test -tags tossos_testseams ./internal/execgw ./cmd/tossctl` | 태그 | ok · ok |
| `go test -tags tossos_testseams -run 'A138\|A136\|A066\|QFinal\|FirstLeg\|Reserve\|Issu' ./internal/journal ./internal/app/engine` | 태그 · 범위 | ok · ok |
| `go test -race -run A138` | journal · execgw · config · risk · console · app/engine | 6/6 ok, DATA RACE 0 |
| `openspec validate a138-max-open-positions --strict --no-interactive` | — | valid |
| `python3 tools/logic-map/check_analysis.py --change a138-max-open-positions` | 번들 34 | 커밋 전: 미추적 시험 이름 경고만(커밋으로 해소) — 커밋 뒤 재실행 기록은 아래 |
| 렌더(규칙 13) | 격리 config · HOME · TOSSOS_DATA_DIR 로 로컬 콘솔(`tossctl console --port 18938`), `/settings/daily` 를 chrome-devtools 로 | 현재 값 행 「최대 보유 포지션 수 미설정 — 기본값 6 종목이 적용된다」, 고급 폼 칸(label 「최대 보유 포지션 수 (종목)」 · `inputmode=numeric` · `aria-describedby=open-positions-help` · 보임), 콘솔 메시지 0, 요청 1(GET 200). 저장 · 토글 · 승인 버튼 미클릭. 콘솔 프로세스 종료 확인 |
| 규칙 12 | `ui-skills-root` → `fixing-accessibility` 1개 | label 안 입력 · 도움말 aria-describedby 연결 · inputmode |

RED 영수증: `analysis/red/red-*.txt`(컴파일 RED — base + a138 시험), `red-journal-behaviour.txt`(판정 함수 빈 몸 — 9 시험 FAIL), `red-console.txt`(콘솔 · tossctl 세 파일만 HEAD 로 되돌린 상태 — 콘솔 시험 2 · tossctl 1 FAIL).
절차 기록: 콘솔 시험은 콘솔 GREEN 뒤에 썼다 — RED 는 production 세 파일을 HEAD 로 되돌린 상태로 따로 관찰했다(순서 역전, review.md).

## 구현 리뷰 수리 뒤 (2026-10-15)

| 명령 | 대상 | 결과 |
|---|---|---|
| 지연 실측(임시 시험, 격리 worktree `86442ad7` — 커밋 안 함) | `checkOpenPositions` 를 원장 tx 안에서 100회, CLOSED 포지션 3000 + 해제 RECONCILE 3000 + HELD 5 | 평균 **0.97 ms/회**(리뷰어 측정 0.85 ms 와 같은 크기) — 쓰기 잠금 영향 무시 수준 |
| `go test ./internal/config -run A138` | P2-1 새 파일 갈래 | RED(`analysis/red/red-review-p2-1.txt` — limits_io.go 만 HEAD) → GREEN |
| `go test ./internal/journal -run A138` · `./internal/execgw -run A138` | ④ 맞춤 갈래 4 · 진술 7 단언 · 비진입 결정 HELD | ok · ok |
