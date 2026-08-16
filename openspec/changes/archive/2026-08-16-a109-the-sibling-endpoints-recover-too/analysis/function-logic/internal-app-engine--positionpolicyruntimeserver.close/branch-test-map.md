# Branch Test Map: `PositionPolicyRuntimeServer.Close`

구현 후 재최신화(§1.7 · §2b.3 G9). 지금 HEAD 의 AST 분기는 **1개**다 — listener 해체와
경로 제거(옛 B2~B5)를 형제(`AlertControlServer.Close`)와 공유하는
`closePrivateEndpointFiles`로 옮겼기 때문이다. 동작은 불변이고, 그 절들의 커버리지는
그 기계가 진다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | nil 수신자 Close | `TestTheSiblingEndpointClosesAreSafeOnANilServer`(cmd/tossctl — 강등 경로의 계약) | no | yes(§2-fix F7) |

## 옮긴 절의 커버리지 (`closePrivateEndpointFiles`)

| 옮긴 절 | Test | 비고 |
|---|---|---|
| listener 를 Close 가 **직접** 닫는다 | **결정적 커버 없음** — late unlink 는 경합이고 a108 도 300라운드 중 3회로 관측했다. 뮤테이션 M26 생존이 그 사실의 측정이다(mutation-ledger-t1.md §B) | 기계가 한 벌이 되면서 두 endpoint 에 같은 측정이 적용된다 |
| `net.ErrClosed` 관용 | `TestTheAlertSocketIsPrivateToThisUser` · `TestRuntimeOnlyUnixControlCrossNamespaceTopologyAndRouteAbsence` 가 정상 Close 의 오류 없음을 확인한다 | |
| descriptor·socket·controlDir 제거 | 같은 두 테스트가 Close 후 경로 부재를 확인한다 | |
| `os.ErrNotExist` 관용 | `…StartsOverALeftover` 의 `t.Cleanup` 이중 Close 가 같은 관용에 의존한다 | |
