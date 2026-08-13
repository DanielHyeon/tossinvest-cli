# Branch Test Map: `Server.Close`

- Source: `internal/strategyprojectionrpc/transport_unix.go` (358-386) — revision `current`
- 첫 라운드는 이 함수를 바꾸지 않았다. Fix 라운드가 listener 소유권을 옮겼다(D2-2).

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | nil 서버를 닫는다 | 없음(무변경·기존 무테스트) | no | no |
| B2 | listener를 Close가 직접 닫는다 | `TestCloseClosesItsOwnListenerWithoutServe` | no (Fix 라운드 신규 핀) | yes |
| B3 | 이미 닫힌 listener는 실패가 아니다 | `TestCloseClosesItsOwnListenerWithoutServe` | no | yes |
| B4 | descriptor → socket → controlDir 순으로 지운다 | `TestCloseToleratesLeftoverAlreadyRemoved` | no (편집 전 GREEN) | yes |
| B5 | 이미 없는 경로의 제거는 실패가 아니다 | `TestCloseToleratesLeftoverAlreadyRemoved` | no (편집 전 GREEN) | yes |

## B2가 재는 것 — 경합을 기다리지 않는 핀

`TestCloseClosesItsOwnListenerWithoutServe`는 `Serve`를 **한 번도 부르지 않은** 서버를
만든다. 그것이 곧 "`Shutdown`이 `Serve`의 등록을 앞지른" 창이다 — A1이 300라운드 중
3회로만 만든 상태를 결정적으로 만든 것이다. `Close`가 돌아온 뒤 두 번째
`listener.Close()`가 `net.ErrClosed`를 돌려주면 첫 번째를 `Close`가 이미 했다는 뜻이다.

후계자 보존은 별도 핀이 잰다: `TestLateListenerCloseCannotUnlinkSuccessorSocket`이 첫
주인의 listener를 손에 쥔 채 후계자를 세우고 **그 다음에** 닫아, 늦은 정리가 후계자의
socket을 건드리지 못하는 것을 inode 동일성으로 확인한다.
`TestPublishedListenerNeverUnlinksTheNameItRemembers`는 listener가 기억하는 이름 자리에
표적 파일을 놓고 닫아서 `SetUnlinkOnClose(false)`를 직접 잰다.

## B5가 재는 것

`TestCloseToleratesLeftoverAlreadyRemoved`는 다른 부팅의 회수가 먼저 지나간 상황을
만든다: 세 경로를 밖에서 지운 뒤 `Close`가 `nil`을 반환하는지, 그리고 그 뒤의 기동이
빈 자리를 보는지까지 본다. `ErrNotExist` 용인이 없으면 이 테스트가 죽고, 그것이
design D2 "경합은 양성이다"의 코드 근거다.
