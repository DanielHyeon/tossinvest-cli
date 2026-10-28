# a164 evidence reconciliation

| 질문 | CodeGraph | CGC | HEAD(AST·git grep) | 결론 |
| --- | --- | --- | --- | --- |
| `terminal()` 호출 자리 | 1(`joinChain`) | 낡음 | 4(record.go 712·722 · projection_reconciled.go 18 · runner.go 1052) | HEAD. 넷째 종결은 술어 한 곳 편집으로 네 자리에 다 닿음 |
| `Reconcile` 생산 호출자 | 8(이름 충돌 포함) | 낡음 | `runVerifyReconcile` 1 | HEAD |
| 미결 HOLD 지점 | — | 낡음 | cmd `validateM0TriggerMode` B9 `:428` · `New` B9 `:346` | AST |
| 전제 판정 순서 | — | 낡음 | cmd: B8 `M0ExactPrerequisites` `:424` 가 B9 `:428` 앞 / `New`: B9 `:346` 이 B15 `:358` 앞 | AST — a164 만으로 cmd 진입은 안 풀림(design D9) |
| Q1 정적 거절 | — | 낡음 | `Reconcile` B8 `:205` → `checkReconcileRetention` B1 `:118` | AST |
| Q6 거절 | — | 낡음 | `checkReconcileAbsence` B5 `:384` | AST |

불일치 둘(CodeGraph 누락·이름 충돌)은 HEAD 로 해소했다. 편집 차단 불일치 없음. 편집 직전 재확인은 tasks 0.9.

## 0.9 재확인 (2026-10-22, 구현 worktree `a164-impl` @ `aec9bde67`)

- `make sdd-sync`: CodeGraph fingerprint 기록(`.sdd/index-state.json` codegraph head `aec9bde67`) — **hard evidence 신선**.
  CodeGraphContext 는 `incomplete`(R5 그대로 — advisory 미반영), GBrain 은 worktree-로컬 홈에 새 source 로 색인(파싱 오류 다수, advisory).
- `codegraph affected … --filter '*_test.go'`(편집 대상 8 파일) → **1,208** 시험 파일(`../codegraph-affected-0.9.txt`). 0.5 의 1,212 와
  같은 규모 — cmd/tossctl 전이 의존 때문. 선택 시험은 패키지 셋(verifylive · official · cmd/tossctl Reconcile 계열)으로 좁힘(review.md).
- callers 재질의(1.6.0): `checkReconcileAbsence` 1 · `checkReconcileWindow` 1(`Reconcile`) · `DecodeReconcileConditionalPage` 1 생산
  (`ReconcileConditionalOrdersPage`) · `ReconciledArtifacts` 2 생산(`BuildReport`·`BuildProgress`) · `writeReconciled` 2(두 `WriteText`) ·
  `terminal` 1(`joinChain`, **여전히 누락**) · `MarshalJSON` 0(인터페이스 호출이라 그래프에 없음).
- HEAD grep 대조: `.terminal()` 생산 4 자리(record.go 712·722 · projection_reconciled.go 18 · runner.go 1052) · `checkReconcileWindow(` 호출 2
  (reconcile.go 245·304) · `checkReconcileAbsence(` 1 · `ReconciledArtifacts(` 2. 불일치는 0.5 와 같은 둘(terminal 누락·이름 충돌)뿐 — HEAD 로 해소, 편집 차단 없음.
- 추가 발견(HEAD): `Reconcile` 본문은 AST 구조 핀 셋이 고정함 — `checkReconcileWindow` 정확히 2회, `checkReconcileInstrument` 2·`sameReconcileMultiset` 1·
  `decodeRecordStrict` 1·`appendLine` 1, lock → `encodeReconcileLine` → 마지막 게이트 → 쓰기 순(`reconcile_green_test.go:52` ·
  `reconcile_codex_test.go:185`). 그래서 만료 모드는 같은 함수 안의 `p.Expired` 분기로 넣고 잠금·쓰기 경로를 복사하지 않음(Pre-Edit).
