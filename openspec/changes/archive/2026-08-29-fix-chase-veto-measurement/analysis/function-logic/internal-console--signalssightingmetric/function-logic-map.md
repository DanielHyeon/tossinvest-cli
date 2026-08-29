# Function Logic Map: `signalsSightingMetric`

- Source: `internal/console/signals.go`
- Change: `fix-chase-veto-measurement`
- AST evidence: `ast.json` (revision `current`, L705–714, 분기 2개)
- Risk scan: `risk-pattern-report.md`

최초 관측의 목록 위치를 렌더한다. 이 change가 **끝의 한 줄**을 바꿨다:

```
-   if s.NewlyListed { text += " · 신규 진입" }
-   return measuredMetric(text)
+   return measuredMetric(text + " · " + signalsNewlyListedText(s.NewlyListed))
```

`if s.NewlyListed`는 **구조적으로 항상 false**였다 — 그 필드는 다섯 타입에 선언되고
생산 소스 어디에서도 대입되지 않았다. 그래서 이 표시는 뜰 수 없었고 아무것도 실패하지
않았다. 부정과 미상까지 렌더하는 것이 그 차이를 보이게 만든다: **아무 말도 하지 않는 칸이
바로 고장 난 판본의 모습**이었다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `s.Measured` | false면 사유만 | `MeasureFirstSighting` | 미측정은 사유 텍스트 |
| `s.Rank`/`s.RankTotal` | 위치 | 저장 칼럼 | 항상 렌더된다 |
| `s.PercentilePct` | 비어 있을 수 있다 | `MeasureFirstSighting` | 비면 괄호를 붙이지 않는다 |
| `s.NewlyListed` | 3-상태 | 저장 칼럼 | 세 상태 모두 문구를 갖는다 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `!s.Measured` | 없음 | 사유를 실은 미측정 metric | `TestARefusedSightingNamesWhyRatherThanShowingTheMarker` · `TestATruncatedReadingIsNamedOnTheScreenToo` |
| B2 | `s.PercentilePct != ""` | 괄호로 백분위 추가 | — | `TestTheNewEntrantMarkerRendersAllThreeStatesDistinctly` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `unmeasuredMetric(string(s.Reason()))` | 사유 — 절대 비지 않는다 | 순수 | ast.json calls |
| `signalsNewlyListedText(s.NewlyListed)` **(신규)** | 3-상태 문구 | 순수 | ast.json calls |
| `measuredMetric(...)` | 측정된 칸 | 순수 | ast.json calls |

## State mutations and fallbacks

- 없음 — 문자열 하나.
- fallback 없음. 미상도 자기 문구를 갖는다 — 빈 칸이 고장의 모습이었기 때문이다.

## Safety conclusion

- Safe edit boundary: 본문 3줄 → 1줄. 미측정 경로와 백분위 괄호 무변경.
- High-risk impact: no (렌더). 재는 성질은 High-risk 인접 — 이 칸이 다시 침묵하면 다섯 층에 선언된 사실이 또 기록되지 않은 채 출하된다.
