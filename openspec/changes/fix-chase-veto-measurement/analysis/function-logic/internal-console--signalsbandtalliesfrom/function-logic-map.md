# Function Logic Map: `signalsBandTalliesFrom`

- Source: `internal/console/signals.go`
- Change: `fix-chase-veto-measurement`
- AST evidence: `ast.json` (revision `base`, L729–753, 분기 4개)
- Risk scan: `risk-pattern-report.md`

**본문 무변경**이다. base 대비 byte 동일하고, 이 change가 위쪽(`signalsSightingMetric`,
`signalsNewlyListedText`, `signalsVetoTallyFrom`)을 편집하면서 diff hunk가 교차해 evidence가
요구되었다. `ast.json`은 base revision에서 뜬 것이다.

그림자 밴드를 D3의 code 순서로 렌더한다 — 맵이었다면 새로고침마다 페이지가 스스로 순서를
바꾼다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `in` | `map[VetoCode]BandTally` | seam | code가 없으면 그 블록은 없다 |
| `candidate.VetoCodes` | D3의 순서 | `internal/candidate` | 순회 순서의 정본 |
| `candidate.BandsFor(code)` | code별 밴드 목록 | `band.go` | 아무도 넘지 않은 밴드도 칸을 갖는다 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `for _, code := range candidate.VetoCodes` | `out` append | — | `internal/console/signals_test.go`의 밴드 블록 |
| B2 | `!ok` — 이 code에 밴드가 없다 | skip | continue | 동상 |
| B3 | `for _, band := range candidate.BandsFor(code)` | 밴드 칸 | — | 동상 |
| B4 | `for why, n := range tally.NotMeasured` | 사유 census | — | 동상 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `candidate.BandsFor(code)` | 밴드 목록 | 순수 | ast.json calls (base) |
| `sortedSignalsCounts(counts)` | census 정렬 | 순수 | ast.json calls (base) |

## State mutations and fallbacks

- 없음 — 새 슬라이스, base와 byte 동일.
- fallback 없음.

## Safety conclusion

- Safe edit boundary: 무변경 — 인접 편집만 존재한다.
- High-risk impact: no (렌더). 이 함수가 한 code를 빠뜨리면 '아무도 넘지 않은 밴드'와 '아무도 세지 않은 밴드'가 같은 화면이 된다 — 이 저장소가 한 번 지불한 실패다.
