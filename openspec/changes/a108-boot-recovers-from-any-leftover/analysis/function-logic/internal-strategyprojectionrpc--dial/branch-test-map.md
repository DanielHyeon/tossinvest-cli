# Branch Test Map: `Dial`

- Source: `internal/strategyprojectionrpc/transport_unix.go` (402-424) — revision `current`
- 첫 라운드는 이 함수를 바꾸지 않았다. Fix 라운드가 connect probe를 넣었다(D4-2).

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | descriptor를 읽지 못한다 | `TestStartRecoversFromDescriptorOnlyLeftover`, `TestAuthenticatedUnixProjectionCrossNamespaceAndReadOnlyAuthority` (성공 쪽) | no (편집 전 GREEN) | yes |
| B2 | socket 파일이 없거나 0600 socket이 아니다 | `TestCloseToleratesLeftoverAlreadyRemoved` — 경합을 배제하기 전 실측에서 `socket is invalid`를 관측했다 | no (편집 전 GREEN) | yes |
| B3 | 파일은 있는데 아무도 수락하지 않는다 (S3) | `TestDialRefusesSocketWithNoListener` | yes (Dial이 client를 돌려줬다) | yes |

## B3의 RED

Fix 라운드 §6.1의 RED다. 죽은 socket(파일 잔존·listener 없음)과 유효한 descriptor를
놓고 `Dial`을 부르면 **성공**했다 — 그것이 A2 F3의 실측이었다. 지금은 즉시 오류다.
뮤테이션 M12(probe 블록 삭제)가 이 테스트 하나로 죽는다.

## 회수 테스트가 `Dial`을 부르는 이유

회수 시나리오는 전부 `Start` 성공 뒤 `Dial` + `Read`까지 간다. "거부는 안 했다"만 보면
**잔재만 지우고 아무것도 세우지 않은 구현**이 통과하기 때문이다. 그래서 B1의 성공 쪽이
모든 회수 테스트의 마지막 단계에 있다. probe가 들어온 지금은 그 마지막 단계가 "세운
endpoint가 실제로 수락한다"까지 함께 잰다.
