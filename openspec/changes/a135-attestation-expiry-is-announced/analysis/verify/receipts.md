# a135 검증 영수증 (2026-10-14, 공유 워크트리 · HEAD 85752cd6 위 미커밋 a135 파일)

잘리지 않은 출력은 `rtk proxy` 로 받음. 병행 change(a128 · a134)의 워크트리 파일이 같은 빌드에 들어 있음.

| 명령 | 결과 |
|---|---|
| `gofmt -l cmd/tossctl/ internal/app/engine/ internal/obs/` | 빈 출력 |
| `go vet ./...` | rc 0 |
| `go vet -tags tossos_testseams ./internal/app/engine/ ./internal/obs/ ./cmd/tossctl/` | rc 0 |
| `go test -count=1 -run TestA135 ./internal/app/engine/` | ok (11 시험) |
| `go test -count=1 -timeout 40m ./internal/app/engine/ ./internal/obs/ ./internal/attest/ ./internal/soak/` | 넷 다 ok(engine 215s · obs 27s) — 배선 전 |
| `go test -count=1 -timeout 40m -tags tossos_testseams ./internal/app/engine/ ./internal/obs/` | ok(engine 309s · obs 26s) — 배선 전 |
| `go test -count=1 ./cmd/tossctl/` | ok 57s — 배선 뒤 |
| `go test -count=1 -tags tossos_testseams ./cmd/tossctl/` | ok 59s — 배선 뒤 |
| `analysis/mutation/run_mutations.py` | CONTROL GREEN · 6/6 CAUGHT · 원복 sha256 일치 |
| `openspec validate a135-attestation-expiry-is-announced --strict --no-interactive` | valid |
| `python3 tools/pm/generate_master_tracker.py --check` | current |
| `python3 tools/logic-map/check_analysis.py --change a135-…`(커밋 전) | rc 1 — 남은 사유: (1) 새 시험이 아직 추적 안 됨(커밋으로 해소), (2) base 뒤 병행 커밋 3개(a128 ×2 · a134)의 기존 함수 26개(`tools/a128-jev-shadow-probe/*` · `newEngineCmd`)가 같은 창에 섞임 — a135 귀속 아님, 커밋 뒤 `--record-landing` 으로 좁힘 |

`TestSupervisedEvidenceOlderThanTheValidityIsNotEvidence`(internal/soak) 는 위 soak 패키지 실행에 포함 — 「기간 밖 감독 증거로 갱신하면
POST 누락 attestation 이 나온다」를 실행으로 고정.

## 배선 · 리뷰 반영 뒤 최종(2026-10-14, HEAD 85752cd6 위)

| 명령 | 결과 |
|---|---|
| `go test -count=1 -race -run TestA135 ./internal/app/engine/` | ok(16 시험) |
| `go test -count=1 -timeout 40m ./internal/app/engine/ ./internal/obs/ ./cmd/tossctl/ ./internal/soak/ ./internal/attest/` | 다섯 다 ok(engine 258s · obs 28s · cmd 61s) |
| `go test -count=1 -timeout 40m -tags tossos_testseams ./internal/app/engine/ ./internal/obs/ ./cmd/tossctl/` | 셋 다 ok(engine 254s · obs 28s · cmd 60s) |
| `analysis/mutation/run_mutations.py` | CONTROL GREEN · 11/11 CAUGHT · 원복 sha256 일치 |
