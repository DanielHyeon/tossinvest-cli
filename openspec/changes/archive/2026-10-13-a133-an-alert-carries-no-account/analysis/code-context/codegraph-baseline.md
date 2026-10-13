# CodeGraph baseline — a133

- 날짜: 2026-10-13 · base `dbb0681c` · codegraph **1.6.0** · 질의 직전 `codegraph sync .` · 모든 질의 `--limit 1000`, 원자료 `raw/`.

| 질의 | 결과 | HEAD 대조 |
|---|---|---|
| `callers Publish` | 19 — 이름 충돌(`httpapi.Stream.Publish` 와 그 시험 9) 섞임. obs 쪽: `publishBestEffort` · `deliver` · `Flush` · `Beat` · `NormalRelay.publish` · `publishNotificationTest` · 시험 2 | **`alertDeliverer.deliverOne`(`alertdelivery.go:327`) 누락** — 인터페이스 필드(`d.Publisher`) 경유 호출을 못 잇음. HEAD grep 으로 보강 → 생산 7 자리(design D1) |
| `callers renderFields` | 1: `notificationFor` | 일치 |
| `callers notificationFor` | 3: `NormalRelay.publish` · `publishBestEffort` · `deliver` | 일치 — 재전송(`deliverOne` · `Flush`)은 이 함수를 안 지남(저장 행으로 직접 `Notification` 을 만듦) |
| `callers encodeFields` | 2: `recordCritical` · `notifyCritical` | outbox payload 쓰기. 그 밖의 outbox 쓰기 `execgw.parkAlert`(`replay.go:553`)는 HEAD grep(`EnqueueAlert(`)으로 확인 — design D4-2 |
| `impact renderFields` | 5: `renderFields` · `notificationFor` · `publishBestEffort` · `deliver` · `NormalRelay.publish` | 일치 |
| `impact notificationFor` | `raw/impact-notificationFor.txt` | 일치 |
| `affected ntfy.go notifier.go` 기본 | 시험 1(`auth-helper/tests/test_cli.py`) | 하네스 주석의 `isTestPath` 결함 재현 — 결론으로 안 씀 |
| `affected … --filter '*_test.go'` | 1127 — `internal/app` 171 · `internal/journal` 141 · `cmd/tossctl` 91 · `internal/console` 62 · `internal/execgw` 49 · … · `internal/obs` 21 | obs 가 거의 모든 패키지에 import 되어 넓음. 검증은 행동이 닿는 패키지(`internal/obs` · `internal/app/engine` · `cmd/tossctl` · `internal/redact` · `internal/flatten`) 전체 `go test`(+ race) |
| 생산 `Publisher` 구현(HEAD `rg 'func \([^)]*\) Publish\('`) | `*obs.Ntfy` 1 (+ `httpapi.Stream` 시그니처 다름) | 구조 핀의 기준(design D2 ①) |
