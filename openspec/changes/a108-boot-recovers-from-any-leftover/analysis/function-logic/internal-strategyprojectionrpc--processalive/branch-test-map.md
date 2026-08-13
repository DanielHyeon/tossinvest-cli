# Branch Test Map: `processAlive`

- Source: `internal/strategyprojectionrpc/transport_unix.go` (169-175) — revision `base`
- 이 함수는 a108에서 **삭제됐다.** 아래 행은 삭제 전 상태의 기록이다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `pid <= 0`이면 사망으로 읽는다 | 없음 — 편집 전에도 직접 테스트가 없었고, 함수가 사라져 새로 만들 대상도 아니다 | no | no |

## 삭제를 무엇이 지키는가

직접 테스트 대신, **삭제된 판정이 되살아나면 죽는 테스트**가 근거다. 뮤테이션 원장
(`../../mutation-ledger-t1.md`)의 M5·M5b·M5c가 kill-0 판정을 되살리고, 다음이 죽는다.

- `TestStartRecoversFromDeadEndpointWhoseDescriptorPIDIsAlive` — 죽은 endpoint를
  "주인 생존"으로 오판해 영구 거부하는 방향.
- `TestStartRefusesLiveSocketWhoseDescriptorPIDIsDead`,
  `TestStartRefusesLiveSocketWithoutDescriptor` — 수락 중인 socket을 "주인 사망"으로
  오판해 남의 파일을 지우는 방향.

즉 이 함수의 커버리지는 "함수를 부르는 테스트"가 아니라 "함수를 되살리면 깨지는
테스트"로 존재한다. 삭제된 코드에 대해 가능한 유일한 형태다.
