# GREEN·선택 시험·lint (a192 tasks 3.9·4.2·4.3, 2026-11-17)

모든 원문은 `rtk proxy` 로 받은 잘리지 않은 출력. 판정 명령은 파이프에 넣지 않고 rc 를 따로 적음(`final-summary.txt`).

## 최종(코드 리뷰 2회 반영 뒤 소스)

| 원문 | 명령 | rc | 결과 |
|---|---|---|---|
| `final-1-reconcile.txt` | `go test -count=1 -v ./internal/reconcile/...` | 0 | PASS 209(그중 TestA192 32) · FAIL 0 · SKIP 0 — **패키지 전체** |
| `final-2-journal-selected.txt` | `go test -count=1 -v -run 'Reconcile\|Release\|RiskBucket\|A192\|OpenPosition\|A138' ./internal/journal/` | 0 | PASS 160(TestA192 8) · FAIL 0 |
| `final-3-engine.txt` | `go test -count=1 -v ./internal/app/engine/` | 0 | PASS 849(TestA192 2) · FAIL 0 · SKIP 1 — **패키지 전체** |
| `final-4-cmd-selected.txt` | `go test -count=1 -v -run 'ReconcileResolve\|EngineRuntime\|EngineFillDetector' ./cmd/tossctl/` | 0 | PASS 17 · FAIL 0 |
| `final-5-reconcile-race.txt` | `go test -count=1 -race ./internal/reconcile/` | 0 | ok(489s) |
| `make-lint-final.txt` | `make lint`(gofmt + `go vet ./...` + 태그 vet) | 0 | 위반 0 |
| `openspec-validate-final.txt` | `openspec validate a192-… --strict` | 0 | valid |
| `check-analysis-final.txt` | `python3 tools/logic-map/check_analysis.py --change a192-…` | 1 | required 함수 **7**(= D11). 오류 9 줄 전부 「시험이 tracked 파일에 없음」 — 새 시험 파일이 **커밋 전**이라서임(지시상 커밋 금지). 그 밖의 번들 오류 0 |

중간 판본 원문(`green-1-a192*.txt` … `green-5-*`, `check-analysis-1.txt`, `make-lint.txt`)은 리뷰 수정 전 소스 기준이며 기록으로 둠.

## 선택 이유

- `codegraph affected internal/reconcile/mismatch.go internal/journal/reconcile_states.go --filter '*_test.go' -d 1` = 24 파일(`../code-context/codegraph-affected.txt`)이 상한.
  그 안에서: reconcile 패키지 전체(편집 함수 4개의 집), engine 패키지 전체(RunOnce·buildGateway·로그 함수, 대사 루프·편입 시험), journal 은 RECONCILE·release cause·
  위험 버킷·열린 포지션 상한 시험(새 원인이 닿는 소비자 — design Q5 ⑥⑦⑧), cmd/tossctl 은 `reconcile-resolve`·엔진 조립 시험(무변경 확인).
- `-race` 는 reconcile 패키지만(결속 기억은 `t.mu` 아래, 드라이버의 사유 기억은 RunOnce 직렬).

## 실행하지 않은 범위

- 전체 `make test`·`make test-seams`·`make test-race`·`make gate`·`make sdd-check`·`make sdd-sync` — 지시상 금지.
- `internal/journal` 의 위 -run 밖 시험, `internal/execgw`(affected 3 파일 — 이 change 는 execgw 무편집), `cmd/tossctl` 의 위 -run 밖 시험.
- engine 패키지 `-race`(무거움 — 기억 「경합 검출기는 여기서 돈 적이 없었다」 의 이름 목록 방식 밖).
- 운영·브로커 실측 0(금지).
