# Branch Test Map: `Console.routes`

편집은 분기를 추가하지 않았다. 새 라우트 셋은 조건 없이 등록되는 문이고, 분기는
remote 모드에 관한 둘 그대로다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | remote이고 trusted-network가 아님 → login/logout 등록 | 기존 remote 테스트 | no | pass |
| B2 | remote 구성됨 → 미들웨어 래핑 | 기존 remote 테스트 | no | pass |

## 추가 시나리오 (분기 밖의 계약)

| 계약 | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| 세션 | 새 라우트 셋이 session0 뒤에 있다 | `TestEveryRouteGoesThroughTheSessionGate` | no | pass |
| CSRF | 새 라우트 셋이 mutating 뒤에 있다 | `TestEveryStateChangingRouteAlsoGoesThroughTheCSRFGate` | **yes** | pass |
| 열거 | 새 라우트가 계좌 동사를 쓰지 않는다 | `TestNoRouteNamesAnAccountMutation` | no | pass |
| 능력 | Options의 새 필드가 능력 목록에 열거되어 있다 | `TestEveryCapabilityTheConsoleReceivesIsEnumeratedAndDeclaresNothingButReads` | **yes** | pass |
| 리터럴 | 세 경로가 메서드 패턴이 아니다 | `TestNoRouteIsRegisteredWithAMethodPattern` | no | pass |
| 카나리아 | 라우트 수 하한이 실제 등록을 따라간다 | `TestEveryRouteGoesThroughTheSessionGate` | no | pass |
| 원점 | 세 POST가 자기 카드로 답을 돌려보낸다 | `TestEverySettingsPostNamesTheFormItReturnsTo` | no | pass |

두 RED는 실제 관측이다 (2026-08-04). 라우트와 Options 필드를 더하고 정적 목록을
갱신하기 전에 `go test ./internal/console/`를 돌렸을 때 정확히 그 둘이 실패했다.

```
static_test.go:410: /settings/notifications/on is a read route behind the CSRF gate
static_test.go:410: /settings/notifications/test is a read route behind the CSRF gate
static_test.go:410: /settings/notifications/off is a read route behind the CSRF gate
static_test.go:1156: console.Options declares "Notifications", which is not in consoleCapabilities
```

`TestNoRouteNamesAnAccountMutation`은 RED가 아니었다 — 새 경로에 계좌 동사가 없으므로
처음부터 통과했다. 표에 남기는 이유는 그 검사가 이 라우트들을 **실제로 보고 있다**는
것이 이 change의 안전 논거 일부이기 때문이다.

이것이 이 저장소의 정적 검사가 설계된 방식이다 — 상태변경 목록의 확장은 스펙 문장과
검사 목록이 같은 커밋에서 함께 움직여야 하고, 둘 중 하나를 잊으면 빌드가 멈춘다.
