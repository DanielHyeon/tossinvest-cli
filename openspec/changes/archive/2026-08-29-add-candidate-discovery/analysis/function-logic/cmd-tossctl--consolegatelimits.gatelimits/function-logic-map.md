# Function Logic Map: `consoleGateLimits.GateLimits`

- Source: `cmd/tossctl/console.go`
- AST evidence: `ast.json` (revision=current, L332–349, 분기 1개)
- Risk scan: `risk-pattern-report.md`
- 이 change의 base: `583772c4` — base에는 이 함수가 없었다. 현재 본문을 diff hunk가 덮으므로 evidence가 요구된다 (revision=current)

config 파일 1회 읽기. 경계를 넘는 값은 **float64 다섯 개 + 통화 문자열**이고 그것이 전부다.

- **타임아웃**: seam이 context를 받지 않으므로 데드라인을 여기서 세운다 — `consoleGateLimitsTimeout` 5초. 개요는 30초마다 자기를 다시 로드하고 운영자는 하루 종일 열어둔다. 멈춘 파일시스템에서 `context.Background()`로 읽으면 HTTP 핸들러가 무한히 잡힌다 — 개요가 유일하게 문장으로 표현할 수 없는 실패다.
- **실패가 화면에 남기는 것**: 에러를 삼키지 않고 반환한다. 개요는 읽지 못한 한도를 0도 무제한도 아닌 "미측정 + 에러 문구"로 렌더한다. 타임아웃도 그 실패 중 하나로 도착한다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `s.svc` | non-nil (`consoleGateLimitsSeam`이 보장) | `configServiceFor` | nil이면 seam 자체가 만들어지지 않는다 |
| `cfg.Engine.AutomationGate` | float64 5개 + 통화 | config.json | 읽기 실패는 에러로 전달 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `s.svc.Load(ctx)` 실패(타임아웃 포함) | 없음 | `console.GateLimits{}, err` | 개요의 미측정 렌더 + `TestTheLimitsReadIsBounded`(무한 대기 금지) |
| (else) | 로드 성공 | 없음 | 다섯 숫자 + 통화 | `TestTheConsoleIsHandedTheLimitsAsNumbersAndNoWayToWriteThem` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `context.WithTimeout(context.Background(), consoleGateLimitsTimeout)` | 5초 상한 | 만료는 Load 에러로 도착 | `TestTheLimitsReadIsBounded`가 소스에서 이 문자열을 검사 |
| `s.svc.Load` | config 파일 읽기 | 에러 그대로 반환 | ast.json calls |
| `cancel` (defer) | context 누수 방지 | — | ast.json defers |

## State mutations and fallbacks

- 쓰기 없음. `config.Service`의 writer는 이 타입에서 도달 불가하며 애초에 콘솔로 넘어가지 않는다.
- 값을 복사해서 넘기므로 콘솔이 `config.AutomationGate` 타입을 알 필요가 없다 — internal/console의 정적 가드가 요구하는 형태.

## Safety conclusion

- Safe edit boundary: 타임아웃 상수와 필드 대응. `context.Background()`를 그대로 넘기는 되돌림은 테스트가 막는다.
- High-risk impact: yes (Guardian 경로) — Guardian 한도의 유일한 읽기 경로다. 읽기 실패를 0으로 대체하는 편집은 개요가 "한도 0"(=아무것도 못 낸다) 또는 "무제한"으로 읽히게 만들고, 둘 다 아무도 읽지 않은 값을 운영자에게 사실처럼 보여준다. 주문 능력은 넘기지 않는다.
