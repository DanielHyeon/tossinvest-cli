# a133 검증 영수증

격리 worktree(저장소 밖 scratchpad, `git worktree add --detach … 35936cc8`, 작업 트리 변경 0) — 병행 a130 의 미커밋 Go 편집이 판정에 들어가지 않게.

| 명령 | 결과 |
|---|---|
| `python3 tools/logic-map/check_analysis.py --change a133-an-alert-carries-no-account` | rc 0 — required 2 · `evidence complete or diff-proven exempt` |
| `make lint` | rc 0 |
| `go test -race -timeout 45m -count=1 ./internal/obs ./internal/redact ./cmd/tossctl ./internal/flatten` — obs | ok 284.173s |
| `go test -race -timeout 45m -count=1 ./internal/obs ./internal/redact ./cmd/tossctl ./internal/flatten` — redact | ok 5.214s |
| `go test -race -timeout 45m -count=1 ./internal/obs ./internal/redact ./cmd/tossctl ./internal/flatten` — cmd/tossctl | ok 180.328s |
| `go test -race -timeout 45m -count=1 ./internal/obs ./internal/redact ./cmd/tossctl ./internal/flatten` — flatten | ok 77.816s |
| `go test -race -timeout 45m -count=1 ./internal/app/engine/` | ok 2433.024s |
| `make test-seams` | rc 2 — 실패 1: `internal/position` `TestExitEligibilityHasOneSpelling` — a130 커밋(`dbb0681c`, base)의 `internal/journal/trade_outcome_backfill.go` · `trade_outcome_refusal.go` 가 `adoption_id` 를 씀. a133 은 이 파일들을 안 만짐(a133 Go 변경은 internal/obs 3 + 새 시험 4). a130 의 미커밋 `eligibility_drift_test.go` 편집이 그 수리. 나머지 패키지 전부 ok |
| `make sdd-check-ci` | rc 0 |


## 최종 판(구현 리뷰 수리 뒤) — 격리 worktree `2c2daae7`(작업 트리 변경 0)

| 명령 | 결과 |
|---|---|
| `check_analysis.py --change a133-…` | rc 0 — `evidence complete or diff-proven exempt`(번들 3: Ntfy.Publish · renderFields · redact.String) |
| `check_analysis.py --change a133-… --record-landing` | `landed-commit.txt` = `2c2daae7` (도구가 증거에서 계산) |
| `make lint` | rc 0 |
| `go test -race -timeout 45m -count=1` redact · obs · audit · soak · verifylive · official · cmd/tossctl · flatten | 전부 ok — redact 6.032s, obs 309.033s, audit 1.507s, soak 1.514s, verifylive 17.760s, official 6.429s, tossctl 199.882s, flatten 90.487s |
| `go test -race -timeout 45m -count=1 ./internal/app/engine/` | ok 2547.398s |
| `make test-seams` | rc 2 — ok 105 패키지, 실패 1 = 앞 판과 같은 `internal/position` `TestExitEligibilityHasOneSpelling`(base 의 a130 파일 원인, a133 무관) |
| `make sdd-check-ci` | rc 0 |

## 저장소 본체(CodeGraph 색인이 있는 곳)

| 명령 | 결과 |
|---|---|
| `make sdd-sync` | CodeGraph sync 완료 · CodeGraphContext update 300s 타임아웃(advisory — a131 선례) · GBrain busy(advisory) → rc 2(incomplete: codegraphcontext) |
| `make sdd-check` (sdd-sync 직후) | **rc 0** — `CodeGraph hard-evidence index matches the worktree`, CGC · GBrain 은 advisory WARN. 앞선 두 번은 a130 의 -race 부하로 `codegraph status` 15s 타임아웃(18s 실측) 또는 a130 편집 직후 지문 불일치 — 부하가 내려간 뒤 통과 |

청산 · 손절 비지연 회귀(design D6): `internal/obs` a092 시험군 · `internal/app/engine` a092 · a124 · a098 비지연 시험은 위 -race 패키지 전체 판에 포함되어 통과.
