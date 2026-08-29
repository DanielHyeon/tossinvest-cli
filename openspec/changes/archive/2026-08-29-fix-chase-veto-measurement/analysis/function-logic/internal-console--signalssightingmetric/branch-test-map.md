# Branch Test Map: `signalsSightingMetric`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 미측정 sighting은 사유를 렌더한다 | `TestARefusedSightingNamesWhyRatherThanShowingTheMarker`(`NEW_ENTRANT_UNKNOWN`) · `TestATruncatedReadingIsNamedOnTheScreenToo`(`READING_TRUNCATED`) | yes (전자는 표식과 사유가 섞이지 않는 것까지 단언) | yes |
| B2 | 백분위가 있으면 괄호로 붙는다 | `TestTheNewEntrantMarkerRendersAllThreeStatesDistinctly`(`12 / 100` + `88%p`) | yes | yes |
