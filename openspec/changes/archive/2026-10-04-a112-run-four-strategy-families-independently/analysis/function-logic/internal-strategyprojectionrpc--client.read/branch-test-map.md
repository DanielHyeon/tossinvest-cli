# Branch Test Map: `Client.Read`

- Source SHA-256: `c2a7eaf8469508e309803d53120f0cfbb7f6ec35c24c42e27c5fbcb8bd55f617`; AST branch locations are authoritative.
- B6 하나만 이 lot 이 바꿨다(조건이 좁아짐). 나머지 일곱은 그대로다.

| Branch | Scenario anchor | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if at 57:2 — nil client/토큰 | `TestTheClientCanBeLetGo` 가 유효 client 로 이 가지를 지나간다. nil/짧은 토큰 거절 경로는 이 lot 이 바꾸지 않았고 직접 재는 테스트가 없다 | 아니오 — 기존 분기 | 예 |
| B2 | if at 61:2 — 요청 생성 실패 | 없음 — 이 lot 이 바꾸지 않았고 `http.NewRequestWithContext` 실패를 만들 실용적 입력이 없다 | 아니오 | 해당 없음 |
| B3 | if at 66:2 — 전송 실패 | `TestADialFailureRendersUnavailableRatherThanNotConfigured` (cmd/tossctl) 가 죽은 endpoint 로 읽기 실패 경로를 만든다 | 아니오 — 기존 분기 | 예 |
| B4 | if at 71:2 — 크기 상한 | `TestUnixClientRejectsOversizedResponse` | 아니오 — 기존 분기, **유지가 요점** | 예 |
| B5 | if at 74:2 — HTTP 상태 | `TestAuthenticatedUnixProjectionCrossNamespaceAndReadOnlyAuthority` 의 인증 실패 경로 | 아니오 | 예 |
| B6 | if at 90:2 — 디코딩 (**완화**) | `TestClientIgnoresAdditiveFieldsFromANewerEngine` (모르는 필드는 이 가지를 타지 않는다) | 예 — `DisallowUnknownFields` 가 살아 있는 상태에서 이 테스트는 `unknown field "coordinators"` 로 실패했다 | 예 |
| B7 | if at 94:2 — 값이 하나인가 | 기존 그대로. 이 lot 은 바꾸지 않았다 | 아니오 | 예 |
| B8 | if at 97:2 — 의미 판정 | `TestClientStillRejectsASemanticallyInvalidSnapshot` | 아니오 — 기존 분기지만 이 lot 이 처음 직접 잰다 | 예 |

## 반증 실측

| 뮤테이션 | 결과 |
|---|---|
| M5: `DisallowUnknownFields()` 를 되돌린다 | KILLED — `TestClientIgnoresAdditiveFieldsFromANewerEngine` 이 `json: unknown field "coordinators"` 로 실패. 원복 후 `DisallowUnknownFields` 잔여 1건(`readDescriptor`) 확인. |

B2 는 이 lot 이 편집하지 않은 분기이고 실용적 입력이 없어 **테스트 없음**으로 남긴다 —
침묵한 생략이 아니라 여기 적어 둔 생략이다.
