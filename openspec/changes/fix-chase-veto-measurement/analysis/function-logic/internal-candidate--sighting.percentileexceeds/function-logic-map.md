# Function Logic Map: `Sighting.PercentileExceeds`

- Source: `internal/candidate/veto.go`
- Change: `fix-chase-veto-measurement`
- AST evidence: `ast.json` (revision `current`, L599–615, 분기 2개)
- Risk scan: `risk-pattern-report.md`

**본문 무변경**이다. base 대비 함수 본문이 byte 동일하고, 이 change가 위쪽 `Sighting` 구조체에
두 필드(`NewlyListed`의 타입 변경, `Truncation` 신설)와 주석을 넣으면서 diff hunk가 교차해
evidence가 요구되었다.

`seen_late`가 쓰는 비교이고, 이 change가 **의도적으로 건드리지 않은** 곳이다.
`percentileOf`의 산식은 `metrics.go`의 순위 이동 백분위와 같아야 하고
(`TestTheSightingPercentileIsTheSameOneTheRankMoveUses`), 바꾸면 축적된 기록의 의미가
바뀐다(design D10).

산식의 비대칭 — 1위가 100에 닿지 못한다 — 은 이 change가 **다른 곳에서** 덮는다:
절단 거부(D4)가 얇은 목록 전체를 걷어낸다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `s.Measured` | true여야 비교가 가능 | `MeasureFirstSighting` | false면 `(false, false)` — 미측정 |
| `thresholdPct` | 평범한 소수 문자열 | `VetoThresholds.SeenLatePercentilePct` | 비어 있거나 파싱 불가면 미측정 |
| `s.Rank`/`s.RankTotal` | 둘 다 양수, `Rank <= RankTotal` | 저장 칼럼 | 위반은 미측정 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `!s.Measured` | 없음 | `(false, false)` | `TestATruncatedReadingsPositionIsNotAPercentile`(거부된 sighting을 `AssessSeenLate`에 넘긴다) · `TestASightingWithNoRankIsUnmeasured` |
| B2 | `why != "" || RankTotal <= 0 || Rank <= 0 || Rank > RankTotal` | 없음 | `(false, false)` | `TestAnAbsentThresholdIsNotAPassedVeto` · `TestAnUnreadableThresholdIsNotAPassedVeto` |

두 번째 반환이 D10의 3-상태다. false는 "이 후보를 재지 못했다"이지 "일찍 봤다"가 아니고,
그것을 떨어뜨린 호출자는 모든 후보에 대해 `seen_late`를 꺼 버린 것이다.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `parseFigure(thresholdPct)` | 임계 파싱 | 실패는 `why`로 나오고 미측정이 된다 | ast.json calls |
| `percentileOf(s.Rank, s.RankTotal)` | **렌더가 아니라 원 유리수**로 비교 | 순수 | ast.json calls |
| `(*big.Rat).Cmp` | 정확 비교 | 순수 | ast.json calls |

## State mutations and fallbacks

- 없음 — 순수 판정, **base와 byte 동일**.
- fallback 없음. 파싱 실패도 임계 부재도 미측정이다.

## Safety conclusion

- Safe edit boundary: 무변경 — 인접 구조체 필드 추가만 존재한다.
- High-risk impact: **yes** — chase veto 셋 중 하나의 비교다. 두 번째 반환을 잃으면 미측정이 통과로 접히고, 그것이 D10이 막는 한 줄짜리 편집이다.
