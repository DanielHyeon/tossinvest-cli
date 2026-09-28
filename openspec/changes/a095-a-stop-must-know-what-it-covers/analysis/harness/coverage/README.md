# a095 3판 커버리지 프로파일 (대상 파일만 남김)

- 측정 HEAD: obs · journal · app/engine = `4fd400ea`, reconcile = `02716357` (두 HEAD 사이에 대상 소스 무변화 —
  각 번들의 `source_sha256` 이 base `02716357` 의 파일과 일치)
- 명령: `go test ./internal/<pkg>/ -count=1 -covermode=set -coverprofile=<pkg>.out` (패키지별 한 번)
- 원본 프로파일에서 FLM 대상 소스 파일의 줄만 남겼음(`mode:` 머리줄 유지). 「진입 실측」 재현:
  `python3 ../render_bundles.py <ast 디렉터리> obs.out journal.out app-engine.out reconcile.out`

## 4판 추가 (r3 N3 · N5 근거 번들 5개)

- `r4-journal.out` · `r4-exitpolicy.out` · `r4-app-engine.out` — HEAD `04a0dd25` 의 깨끗한 연결 worktree 에서
  `go test ./internal/<pkg>/ -count=1 -covermode=set` (공유 워킹트리의 병행 로트 미커밋 편집을 피함). 대상 파일
  `internal/journal/outbox.go` · `internal/exitpolicy/recovery.go` · `internal/app/engine/alertdelivery.go` 의 줄만 남김
- 재현: 그 다섯 stem 의 AST 만 든 디렉터리로 `render_bundles.py <ast4> r4-journal.out r4-exitpolicy.out r4-app-engine.out`
  (생성기는 주어진 디렉터리에 AST 가 있는 stem 만 그림)
- 셋 다 base `02716357` 와 다를 수 있는 파일이다: `outbox.go`(a124 가 :512 뒤에 함수 하나 추가 — 인용한 두 함수의 줄은
  같음) · `alertdelivery.go`(a124 가 크게 바꿈) · `recovery.go`(base 와 동일). 번들 해시는 HEAD 파일에 묶였다
