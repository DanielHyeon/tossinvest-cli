# a201 편집 표면 · 고정 장치 · 겹침 (tasks 1.7)

HEAD `86a131a51`, CodeGraph 1.6.0(code-context/codegraph-baseline.md), AST `cited-ast/`.

## 1. 편집할 기존 함수 — 1 개(확정 조건부)

| 함수 | 파일 | 편집 | 근거 |
|---|---|---|---|
| `ReconcileDriver.RunOnce` | internal/app/engine/reconcileloop.go:425-479 | call@434 앞 워터마크 읽기 1문장, call@464 앞 재판정 호출 1자리(인자를 사본으로) | design D1 · D2. 전제 (a) 참 — 조정 반영 종목은 `converged.Converged` 에 이미 있음 |

- 새 파일 2(기존 함수 편집 아님): `internal/reconcile/<재판정>.go`(비공개 `quantitiesAgree` · `canonicalPresentQuantity` · `HeldBySymbol` 을 그대로 부르려면 이 패키지여야 함), `internal/app/engine/<로그>.go`(D7 로그 · 이벤트 상수 — a192 의 `a192_order_terminal_release_log.go` 와 같은 모양, `obs/event.go` 무편집).
- **무편집 목표 유지 가능**: `Observe` · `ConvergeQuantities` · `Compare` · `LocalStateFromJournal` · `HeldBySymbol` · 크레딧 함수 · `blocksFor` · 스키마.
- 둘째 함수가 필요해지는 조건(지금은 아님): Manager 가 (i) 판 중단 뒤 종목(claim-check C-R1)을 범위에 넣으면 `ConvergeQuantities`(converge.go:228-229), (ii) 뺀 결과가 해제 갈래로 가는 것(C-D6)을 막기로 하면 `Observe` 또는 재판정 조건에 「그 종목에 활성 블록 · 쓸 수 있는 크레딧이 없을 때만」을 더함(후자는 RunOnce 가 `Tracker.Blocks()` 로 읽을 수 있어 기존 함수 추가 편집 없이 가능).

## 2. 고정 장치 적중

| 장치 | 대상 | 적중 | 처분 |
|---|---|---|---|
| `a071_readiness_assembly_test.go:27-40` | reconcileloop.go 에 `ProtectionReadiness` · `ReadinessAdapter` · `checkProtection` 문자열 금지 | 편집이 이 문자열을 안 씀 → 비적중 | 2.6 선택 시험에 포함 |
| a166 census `internal/accountbudgetruntime/testdata/a166-wiring/{shapes,sites}.json` | 배선 함수 본문 | reconcileloop.go · reconcile/{compare,converge,mismatch}.go 키 0 · 자리 0(grep) | 배선(gateway.go `Tracker{}` · `Converger{}` 리터럴 · `buildGateway`) 무편집이면 비적중 |
| engine 패키지 전체를 파싱하는 구조 시험(a112 · a163 · a124 · a091 · a087 · entryreach · strategy_* 등 ~30) | 특정 식별자 · 호출(`ErrStrategyCentralIntegrity`, `SourceTimestamp`, 주문 · 전략 경로 등) 센서스 | RunOnce 편집이 쓰는 식별자(`MaxFillEventID` · `FillWatermark` · 새 재판정 함수 · 로그 이벤트)와 겹치는 금지 · 센서스 이름 0(grep `"FillWatermark"` 등 리터럴 — strategy_dispatch_handoff_guard_test.go:539 의 `Observe` 접두는 다른 셀렉터) | 2.6 에서 `go test ./internal/app/engine` 선택 범위로 실측 |
| `internal/reconcile` 패키지 파싱 시험 | a094_acked_boot_test.go:279(acked_boot.go 만) | 비적중 | — |
| 골든 · 동결 표 | `analysis/goldens/`, `ValidReconcileReleaseCause` 집합 | release cause 무편집 → 비적중 | — |
| a163 핀 | engine judgement 함수 render hash | `ExitObserver.*` 만 → 비적중 | — |

## 3. 진행 중 change 와의 겹침

| change | 위치 | 편집(커밋) | 겹침 | 처분 |
|---|---|---|---|---|
| a199 | `~/tossos-work/a199-impl` HEAD `815e378dd`(mb `7fec501f1`) | reconcileloop.go **`Context.ReconcileDriver`** :368 한 줄(`Automation.Verified` → `RunsExitSide()`) + engine.go · interlock.go 등 | **같은 파일 · 다른 함수**(RunOnce 와 57 줄 떨어짐 — 텍스트 충돌 없음 예상) | 종결 때 「착지 기록 → merge」. a199 가 먼저 main 에 오면 그 위에서 구현 시작(design §5). gate ⑤ 창: a199 가 a201 base 뒤 main 에 착지하고 a201 이 merge 하면 `Context.ReconcileDriver` 가 a201 창에 들어옴 — 착지 기록으로 좁힘 |
| a200 | `~/tossos-work/a200-impl` HEAD `8ad6c3e79` | internal/journal/position_projection.go · a200_adjustment_close_fill_recognition.go | 파일 겹침 없음. 의미 결합: a200 은 W4(조정 뒤 체결)를 인정 — (나)가 그 창을 건드리지 않음(C-D3b). 판 중단 뒤 종목이 수렴됐더라면 W4 → a200 경로였음 | 없음 |
| a197 | `~/tossos-work/a197-impl`(mb `064f59fd9`) | 비시험 Go 커밋 0 | 없음 | — |
| a087 | `~/tossos-work/a087-impl`(mb = HEAD) | 비시험 Go 커밋 0 | 없음 | — |
| a167 · a171(main `openspec/changes/`) | 문서 | `reconcileloop` · `RunOnce` · `ConvergeQuantities` · `Tracker.Observe` 언급 0(grep) | 없음 | — |
| a100(main) | 문서 | converge.go 인용만(사전 증거) | 없음 | — |
| PM | `_registry.yaml` · `FEAT-TOS-005.yaml` | a199 · a200 도 같은 목록 끝에 한 줄씩 추가 | 종결 merge 때 목록 충돌(기계적) | merge 때 둘 다 남김 |

## 4. 잔여 창(1.8) — `analysis/residual-window.md`
