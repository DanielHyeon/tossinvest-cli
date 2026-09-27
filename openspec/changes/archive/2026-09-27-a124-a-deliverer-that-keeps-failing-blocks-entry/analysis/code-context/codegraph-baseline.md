# CodeGraph baseline — a124

- 날짜: 2026-09-26 · 기준 HEAD/base `4798d399` (Go 소스는 AST 추출 HEAD `463cc895` 와 diff 0 —
  `git diff --stat 463cc895 4798d399 -- '*.go'` 빈 출력) · codegraph **1.6.0** · 질의 직전 `codegraph sync .` 완료
- 모든 질의는 `--limit 1000 --json` 으로 돌렸다. **기본 `--limit` 은 20** 이라 `PendingAlerts`·`Block` 은 기본 실행에서
  잘려 「(20)」 으로 보였다(잘린 수를 전체로 읽지 말 것). 원자료: `raw/*.json`.

| 질의 | 결과 (비테스트 / 전체) | 비고 |
|---|---|---|
| `callers alertDeliverer.deliverOne` | 1 / 1: `alertDeliverer.cycle` (alertdelivery.go:145) | HEAD 와 일치 |
| `callers alertDeliverer.cycle` | 1 / 11: `invokeStrategyCycle` (strategy_entry_supervisor.go:1019) + a098·strategy 시험 10 | **생산 호출자 `alertDeliverer.Run` (alertdelivery.go:124) 누락**, `invokeStrategyCycle` 은 이름 해소 오탐 — reconciliation R1 |
| `callers Journal.PendingAlerts` | 3 / 22: `alertOps.Pending` (alertops.go:116) · `Notifier.Flush` (notifier.go:727) · `Notifier.Acknowledge` (notifier.go:840) | **`alertDeliverer.cycle` (alertdelivery.go:150) 누락** — R2 |
| `callers EntryGate.Block` | 2 / 18: `RebuildReconcileProjection` · `Tracker.syncGate` + execgw 시험 16 | **`EntryGate.BlockSymbol` 로 해소됐다**(callees 에 `symbolKey`·`SymbolBlock` — BlockSymbol 본문) — R3 |
| `callers Block` (무수식) | 28 / 50 | `reconcile.Block` 타입·`Blocks` 메서드까지 섞임. 이 중 게이트 `Block` 호출자는 grep 18 자리와 대조(R3) |
| `callers Journal.EscalateOperatingMode` | 3 / 3: `Guardian.escalateFor` (riskguardian.go:640) · `Notifier.escalate` (notifier.go:378) · `Retrier.EscalateOperatingMode` (retry.go:299) | 인터페이스 경유 호출 누락 — R4 |
| `callers EscalateOperatingMode` (무수식) | 6 / 6: 위 3 + `escalateCredentialFailure` (retry.go:409) · `exitObserver.checkOutage` (exitloop.go:817) · `Runtime.escalate` (runtime.go:442) | HEAD grep 6 자리와 일치 |
| `callees alertDeliverer.deliverOne` | 7: `forgetHeld` · `reportHeld` · 타입 참조 5 | **`ClaimAlertByID`·`Publish`·`MarkAlertAttemptFailed`·`MarkAlertDelivered`·`release`·`logf` 누락** (필드 경유 메서드) — R5 |
| `callees alertDeliverer.cycle` | 4 비테스트: `forgetLapsedHeld` · `batch` · `deliverOne` · `Context` | `Journal.PendingAlerts`·`Clock.Now` 누락 — R5 |
| `callees Journal.PendingAlerts` | `scanAlerts` · `AlertPending` · `alertSelect` · 타입 | HEAD 와 일치 |
| `callees Journal.EscalateOperatingMode` | `TargetModeForTrigger` · `AutomaticTriggers` · `TransitionOperatingMode` · 타입 | HEAD 와 일치 (operating_mode.go:496-511) |
| `callees EntryGate.Block` | `Block`(retry.go:526) · `symbolKey` · `Clock.Now` · `SymbolBlock` · `ReasonCode` | BlockSymbol 본문이다(R3). 실제 `Block` 은 map 삽입만(retry.go:526-533) |
| `affected alertdelivery.go outbox.go auxiliary.go` 기본 | 시험 **1** (`auth-helper/tests/test_cli.py`, Go 아님) | 하네스 주석의 `isTestPath` 결함 재현 |
| `affected … --filter '*_test.go'` | 시험 **854** 파일 (journal 117 · app/engine 95 · cmd/tossctl 78 · console 59 · execgw 39 · … · obs 12) | a098_* 15 · a099_*/a096_* 11 포함. 원자료 `raw/affected-filter.txt` |
