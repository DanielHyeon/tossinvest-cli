# Function Logic Map: `qualifiesFirstRank`

- Source: `internal/candidate/scan.go`
- Change: `fix-chase-veto-measurement`
- AST evidence: `ast.json` (revision `current`, L639–641, 분기 0개)
- Risk scan: `risk-pattern-report.md`

**신규 함수** (2026-07-28 pre-gate 리뷰 P1-3, issues.md I17). 한 읽기가 **저장 가능한 최초
관측 위치**를 실을 자격이 있는지를 판정하는 순수 술어다.

```go
r.NewlyListed.Known() && r.RankRequested > 0
```

두 절 모두 **시장에 대한 판단이 아니라 소스가 자기 읽기에 대해 declare한 사실**이다:
"직전 읽기를 갖고 있었는가"와 "몇 행을 요청했는가". 둘 중 하나가 없는 읽기는 자기 위치가
무엇을 뜻하는지 말할 수 없고, `first_rank`는 write-once이므로 그 침묵이 후보의 남은 생명
전체를 답한다.

**왜 함수인가**: 같은 두 절이 두 자리에서 필요하고, 두 자리가 **반드시 같은 답을 내야**
한다. `Collect`는 이것으로 tick 안의 어떤 읽기가 위치를 차지하는지 정하고, `recordFirsts`는
그 위치를 저장할지 정한다. 두 규칙이 갈라지면 — 실제로 갈라져 있었다 — `Collect`가 채택한
읽기를 `recordFirsts`가 보류하면서 같은 tick의 자격 있는 읽기를 버린다. 술어가 하나면 그
상태가 표현 불가능해진다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `r.NewlyListed` | 3-상태(`unknown`/`yes`/`no`) | 소스 어댑터의 직전 읽기 비교 | `unknown`이면 false — zero value가 `unknown`이다 |
| `r.RankRequested` | `> 0`이면 기록됨, `0`은 미기록 | 어댑터의 `count`/`size` | `0`이면 false. 음수는 `Observation.validate`가 경계에서 이미 거부한다(I14) |

불변식: 이 함수는 `Rank`·`RankTotal`을 보지 않는다. 위치가 있는지는 호출자가 이미 확인했고
(`Collect` B16, `recordFirsts` B9), 여기서 묻는 것은 **그 위치가 무엇을 뜻하는지 말할 수
있는가**다. 절단 여부도 보지 않는다 — 절단된 읽기는 두 사실을 다 갖고 있으므로 자격이
있고, 그것이 만드는 거부(`READING_TRUNCATED`)는 운영자가 쫓을 수 있는 진단이다(I7).

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | 분기 없음 — 두 절의 논리곱 하나를 반환 | 없음(순수) | `bool` | `TestAQualifiedReadingTakesTheFirstSightingFromAPanelEarlierUnqualifiedOne`(둘 다 참) · `TestWhenNoReadingInTheTickCanQualifyThePositionIsHeld`(첫 절 거짓) · `TestAReadingThatNeverRecordedItsRequestIsHeldToo`(둘째 절 거짓) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `r.NewlyListed.Known()` | 3-상태에서 "측정됨"을 읽는다 — `unknown`을 false로 접지 않기 위한 유일한 접근자 | 순수, 오류 없음 | ast.json calls |

## State mutations and fallbacks

- 없음. 인자를 값으로 받는 순수 함수이고 아무것도 쓰지 않는다.
- fallback 없음. 자격이 없으면 호출자가 **쓰지 않고 센다**(`FirstRanksHeld`) — 값을
  지어내는 자리가 없다.

## Safety conclusion

- Safe edit boundary: 신규 leaf 함수. 기존 동작 없음. 이 함수가 대체한 것은
  `recordFirsts`에 인라인되어 있던 같은 두 절이고, 그 절의 의미는 바뀌지 않았다.
- High-risk impact: **yes 인접** — write-once 칼럼의 유일한 생산 writer를 문지기한다.
  두 절 중 하나를 지우면 자격 없는 위치가 저장되고 그 후보의 `seen_late`는 영구 미측정이
  된다. 반대로 절을 더하면(예: 절단까지 요구) 저장이 더 줄어 노출은 늘지 않는다 — 이
  함수의 안전 방향은 **더 엄격한 쪽**이다.
