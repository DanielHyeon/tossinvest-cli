# Branch Test Map: `TestTheTiersTranscribeStockOS`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 티어 수가 코퍼스 합과 다르다 | 자기 자신 | yes (4 vs 5) | yes |
| B2 | 등록 티어를 순회한다 | 자기 자신 | yes | yes |
| B3 | 라벨 없는 티어 | 자기 자신 | no (전 티어 라벨 보유) | yes |
| B4 | 전사 코퍼스 밖 ID | 자기 자신 | yes | yes |
| B5 | 실측 코퍼스로 위임 | 자기 자신 | no (RED 시점엔 미등록) | yes |
| B6 | 전사 값 드리프트 | 자기 자신 | no | yes |
| B7 | 실측 코퍼스 순회 | 자기 자신 | yes | yes |
| B8 | 실측 티어 미등록 | 자기 자신 | yes (RED) | yes |
| B9 | 실측 값 불일치 | 자기 자신 | no | yes |
