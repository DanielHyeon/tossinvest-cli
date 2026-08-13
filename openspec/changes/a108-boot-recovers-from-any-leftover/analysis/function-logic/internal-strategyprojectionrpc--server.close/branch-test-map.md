# Branch Test Map: `Server.Close`

- Source: `internal/strategyprojectionrpc/transport_unix.go` (232-248)
- 이 change는 이 함수를 **바꾸지 않았다.** 잔재의 생산자로서 회수 설계의 입력이다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | nil 서버를 닫는다 | 없음(무변경·기존 무테스트) | no | no |
| B2 | descriptor → socket → controlDir 순으로 지운다 | `TestCloseToleratesLeftoverAlreadyRemoved` | no (편집 전 GREEN) | yes |
| B3 | 이미 없는 경로의 제거는 실패가 아니다 | `TestCloseToleratesLeftoverAlreadyRemoved` | no (편집 전 GREEN) | yes |

## B3이 재는 것

`TestCloseToleratesLeftoverAlreadyRemoved`는 다른 부팅의 회수가 먼저 지나간 상황을
만든다: 세 경로를 밖에서 지운 뒤 `Close`가 `nil`을 반환하는지, 그리고 그 뒤의 기동이
빈 자리를 보는지까지 본다. `ErrNotExist` 용인이 없으면 이 테스트가 죽고, 그것이
design D2 "경합은 양성이다"의 코드 근거다.

측정 중 발견한 별건(이 change 밖): `Shutdown`이 `Serve`의 listener 등록을 앞지르면
listener 정리가 뒤로 밀리고 그 늦은 unlink가 **다음 기동의 socket**을 지운다. 테스트는
`Close` 전에 Dial+Read로 수락을 확인해 그 경합을 결정적으로 배제했다(`-count=300` 통과).
