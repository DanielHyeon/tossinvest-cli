# Function Logic Map: `signalsNewlyListedText`

- Source: `internal/console/signals.go`
- Change: `fix-chase-veto-measurement`
- AST evidence: `ast.json` (revision `current`, L722–731, 분기 4개)
- Risk scan: `risk-pattern-report.md`

**신규 함수**. 신규 진입 사실의 세 상태를 각각의 문구로 만든다.

미상 arm이 오늘 도달 불가인 것은 구조상 그렇다 — `MeasureFirstSighting`이 미상을 거부하므로
측정된 `Sighting`은 항상 알려진 사실을 갖는다. 그래도 두는 이유는 **다른 producer의
`Sighting`이 여기 도달하면** 빈 칸이 되고, 빈 칸은 출하된 결함과 구분되지 않기 때문이다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `n` | `candidate.NewlyListed` 3-상태 | 저장 칼럼 | 세 상태 모두 문구가 있다 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | 상태 switch | 없음 | — | `TestTheNewEntrantMarkerRendersAllThreeStatesDistinctly` |
| B2 | `n.Yes()` | 없음 | `"신규 진입"` | 동상 |
| B3 | `n.No()` | 없음 | `"직전 읽기에 있었음"` | 동상 |
| B4 | default (미상) | 없음 | `"신규 진입 미상"` | 동상 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `n.Yes()` / `n.No()` | 3-상태 술어 — **둘 다 measurement를 요구한다** | 순수 | ast.json calls |

## State mutations and fallbacks

- 없음 — 문자열 하나.
- fallback: default가 미상이다. 값을 지어내지 않고 미상을 **이름 붙인다**.

## Safety conclusion

- Safe edit boundary: 신규 함수. 기존 동작 없음.
- High-risk impact: no (렌더).
