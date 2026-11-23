# 증거 조정 (a194, tasks 1.2)

기준: HEAD `064f59fd9`, CodeGraph 1.6.0, AST `../cited-ast/*.ast.json`(17개, 같은 HEAD), 비시험 `grep -rn`.

## CodeGraph ↔ HEAD 불일치

| # | CodeGraph 답 | HEAD 실측 | 결론 |
|---|---|---|---|
| 1 | `alertUnmanaged` 호출자에 `ExitObserver.workingSet`(exitloop.go:573) | 그 호출은 동명 `ExitObserver.alertUnmanaged`(exitloop.go:1848)의 것. `ReconcileDriver.alertUnmanaged` 호출 자리는 adoption.go:191 하나 | 이름으로 묶은 오보. a194 편집 표면에 exit 루프 없음 |
| 2 | `Tracker.Blocks` 호출자에 `alertops.go entryBlockReasons` | alertops.go:191 은 `*execgw.EntryGate` 의 `Blocks()` | 오보. Tracker.Blocks 비시험 자리는 5(position_policy_command.go:133 · reconcileloop.go:460 · mismatch.go:1327 · :1361 · cmd/tossctl/engine_reconcile.go:303) |
| 3 | `Tracker.EntryAllowed`·`Tracker.Permanent` 비시험 호출자 = `blocked` 하나 | exitwiring.go:197 `reconcileFloor.ConfirmedFloor` 에도 둘 다 있음(필드 경유 `f.tracker.…` — 해소 누락) | 자리 2. a194 는 이 자리를 건드리지 않음 — 다만 「RECONCILE 판정」의 두 번째 소비자라 D1 판정을 바꾸면 이 자리와 갈라질 수 있음(같은 술어를 쓰는 것이 안전) |
| 4 | `callees judgeHoldings` 에 `prewire_fence_linux.go Context`·`filldetect Snapshot` | 실제 타입은 `context.Context`·`reconcile.Snapshot` | 타입 해소 오보. 영향 없음 |
| 5 | `affected` = 시험 파일 1,037 | 패키지·cmd 전체가 의존으로 걸림 | 선택 근거로 못 씀. 선택 시험은 design §4 와 1.4 의 grep 전수로 정함 |

## proposal·design 주장 ↔ HEAD (요약 — 상세 대조는 review.md 「1.3 주장 대조」)

- 참: F1(B10 if@152 → continue@153, B16 range@188 에 안 닿음), F2(RunOnce B7 if@463 → return@467), F7(`reconcile.mismatch_permanent` = `obs.EventReconcilePermanent` 생산 호출 0 — obs 시험 2곳만), F6 의 결론(쓰기 자리에 알림 없음 — `internal/journal` 은 `internal/obs` 를 import 하지 않고(obs 가 journal 을 import), `internal/reconcile` 의 유일한 알림 호출은 `acked_boot.go:98` ACKED attempt 보고).
- 정정(결정 불변): F5 — 「나머지는 `Permanent`」가 원인별로 다름(아래). F6 의 「8 자리」 — 이번 grep 은 쓰기 진입 9 자리 + `Tracker.persist` 경로(세는 방법이 달라 숫자를 재현 못 함 → 숫자를 빼고 grep 표로 바꿈).
- Manager 결정 필요: D1 의 판정 재료(`EntryAllowed` 반환에 원인·시각이 없음), D5 의 기록 입구(드라이버에 `RecordCritical` 없음)·알림 꺼짐 등급, engine-safety 정본과의 관계, `cycle.Unmanaged` 셈. 보고서·review.md 참조.

## F5 상세 — 원인별 블록 모양(AST `blockFromReconcileState` B1 if@1411 · B2 if@1414 · B3 if@1417, `ReconcileReasonFor` B2 case@159 · B3 case@163)

| 원인(cause) · 범위 | Reason | Release | Permanent | 생산 쓰기 자리 |
|---|---|---|---|---|
| QUANTITY_MISMATCH · 종목 | `ReasonReconcileMismatch` | adjusted_reconcile(자동) | false | Tracker(blocksFor) |
| QUANTITY_MISMATCH · 계좌(빈 종목) | `ReasonReconcilePermanent` | operator_only | true | Tracker 영구 승격 |
| SNAPSHOT_UNAVAILABLE · SNAPSHOT_STALE · 종목 | `ReasonReconcileMismatch` | **operator_only** | **false** | **0**(상수 정의·journal 검증 목록뿐) |
| IDENTIFIER_CONFLICT · ATTRIBUTION_FAILED · 알 수 없는 원인 · 종목 | `ReasonReconcilePermanent` | operator_only | true | journal 여러 자리(lineage·fills·position_projection·position_campaign·reservation_release) |
| 위 원인 · 계좌(빈 종목) | `ReasonReconcilePermanent` | operator_only | true | position_campaign·reservation_release·position_projection 의 빈 종목 호출 |

결론: 오늘 생산이 쓰는 원인에서는 「사유가 `ReasonReconcilePermanent`」 ⇔ 「Release = operator_only」가 성립한다. 성립하지 않는 칸(SNAPSHOT_*)은 생산 쓰기 자리가 0 이다. `Tracker.Permanent()` 는 계좌 전체 **수량** 영구 승격만 뜻하고(mismatch.go:1065 `hasPermanentQuantityAccountBlock`), 계좌 전체의 다른 원인 행은 `EntryAllowed` 가 계좌 범위 `Covers`(mismatch.go:264-267)로 잡는다.
