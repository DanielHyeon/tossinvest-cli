# Function Logic Map: `BuildReport`

- Source: `internal/verifylive/report.go` (166-216)
- Qualified function: `BuildReport`
- Revision: `current` (구현 base `de147cc2`, `source_sha256` 4da7dc92…)
- AST evidence: `ast.json` — AST branches 10, 반환 1(215:2), 호출 11
- Risk scan: `risk-pattern-report.md`
- 편집 예정: design Revision 1 G2 — report 가 대사된 artifact 를 **reconciled absent** 로 표시하고 cancelled·filled 로
  쓰지 않음 (task 3.1/3.2)

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `recordPath string` | 표시용 경로, 해석 안 함 | 호출자(`verify.go:727` · `internal/console/data.go:428`) | 없음 |
| `entries []Entry` | 기록 전 줄(종류 무관), nil 가능 | `ReadRecord` 결과 | nil 이면 Steps 없음·전 속성 미검증 |
| `now time.Time` | 생성 시각 | 호출자 | UTC 로만 정규화 |

불변식: (1) 측정 안 된 속성은 빠지지 않고 **미검증**으로 나온다(B10). (2) 스텝이 아닌 줄(`isStepEntry` 거짓 —
approval·cleanup·m0-checkpoint, 그리고 장래 `KindReconcile`)은 Steps·관측에 들어가지 않는다(B3). (3) 남은 객체
목록은 `Outstanding(entries)` 한 곳에서 온다(212행, 분기 아님) — 종결(취소·체결) artifact 는 여기서 이미 빠지며,
현재 report 에는 **종결된 artifact 를 보여 주는 칸이 없다**.

## Branches and early returns

| Branch | AST kind | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|---|
| B1 | `range at 177:2` | 모든 줄을 순회 | `rep.AccountRef`·`rep.Steps`·`latest` 누적 | 없음 | `TestUnverifiedValuesAreNotCountedAsAnswers` |
| B2 | `if at 178:3` | 아직 계좌 참조가 비었음 | 첫 줄(종류 무관)의 `AccountRef` 를 채택 | 없음 | `TestUnverifiedValuesAreNotCountedAsAnswers` |
| B3 | `if at 181:3` | `!isStepEntry(e)` — 스텝이 아닌 줄 | Steps·관측에 넣지 않고 `continue` | 없음 | `TestAFullRunProducesAReportWithNoIdempotencyGaps` |
| B4 | `range at 190:3` | 스텝 줄의 관측 순회 | 같은 키는 뒤 줄이 이김(`latest`) | 없음 | `TestReplayStaysDisabledUnlessBothHalvesArePositive` |
| B5 | `range at 194:2` | 체크리스트 그룹 순회 | `rep.Groups` 추가 | 없음 | `TestEveryChecklistPropertyIsReportedEvenWhenUnmeasured` |
| B6 | `range at 196:3` | 그룹 속성 순회 | `g.Attributes` 추가 | 없음 | `TestEveryChecklistPropertyIsReportedEvenWhenUnmeasured` |
| B7 | `if at 198:4` | 관측이 있고 측정값임(`isMeasured`) | 값·Verified=true | 없음 | `TestUnverifiedValuesAreNotCountedAsAnswers` |
| B8 | `else at 200:11` | B7 거짓 | B9 로 | 없음 | `TestUnverifiedValuesAreNotCountedAsAnswers` |
| B9 | `if at 200:11` | 관측은 있으나 "unverified" 류 | 값만 싣고 Verified=false | 없음 | `TestUnverifiedValuesAreNotCountedAsAnswers` |
| B10 | `if at 204:4` | `!a.Verified` | `rep.Unverified` 에 키 추가 | 없음 | `TestBuildReportOnAnEmptyRecord` |

조기 반환 없음. 반환은 215행 한 곳.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `now.UTC` | 생성 시각 정규화 | 실패 없음 | AST |
| `strings.TrimSpace` | 계좌 참조가 비었는지 | 실패 없음 | AST |
| `isStepEntry` | 스텝 줄만 걸러냄(`record.go:457` — `Kind == "" \|\| Kind == KindStep`) | 순수 | AST + CodeGraph callees |
| `append` | Steps·Unverified·Attributes·Groups 누적 | 해당 없음 | AST |
| `requiredProperties` | 체크리스트(`report.go:77`) | 순수 | AST + CodeGraph callees |
| `isMeasured` | 측정값 판정(`report.go:220`) | 순수 | AST + CodeGraph callees |
| `Outstanding` | 남은 객체 목록(`record.go:506` → `outstandingLines` → `Artifact.terminal`) | 순수 | AST + CodeGraph callees |
| `replayEnabled` | 멱등 재생 허용 판정(`report.go:234`) | 순수, 기본 false | AST + CodeGraph callees |

CodeGraph 1.6.0 callees 는 `Report`·`Outcome`·`Entry` 구조체를 다른 패키지(`internal/doctor`·`internal/continuationlane`·
`internal/audit`)로 잘못 해소했다 — 실제는 같은 패키지 `verifylive` 의 타입(HEAD 확인). 호출자: `runVerifyReport`
(`verify.go:727`), `Console.report`(`internal/console/data.go:428`), 시험 7자리.

## State mutations and fallbacks

- 입력을 바꾸지 않는다. 새 `Report` 값만 만든다.
- B3 때문에 `KindReconcile` 줄의 관측(`reconcile.basis`)은 어떤 속성 값에도 들어가지 않는다 — 대사가 측정 판정을
  바꾸지 못한다는 design 「Safety and attestation boundary」 와 맞는다. 단 B2 는 종류를 보지 않으므로 대사 줄이
  **첫 줄**이면 그 `AccountRef` 가 표시 계좌가 된다(대사 줄은 기존 줄 뒤에 붙으므로 실제로는 첫 줄이 아님).
- "reconciled absent" 표시: 212행의 `Outstanding` 은 종결 artifact 를 내지 않으므로, 표시는 **새 필드**(예: 대사된
  artifact 목록)를 이 함수에서 채우고 `Report.WriteText`(`report.go:239-291`)가 그려야 한다. `WriteText` 는 task 1.2 의
  편집 대상 목록에 없다 — evidence-reconciliation.md 「범위 차이」 참조.

## Safety conclusion

- Safe edit boundary: 기존 10 분기와 `Outstanding`·`ReplayEnabled` 계산은 그대로 두고, 대사된 artifact 를 별도 필드로
  더하는 것만. 대사 줄은 B3 로 계속 걸러져야 하며 Steps·Unverified·ReplayEnabled 를 바꾸면 안 된다.
- High-risk impact: no (읽기 전용 표시) — 다만 report 는 사람이 자동 진입 금지 목록(Unverified)을 읽는 화면이라
  대사가 Unverified 를 줄이면 안전 경계 위반이다. RED(task 2.4)에서 "대사 뒤 Unverified·ReplayEnabled 불변" 을 단언.
