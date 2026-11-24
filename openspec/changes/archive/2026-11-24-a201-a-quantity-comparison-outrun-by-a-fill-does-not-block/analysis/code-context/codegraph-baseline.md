# a201 CodeGraph baseline (tasks 1.6)

- 도구: CodeGraph CLI **1.6.0**, 이 worktree 에서 `codegraph init .`(2026-11-24, HEAD `86a131a51` — production Go 는 base `f9b818a0e` · origin/main `7fec501f1` 과 동일), `codegraph status .` 「Index is up to date」. 원문 `codegraph-callers.txt`.
- CodeGraph 는 필드 · 인터페이스 경유 호출(`d.opts.Converge.ConvergeQuantities`, `d.opts.Tracker.Observe`, `c.Journal.FillWatermark`)과 같은 이름의 다른 메서드(`Observe` 20 중 대사 아닌 것 다수)를 가르지 못함 → **생산 호출 자리 수는 grep**(비시험 `.go`, 주석 · `func` 줄 제외)으로 셈. `affected` 는 편집 대상 파일이 아직 없어 돌리지 않음(2.x 에서 `--filter '*_test.go'`).

| 심볼 | CodeGraph 호출자(표시 상한 20) | 생산 호출 함수 / **자리**(grep) |
|---|---|---|
| `ReconcileDriver.RunOnce` | `Run`(reconcileloop.go:403) 1 | 1 / 1 — reconcileloop.go:412 |
| `Converger.ConvergeQuantities` | 14(RunOnce + 시험 13) | 1 / 1 — reconcileloop.go:451 |
| `Tracker.Observe` | 20(이름 충돌 다수 — filldetect · enginewatch · positioncampaign 의 다른 `Observe`) | 1 / 1 — reconcileloop.go:464 |
| `blocksFor` | `Tracker.Observe` 1 | 1 / 1 — mismatch.go:704 |
| `LocalStateFromJournal` | 20(상한) | 3 / 3 — reconcileloop.go:434, recovery.go:324, cmd/tossctl/engine_reconcile.go:304 |
| `Comparer.Compare` | 20(상한, 시험) | 2 / 2 — reconcileloop.go:439, recovery.go:329 |
| `HeldBySymbol` | `LocalStateFromJournal` 1 | 1 / 1 — compare.go:197 |
| `quantitiesAgree`(비공개) | `Compare` 1 | 1 / 1 — compare.go:426 |
| `canonicalPresentQuantity`(비공개) | `Compare` 1 | 1 / 2 — compare.go:392 · 393 |
| `Journal.FillWatermark` | 6(Ingest + 시험 + 인터페이스) — 수렴 호출은 인터페이스 경유라 안 잡힘 | 2 / 2 — converge.go:205, external.go:217 |
| `Journal.MaxFillEventID` | `enterReconcileWithBinding` 1 | 1 / 1 — order_terminal_release.go:351 |
| `promotionKeys` · `symbolsInDispute` | `Tracker.Observe` 1 각각 | 1 / 1 각각 — mismatch.go:645 · :644 |

## 결론

- 편집 예정 함수 `RunOnce` 의 생산 호출 자리 1(`Run`) — 이 change 가 `RunOnce` 의 계약(반환 · 순서)을 바꾸지 않으면 호출자 영향 0.
- `Observe` 생산 호출 1 자리 = RunOnce:464 → design D1 「`Observe` 의 다른 호출자 없음」 참(시험 호출자는 cmd/tossctl a195 시험 · reconcile 시험).
- 재판정이 `Compare` 의 일치 판정을 그대로 쓰려면 `quantitiesAgree` · `canonicalPresentQuantity`(둘 다 비공개, 호출자 Compare 하나)를 부르는 새 함수가 **`internal/reconcile` 패키지 안**에 있어야 함 — engine 패키지 새 파일로는 불가(공개로 바꾸면 기존 함수 편집 아님이지만 이름 변경 = 두 자리 편집).
