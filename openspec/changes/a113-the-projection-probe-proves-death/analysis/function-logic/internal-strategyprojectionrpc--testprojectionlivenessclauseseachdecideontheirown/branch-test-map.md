# Branch Test Map: `TestProjectionLivenessClausesEachDecideOnTheirOwn`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 6행 순회 | `TestProjectionLivenessClausesEachDecideOnTheirOwn` | yes — base 코드에서 a113 두 행 FAIL | yes |
| B2 | 산 0400 행 root skip | `TestProjectionLivenessClausesEachDecideOnTheirOwn` (비root 실행이라 skip 안 됨) | no | yes |
| B3 | fixture chmod 실패 | `TestProjectionLivenessClausesEachDecideOnTheirOwn` (fixture 가드) | no | yes |
| B4 | 죽은 0500 행 root skip | `TestProjectionLivenessClausesEachDecideOnTheirOwn` | no | yes |
| B5 | fixture 쓰기 실패 | `TestProjectionLivenessClausesEachDecideOnTheirOwn` (fixture 가드) | no | yes |
| B6 | 판정 불일치 | `TestProjectionLivenessClausesEachDecideOnTheirOwn` — 뮤테이션 N1·N4 사망 | yes | yes |
