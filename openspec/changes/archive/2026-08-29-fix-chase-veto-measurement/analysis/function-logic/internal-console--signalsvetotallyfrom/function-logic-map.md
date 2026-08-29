# Function Logic Map: `signalsVetoTallyFrom`

- Source: `internal/console/signals.go`
- Change: `fix-chase-veto-measurement`
- AST evidence: `ast.json` (revision `current`, L753–780, 분기 4개)
- Risk scan: `risk-pattern-report.md`

veto tally를 화면 값으로 옮긴다. 이 change가 **루프 하나**(B2)를 더했다 — `t.Anomalies()`가
찾은 모순을 이 화면의 문장으로.

`PassedNote`는 **그대로 둔다**(§4.5). 임계가 없는 상태가 유지되므로 그 단언은 계속 참이고,
항등식 경보는 그 옆에 추가되지 그것을 대체하지 않는다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `t` | `candidate.VetoTally` | seam | 정상 상태는 경보 0 |
| `t.Passed` | 구조적으로 0 | `TallyVetoes` | 0이 아니면 note가 바뀐다 |
| `candidate.VetoCodes` | D3의 순서 | `internal/candidate` | 맵이 아니라 목록이라 화면이 흔들리지 않는다 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `t.Passed != 0` | `PassedNote`를 예상 밖 문구로 | — | `internal/console/signals_test.go`의 `signalsPassedUnexpected` 케이스 |
| B2 | `for _, a := range t.Anomalies()` **(신규)** | `out.Alarms` append | — | `TestTheSignalsScreenSaysSoWhenTheTallyContradictsItself` · `TestTheOrdinarySignalsScreenRaisesNoAlarm` |
| B3 | `for _, code := range candidate.VetoCodes` | code별 칸 | — | `signals_test.go` |
| B4 | `for why, n := range t.Reasons` | 사유 census | — | 동상 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `t.Anomalies()` | **판정** — 이 파일이 하지 않는다 | 순수 | ast.json calls |
| `signalsTallyAlarm(a)` | 문장 | 순수 | ast.json calls |
| `sortedSignalsCounts(counts)` | census 정렬 | 순수 | ast.json calls |

## State mutations and fallbacks

- 없음 — view 값 하나.
- fallback 없음. 빈 `Alarms`는 '숫자가 맞다'는 주장이 아니라 '증명 가능하게 틀리지는 않았다'이며, 그래서 경보가 숫자 옆에 붙는다.

## Safety conclusion

- Safe edit boundary: 루프 1개 + view 필드 1개 가산. 기존 세 루프와 note 로직 무변경.
- High-risk impact: no (렌더). `TestNeitherSurfaceDecidesWhetherTheTallyIsConsistent`가 이 파일이 tally count 둘을 서로 비교하지 못하게 AST로 막는다.
