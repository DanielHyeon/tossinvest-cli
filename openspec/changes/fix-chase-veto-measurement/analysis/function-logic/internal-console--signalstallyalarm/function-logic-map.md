# Function Logic Map: `signalsTallyAlarm`

- Source: `internal/console/signals.go`
- Change: `fix-chase-veto-measurement`
- AST evidence: `ast.json` (revision `current`, L547–558, 분기 3개)
- Risk scan: `risk-pattern-report.md`

**신규 함수**. 하나의 산술 모순을 이 화면의 문장으로 만든다.

판정은 `candidate.VetoTally.Anomalies`이고 여기서 반복하지 않는다. 그 분리가 네 번째 그림자
밴드의 교훈이다 — 두 번 구현한 규칙은 언젠가 자기 자신과 어긋나고, 그 어긋남은 **평온해
보이는 페이지**로 나타난다.

숫자를 **대체하지 않고 옆에** 붙는다. 경보의 근거가 그 숫자이기 때문이다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `a.Kind` | 두 kind | `Anomalies()` | 미인식은 default 문장 |
| `a.Arithmetic()` | 실패한 합 | `tallycheck.go` | 조언이 들어 있지 않다 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `switch a.Kind` | 없음 | — | `TestTheSignalsScreenSaysSoWhenTheTallyContradictsItself` |
| B2 | `TallyPassedWithoutThreshold` | 없음 | D18을 인용한 한국어 문장 | 동상 |
| B3 | default (`OVERCOUNTED`) | 없음 | 세 버킷이 서로소라는 한국어 문장 + D10 | 동상 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `a.Arithmetic()` | 숫자와 실패한 비교 | 순수 | ast.json calls |

## State mutations and fallbacks

- 없음 — 문자열 하나.
- fallback: 미인식 kind는 일반 문장. 침묵하지 않는다.

## Safety conclusion

- Safe edit boundary: 신규 함수. 기존 동작 없음.
- High-risk impact: no (렌더). 재는 성질은 High-risk — 이 문장이 없으면 tally 모순이 화면에서 평범한 숫자로 보인다.
