# Function Logic Map: `tallyAlarm`

- Source: `cmd/tossctl/candidate.go`
- Change: `fix-chase-veto-measurement`
- AST evidence: `ast.json` (revision `current`, L621–633, 분기 3개)
- Risk scan: `risk-pattern-report.md`

**신규 함수**. 하나의 산술 모순을 이 표면의 문장으로 만든다.

판정은 `candidate.VetoTally.Anomalies`이고 **여기서 반복하지 않는다** — 두 화면이 숫자가
맞는지에 대해 서로 다른 답을 낼 수 있으면 안 된다. 여기 있는 것은 문장뿐이며, 읽는 사람이
다르기 때문이다: 이쪽은 숫자 옆의 터미널 영어이고 `/signals`는 브라우저의 한국어다.

문장이 명명하는 것은 **D18이 금지하는 그 수리**다 — 임계 없는 veto를 통과로 세는 것.
그것이 아무도 임계를 승인하지 않은 채 이 상태를 만드는 유일한 편집이기 때문이다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `a.Kind` | `OVERCOUNTED` 또는 `PASSED_WITHOUT_THRESHOLD` | `VetoTally.Anomalies` | 미인식 kind는 default 문장 |
| `a.Arithmetic()` | 실패한 합, 숫자 포함 | `tallycheck.go` | 문장이 아니라 산술이다 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `switch a.Kind` | 없음 | — | `TestEveryAnomalyKindStatesItsOwnArithmetic` |
| B2 | `candidate.TallyPassedWithoutThreshold` | 없음 | THRESHOLD_ABSENT를 통과로 센 것이라는 문장 | `TestTheScanOutputSaysSoWhenTheTallyContradictsItself` |
| B3 | default (`OVERCOUNTED`) | 없음 | 세 버킷이 서로소라는 문장 + D10 인용 | 동상 |

default가 `OVERCOUNTED`를 잡는 형태인 것은 의도다 — 새 kind가 생기면 사라지는 대신
일반 문장을 얻는다. 빈 문자열은 화면에서 '경보 없음'과 구분되지 않는다.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `a.Arithmetic()` | 실패한 합과 숫자 | 순수 — 조언이 들어 있지 않다 | ast.json calls |

## State mutations and fallbacks

- 없음 — 문자열 하나.
- fallback: 미인식 kind는 일반 문장. 침묵하지 않는 것이 요점이다.

## Safety conclusion

- Safe edit boundary: 신규 함수. 기존 동작 없음.
- High-risk impact: no (렌더). 재는 성질은 High-risk — 이 경보가 침묵하면 미측정 veto가 통과로 세어지는 것이 화면에서 평범해 보인다.
