# Function Logic Map: `Console.routes`

- Source: `internal/console/console.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`
- Narrative context: `../../function-logic-map.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 등록 테이블 | 라우트 22개 | `internal/console/console.go`의 이 함수뿐 | 등록 누락은 404이고 정적 검사가 먼저 실패한다 |
| wrapper 조합 | 읽기=`session0`, 행위=`session0(mutating(...))` | 이 함수 | 조합이 틀리면 `TestEveryStateChangingRouteAlsoGoesThroughTheCSRFGate`가 실패 |
| 경로 어휘 | 계좌·게이트 어휘 금지 | `routeOnlyAccountVerbs` | `gate`를 담은 경로는 `TestNoRouteNamesAnAccountMutation`이 거부 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | 분기 없음 — 순차 등록만 한다 | `mux`에 핸들러를 붙인다 | `http.Handler` | `TestEveryRouteGoesThroughTheSessionGate` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `http.NewServeMux` | 라우트 표 | 없음 | CodeGraph + AST |
| `Console.session0` | 세션 게이트 — 전 라우트 | 미인증은 403 | CodeGraph + AST |
| `Console.mutating` | CSRF+POST 게이트 — 상태변경 라우트만 | 미충족은 403/405 | CodeGraph + AST |
| `Console.handleSettingsLimits` | 이 change의 개별 한도 저장 | 핸들러가 자체 처리 | CodeGraph + AST |
| `Console.handleSettingsLimitPreset` | 이 change의 프리셋 적용 | 핸들러가 자체 처리 | CodeGraph + AST |

## State mutations and fallbacks

- 이 change의 변경은 `/settings/limits`·`/settings/limits/preset` 등록 **2줄 추가**뿐이다. 기존 등록의 순서·경로·wrapper는 무변경이다.
- 두 경로 모두 `mutating` 안에 있다. 읽기로 등록하면 CSRF 검사가 붙지 않아 정적 검사가 실패한다.
- 경로 이름에 `gate`를 쓰지 않았다 — `routeOnlyAccountVerbs`의 금지어이고, 이 change 이후에도 게이트를 여는 라우트는 없어야 하므로 그 금지어를 유지했다.

## Safety conclusion

- Safe edit boundary: 라우트 2줄. wrapper 조합이 `/settings/exclude`와 같아야 하고, 정적 검사 4종이 그 사실을 읽는다.
- High-risk impact: no — 이 함수 자체는 계좌·원장·주문에 닿지 않는다. 등록되는 핸들러의 위험은 그 핸들러의 map이 다룬다.
