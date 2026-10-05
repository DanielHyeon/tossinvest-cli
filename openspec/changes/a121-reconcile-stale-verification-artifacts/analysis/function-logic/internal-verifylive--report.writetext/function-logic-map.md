# Function Logic Map: `Report.WriteText`

- Source: `internal/verifylive/report.go` (239-291)
- Qualified function: `Report.WriteText`
- Revision: `current` (구현 base `de147cc2` = 고정 사본 `2c6ef1ef` 의 같은 파일, `source_sha256` 4da7dc92…)
- AST evidence: `ast.json` — AST branches 11, 반환 1(247:3 조기 반환), 호출 28
- Risk scan: `risk-pattern-report.md`
- 편집 예정: design 「로트 1 처분」 S1 — 대사된 artifact 를 **reconciled absent** 로 텍스트 출력에 표시(task 3.1/3.2). tasks 1.2 유보 번들.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `rep Report` (값 수신자) | `BuildReport` 결과 | `BuildReport`(`report.go:166-216`) | 없음 — 렌더만 |
| `w io.Writer` | 비-nil | 호출자(`runVerifyReport` `verify.go:733` · 콘솔 `internal/console/pages.go:357`) | 쓰기 오류는 무시(`fmt.Fprintf` 반환값 미검사) |

불변식: (1) 이 함수는 `rep` 를 읽기만 한다 — 판정을 만들지 않는다. (2) "살아 있다" 절(B9~B11)은 `rep.Outstanding` 한 칸만
본다 — 그 칸은 `BuildReport` 의 `Outstanding(entries)`(`report.go:211`)에서 오고, 대사된 artifact 가 그 칸에서 빠지는 것은
`Artifact.terminal()` 편집의 결과다(이 함수 편집이 아니다). (3) 종결된 artifact 를 담는 칸은 `Report` 에 없다 — "reconciled
absent" 를 **보이게** 하려면 새 칸 + 이 함수의 새 출력 절이 필요하다(S1).

## Branches and early returns

| Branch | AST kind | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|---|
| B1 | `if at 245:2` | `len(rep.Steps) == 0` | "아직 기록된 검증이 없다" 출력 | **조기 반환 247:3** — 뒤의 모든 절(Outstanding 포함) 생략 | `TestBuildReportOnAnEmptyRecord` |
| B2 | `range at 251:2` | 단계 순회 | 단계·판정·이유·라벨 출력 | 없음 | `TestReportTextNamesTheUnverifiedProperties` |
| B3 | `range at 256:2` | 그룹 순회 | 그룹 이름 출력 | 없음 | 같음 |
| B4 | `range at 258:3` | 속성 순회 | — | 없음 | 같음 |
| B5 | `if at 260:4` | 값이 빈 문자열 | 표시값 `unverified` 로 대체 | 없음 | 같음 |
| B6 | `if at 264:4` | `!a.Verified` | 표지 `!` | 없음 | 같음 |
| B7 | `if at 268:4` | `a.Detail != ""` | 세부 줄 출력 | 없음 | 같음 |
| B8 | `range at 277:2` | 미검증 키 순회 | 키 출력 | 없음 | 같음 |
| B9 | `if at 281:2` | `len(rep.Outstanding) > 0` | "⚠ … 아직 계좌에 살아 있다" 머리줄 | 없음 | `TestVerifyReportKeepsReplayDisabledWithoutEvidence`(cmd) |
| B10 | `range at 283:3` | outstanding 순회 | 종류·id·심볼·노트 출력 | 없음 | 같음 |
| B11 | `if at 285:4` | `a.Deliberate` | 노트 앞에 "존속 측정을 위해 의도적으로 남긴 것" | 없음 | 같음 |

조기 반환은 B1 한 곳(247:3). 그 외에는 끝까지 흐른다.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `fmt.Fprintf` / `fmt.Fprintln` | 출력 | 오류 무시 | AST |
| `orNone` | 빈 계좌 참조 표시 | 순수 | AST |
| `rep.GeneratedAt.Format` | 생성 시각 | 순수 | AST |
| `VerdictLabel` · `StepLabel` | 판정·단계 한글 라벨 | 순수 | AST |
| `truncate` | 긴 이유·노트 자르기 | 순수 | AST |
| `replayVerdict` | 멱등 재생 문장 | 순수 | AST |
| `len` | 빈 판정 | 순수 | AST |

호출자(CodeGraph 1.6.0 + HEAD 확인): `runVerifyReport`(`cmd/tossctl/verify.go:733`), 콘솔 보고 페이지
(`internal/console/pages.go:357`), 시험 `report_test.go`. 외부 I/O·브로커 호출 없음.

## State mutations and fallbacks

- 입력을 바꾸지 않는다. 출력만 낸다.
- **B1 조기 반환의 귀결**: 단계 줄이 0 인 기록에서는 Outstanding 절이 렌더되지 않는다. 대사 줄(`KindReconcile`)은
  `BuildReport` B3(`!isStepEntry`)에서 Steps 에 들어가지 않으므로, 대사 줄만 더해진 기록은 이 분기를 바꾸지 않는다 —
  그러나 S1 의 새 "reconciled absent" 절을 B1 뒤에 두면 단계 줄이 없는 기록에선 그 절도 숨는다(실제 a063 기록은 단계 줄이
  있으므로 영향 없음, 편집 시 위치를 B1 뒤로 두는 것이 기존 관례).
- **표시 위험**: 대사된 artifact 를 B9~B11 의 "살아 있다" 절에 남기거나 "취소"·"체결" 로 쓰면 design G2("cancelled·filled 로
  쓰지 않는다")를 어긴다.

## Safety conclusion

- Safe edit boundary: 기존 11 분기·조기 반환은 그대로 두고, `rep` 의 새 칸(대사된 artifact)을 렌더하는 절만 추가한다. B9~B11
  은 `Outstanding` 투영 변화(terminal 편집)로만 달라진다.
- High-risk impact: no(읽기 전용 렌더) — 단 운영자가 수동 취소를 판단하는 화면이므로 라벨 오류는 안전 문제로 다룬다.
- RED(로트 2): `TestReconcileTextLabelsReconciledAbsentInReportAndStatus`(`internal/verifylive/reconcile_projection_test.go`).
