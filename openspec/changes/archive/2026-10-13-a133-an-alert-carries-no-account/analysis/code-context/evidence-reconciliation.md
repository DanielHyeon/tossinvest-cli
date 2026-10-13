# Evidence reconciliation — a133

- 기준: base `dbb0681c`. 확인한 파일·심볼: `internal/obs/ntfy.go`(`Publisher` · `Notification` · `Ntfy.Publish` · `priorityFor`) ·
  `internal/obs/notifier.go`(`Notify` · `logEvent` · `publishBestEffort` · `notifyCritical` · `claimAndDeliver` · `deliver` · `Flush` · `eventKey` · `notificationFor` · `encodeFields` · `renderFields`) ·
  `internal/obs/record_only.go`(`recordCritical` · `operatingModeEvent` · `OperatingModeEventKey`) · `internal/obs/normal_relay.go` · `internal/obs/heartbeat.go` · `internal/obs/topic.go` ·
  `internal/app/engine/{alertdelivery,auxiliary,notifications,engine,exitwiring}.go` · `cmd/tossctl/notificationsettings.go` · `internal/execgw/replay.go`(`parkAlert`) ·
  `internal/journal/outbox.go`(스키마 · `recordAlertTx` 재무장 UPDATE · `alertSelect`) · `internal/redact/*`.

| # | 불일치 | 현재 HEAD 로 내린 결론 |
|---|---|---|
| R1 | CodeGraph `callers Publish` 가 `deliverOne` 을 놓침 · CGC 가 `NormalRelay.publish` 를 놓침 | 인터페이스 필드 경유 호출. HEAD grep `\.Publish\(` 비시험 obs 계열 7 자리가 정본(design D1). 경계가 `Ntfy.Publish` 안이므로 호출 자리 수와 무관하게 덮임 |
| R2 | `affected` 기본 실행 시험 1 | 알려진 판별식 결함(CLAUDE.md 하네스 주석). `--filter` 결과를 씀 |
| R3 | CGC 결과에 옛 체크아웃 경로 행 | advisory — 현재 경로 행만 씀 |
| R4 | outbox 쓰기 자리: CodeGraph(`encodeFields` 호출자 2) vs HEAD | HEAD grep `EnqueueAlert(`·`RecordAlert(`·`ClaimAlertForDelivery(` → obs 2 + `execgw.parkAlert` 1. 저장 시점 치환이 보장 지점이 될 수 없는 근거(design D4-2) |
| R5 | `Heartbeat` · `Flush` 의 생산 배선 | HEAD grep: `obs.Heartbeat{`·`Notifier.Flush(` 비시험 0 — 생산에서 안 돎. 그래도 같은 `Publish` 를 지나므로 경계가 덮음 |
| R6 | 실측 vs 코드 | 운영 outbox(`analysis/verify/outbox-census.md`)의 payload 계좌 · pos- 는 `Fields` 에서 옴 — `deliver` 가 `notificationFor` 로 본문에 펼침(P2). title · body 계좌 0 |
