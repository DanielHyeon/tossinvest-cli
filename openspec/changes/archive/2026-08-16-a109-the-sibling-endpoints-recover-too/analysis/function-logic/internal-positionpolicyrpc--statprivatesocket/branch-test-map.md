# Branch Test Map: `statPrivateSocket`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 빈 경로 | 커버 없음 — 두 호출자 모두 구성된 경로를 넘긴다 | no | no |
| B2 | 디렉터리 위생 실패 | `TestTheAlertSocketIsPrivateToThisUser`가 성립 방향(0700 디렉터리)을 실측한다. 거부 방향의 직접 커버는 없다 | no | no |
| B3 | 경로 부재 → ErrNotExist | `TestAlertControlServerStartsOverALeftover`의 "빈 디렉터리" 케이스가 이 경로로 기동한다 | no | no |
| B4 | pre-chmod 0700 socket 잔재를 "not an exact 0600"으로 거부 | a109 §1.1 RED가 alert control 기동에서 이 거부를 관측한다. GREEN 이후 이 절은 기동 경로에서 더 이상 불리지 않는다(회수가 대신한다) | yes(§1.1) | yes(§1.4) |
