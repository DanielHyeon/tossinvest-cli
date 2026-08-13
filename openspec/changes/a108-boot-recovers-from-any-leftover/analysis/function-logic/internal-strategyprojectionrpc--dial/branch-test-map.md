# Branch Test Map: `Dial`

- Source: `internal/strategyprojectionrpc/transport_unix.go` (216-230)
- 이 change는 이 함수를 **바꾸지 않았다.** 잔재의 소비자로서 겹3(T2)의 입력이다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | descriptor를 읽지 못한다 | `TestStartRecoversFromDescriptorOnlyLeftover`, `TestAuthenticatedUnixProjectionCrossNamespaceAndReadOnlyAuthority` (성공 쪽) | no (편집 전 GREEN) | yes |
| B2 | socket 파일이 없거나 0600 socket이 아니다 | `TestCloseToleratesLeftoverAlreadyRemoved` — 경합을 배제하기 전 실측에서 이 오류(`socket is invalid`)를 관측했다 | no (편집 전 GREEN) | yes |

## 회수 테스트가 `Dial`을 부르는 이유

회수 시나리오는 전부 `Start` 성공 뒤 `Dial` + `Read`까지 간다. "거부는 안 했다"만 보면
**잔재만 지우고 아무것도 세우지 않은 구현**이 통과하기 때문이다. 그래서 B1의 성공 쪽이
모든 회수 테스트의 마지막 단계에 있다.
