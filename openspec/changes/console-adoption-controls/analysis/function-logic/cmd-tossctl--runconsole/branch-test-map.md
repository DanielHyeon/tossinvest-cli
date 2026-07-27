# Branch Test Map: `runConsole`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | AST if 분기 — 위 목적·경계 서술의 일부 | internal/console settings_test.go 전체(주입된 seam 동작) + cmd/tossctl 기존 콘솔 테스트 green | 해당(아래 기록) | yes |
| B2 | AST if 분기 — 위 목적·경계 서술의 일부 | internal/console settings_test.go 전체(주입된 seam 동작) + cmd/tossctl 기존 콘솔 테스트 green | 해당(아래 기록) | yes |
| B3 | AST if 분기 — 위 목적·경계 서술의 일부 | internal/console settings_test.go 전체(주입된 seam 동작) + cmd/tossctl 기존 콘솔 테스트 green | 해당(아래 기록) | yes |
| B4 | AST if 분기 — 위 목적·경계 서술의 일부 | internal/console settings_test.go 전체(주입된 seam 동작) + cmd/tossctl 기존 콘솔 테스트 green | 해당(아래 기록) | yes |
| B5 | AST if 분기 — 위 목적·경계 서술의 일부 | internal/console settings_test.go 전체(주입된 seam 동작) + cmd/tossctl 기존 콘솔 테스트 green | 해당(아래 기록) | yes |
| B6 | AST if 분기 — 위 목적·경계 서술의 일부 | internal/console settings_test.go 전체(주입된 seam 동작) + cmd/tossctl 기존 콘솔 테스트 green | 해당(아래 기록) | yes |
| B7 | AST else 분기 — 위 목적·경계 서술의 일부 | internal/console settings_test.go 전체(주입된 seam 동작) + cmd/tossctl 기존 콘솔 테스트 green | 해당(아래 기록) | yes |

RED: config/engine 신규 테스트 4건 실패 관측(2026-07-27, 구현 전 — adopted=0·사유 미표기·'off' 오표기·audit 항목 부재), CSRF 가드 목록 불일치 실패 관측.
GREEN: `go test ./internal/config/ ./internal/app/engine/ ./internal/console/ ./cmd/tossctl/ -race -count=1` — 629 passed (2026-07-27).
