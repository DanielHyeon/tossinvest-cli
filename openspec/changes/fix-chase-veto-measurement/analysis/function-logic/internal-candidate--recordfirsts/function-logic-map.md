# Function Logic Map: `recordFirsts`

- Source: `internal/candidate/scan.go`
- Change: `fix-chase-veto-measurement`
- AST evidence: `ast.json` (revision `current`, L528–623, 분기 13개)
- Risk scan: `risk-pattern-report.md`

승격된 후보 중 아직 stored first가 없는 것에 대해 첫 가격과 첫 순위를 쓴다. 이 change가
**보류 분기 하나**(B10)와 `NoteFirstRank`의 구조체 인자를 더했다.

보류가 이 change의 핵심 수리다. `first_rank`는 **한 번만** 쓰이므로 그것을 저장하는 것은
후보의 남은 생명 전체에 대한 결정이다. 직전 읽기를 갖고 있지 않은 소스의 읽기 — 세션 첫
tick의 모든 읽기, 그리고 `tossctl candidate scan`이 취하는 **모든** 읽기(패널을 만들고 한
번 읽고 종료하므로) — 는 두 자격 칼럼이 전부 NULL이다. 저장하면 그 생명의 `seen_late`가
영구 미측정이 되고 나중에 채울 수도 없다(사실은 사라진 읽기의 것이다).

저장하지 않으면 그 tick만 `NO_FIRST_RANK`이고 다음 자격 있는 읽기가 최초 관측이 된다 —
±TTL 동일성 창(10분) 안이고 공식 간격은 15초다. 하나는 회복 가능하고 하나는 아니므로
스캔은 회복 가능한 쪽을 택한다.

**절단된 읽기는 저장한다.** 자격 없는 위치가 아니다 — 두 사실 모두 기록되었고, 그것이
만드는 거부는 이름이 있으며, 그 후보를 처음 봤을 때 어디 있었는지에 대한 정직한 기록이다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `promoted` | 이번 pass가 승격한 심볼 | `Collect` | 비면 no-op |
| `summaries` | **승격 이후에** 읽는다 | `store.Summaries` | 만료로 초기화된 생명이 '이미 기록됨'으로 읽히지 않게 |
| `ranked[symbol]` | 이번 tick에서 **자격이 가장 높은** 순위 행(동률이면 패널 순서) | `Collect`의 `firstRanked` | 없으면 skip |
| `o.Reported.NewlyListed` / `RankRequested` | 자격 두 사실 | 소스 어댑터 | 둘 중 하나라도 없으면 **보류** |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `len(promoted) == 0` | 없음 | `nil` | `TestAScanDoesNotCoolASymbolItDidNotLookFor`(승격 없는 pass) |
| B2 | `Summaries` 오류 | 없음 | 오류 | 커버 없음 — I/O |
| B3 | `for _, s := range summaries` | `needPrice`/`needRank` 구성 | — | `scan_test.go` 전반 |
| B4 | `s.Market != market` | skip | — | **커버 없음** — 두 시장을 한 저장소에 넣고 한쪽만 스캔하는 fixture가 없다 |
| B5 | `for _, symbol := range promoted` | 심볼별 쓰기 | — | `scan_test.go` 전반 |
| B6 | `priced[symbol]` 있고 `needPrice[symbol]` | `NoteFirstPrice` | — | `TestOfficialSourcesAloneProduceCandidates` 등 |
| B7 | `NoteFirstPrice` 오류 | `Rejected` | 계속 | **커버 없음** |
| B8 | else — 성공 | `result.FirstPrices++` | — | `TestAScanReportsTheShadowRecordForEveryCodeThatHasOne` |
| B9 | `!ok || !needRank[symbol]` | skip | continue | `TestAPositionStoredBeforeTheFactsExistedIsNotFilledInByALaterScan`(이미 있음) |
| B10 | `!qualifiesFirstRank(o.Reported)` **(신규)** | `result.FirstRanksHeld++`, **쓰지 않는다** | continue | `TestASessionStartDoesNotStampThePanelAsSeenLate` · `TestWhenNoReadingInTheTickCanQualifyThePositionIsHeld` · `TestAReadingThatNeverRecordedItsRequestIsHeldToo` · `TestTheHeldCountIsRenderedAndSaysWhichCommandCanNeverQualifyAPosition` |
| B11 | `NoteFirstRank` 결과 switch | — | — | `TestARePromotionAfterExpiryIsQualifiedByTheReadingThatSawTheSymbolReturn` |
| B12 | `err != nil` | `Rejected` | 계속 | **커버 없음** |
| B13 | `stored.Recorded()` | `result.FirstRanks++` | — | `TestAScanReportsTheShadowRecordForEveryCodeThatHasOne` |

B11의 세 번째 경우(오류도 아니고 `Recorded()`도 아닌 것)는 **의도적으로 아무것도 하지
않는다** — 동일성 창 밖 읽기가 그것이고, 순위를 싣지 않는 소스가 처음 올린 후보의 일상
상태다. 함수 끝의 주석이 그 침묵을 설명한다.

**B10의 조건이 2026-07-28에 술어로 바뀌었다** (pre-gate 리뷰 P1-3, issues.md I17). 같은 두
절이 `Collect`의 `firstRanked` 선택에도 필요해졌고, 두 자리가 서로 다른 규칙을 갖는 것이
바로 그 결함이었다 — `Collect`가 채택한 읽기를 여기서 보류하면 같은 tick의 자격 있는
읽기가 버려진다. 이제 `qualifiesFirstRank` 하나를 둘이 공유하므로 어긋날 수 없고,
이 분기에 도달한다는 것은 **패널의 어떤 소스도** 그 심볼을 자격 부여하지 못했다는 뜻이다.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `store.Summaries(ctx, at)` | 아직 없는 것의 집합 | 오류는 pass 중단 | ast.json calls |
| `store.NoteFirstPrice(...)` | 첫 가격 | 실패는 `Rejected` | ast.json calls |
| `store.NoteFirstRank(ctx, market, symbol, FirstRank{...})` | 위치 + 두 자격을 **한 문장에** | 실패는 `Rejected`, 창 밖은 no-op | ast.json calls |

## State mutations and fallbacks

- `candidates`의 baseline 3칼럼과 first_rank 6칼럼(자격 2개 포함).
- `result.FirstPrices` / `result.FirstRanks` / **`result.FirstRanksHeld`**(신규) 카운터.
- fallback 없음. 보류는 값을 지어내는 것이 아니라 **쓰지 않는 것**이다.

## Safety conclusion

- Safe edit boundary: 보류 분기 4줄 + `NoteFirstRank` 호출을 구조체 인자로. 기존 needPrice/needRank·창 규칙 무변경.
- High-risk impact: **yes 인접** — write-once 칼럼의 유일한 생산 writer다. 이 change의 방향은 **덜 쓰는 쪽**이고, 덜 쓰는 것의 비용은 한 tick의 `NO_FIRST_RANK`(회복 가능)이며 더 쓰는 것의 비용은 그 생명 전체의 영구 미측정(회복 불가)이다.
