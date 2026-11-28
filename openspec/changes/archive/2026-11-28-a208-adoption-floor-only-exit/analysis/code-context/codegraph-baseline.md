# CodeGraph baseline — a208

- 도구: codegraph **1.6.0**, 질의 위치: 주 체크아웃 `/mnt/D/Axipient/workspace/TossOS`(HEAD `baeddfe31` = 이 worktree 의 base, `codegraph status .` →
  `Index is up to date`). worktree 에는 `.codegraph/` 가 없어 주 체크아웃에서 읽기 질의만 함. `make sdd-sync` 는 **안 돌림**(주 체크아웃의
  `.sdd` 상태를 바꾸므로) — 구현 착수 전에 worktree 에서 돌림(tasks 1.0).
- 날짜: 2026-11-28.

## affected — Go 시험 필터(.claude/CLAUDE.md 하네스 주석대로 `--filter '*_test.go'`)

| 파일 | 영향 시험 파일 수 |
|---|---|
| internal/exitpolicy/ladder.go | 1427 |
| internal/exitpolicy/snapshot.go | 1432 |
| internal/app/engine/exitloop.go | 1389 |
| internal/app/engine/adoption.go | 784 |
| internal/config/engine.go | 1359 |
| internal/config/adoption_io.go | 589 |
| internal/console/settings.go | 1356 |
| internal/app/engine/interlock.go | 1332 |
| cmd/tossctl/adoptionsettings.go | 1264 |
| 9 개 합집합 | 1444 |

전이 폐포라 거의 전 저장소임 — 선택 시험은 아래 직접 호출자와 패키지 시험으로 고름(design 8.3).

## callers (직접)

| 심볼 | 생산 호출자 | 시험 호출자(일부) |
|---|---|---|
| `exitpolicy.EvaluateLadder` | `EvaluateLadderSnapshot`(snapshot.go:120) | common_policy_test.go · ladder_test.go(10) |
| `exitpolicy.canonicalLadderInput` | `EvaluateLadderSnapshot` | — |
| `exitpolicy.EvaluateLadderSnapshot` | `ExitObserver.judgeLadder`(exitloop.go:1173) · `ValidateRecoveryDerivation`(recovery.go:116) | snapshot_test.go · recovery_validation_test.go |
| `exitpolicy.ValidateRecoveryDerivation` | `StoredExitSnapshot.validate`(journal/exit_snapshot.go:182) | recovery_validation_test.go · a111 |
| `exitpolicy.CommonLadderForPosition` | `ExitObserver.ladderFor` | `TestExternalRunnerHasNoAutomaticPartials`(common_policy_test.go:70) · policy_identity_contract_test.go · journal a078 |
| `ExitObserver.judgeLadder` | `ExitObserver.judge` | — |
| `ExitObserver.ladderFor` | `ExitObserver.judgeLadder` | — |
| `ReconcileDriver.adoptOne` | `ReconcileDriver.adopt`(adoption.go:215) | reconcileloop_test.go(편입 시험군) |
| `Journal.OpenAdoptedExitState` | `adoptOne` · `ExitObserver.openAdoptedState` | journal adoption_outcome · a130 · a200 · cmd a111 |
| `config.mergeAdoption` | `mergeEngine`(engine.go:336) | — |
| `Service.LoadRawEngineAdoption` · `SaveEngineAdoption` | **CodeGraph: 호출자 없음** (불일치 — evidence-reconciliation 참조) | — |
| `Console.handleSettingsSave` | route `ANY /settings/save`(console.go:796) | — |
| `recordGateSettings` | `NewContext`(engine.go:452) | a074 시험 4 · `TestA163TheReportToggleIsTheLastAuditedSettingAndOnlyItsChangeIsWritten` |
| `recordAdoptionSave` | `consoleAdoptionSettings.Save` | — |

원본 출력: 세션 스크래치(`callers.txt`, `a208-affected.txt`) — 커밋하지 않음(수치는 위 표로 옮김).

## 구현 착수 재확인(2026-11-28, 설계 개정 1 뒤 — tasks 1.0)

- 주 체크아웃 HEAD `9bb1a727c`(codegraph 1.6.0). 편집 5 파일의 본문은 base `baeddfe31` 와 같음(`git diff --stat baeddfe31 HEAD -- <5 파일>` 0 줄).
- `codegraph affected <file> --filter '*_test.go'`: adoption.go 785 · interlock.go 1335 · config/engine.go 1363 · adoption_io.go 589 · console/settings.go 1360(전이 폐포 — 선택 시험은 design 8.3).
- callers: `adoptOne` ← `adopt`(adoption.go:215) 1 · `handleSettingsSave` ← route `ANY /settings/save`(console.go:796) 1 · `recordGateSettings` ← `NewContext` + a074 4 · a163 1.
- a166 census(`internal/accountbudgetruntime/testdata/a166-wiring/shapes.json`): 편집 함수 6 개 중 고정 본문 0(`recordGateSettings` 는 `NewContext` 본문의 호출 줄로만 등장 — 호출 줄 무변).
- 초판의 `EvaluateLadder` · `canonicalLadderInput` · `judgeLadder` · `Context.ExitObserver` 는 편집 표면에서 빠짐(유예 삭제).
