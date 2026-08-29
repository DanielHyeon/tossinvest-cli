# Function Logic Map: `writeCandidateTable`

- Source: `cmd/tossctl/candidate.go`
- Change: `fix-chase-veto-measurement`
- AST evidence: `ast.json` (revision `current`, L805–938, 분기 29개)
- Risk scan: `risk-pattern-report.md`

리포트를 터미널 표로 렌더한다. 이 change가 **네 블록**을 더했다.

- 소스별 `reading … N requested, M arrived (whole|short)` (B9·B10)
- `held   N position(s) not stored …` (B11) — 그리고 **왜** 그것이 안전한 방향인지를 같은
  줄에 적는다(write-once)
- `ALARM …` 줄들 (B16)
- `first sightings by source` 블록 (B17·B18·B19)

두 숫자를 비교로 **대체하지 않는다**. 100 요청 99 도착과 100 요청 3 도착은 같은
`Truncation`이고 전혀 다른 대화다 — `TestTheScanReportSaysWhatEachSourceAskedForAndWhatArrived`가
`truncated: true` 형태를 금지한다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `report` | `buildCandidateReport`의 출력 | 같은 파일 | 빈 블록은 줄이 없다 |
| `res` | 원본 `CycleResult` | `Cycle` | `Rejected` 상세만 여기서 읽는다 |
| `report.Readings` / `report.Sightings` / `report.Veto.Alarms` | 신규 블록 | 리포트 | 비면 렌더하지 않는다 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `res.Halted` | 중단 안내 | — | `candidate_test.go` |
| B2 | `report.Quiet` | 조용한 turn 요약 | — | 동상 |
| B3 | else | 보통 요약 | — | 동상 |
| B4 | `for _, m := range report.Sources.Missing` | 실패한 소스 줄 | — | 동상 |
| B5 | `m.RateLimited` | 429 표시 | — | 동상 |
| B6 | `len(report.Sources.NotDue) > 0` | not-due 줄 | — | 동상 |
| B7 | `for _, b := range report.Backoff` | backoff 줄 | — | 동상 |
| B8 | `report.EngineYield` | 엔진 양보 줄 | — | 동상 |
| B9 | `for _, rd := range report.Readings` **(신규)** | 소스별 요청·도착 줄 | — | `TestTheScanReportSaysWhatEachSourceAskedForAndWhatArrived` |
| B10 | `rd.Whole` **(신규)** | `whole` vs `short` 라벨 | — | 동상(두 케이스 모두) |
| B11 | `report.Recorded.FirstRanksHeld > 0` **(신규)** | 보류 줄 + 이유 | — | `TestTheHeldCountIsRenderedAndSaysWhichCommandCanNeverQualifyAPosition`(양·음 대조) |
| B12 | `report.Recorded.Rejected > 0` | 거부 요약 | — | `candidate_test.go` |
| B13 | `for _, r := range res.Scan.Rejected` | 거부 상세 | — | 동상 |
| B14 | `for _, code := range candidate.VetoCodes` | code별 줄 | — | 동상 |
| B15 | `for _, line := range sortedCounts(report.Veto.Reasons)` | 사유 census | — | 동상 |
| B16 | `for _, line := range report.Veto.Alarms` **(신규)** | 경보 줄 | — | `TestTheScanOutputSaysSoWhenTheTallyContradictsItself` |
| B17 | `len(report.Sightings) > 0` **(신규)** | 블록 헤더 | — | `TestTheScanReportAttributesTheRefusalsToASource` |
| B18 | `for _, s := range report.Sightings` **(신규)** | 소스별 `M of N measured` | — | 동상 |
| B19 | `for _, line := range sortedCounts(s.NotMeasured)` **(신규)** | 소스별 거부 사유 | — | 동상 |
| B20 | `for _, line := range sortedCounts(report.ShadowAcceleration.NotComputed)` | 가속 미계산 | — | `candidate_test.go` |
| B21 | `for _, code := range {seen_late, extended}` | 밴드 블록 | — | `TestAScanReportsTheShadowRecordForEveryCodeThatHasOne` |
| B22 | `!ok` — 밴드가 없는 code | skip | — | 동상 |
| B23 | `for _, line := range sortedCounts(band.NotMeasured)` | 밴드별 미측정 | — | 동상 |
| B24 | 보존 switch | — | — | `candidate_test.go` |
| B25 | `report.Retention.Why != ""` | 미측정 사유 | — | 동상 |
| B26 | `report.Retention.Busy` | busy 표시 | — | 동상 |
| B27 | default | 보통 보존 줄 | — | 동상 |
| B28 | `report.Space.Checked` | 여유 공간 줄 | — | 동상 |
| B29 | else | 미측정 공간 줄 | — | 동상 |

B11의 음성 대조군이 이 change에서 중요하다 — 아무것도 보류하지 않은 스캔에 그 줄이 나오면
"보류"가 상태가 아니라 장식이 된다. `TestTheHeldCountIsRenderedAndSaysWhichCommandCanNeverQualifyAPosition`이
양쪽을 모두 단언한다.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `fmt.Fprintf(w, ...)` | 줄마다 | 쓰기 실패는 무시(기존 동작) | ast.json calls |
| `sortedCounts(...)` | 결정적 census 순서 | 순수 | ast.json calls |
| `orderedCounts(...)` | 선언 순서 밴드 칸 | 순수 | ast.json calls |

## State mutations and fallbacks

- `w`에 쓰는 것뿐. 저장소·계좌 무접촉.
- 판정을 하지 않는다 — `TestNeitherSurfaceDecidesWhetherTheTallyIsConsistent`가 이 파일이 tally count 둘을 서로 비교하지 못하게 AST로 막는다.
- fallback 없음.

## Safety conclusion

- Safe edit boundary: 블록 4개(대략 25줄) 가산. 기존 25개 분기 무변경.
- High-risk impact: no (렌더). 재는 성질은 High-risk 인접 — 경보 줄이 사라지면 미측정이 통과로 세어지는 상태가 화면에서 평범해 보인다.
