# a137 검증 영수증

연결 워크트리 `4fc5d4a4`(고정 — 병행 a136 커밋이 측정 중 HEAD 를 옮기지 못하게), `GOFLAGS` 비움(`-trimpath` 는 소스 경로를 읽는 시험을 깨므로 쓰지 않음 — 1차 시도의 49 FAIL 은 그 결함이었고 진짜 실패 2(정적 가드)는 `4fc5d4a4` 로 고침).

| 명령 | 결과 |
|---|---|
| `go vet ./...` · `go vet -tags tossos_testseams ./...` | rc 0 · rc 0 |
| `gofmt -l internal cmd` | 빈 출력 |
| `go test -timeout 30m -count=1 ./...` | rc 0 — ok 패키지 108, FAIL 0 |
| `go test -timeout 45m -count=1 -tags tossos_testseams ./...` | rc 0 — ok 패키지 109, FAIL 0 |
| `make test-race` | rc 0 |
| `go test -race -tags tossos_testseams -run TestA137 ./internal/{execgw,app/engine,journal,dailyloss}/` | rc 0 |
| `openspec validate a137-daily-loss-is-measured --strict` | valid |
| 변이 원장(`analysis/mutation/a137_mutate.py`) | 28/28 CAUGHT(`ledger.md`) |
