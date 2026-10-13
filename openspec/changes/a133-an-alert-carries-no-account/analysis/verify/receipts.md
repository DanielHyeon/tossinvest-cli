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
