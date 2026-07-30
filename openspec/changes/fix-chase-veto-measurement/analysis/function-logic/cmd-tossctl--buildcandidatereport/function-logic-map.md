# Function Logic Map: `buildCandidateReport`

- Source: `cmd/tossctl/candidate.go`
- Change: `fix-chase-veto-measurement`
- AST evidence: `ast.json` (revision `current`, L644–785, 분기 13개)
- Risk scan: `risk-pattern-report.md`

`CycleResult`를 리포트 구조체로 옮긴다. 이 change가 **네 곳**을 더했다.

- `r.Recorded.FirstRanksHeld` — 스캔이 저장을 보류한 위치 수. `0 first ranks`만으로는
  "기록할 것이 없었다"와 구분되지 않는다.
- `r.Readings` (B4) — 소스별 요청·도착. 이 change가 열어 두는 질문의 직접 측정이다.
- `r.Sightings` (B5·B6) — 소스별 최초 관측 census.
- `r.Veto.Alarms` — tally 항등식 경보. `passedNote` **옆에** 붙고 대체하지 않는다.

경보가 note를 대체하지 않는 것이 §4.5다. note는 임계가 승인되지 않은 동안 통과 수가
구조적으로 0이라는 이야기이고, 경보는 tally가 자기 자신과 모순되는지에 대한 것이며 임계와
무관하게 참·거짓이다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `res.Scan.Readings` | 소스별 요청·도착 | `Collect` | 빈 맵이면 블록 없음 |
| `res.Sightings` | 소스별 census | `assessInto` | 빈 목록이면 블록 없음 |
| `res.Vetoes` | `VetoTally` | `TallyVerdicts` | `Anomalies()`가 경보의 유일한 판정 |
| `res.Scan.FirstRanksHeld` | 보류 수 | `recordFirsts` | 0이면 표에 줄이 없다 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `for _, id := range res.NotDue` | `r.Sources.NotDue` | — | `cmd/tossctl/candidate_test.go`의 not-due 렌더 |
| B2 | `for _, m := range res.Scan.Missing` | `r.Sources.Missing` | — | 동상 |
| B3 | `for _, b := range res.Backoff` | `r.Backoff` | — | 동상 |
| B4 | `for _, id := range sortedSourceIDs(res.Scan.Readings)` **(신규)** | `r.Readings` append | — | `TestTheJSONReportCarriesBothBlocks` |
| B5 | `for _, s := range res.Sightings` **(신규)** | `r.Sightings` append | — | `TestTheScanReportAttributesTheRefusalsToASource` |
| B6 | `for why, n := range s.NotMeasured` **(신규)** | 사유 맵을 문자열 키로 | — | 동상 |
| B7 | `res.Vetoes.Passed != 0` | `PassedNote`를 예상 밖 문구로 | — | `candidate_review_test.go`의 `structurally 0` 케이스 |
| B8 | `for code, n := range res.Vetoes.Raised` | `r.Veto.Raised` | — | `candidate_test.go` |
| B9 | `for code, n := range res.Vetoes.NotMeasured` | `r.Veto.NotMeasured` | — | 동상 |
| B10 | `for why, n := range res.Vetoes.Reasons` | `r.Veto.Reasons` | — | 동상 |
| B11 | `for why, n := range res.Crossings.NotComputed` | 가속 미계산 census | — | 동상 |
| B12 | `for code, tally := range res.Bands` | 밴드 블록 | — | `TestAScanReportsTheShadowRecordForEveryCodeThatHasOne` |
| B13 | `for why, n := range tally.NotMeasured` | 밴드별 미측정 census | — | 동상 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `sortedSourceIDs(res.Scan.Readings)` | 결정적 순서 | 순수 | ast.json calls |
| `facts.Truncation().No()` | `whole` 필드 — **`truncationOf`와 같은 규칙** | 순수 | ast.json calls |
| `candidateTallyAlarms(res.Vetoes)` | 경보 문장들 | 판정은 `Anomalies()` | ast.json calls |

## State mutations and fallbacks

- 없음 — 리포트 값 하나를 만든다. 저장소·계좌 무접촉.
- fallback 없음. 비어 있는 블록은 `omitempty`로 사라지고, 그것이 '측정하지 않았다'가 아니라 '이번 스캔에 항목이 없다'라는 것은 각 블록의 주석이 말한다.

## Safety conclusion

- Safe edit boundary: 필드 4묶음 + 루프 3개 가산. 기존 13개 필드와 note 로직 무변경.
- High-risk impact: no (리포트 조립). 재는 성질은 High-risk 인접 — `Alarms`가 조립되지 않으면 미측정이 통과로 세어지는 상태가 화면에서 평범해 보인다.
