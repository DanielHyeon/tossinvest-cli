# Branch Test Map: `Saga.event`

측정: `go test -covermode=set -coverprofile ./internal/flatten/` RED 프로파일에서 블록
단위로 직접 읽었다.

**a097은 이 함수를 편집하지 않는다.** 표가 필요한 이유는 proposal R2가 이 함수를
"알림 오류를 버리는 호출자"의 실증으로 인용하기 때문이다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:690` Notifier 미배선 → 로그로 강등 | flatten saga 테스트 다수 (Notifier 없이 조립) | 진입 (`690-693 count=1`) | 변화 없음 (편집 대상 아님) |

정상 경로 `s.Notifier.Notify@694`: **미진입** (`694-699 count=0`).

## 이 측정이 R2에 보태는 것

flatten 스위트의 어떤 테스트도 Notifier를 배선하지 않는다. 따라서 **오류를 버리는 그 줄은
한 번도 실행된 적이 없다.** 그것이 P2 ③이 지금까지 보이지 않았던 이유다 — 코드 리뷰로만
볼 수 있고 테스트로는 볼 수 없었다.

a097은 이 함수에도 flatten 스위트에도 손대지 않는다. 대신 생산자
(`obs.Notifier.claimAndDeliver` B1@227)에서 gate를 잠가, 이 줄이 언젠가 실행될 때 그
버려진 오류가 **이미 결과를 남긴 뒤**가 되게 한다.

GREEN에서 이 표의 값은 바뀌지 않아야 한다. 바뀐다면 a097이 flatten 경로를 건드렸다는
뜻이고, 그것은 범위 위반이다.
