# Branch Test Map: `signalsNewlyListedText`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 세 상태가 서로 다른 문자열로 렌더된다 | `TestTheNewEntrantMarkerRendersAllThreeStatesDistinctly`(세 행이 서로 다름까지 단언) | yes | yes |
| B2 | `yes` | 동상(005930) | yes (이 표식은 이 change 이전에 뜰 수 없었다) | yes |
| B3 | `no` | 동상(000660 — 미상 문구와 섞이지 않는 것까지) | yes | yes |
| B4 | 미상 | 동상(035720) | yes | yes |
