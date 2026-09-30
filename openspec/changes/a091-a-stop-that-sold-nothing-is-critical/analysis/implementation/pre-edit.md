# a091 Pre-Edit 선언 (tasks 2.0 · 3.0) — 2026-10-01

구현 워크트리 `/tmp/claude-1000/a091-impl`(detached @ `1602fd45`), 대상 파일은 base `b30318d6` 과 바이트 동일(`git diff --quiet`). 편집 전 AST 사본은 `pre-edit/`.

| 대상 | 편집 | High-risk | 경계(바꾸지 않는 것) |
|---|---|---|---|
| `internal/obs/event.go` 종류 상수 · `criticalEvents` | `EventExitStopSoldNothing = "exit.stop_sold_nothing"` 추가 · 등급표 한 줄 | 예(등급표 = 진입 차단 스위치) | 기존 19 종의 등급 · 값 |
| `ExitObserver.submit`(`exitloop.go:1395`) | `applyFloor` 에 `isProtective(proposal)` 전달 | 예 | 분기 B1~B13 · 제출 수량 · 시점 |
| `ExitObserver.applyFloor`(`exitloop.go:1617`) | 보호 여부 인자 · B2 의 보고(원인 분류 · 게이트 · 가린 로그) · 끝의 0주 보고 분리(부분 캡 알림은 그대로) | 예(손절 경로) | **반환값 `(수량, capped, err)` — B1~B6 · 끝 전부**, 브로커 요청(`ConfirmedFloor` 호출 1회) |
| `ExitObserverOptions` | `NotificationsEnabled bool` 필드 | — | 기존 필드 |
| `Context.ExitObserver`(`exitwiring.go`) | 로드된 설정 `c.Config.Engine.Notifications.Enabled` 로 `opts.NotificationsEnabled` 덮기 | 예(생산 배선) | 기존 덮기 · 주입 |
| `obs.Notifier.escalate`(`notifier.go:425`) | 로그 두 줄(`:433` · `:440`)의 `FieldAccount` 제거 | 아니오(로그) — 모드 승격 경로 위 | 판정 · 반환 · 원장 호출 |
| 새 파일 `internal/app/engine/exit_stop_sold_nothing.go` | 원인 분류 · 게이트 · 문구 · 가린 로그 · `WithoutCancel` 기록 | 예 | — |

Function Logic Map: 편집 전 번들(`analysis/function-logic/`)은 base AST · 커버리지로 이미 있음(0.3). `Context.ExitObserver` 는 편집 전 AST 를 `pre-edit/` 에 두고, 구현 뒤
모든 편집 함수의 번들을 편집 뒤 소스로 재생성한다(a094 선례).
