# Function Logic Map: `Progress.WriteText`

- Source: `internal/verifylive/report.go` (346-379)
- Qualified function: `Progress.WriteText`
- Revision: `current` (구현 base `de147cc2` = 고정 사본 `2c6ef1ef` 의 같은 파일, `source_sha256` 4da7dc92…)
- AST evidence: `ast.json` — AST branches 8, 반환 1(351:3 조기 반환), 호출 26
- Risk scan: `risk-pattern-report.md`
- 편집 예정: design 「로트 1 처분」 S1 — `verify status` 텍스트가 대사된 artifact 를 **reconciled absent** 로 표시(task 3.1/3.2).
  tasks 1.2 유보 번들.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `p Progress` (값 수신자) | `BuildProgress` 결과 | `BuildProgress`(`report.go:315-343`) | 없음 — 렌더만 |
| `w io.Writer` | 비-nil | 호출자 `runVerifyStatus`(`cmd/tossctl/verify.go:714`) | 쓰기 오류 무시 |

불변식: (1) 읽기만 한다. (2) "살아 있다" 절(B5~B6)은 `p.Outstanding` 만 본다 — `BuildProgress` 의 `Outstanding(entries)`
(`report.go:341`). (3) B5 절의 마지막 줄은 "`tossctl verify run --resume` 가 … 이들을 취소한다" 로 **취소를 권한다** — 대사된
artifact 가 이 절에 남으면 운영자에게 없는 객체의 취소(=같은 DELETE 재시도)를 권하는 셈이다.

## Branches and early returns

| Branch | AST kind | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|---|
| B1 | `if at 348:2` | 단계 0 이고 M0 체크포인트 0 | 시작 안내 2줄 | **조기 반환 351:3** | `TestProgressOnAnUnstartedVerification` |
| B2 | `range at 356:2` | 단계 순회 | 단계·판정·이유 출력 | 없음 | `TestProgressPointsAtTheRestartWhenOneIsPending` |
| B3 | `if at 359:2` | `len(p.Pending) > 0` | "남은 단계" 줄 | 없음 | `TestProgressPointsAtTheRestartWhenOneIsPending` |
| B4 | `if at 362:2` | `p.AwaitingRestart != ""` | 재시작 안내 | 없음 | `TestProgressPointsAtTheRestartWhenOneIsPending` |
| B5 | `if at 366:2` | `len(p.Outstanding) > 0` | "⚠ … 살아 있다" 머리줄 + 취소 권고 꼬리줄 | 없음 | `TestProgressPointsAtTheRestartWhenOneIsPending` |
| B6 | `range at 368:3` | outstanding 순회 | 종류·id·심볼 | 없음 | 같음 |
| B7 | `if at 373:2` | `len(p.M0Checkpoints) > 0` | "M0 복구 체크포인트 (취소 대상 아님)" | 없음 | `TestM0CheckpointsAreVisibleInStatusButNeverAbortTargets` |
| B8 | `range at 375:3` | 체크포인트 순회 | kind·client·parent·child | 없음 | 같음 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `fmt.Fprintf` / `fmt.Fprintln` | 출력 | 오류 무시 | AST |
| `len` | 빈 판정 | 순수 | AST |
| `orNone` | 빈 값 표시 | 순수 | AST |
| `VerdictLabel` · `truncate` · `joinSteps` | 라벨·자르기·단계 목록 | 순수 | AST |

호출자(CodeGraph 1.6.0 + HEAD): `runVerifyStatus`(`cmd/tossctl/verify.go:714`), 시험 `report_test.go`·`m0_status_test.go`.
콘솔은 `Progress` 를 `html/template` 으로 그리며 이 함수를 부르지 않는다(design S1 — 콘솔 비편집).

## State mutations and fallbacks

- 입력을 바꾸지 않는다.
- **B1 조기 반환**: 단계 0·체크포인트 0 이면 Outstanding 절도 숨는다. 대사 줄은 `LastEntry` 로 카탈로그 단계를 찾는
  `BuildProgress` B3/B4 에 걸리지 않아야 한다 — 대사 줄의 `StepID` 가 카탈로그 밖이어야 하는 이유(design R1). 그 조건이
  지켜지면 대사 줄만 더해진 기록은 B1 을 바꾸지 않는다.
- **표시 위험**: 대사된 artifact 를 B5~B6 에 남기면 꼬리줄이 그것의 취소를 권한다.

## Safety conclusion

- Safe edit boundary: 기존 8 분기·조기 반환 유지, 대사된 artifact 를 보이는 새 절(예: "reconciled absent — 취소·체결 아님")만
  추가. B5 의 취소 권고는 `Outstanding` 에 남은 것에만 붙는다.
- High-risk impact: no(읽기 전용 렌더) — 단 B5 꼬리줄이 운영자 행동(재개=DELETE)을 권하므로 라벨 정확성은 안전 문제.
- RED(로트 2): `TestReconcileTextLabelsReconciledAbsentInReportAndStatus`(`internal/verifylive/reconcile_projection_test.go`).
