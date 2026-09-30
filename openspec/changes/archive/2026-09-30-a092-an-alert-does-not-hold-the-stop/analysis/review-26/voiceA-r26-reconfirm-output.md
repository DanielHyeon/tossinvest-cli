# 26라운드 보이스 A 재확인 (55b435ea, 읽기 전용) — 판정 유지: APPROVE

- **#1 mode-release ctx — 닫힘.**
  - 커밋 앞은 요청 ctx: `o.journal.TransitionOperatingMode(ctx, …)`(modeops.go:106). 판정 · audit · commit 이 같은 ctx.
  - 커밋 뒤 통지: `detachedAnnouncer` 가 `context.WithoutCancel(ctx)`(:184). 재읽기 둘도 `after := context.WithoutCancel(ctx)`(:129).
  - 재읽기 실패는 `ReReadError` · `NoticeReadError` 칸으로 보고.
  - 사본 `-race` ok(engine 124s · obs 90s).
  - 변이 1개: `:184` 의 `WithoutCancel` 제거 → `TestA092ACommittedReleaseIsNotHostageToTheRequestContext` FAIL(`Notified:false`, `EntryBlocks:[critical_alert_undelivered]`).
- **#4 event 키 중복 — 닫힘.**
  - `grep -n 'FieldEvent,' internal/obs/*.go` 의 비시험 결과는 `log.go:195`(emit 자신) 하나.
  - severity 도 `"trigger_severity"`.
- **#2 · #3 · #5 부수 확인 — 닫힘 동의.**
  - `Offer` 가 `stopped` 를 r.mu 아래에서 봄, `stop()` 은 r.mu 아래에서 표시 · 비우기는 밖 → 틈 없음.
  - 패닉 · 종료 배수는 `Run` 의 defer, 발행 실패 · 발행기 없음은 자체 `publish`.
- **잠금 · 교착 · 오개방 — 새로 만든 것 없음.**
  - `logFailure` 는 `Log.Error` 만.
  - `recordCritical` 은 문자열만 바뀜.
  - `NormalRelay.mu` 는 잎 잠금.
  - `ErrModeAnnouncementFailed` 갈래 · 선점 기록 줄은 로그 · 오류 문구만(판정 불변).
  - 통지 기록이 거의 항상 성공하므로 행 없이 래치만 남는 경우가 줄었음(보수 방향).

## 새 발견

| # | 등급 | 파일:줄 | 무엇 | 근거 | 제안 |
|---|---|---|---|---|---|
| N1 | P2 | modeops.go:129,184 · mode_control_transport_unix.go:137-140 | 떨어진 ctx 에 기한 없음 — 원장 단일 연결 대기에 상한 없음(추론, 실행 확인 안 함). 엔진 종료 때 `Shutdown(2s)` 가 먼저 끝나면 처리기가 닫힌 원장에 통지를 쓰다 실패할 수 있음(래치만 남고 종료, 재시작 뒤 NORMAL 인데 통지 행 없음). 창은 「종료 중 완화」로 좁음. 루프 영향 없음. a066 선례도 같은 모양. | `after := context.WithoutCancel(ctx)`, `s.server.Shutdown(ctx)`(2s) 뒤 `closePrivateEndpointFiles` | 차단 사항 아님. 원하면 종료 순서에서 완화 처리기를 원장 close 앞에 끝까지 기다리게. 기한을 다는 것은 유실 창을 다시 여는 쪽이라 권하지 않음. |

사본 `/tmp/claude-1000/a092-r26-A2-<pid>`(`set -euo pipefail` + `git -C` rev-parse 실패 단언, 끝난 뒤 사본 · GOCACHE · TMPDIR 삭제).

Recommendation: #1 · #4 닫힘을 받아들이고 보이스 A 를 APPROVE 로 종결 — 커밋 앞/뒤 ctx 경계가 코드와 변이 시험 양쪽으로 서고, 이번 diff 는 잠금 그래프에 잎 잠금 하나만 더함.
