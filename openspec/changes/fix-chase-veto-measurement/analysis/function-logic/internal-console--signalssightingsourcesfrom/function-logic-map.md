# Function Logic Map: `signalsSightingSourcesFrom`

- Source: `internal/console/signals.go`
- Change: `fix-chase-veto-measurement`
- AST evidence: `ast.json` (revision `current`, L834–847, 분기 2개)
- Risk scan: `risk-pattern-report.md`

**신규 함수**. 소스별 최초 관측 기록을 화면 값으로 옮긴다.

축약(reduction)은 `candidate.TallySightingSources`이고 여기서 반복하지 않는다 — veto tally를
반복하지 않는 것과 같은 이유다: 스캔 출력과 이 페이지는 **문구는 달라도 되고 건수는 달라서는
안 된다**.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `in` | `[]candidate.SourceSightings` | seam | 비면 nil — 템플릿이 블록을 그리지 않는다 |
| `s.NotMeasured` | 사유별 건수 | `TallySightingSources` | 빈 맵이면 '없음' |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `for _, s := range in` | `out` append | — | `TestTheSignalsPageAttributesTheRefusalsToTheSourceThatProducedThem` |
| B2 | `for why, n := range s.NotMeasured` | 사유 맵을 문자열 키로 | — | 동상 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `sortedSignalsCounts(counts)` | 큰 것부터 — veto 사유와 같은 순서 | 순수 | ast.json calls |

## State mutations and fallbacks

- 없음 — 새 슬라이스.
- fallback 없음. 아무 소스도 최초 관측을 갖지 않으면 nil이고, 템플릿의 `{{if .Sightings}}`가 블록을 통째로 생략한다 — **빈 표가 아니라 부재**다.

## Safety conclusion

- Safe edit boundary: 신규 함수. 기존 동작 없음.
- High-risk impact: no (렌더).
