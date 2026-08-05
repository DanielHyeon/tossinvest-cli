# Branch Test Map: `TestTheOverviewReloadsAtTheCacheTTL`

base 리비전의 분기와, a080 이후 그 판정이 어디서 유지되는지를 함께 적는다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `overviewPage.RefreshSeconds()`가 기대 주기가 아니다 | 갱신 후 동일 함수 | yes (변이 6.1) | yes |
| B2 | `/dashboard`가 meta refresh를 잃었다 | 갱신 후 동일 함수 | no | yes |
| B3 | 렌더된 주기 값이 기대와 다르다 | 갱신 후 동일 함수 | yes (변이 6.1) | yes |
| B4 (신규) | 새 상수가 `holdingsTTL`과 같아 분리를 구분할 수 없다 | 갱신 후 동일 함수 | yes (변이 6.1) | yes |

세 분기 전부 갱신 후에도 같은 함수가 판정한다. 잃은 판정이 없고 B4가 늘었다.
변이 6.1(`lineRefreshInterval`을 30초로 되돌림)에서 B1·B3·B4가 동시에 RED가 되는
것을 확인하고 되돌렸다.
