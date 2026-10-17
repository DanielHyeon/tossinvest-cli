# a146 검증 영수증 (GREEN 커밋 직전 워킹트리 = 50aea19d + a146 편집, 2026-10-18T00:17:42+09:00)

| 명령 | 결과 |
|---|---|
| `rtk proxy make lint` (gofmt -l cmd internal tools/logic-map · go vet · go vet -tags tossos_testseams) | 출력 없음(포맷 0 · vet 0) |
| `rtk proxy go test -timeout 30m ./...` | FAIL 줄 0, rc 0 |
| `rtk proxy go test -timeout 45m -tags tossos_testseams ./...` | FAIL 줄 0 |
| `rtk proxy go test -race -count=1 -tags tossos_testseams -run 'A146\|A137\|EntrySkip' ./internal/app/engine/` | ok 39.5s |
| `rtk proxy go test -count=1 ./internal/strategyworker/ ./internal/obs/` | ok · ok |
| `openspec validate a146-guardian-cap-zero-skips-the-entry --strict --no-interactive` | valid |
| `python3 tools/logic-map/check_analysis.py --change a146-guardian-cap-zero-skips-the-entry` (커밋 전) | required 8 = 번들 8; 남은 줄은 새 시험 파일이 미추적이라는 인용 오류뿐 — 커밋 뒤 재실행은 tasks 1.3 |

RED: `analysis/red/red-compile.txt`(편집 전 production + RED 시험 → 컴파일 RED 11+ 자리, census 행동 RED) · `analysis/red/characterization-head.txt`(편집 전 동작 4/4).
