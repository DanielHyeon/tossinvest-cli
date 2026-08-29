# Function Logic Map: `TestPerformanceResourceProjectsExactMarketCampaignAttributionWithoutInventedNumbers`

- Source: `internal/httpapi/performance_attribution_test.go` (lines 11–35)
- AST evidence: `ast.json` (`source_sha256: 1e208590b42b6840821c28479de2d67db2447cbae5a62b86944beb3a39674f0b`)
- Risk scan: `risk-pattern-report.md`
- 위험 등급: **Normal**

## What it does

성과 귀속 리소스가 측정되지 않은 값을 지어내지 않는지 고정하는 테스트.
개정 4의 `gofmt`가 파일 끝 빈 줄 하나를 지웠을 뿐, 단언은 바뀌지 않았다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `performance.NewUnavailableAttribution` | 측정 불가 귀속 행 | 테스트 픽스처 | — |
| `PerformanceFrom` | 도메인 뷰 → HTTP 리소스 | 운영 코드 | 투영이 어긋나면 `t.Fatalf` |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Required test |
|---|---|---|---|
| B1 | (17) `if` — if len(resource.Attributions) != 1 { | 본문 참조 | 아래 Branch Test Map |
| B2 | (21) `if` — if got.Market != "US" \|\| got.CampaignID != "campaign-1" \|\| got.LegID != "leg-1" \|\| | 본문 참조 | 아래 Branch Test Map |
| B3 | (27) `if` — if err != nil { | 본문 참조 | 아래 Branch Test Map |
| B4 | (30) `range` — for _, want := range []string{`"attributions"`, `"campaignId":"campaign-1"`, `"missingMeasurements":["fees","fx"]`, `"netPnl":{"status":"not_measured","value":""`} { | 본문 참조 | 아래 Branch Test Map |
| B5 | (31) `if` — if !strings.Contains(string(body), want) { | 본문 참조 | 아래 Branch Test Map |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `PerformanceFrom` | 투영 대상 | — | AST `calls` |
| `json.Marshal` | 직렬화 표면 고정 | 오류면 `t.Fatal` | AST `calls` L26 |

## State mutations and fallbacks

- 없음. 테스트.

## Safety conclusion

- Safe edit boundary: 단언 문자열. 상태·형식이 아니라 공백만 바뀌었다.
- High-risk impact: no — 테스트 파일이고 §0.8 표면(계좌·시크릿)을 만지지 않는다.
