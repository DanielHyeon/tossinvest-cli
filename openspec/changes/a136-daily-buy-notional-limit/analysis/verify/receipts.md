# a136 검증 영수증 (2026-10-14)

모두 `rtk proxy` 또는 원래 go 출력(잘리지 않음). 병행 a137 의 미커밋 파일(internal/dailyloss · a137_* 시험 · 공유 파일 편집)이 공유 워킹트리에
있어 그 패키지가 컴파일되지 않는 때가 있으므로, 태그 시험 · race 는 **HEAD + a136 파일만** 담은 격리 사본(`git archive HEAD`)에서 돌렸다.

| 검사 | 범위 | 결과 |
|---|---|---|
| `gofmt -l internal cmd tools` | 워킹트리 | 빈 출력 |
| `go vet ./...` · `go vet -tags tossos_testseams ./...` | 워킹트리 | 이슈 0 |
| `openspec validate --all --strict` | 워킹트리 | 46/46 |
| `go test -count=1`(무태그) | journal · app/engine · execgw · manualbuy · measure · cmd/tossctl · config · risk · clock · obs | 10/10 ok(1.5 착지 직후, 워킹트리) |
| `go test -count=1` | console · config · cmd/tossctl | ok(1.3 착지 직후) |
| `go test -tags tossos_testseams` | journal · execgw · cmd/tossctl · app/engine · manualbuy · measure · config · risk · clock · console · obs · strategydispatch | journal 의 실패 1(`TestAdoptionIDIsWrittenOnlyByTheAdoptionAPI` — a137 미커밋 `a137_daily_loss_rows.go`) 외 a136 쪽 실패는 testseam fixture 의 보드 누락 · USD fixture → 수리 뒤 격리 사본에서 app/engine · cmd/tossctl 태그 ok, journal · execgw 태그 ok |
| `go test -race -run TestA136` | journal · execgw · app/engine · manualbuy · console · config | 6/6 ok, DATA RACE 0(격리 사본) |
| 변이(`analysis/harness/a136_mutate.py`) | 32 변이 | round4 31 CAUGHT + M14 재측정 CAUGHT = **32/32**(무변이 대조군 GREEN 선행, 컴파일 실패는 CAUGHT 로 세지 않음) |
| 렌더(규칙 13) | 격리 config(USD 무값) 로컬 콘솔 `/settings/daily` | 행 · 칸 · 도움말 렌더, 콘솔 오류 0, 요청 1(200), 버튼 미클릭 |

주의: `GOFLAGS=-trimpath` 로 돌리면 소스 경로를 읽는 a111 시험 둘이 경로를 못 찾아 실패한다(시험 인프라의 한계 — 무 trimpath 실행에서는 통과).
