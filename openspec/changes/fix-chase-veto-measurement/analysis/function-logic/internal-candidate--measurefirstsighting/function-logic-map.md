# Function Logic Map: `MeasureFirstSighting`

- Source: `internal/candidate/veto.go`
- Change: `fix-chase-veto-measurement`
- AST evidence: `ast.json` (revision `current`, L669–753, 분기 9개)
- Risk scan: `risk-pattern-report.md`

후보를 처음 봤을 때 목록 어디에 있었는지를 보고한다. 이 change가 **case 두 개**를 더했다 —
B7 `NEW_ENTRANT_UNKNOWN`(design D3)과 B8 `REQUEST_UNRECORDED`(리뷰 F4) — 그리고 B9
`READING_TRUNCATED`(design D4)도 이 change의 것이다.

세 거부의 공통점: **숫자를 지어내지 않는다.** 하나는 소스에게 직전 읽기를 갖고 있었느냐고
묻고, 둘은 소스가 스스로 선언한 두 수를 비교한다. 대안이었던 "한 스캔이 패널의 N%를
승격했으면 일괄로 본다"는 그 N이 또 출처 없는 숫자이고, 이 change가 치료하는 병을 치료
과정에서 재발시킨다 — `TestTheRefusalCountsNoRatioOfThePanel`이 AST로 그것을 막는다.

자격 사실이 **저장 칼럼**에서 오는 것도 이 change다. 전에는 관측 슬라이스에서 (source,
instant, rank, total) 일치로 찾았는데, `Assess`가 최근 10분만 읽으므로 그 행은 오래된 후보
전부에 대해 슬라이스 **밖**이었다 — 다섯 층에 선언된 사실이 한 번도 기록되지 않은 두 번째
이유다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `c.FirstSeenAt` | non-zero | `candidates` 행 | zero면 `NO_CANDIDATE` |
| `first` | `Recorded()`면 위치 채택 후보 | `candidates.first_rank*` 6칼럼 | 아니면 `unstoredFirstSighting` |
| `first.NewlyListed` | 3-상태 | `first_rank_newly_listed` | `unknown`이면 `NEW_ENTRANT_UNKNOWN` |
| `first.Truncation()` | 3-상태 파생 | `first_rank_requested` + `first_rank_total` | `unknown`→`REQUEST_UNRECORDED`, `yes`→`READING_TRUNCATED` |
| `observations` | 한 후보의 것이어야 | `Assess`의 창 | 섞이면 `MIXED_CANDIDATES` |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `len(observations) > 0 && !oneCandidate(observations)` | 없음 | `MIXED_CANDIDATES` | `TestASightingWithNoRankIsUnmeasured`(같은 파일 L753의 mixed 단언) |
| B2 | `c.FirstSeenAt.IsZero()` | 없음 | `NO_CANDIDATE` | `TestSeenLateWithoutACandidateCannotKnowWhenWeFirstSaw` |
| B3 | `!first.Recorded()` | 없음 | `unstoredFirstSighting`으로 위임 | `TestTheUnqualifiedReasonIsNotOneOfTheStoresOwnGaps` · `TestASightingWithNoStoredRankIsNotMeasuredFromARow` |
| B4 | 자격 switch | `out`에 위치·instant·source·두 자격 사실을 먼저 싣는다 | — | 아래 다섯 case |
| B5 | `first.At.IsZero()` | 없음 | `FIRST_RANK_UNDATED` | `TestTheFirstSightingBoundaryIsTheStalenessTTL`(L685) |
| B6 | `!nearFirstSighting(first.At, c.FirstSeenAt)` | 없음 | `FIRST_RANK_NOT_FIRST` | `TestTheFirstSightingBoundaryIsTheStalenessTTL` · `TestAPreviousLifesReadingsAreNotThisCandidatesFirstSighting` |
| B7 | `!out.NewlyListed.Known()` **(신규 — D3)** | 없음 | `NEW_ENTRANT_UNKNOWN` | `TestAPositionFromASourcesFirstReadingCannotAnswerSeenLate` · `TestAPositionStoredBeforeTheFactsExistedIsNotFilledInByALaterScan` |
| B8 | `!out.Truncation.Known()` **(신규 — F4)** | 없음 | `REQUEST_UNRECORDED` | `TestAPositionWithNoRecordedRequestIsRefusedUnderItsOwnReason` · `TestTheOneRowReadingWithNoRecordedRequestIsTheCaseThatMattered` |
| B9 | `out.Truncation.Yes()` **(신규 — D4)** | 없음 | `READING_TRUNCATED` | `TestATruncatedReadingsPositionIsNotAPercentile` · `TestTheOneRowReadingIsCaughtByTheSameRefusal` |

거부된 경우에도 **읽은 것은 보고한다** — `out.Rank`, `out.RankTotal`, `out.At`, `out.Source`,
두 자격 사실이 switch **앞에서** 채워진다. 화면이 "왜 못 쟀는지"와 "무엇을 읽었는지"를
함께 보여줄 수 있어야 하고, 그것이 D6("읽은 값과 결론은 다른 필드")의 이 자리 적용이다.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `oneCandidate(observations)` | 슬라이스 정합성 | 순수 | ast.json calls |
| `first.Recorded()` / `first.Truncation()` | 저장 위치의 두 술어 | 순수 | ast.json calls |
| `nearFirstSighting(first.At, c.FirstSeenAt)` | ±TTL 동일성 창 | 순수 | ast.json calls |
| `percentileOf(out.Rank, out.RankTotal)` / `formatDecimal` | 백분위 — **산식 무변경** | 순수 | ast.json calls |
| `unstoredFirstSighting(out, c, observations)` | 저장 위치가 없을 때의 네 가지 이유 | 순수 | ast.json calls |

## State mutations and fallbacks

- 없음 — `Sighting` 하나를 만든다.
- fallback **없음**, 그리고 그것이 요점이다. 세 새 거부는 전부 값을 지어내지 않고 거부한다. `AssessSeenLate`·`thresholdReason` 쪽에도 fallback을 넣지 않는다(design D10).

## Safety conclusion

- Safe edit boundary: switch case 2개(B7·B8) + B9 가산, `out.NewlyListed`/`out.Truncation` 대입 1줄. 백분위 산식·창 규칙·기존 두 case 무변경.
- High-risk impact: **yes** — `seen_late`의 유일한 측정이고 chase veto의 3분의 1이다. 방향은 **더 자주 미측정**이며, 미측정은 통과가 아니므로(D10) 노출을 늘리지 않는다. `Chase.Passed()`는 임계가 없어 여전히 도달 불가이고 이 change가 그것을 바꾸지 않는다.
