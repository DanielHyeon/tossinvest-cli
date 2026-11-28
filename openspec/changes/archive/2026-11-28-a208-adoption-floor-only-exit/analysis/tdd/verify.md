# a208 VERIFY — 선택 시험(design 8.3), 2026-11-28, 작업 worktree(HEAD a547fa67f + 리뷰 F1 시험 강화)

| 명령 | 결과 |
|---|---|
| `rtk proxy go test ./internal/config/ ./internal/console/ ./internal/app/engine/ -run 'A208\|A163\|A074\|A197\|Adopt\|Include\|Driver\|Reconcile\|ExistingAuditEntries\|Notification' -count=1` | ok · ok · ok(53.1s) |
| `rtk proxy go test ./internal/config/... ./internal/console/... ./internal/exitpolicy/... ./internal/accountbudgetruntime/... ./cmd/tossctl/ -count=1` | 5 패키지 ok(console 92.6s · accountbudgetruntime 25.3s · cmd/tossctl 72.9s) |
| `rtk proxy go test ./internal/app/engine/ -count=1`(기본 시한 10분) | **FAIL — 600s 시한 초과**(변이 하네스 · 리뷰 2 보이스와 동시 실행 중). 시한 순간 돌던 `TestTheStartupSaysProtectionDiesWithTheProcess` 단독 재실행 ok(12.2s) |
| `rtk proxy go test ./internal/app/engine/ -count=1 -timeout 40m`(부하 없이 재실행) | ok(359.95s) — 엔진 패키지 전체 통과 |
| `rtk proxy go test ./internal/config/... -count=1`(F1 시험 강화 뒤) | ok |
| `rtk proxy make lint`(gofmt + vet + tags vet) | rc 0(2 회: GREEN 뒤 · F1 뒤) |
| `python3 tools/logic-map/check_analysis.py --change a208-adoption-floor-only-exit` | required 7 · evidence complete |
| `openspec validate a208-adoption-floor-only-exit --strict` | valid |

미실행(사용자 미요청): 저장소 전체 `make test` · `make test-seams` · `make test-race` · `make gate` · 실계좌 · 엔진 · 컨테이너.
`make test-race` 대상 판단: 이 change 는 새 고루틴 · 공유 상태를 만들지 않음(`adoptOne` 은 대사 루프 고루틴 하나, 새 함수는 순수 함수) — not-applicable.
