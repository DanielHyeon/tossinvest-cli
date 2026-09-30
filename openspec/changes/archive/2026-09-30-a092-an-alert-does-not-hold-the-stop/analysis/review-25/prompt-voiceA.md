(prompt-common.md 전문을 먼저 읽는다 — 아래는 보이스 A 의 초점.)

## 보이스 A — 동시성 · 잠금 · 원칙 E

`internal/obs/notifier.go` · `record_only.go` 를 중심으로:
- 잠금을 좁힌 뒤 가능한 모든 끼어듦(두 동기 발송 · 동기 발송 + 기록 전용 입구 · 동기 발송 + `Acknowledge` · 배달 실행자 `internal/app/engine/alertdelivery.go` + 동기 발송)을 열거하고, 각각이 (a) 이중 발송 (b) 게이트 오개방(미전달 critical 행이 있는데 진입 허용) (c) 영구 잠김(사람이 풀 수 없는 래치) (d) 교착 중 무엇을 만들 수 있는지 판정하라.
- `judge` 가 호출되는 문맥에서 announcer 재진입(`escalate` → 모드 전이 → 통지 → `Notify` → `n.mu`)이 교착을 만들 수 있는지. `recordCritical` 의 실패 갈래도.
- `readVerdict` 의 세대 읽기 위치가 「근거가 확정된 순간」인지 세 자리 각각(특히 `relErr != nil` 로 소진 판정으로 떨어지는 경로).
- 필요하면 `go test -race` 를 대상 트리 **사본**에서 돌려라(사본은 `/tmp/claude-1000/a092-r25-A-$$` 에 만들고, 만들기 전에 `set -euo pipefail`, 사본의 `git rev-parse` 가 실패하는지 — 즉 실제 저장소가 아닌지 — 단언).
