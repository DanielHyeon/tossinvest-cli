# a201 v3 편집 표면 · 고정 장치 · 겹침 (tasks 1.7)

HEAD `6e36fd2c0`(origin/main `d66536d9b` merge). CodeGraph 1.6.0(`v3-evidence/codegraph-callers.txt` · `codegraph-affected.txt`), AST `cited-ast/` · `v3-evidence/ast/` · `function-logic/*/ast.json`.

## 1. 편집할 기존 함수 — 넷(Manager 예상 다섯 중 `Restore` 무편집)

| 함수 | 파일 | 편집 | 생산 호출 자리 | 번들 |
|---|---|---|---|---|
| `ReconcileDriver.RunOnce` | internal/app/engine/reconcileloop.go:425-479 | call@434 앞 표식 1문장 · call@477 앞 부여 1자리 | `Run` call@412 — 1 | `function-logic/internal-app-engine--reconciledriver.runonce/` |
| `Tracker.Observe` | internal/reconcile/mismatch.go:629-856 | Cleared@701 직전 원인 표지 1문장 | RunOnce call@464 — 1 | `…--tracker.observe/`(a192 번들 함수 — 편집 지점 뒤 분기 번호 변화 없음 예상: 대입식은 분기 아님) |
| `Tracker.persist` | internal/reconcile/mismatch.go:878-954 | call@935 의 `Cause` · `Evidence` 를 표지로 고름 | Observe call@786 — 1 | `…--tracker.persist/` |
| `journal.ValidReconcileReleaseCause` | internal/journal/reconcile_states.go:111-119 | case 1 추가 | 쓰기 검사 3 자리(:269 · :351 · reconcile_order_terminal_release.go:106) | `…--validreconcilereleasecause/` |

- `Restore` 무편집: `t.adjusted = nil`(:1001)이 체결 credit 도 지움. 곁 맵 `t.filled` 의 낡은 항목은 「`t.adjusted` 와 같은 as-of」 일 때만 원인을 바꾸고, 재기동 뒤 그 as-of 의 credit 은 없으므로 해제를 만들지 못함(부여 메서드가 낡은 항목을 정리). 재기동 뒤 같은 as-of(초 단위)의 조정 credit 이 생길 수 있는지는 시계가 뒤로 가지 않는 한 0.
- 구조체 칸(함수 편집 아님): `Tracker.filled map[string]string`, `Block` 의 비공개 원인 표지(bool).
- 새 파일: `internal/reconcile/<a201>.go`(로컬 읽기 전 표식 타입 · 표식 읽기 · 부여 대상 계산 · `Tracker` 부여 메서드), `internal/journal/<a201>.go`(상수 `ReconcileReleaseFillApplied = "FILL_APPLIED"` — a192 의 `ReconcileReleaseOrderTerminalConfirmed` 처럼 별도 파일).
- 무편집 확인 대상: `ConvergeQuantities` · `Compare` · `LocalStateFromJournal` · `AdjustmentApplied` · `refuteUsableAdjustmentCredits` · `creditUsableBy` · `hasBlockFor` · `orderTerminalCandidate` · `enterReconcileWithBinding` · `Refresh` · `Restore` · `buildGateway` · `logCycle` · `note` · 스키마 36.

## 2. 고정 장치

| 장치 | 적중 | 처분 |
|---|---|---|
| `a071_readiness_assembly_test.go:27-40`(reconcileloop.go 금지 문자열) | 비적중(쓰지 않음) | 2.6 선택 시험 |
| a166 census(`internal/accountbudgetruntime/testdata/a166-wiring/`) | reconcileloop.go · mismatch.go · reconcile_states.go 키 0, `buildGateway` 무편집 | 2.6 `./internal/accountbudgetruntime` |
| `TestUnknownReleaseCauseIsRefused`(reconcile_states_test.go:334-356) | 유효 목록 열거 — 새 상수는 추가해야 하는 줄(안 깨짐) | RED/GREEN 에 추가 |
| broker-zero 허용 목록(risk_bucket_owner.go:280 · :677) | 무편집 — 새 원인을 거절(의도) | v3 시험 §3-10 |
| engine 패키지 전체를 파싱하는 구조 시험(~30) | 새 식별자(표식 · 부여 메서드) 금지 · 센서스 이름 0(1단계 grep) | 2.6 선택 시험으로 실측 |
| a192 아카이브 FLM 번들(Observe · persist · ValidReconcileReleaseCause · RunOnce) | 다시 고치는 함수 — 이 change 의 새 번들 | 1.9 pre-edit · 2.4 post-edit |
| gate ⑤ 창 | base `f9b818a0e` 뒤 merge 로 들어온 a200 `ProjectPosition` 이 지금 required(`code-context/check-analysis-v3-pre.txt` rc 1) | 종결 때 착지 기록으로 좁힘(자기 Go 커밋 뒤), 또는 「사람 승인 base 재고정」 — Manager 질문 |

## 3. 진행 중 change 겹침

| change | 겹침 | 처분 |
|---|---|---|
| a199(종결 중) | `reconcileloop.go` 의 `Context.ReconcileDriver`(:368) · 패키지 주석 — RunOnce 와 다른 함수 | 구현 전 origin/main 재확인 · merge, `reconcileloop.go` 가 바뀌면 RunOnce 번들 재추출 |
| a200(착지) | Observe · persist 를 지나는 시험(`internal/reconcile/a200_*` · `internal/journal/a200_*`) | 2.6 선택 시험에 포함 |
| a197 · a087 | 대상 파일 무편집 | — |
| PM | `_registry.yaml` · `FEAT-TOS-005.yaml` 목록 끝 | merge 때 기계적 해소(이번 merge 에서 해 봄) |
