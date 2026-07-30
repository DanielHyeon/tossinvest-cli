# Branch Test Map: `Sighting.PercentileExceeds`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 미측정 sighting은 비교하지 않는다 | `TestATruncatedReadingsPositionIsNotAPercentile` · `TestTheOneRowReadingWithNoRecordedRequestIsTheCaseThatMattered`(둘 다 `AssessSeenLate`가 clear도 dangerous도 아님을 요구) · `TestASightingWithNoRankIsUnmeasured` | — (동작 무변경) | yes |
| B2 | 임계가 없거나 위치가 불가능하면 미측정 | `TestAnAbsentThresholdIsNotAPassedVeto` · `TestAnUnreadableThresholdIsNotAPassedVeto` · `TestAMeasuredInputWithUnusableFiguresIsStillNotAPass` | — (동작 무변경) | yes |
| (꼬리) | 자격 있는 위치는 임계와 비교된다 | `TestASessionStartDoesNotStampThePanelAsSeenLate`(97%가 80 임계를 넘어 dangerous) · `TestARankedRowFromTheGapBetweenTwoLivesIsNotThisLifesFirstSighting` | — (동작 무변경) | yes |
