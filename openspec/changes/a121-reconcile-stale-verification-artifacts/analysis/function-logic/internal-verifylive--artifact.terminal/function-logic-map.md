# Function Logic Map: `Artifact.terminal`

- Source: `internal/verifylive/record.go` (588-588)
- Qualified function: `Artifact.terminal`
- Revision: `current` (구현 base `de147cc2`, `source_sha256` ac81738d… (GREEN 편집 뒤 재추출; 편집 전 df526d2c…))
- AST evidence: `ast.json` — AST branches 0, 반환 1(575:37), 호출 0
- Risk scan: `risk-pattern-report.md`
- 편집 예정: design Revision 1 G2 — 셋째 종결 `ReconciledAbsent` 를 이 술어에 더함 (task 3.1)

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 수신자 `a Artifact` (값 복사) | 기록 한 줄의 artifact 한 개 | `record.go:182-244` `Artifact` 정의 | 부작용 없음 — 순수 술어 |
| `a.Cancelled` | bool, 이 도구가 취소를 확인했을 때만 true | `record.go:193-194` | false 면 다음 항으로 |
| `a.Filled` | bool, 체결 관측 시 true. 2026-07-30 이전 줄은 필드 없음 = false | `record.go:195-218` | false 면 비-종결 |

불변식: 종결 판정은 저장소 전체에서 이 술어 한 곳이다(주석 `record.go:569-574`). 종결은 단조다 — 소비자
`outstandingLines` B4(`record.go:552`)가 종결 뒤의 비-종결 줄을 되살리지 않는다.

## Branches and early returns

AST 가 분기 0 을 낸다. 단락 평가(`||`)는 AST 분기가 아니므로 행복 경로 한 행으로 적는다.

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | 분기 없음 — `return a.Cancelled \|\| a.Filled` 단일 식 | 없음 | 두 필드 중 하나라도 true 면 true | `TestAFilledObjectIsNotOutstanding` · `TestTheRealRecordsAreJudgedExactlyAsBefore` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| (없음) | 호출 0 — AST `calls: null` | 해당 없음 | AST |

호출자(CodeGraph 1.6.0 `callers terminal` 3건 = HEAD grep 5자리):

| 호출 함수 | 자리 | 의미 |
|---|---|---|
| `outstandingLines` | `record.go:552`(B4, 두 번) · `record.go:562`(B6) | 투영의 단조성·비-종결만 투영 |
| `Runner.joinChain` | `runner.go:1052` | 진행 중 단계의 `sr.artifacts` 중 비-종결만 체인에 붙임 |
| `TestTheRealRecordsAreJudgedExactlyAsBefore` | `record_replay_test.go:109`·`:111` | 실기록에서 `terminal() == Cancelled` 를 단언 |

`outstandingLines` 를 거쳐 따르는 생산 소비자(HEAD grep `Outstanding(`): `abort.go:66`·`:143`,
`cleanup.go:125`(`cleanupFrom`)·`:282`, `mutate.go:679`(`liveCount`), `redo.go:122`, `report.go:212`·`:341`,
`runner.go:346`·`:877`, `steps.go:1094`·`:1138`, `cmd/tossctl/verify.go:427`.

## State mutations and fallbacks

- 상태 변경 없음, fallback 없음. 값 수신자라 호출자 상태를 바꿀 수 없다.
- 셋째 종결을 더하면 위 소비자 전부가 같은 판정을 받는다(설계 의도, 주석 `record.go:573-574`).
- `Runner.joinChain` 은 진행 중 단계의 artifact 만 보므로 대사 줄(별도 `KindReconcile` 줄)은 거기 들어오지 않는다.
- `record_replay_test.go:109` 의 `terminal() != Cancelled` 단언은 실기록(대사 줄 없음)에서 새 필드가 늘 false
  이므로 그대로 성립해야 한다 — 성립하지 않으면 셋째 종결이 기존 기록의 판정을 바꾼 것이다.

## Safety conclusion

- Safe edit boundary: 술어에 `|| a.ReconciledAbsent` 만 더하고 기존 두 항·값 수신자 형태는 유지. 새 필드가
  false 인 모든 기존 줄의 판정이 불변이어야 한다(`TestTheRealRecordsAreJudgedExactlyAsBefore`).
- High-risk impact: yes — 이 술어가 정리(취소) 대상 선택(`cleanupFrom`)과 노출 상한(`liveCount`)을 정한다.
  잘못 넓히면 살아 있는 조건주문이 정리·상한에서 빠진다. 그래서 새 필드를 세우는 경로는 대사 명령 하나뿐이어야 한다.

## GREEN 로트 재추출 (a121 tasks 3.1·3.2, worktree a99a9059 + GREEN 편집)

- 편집: `|| a.ReconciledAbsent` 셋째 종결 추가(분기 0 유지). GREEN 관측: TestReconcileOnlyRemovesItsExactArtifactFromCleanupPlanning·TestReconcilePinsRedoSetBeforeAndAfter·TestReconcileLeavesTheReportAttributesAndVerdictsUnchanged PASS; 변이 V-terminal-third 원장 참조.
- AST 재추출: `go run ./tools/logic-map` — 분기 0→0, 편집 전후 분기 열을 difflib 로 정렬한 결과 equal 뿐(재번호 없음), 앵커는 줄 이동만 반영.
