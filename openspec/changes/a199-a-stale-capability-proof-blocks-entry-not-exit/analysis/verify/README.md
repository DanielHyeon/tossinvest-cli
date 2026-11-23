# a199 2.8 검증 (HEAD b856fdc08 + 변이 원장 문서)

| 무엇 | 명령(원문 파일) | 결과 |
|---|---|---|
| 선택 시험 engine | `go test -count=1 -timeout 30m -run 'A199\|Interlock\|Gate\|Attestation\|A135\|A196\|ExitObserver\|Reconcile\|StrategyEntry\|A071' ./internal/app/engine` (`selected-engine.txt`) | ok 56.0s |
| 선택 시험 cmd | `-run 'A199\|Engine\|Interlock\|Soak\|A094\|A109' ./cmd/tossctl` (`selected-cmd.txt`) | ok |
| 선택 시험 execgw | `-run 'A199\|Entry\|Latch\|Reason\|Mode' ./internal/execgw` (`selected-execgw.txt`) | ok |
| attest · census · audit | `./internal/attest ./internal/accountbudgetruntime ./internal/audit` (`selected-attest-census-audit.txt`) | ok 전부 |
| race | `go test -race -run 'A199\|TestGateOnRefusals\|TestRefusedStartIsAudited\|TestGateDecisionIsLoggedStructurally\|A135\|A196' ./internal/app/engine ./internal/execgw ./internal/attest ./cmd/tossctl` (`race.txt`) | ok 넷, DATA RACE 0 |
| lint | `make lint`(gofmt · `go vet ./...` · `go vet -tags tossos_testseams ./...`) (`make-lint.txt`) | rc 0 |
| strict validate · tracker | `openspec validate … --strict` · `generate_master_tracker.py --check` (`validate.txt`) | valid · current |
| check_analysis | (`check-analysis-post.txt`) | rc 0, required 20 = 번들 20 |

미실행: 전체 스위트(`make test` · `make test-seams` · `make test-race`) · `make gate` · `make sdd-sync` — 지시로 금지. 위 필터 밖의 패키지(journal · reconcile · filldetect · console · httpapi · strategy* 등)와 같은 패키지 안 필터 밖 시험은 돌리지 않았다. 운영 실측 없음.
