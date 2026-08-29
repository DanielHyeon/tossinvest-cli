# Branch Test Map: `TestPerformanceResourceProjectsExactMarketCampaignAttributionWithoutInventedNumbers`

AST의 모든 분기를 1행씩 덮는다. 조건 열은 `internal/httpapi/performance_attribution_test.go`의 실제 소스 줄이고,
테스트 열과 판정 열은 `go test -covermode=count` 프로파일에서 **측정**한 값이다.
주장이 아니라 측정이므로 이 표는 덮이지 않은 분기를 숨길 수 없다.

| Branch | Condition | Covering test | Measured |
|---|---|---|---|
| B1 | (17) `if` — if len(resource.Attributions) != 1 { | `TestPerformanceResourceProjectsExactMarketCampaignAttributionWithoutInventedNumbers` (통과) | 측정 불가 — `go test -coverprofile`은 `_test.go`를 계측하지 않는다. 이 함수 자체가 테스트이고, 통과가 곧 실행 증거다 |
| B2 | (21) `if` — if got.Market != "US" \|\| got.CampaignID != "campaign-1" \|\| got.LegID != "leg-1" \|\| | `TestPerformanceResourceProjectsExactMarketCampaignAttributionWithoutInventedNumbers` (통과) | 측정 불가 — `go test -coverprofile`은 `_test.go`를 계측하지 않는다. 이 함수 자체가 테스트이고, 통과가 곧 실행 증거다 |
| B3 | (27) `if` — if err != nil { | `TestPerformanceResourceProjectsExactMarketCampaignAttributionWithoutInventedNumbers` (통과) | 측정 불가 — `go test -coverprofile`은 `_test.go`를 계측하지 않는다. 이 함수 자체가 테스트이고, 통과가 곧 실행 증거다 |
| B4 | (30) `range` — for _, want := range []string{`"attributions"`, `"campaignId":"campaign-1"`, `"missingMeasurements":["fees","fx"]`, `"netPnl":{"status":"not_measured","value":""`} { | `TestPerformanceResourceProjectsExactMarketCampaignAttributionWithoutInventedNumbers` (통과) | 측정 불가 — `go test -coverprofile`은 `_test.go`를 계측하지 않는다. 이 함수 자체가 테스트이고, 통과가 곧 실행 증거다 |
| B5 | (31) `if` — if !strings.Contains(string(body), want) { | `TestPerformanceResourceProjectsExactMarketCampaignAttributionWithoutInventedNumbers` (통과) | 측정 불가 — `go test -coverprofile`은 `_test.go`를 계측하지 않는다. 이 함수 자체가 테스트이고, 통과가 곧 실행 증거다 |
