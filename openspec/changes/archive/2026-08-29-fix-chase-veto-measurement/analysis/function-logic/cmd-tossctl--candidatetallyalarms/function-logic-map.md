# Function Logic Map: `candidateTallyAlarms`

- Source: `cmd/tossctl/candidate.go`
- Change: `fix-chase-veto-measurement`
- AST evidence: `ast.json` (revision `current`, L636–642, 분기 1개)
- Risk scan: `risk-pattern-report.md`

**신규 함수**. tally의 모든 anomaly를 이 표면의 문장 목록으로 만든다.

boolean이 아니라 목록인 이유는 각 항목이 **그것을 만든 숫자**를 싣기 때문이다. "뭔가
일관되지 않다"만 들은 소비자에게는 확인할 것이 없다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `t` | `candidate.VetoTally` | `CycleResult.Vetoes` | 정상 상태는 빈 목록 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `for _, a := range t.Anomalies()` | `out` append | — | `TestTheScanOutputSaysSoWhenTheTallyContradictsItself` · `TestTheOrdinaryNoThresholdScanRaisesNoAlarm` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `t.Anomalies()` | **판정** — 이 파일이 하지 않는다 | 순수 | ast.json calls |
| `tallyAlarm(a)` | 문장 | 순수 | ast.json calls |

## State mutations and fallbacks

- 없음 — 문자열 슬라이스 하나. 빈 결과는 nil이고 JSON에서 `omitempty`로 사라진다.
- fallback 없음.

## Safety conclusion

- Safe edit boundary: 신규 함수. 기존 동작 없음.
- High-risk impact: no (렌더). 빈 결과가 '숫자가 맞다'는 주장이 **아니라는** 것이 `Anomalies`의 문서에 적혀 있고, 그래서 경보는 숫자를 대체하지 않고 옆에 붙는다.
