# Branch Test Map: `assessInto`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | due한 것이 없는 turn도 지난 결과를 유지한다 | `TestATurnWithNothingDueIsNotAMarketFailure` · `TestASourceThatWasNotAskedDoesNotVouchForTheCandidatesItRaised` | — (기존 동작) | yes |
| B2 | Assess 실패는 그대로 상승 | **커버 없음** | no | no |
| (신규 1줄) | `res.Sightings`가 채워져 두 표면이 같은 census를 본다 | `TestTheScanReportAttributesTheRefusalsToASource`(스캔 표면) · `TestTheSignalsPageAttributesTheRefusalsToTheSourceThatProducedThem`(/signals 표면) | yes | yes |

대입 자체를 mutation으로 확인했다: `res.Sightings = ...`를 `_ = ...`로 바꾸면
`TestATruncatedReadingReachesTheVerdictAsTruncated`가 `truncation_wiring_test.go:133`에서
"0 sources in the sighting reduction, want 1"로 실패한다(2026-07-28 실행).
두 표면 테스트는 `CycleResult.Sightings`/`SignalsMarket.Sightings`를 직접 채워 렌더만
확인하므로, 이 한 줄을 지키는 것은 `Cycle`을 지나는 그 배선 테스트 하나다.
