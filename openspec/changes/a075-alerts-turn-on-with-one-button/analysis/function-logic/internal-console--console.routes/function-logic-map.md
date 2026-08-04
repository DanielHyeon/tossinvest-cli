# Function Logic Map: `Console.routes`

- Source: `internal/console/console.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`
- Base commit: `openspec/changes/a075-alerts-turn-on-with-one-button/base-commit.txt`
- 위험 등급: High — 이 함수가 **콘솔이 무엇을 할 수 있는가**를 정의한다. 여기에
  등록되지 않은 것은 도달할 수 없고, 여기에 잘못 등록된 것은 게이트 밖에 놓인다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `c.remote` | nil 또는 구성됨 | 기동 옵션 | nil이면 login/logout 미등록 |
| `c.remote.trustedNetwork` | bool | 기동 옵션 | true면 application login 없음 |
| 등록되는 handler들 | 메서드 값 | 같은 패키지 | — |

**불변식 (유지)**: 모든 라우트는 `session0` 뒤에 있다 (`/healthz`와 `/login` 제외).
`TestEveryRouteGoesThroughTheSessionGate`가 소스에서 라우트 표를 읽어 강제한다.

**불변식 (유지)**: 상태를 바꾸는 모든 라우트는 `c.mutating` 뒤에 있고, 읽기 라우트는
그 뒤에 있지 **않다**. `TestEveryStateChangingRouteAlsoGoesThroughTheCSRFGate`가
양방향으로 확인한다 — 목록에 있는데 게이트가 없으면 실패하고, 게이트가 있는데 목록에
없어도 실패한다.

**불변식 (유지)**: 경로는 **문자열 리터럴**이어야 한다. `registeredRoutes`가 소스를
파싱해 라우트 표를 만들고, 리터럴이 아닌 경로는 모든 정적 검사가 조용히 건너뛴다.
같은 이유로 메서드 패턴(`"GET /x"`)을 쓰지 않는다.

**a075가 바꾸는 것**: `/settings/notifications/{on,test,off}` 셋을
`session0(mutating(...))`로 등록한다. 셋 다 `/settings/autostart` **앞**에 온다 —
위치는 의미가 없고 읽는 순서만 결정한다.

**a075가 바꾸지 않는 것**: 기존 라우트의 경로·게이트·핸들러, 두 분기의 조건,
`/` catch-all의 위치.

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (718) | `c.remote != nil && !c.remote.trustedNetwork` | `/login`·`/logout` 등록 | — | 기존 remote 테스트 |
| B2 (841) | `c.remote != nil` | remote 미들웨어 래핑 | — | 기존 remote 테스트 |

새 라우트 셋은 분기가 아니라 문이다. 조건 없이 항상 등록되고, seam이 없는 빌드에서는
handler가 501을 답한다 — 라우트를 조건부로 등록하면 정적 검사가 그 라우트를 보지
못하는 빌드가 생기고, 그것은 게이트 검사가 침묵으로 통과하는 경로다.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `http.NewServeMux` | 라우트 표 | — | AST |
| `c.session0` | 세션 게이트 | 미인증 → 거부 | AST |
| `c.mutating` | CSRF 게이트 | 토큰 불일치 → 거부 | AST |
| `c.readOnly` | GET/HEAD 외 거부 | — | AST |
| **`c.handleSettingsNotificationsOn/Test/Off`** (신규) | 알림 카드의 세 행위 | seam nil → 501 | 신규 |

## State mutations and fallbacks

- `mux`만 만든다. 요청을 처리하지 않고, 파일도 네트워크도 건드리지 않는다.
- 새 handler 셋은 `c.opts.Notifications == nil`이면 `c.refuse(501)`로 답한다 —
  라우트는 존재하지만 능력은 없다는 것이 화면과 라우트 표 양쪽에서 참이다.

## Safety conclusion

- Safe edit boundary: `mux.HandleFunc` 세 줄. 기존 라우트의 게이트 체인 무편집.
- High-risk impact: **no** — 주문·정정·취소·자격증명 라우트를 만들지 않는다.
  세 경로 중 어느 것도 계좌·원장·브로커에 닿지 않는다.
- 정적 검사: 스펙 문장(`operator-console` 상태변경 목록)과 `consoleStateChanging`,
  CSRF 목록이 **같은 커밋에서** 함께 움직였다 — 스펙이 요구하는 것이다.
- 라우트 이름은 무엇을 하는지 말한다. `/settings/notifications/on`이
  `/settings/alerts`보다 감사에서 읽히고, `/settings/gate`가 이름을 숨기지 않는 것과
  같은 규칙이다.
