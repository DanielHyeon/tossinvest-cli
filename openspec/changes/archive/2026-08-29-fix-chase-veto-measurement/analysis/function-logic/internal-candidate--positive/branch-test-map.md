# Branch Test Map: `positive`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 0은 NULL, 양수는 그대로 | `TestNoteFirstRankKeepsTheStoredPositionWhateverIsOfferedNext`(미기록 왕복) · `TestATruncatedReadingReachesTheVerdictAsTruncated`(양수 왕복) · `TestAFreshStoreCarriesTheSameFourConstraints`(CHECK) | yes | yes |

음수 입력이 여기서 NULL이 되어 숨는다는 사실 자체는
`TestANegativeRequestedCountIsRefusedByTheObservationBoundary`가
`truncationOf(-1, 100) == TruncationUnknown()`으로 명시한다 — 이 함수의 하류가 아니라
상류에서 막는 이유다.
