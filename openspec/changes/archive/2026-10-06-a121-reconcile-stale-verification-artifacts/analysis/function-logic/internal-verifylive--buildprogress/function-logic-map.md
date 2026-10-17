# Function Logic Map: `BuildProgress`

- Source: `internal/verifylive/report.go` (333-362)
- Qualified function: `BuildProgress`
- Revision: `current` (구현 base `de147cc2`, `source_sha256` e96b2103… (GREEN 편집 뒤 재추출; 편집 전 4da7dc92…))
- AST evidence: `ast.json` — AST branches 7, 반환 1(342:2), 호출 9
- Risk scan: `risk-pattern-report.md`
- 편집 예정: design Revision 1 G2 — status 가 대사된 artifact 를 **reconciled absent** 로 표시 (task 3.1/3.2)

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `recordPath string` | 표시용 | 호출자(`verify.go:708` · `internal/console/data.go:375`) | 없음 |
| `entries []Entry` | 기록 전 줄, nil 가능 | `ReadRecord` 결과 | nil 이면 전 단계 Pending |

불변식: (1) 단계 상태는 카탈로그(`Steps()`) 순서로, 각 단계의 **마지막 줄**(`LastEntry`)로 정한다. (2) `LastEntry`
(`record.go:476-483`)는 `StepID` 만 보고 **`Kind` 를 보지 않는다** — 카탈로그 단계 ID 를 가진 줄이면 종류와 무관하게
그 단계의 판정이 된다. (3) 남은 객체는 `Outstanding(entries)` 한 곳(341행, 분기 아님)에서 온다.

## Branches and early returns

| Branch | AST kind | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|---|
| B1 | `range at 335:2` | 모든 줄 순회(계좌 참조) | — | 없음 | `TestProgressPointsAtTheRestartWhenOneIsPending` |
| B2 | `if at 336:3` | 계좌 참조가 비었음 | 첫 줄(종류 무관)의 `AccountRef` 채택 | 없음 | `TestProgressPointsAtTheRestartWhenOneIsPending` |
| B3 | `range at 340:2` | 카탈로그 단계 순회 | — | 없음 | `TestProgressOnAnUnstartedVerification` |
| B4 | `if at 342:3` | 그 단계 줄이 없음 | `Pending` 추가 후 `continue` | 없음 | `TestProgressOnAnUnstartedVerification` |
| B5 | `if at 349:3` | 마지막 판정이 비-종결 | `AwaitingRestart`·`Pending` | 없음 | `TestProgressPointsAtTheRestartWhenOneIsPending` |
| B6 | `range at 354:2` | 모든 줄 순회(M0 체크포인트) | — | 없음 | `TestM0CheckpointsAreVisibleInStatusButNeverAbortTargets` |
| B7 | `if at 355:3` | `Kind == KindM0Checkpoint` 이고 값 있음 | `M0Checkpoints` 추가 | 없음 | `TestM0CheckpointsAreVisibleInStatusButNeverAbortTargets` |

조기 반환 없음. 반환은 342행 한 곳.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `strings.TrimSpace` | 계좌 참조가 비었는지 | 실패 없음 | AST |
| `Steps` | 카탈로그(`verifylive.go:172`) | 순수 | AST + CodeGraph callees |
| `LastEntry` | 단계별 마지막 줄 — Kind 무관(`record.go:476`) | 순수 | AST + CodeGraph callees |
| `append` | Pending·Steps·M0Checkpoints 누적 | 해당 없음 | AST |
| `e.Verdict.Terminal` | 판정 종결 여부 | 순수 | AST |
| `Outstanding` | 남은 객체(`record.go:506` → `outstandingLines` → `Artifact.terminal`) | 순수 | AST + CodeGraph callees |

호출자: `runVerifyStatus`(`verify.go:708`), `readVerify`(`internal/console/data.go:375`), 시험 4자리.
CodeGraph 1.6.0 은 `Outcome`·`Entry` 를 다른 패키지로 잘못 해소했다(HEAD 에서 같은 패키지 타입 확인).

## State mutations and fallbacks

- 입력을 바꾸지 않는다.
- **대사 줄의 `StepID` 위험**: 대사 줄이 카탈로그 단계 ID(예: `conditional-cancel`)를 달면 B3/B4/B5 에서 그 단계의
  판정을 대사 줄 판정으로 바꾸고, 같은 원리로 `Settled`·`Passed`(`record.go:487-498`)와 `heldAfter`(`cleanup.go:183-190`,
  `StepID == gate` 만 비교)까지 움직인다. 그러므로 대사 줄은 카탈로그 밖의 `StepID`(빈 값 또는 전용 값)를 가져야 한다
  — design 이 이 값을 아직 적지 않았다(evidence-reconciliation.md 「구현 위험」).
- "reconciled absent" 표시는 341행 `Outstanding` 이 종결 artifact 를 내지 않으므로 새 필드 + `Progress.WriteText`
  (`report.go:346-379`) 편집이 필요하다 — `WriteText` 는 task 1.2 목록에 없다.

## Safety conclusion

- Safe edit boundary: 기존 7 분기는 그대로, 대사 artifact 표시 필드만 추가. 대사 줄이 Steps·Pending·AwaitingRestart 를
  바꾸지 않아야 한다.
- High-risk impact: no (읽기 전용 표시) — 단 status 의 "아직 살아 있다" 목록은 사람이 수동 취소를 판단하는 화면이므로,
  대사된 artifact 를 "취소됨"·"체결됨" 으로 쓰면 안 되고 Outstanding 에서 빠진 이유를 표시해야 한다.

## GREEN 로트 재추출 (a121 tasks 3.1·3.2, worktree a99a9059 + GREEN 편집)

- 편집: `p.Reconciled = ReconciledArtifacts(entries)` 한 줄(분기 7 불변, 번호 불변). GREEN 관측: TestReconcileTextLabelsReconciledAbsentInReportAndStatus PASS; 변이 V-buildprogress-reconciled.
- AST 재추출: `go run ./tools/logic-map` — 분기 7→7, 편집 전후 분기 열을 difflib 로 정렬한 결과 equal 뿐(재번호 없음), 앵커는 줄 이동만 반영.
