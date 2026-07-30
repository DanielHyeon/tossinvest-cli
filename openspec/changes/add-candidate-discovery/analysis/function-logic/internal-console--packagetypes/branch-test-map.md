# Branch Test Map: `packageTypes`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 전 파일 순회 | `checkCapability`의 모든 해석 경로 | — | yes |
| B2 | TypeSpec 수집 | 같은 위 + 별칭 사슬 변이 | yes | yes |
