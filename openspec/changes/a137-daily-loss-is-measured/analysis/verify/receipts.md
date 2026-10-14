# a137 검증 영수증

연결 워크트리(고정 sha — 병행 a136 커밋이 측정 중 HEAD 를 옮기지 못하게), `GOFLAGS` 비움(`-trimpath` 는 소스 경로를 읽는 시험을 깨므로 쓰지 않음).

## 구현 리뷰 수리 뒤 (`c83241ff` 전 스위트 + `3d48c22b` 재확인)

| 명령 | 결과 |
|---|---|
| `go vet ./...` · `go vet -tags tossos_testseams ./...` | rc 0 · rc 0 |
| `gofmt -l internal cmd` | 빈 출력 |
| `go test -timeout 30m -count=1 ./...`(c83241ff) | ok 107 · FAIL 1 = `internal/position` TestExitEligibilityHasOneSpelling(새 태그 seam 파일) → `3d48c22b` 에서 등재로 수리, 재실행 ok |
| `go test -timeout 45m -count=1 -tags tossos_testseams ./...`(c83241ff) | ok 108 · FAIL 1 = 같은 가드 → `3d48c22b` 재실행 ok(position · journal) |
| `make test-race`(c83241ff) | rc 0 |
| `go test -race -tags tossos_testseams -run TestA137 ./internal/{execgw,app/engine,journal,dailyloss}/`(c83241ff) | rc 0 |
| 변이 원장(`3d48c22b`) | 37/37 CAUGHT(`analysis/mutation/ledger.md`) |
| RED(수리 시험을 수리 전 `6cd36e6d` 에 얹음) | 컴파일 RED — `analysis/red/red-review-fixes-engine.txt` · `red-review-fixes-execgw.txt`(`DailyLossUnmeasurable` 미정의). 시험 공백(P1-1 · P1-2 · P2-1)은 지목 변이 M2c · M1b · M3b 에서 CAUGHT 로 확인 |

## 이전(`4fc5d4a4`)

test ok 108 · seams ok 109 · make test-race rc 0 · 변이 28/28.
