# Function Logic Map: `TestEveryRouteGoesThroughTheSessionGate`

- Source: `internal/console/static_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`
- Base commit: `openspec/changes/a075-alerts-turn-on-with-one-button/base-commit.txt`
- 위험 등급: High (guard)

## Inputs and invariants

이 검사는 두 가지를 한다: 모든 라우트가 세션 게이트 뒤에 있는지, 그리고 라우트
표 추출기가 **표 전체를 읽고 있는지**. 두 번째가 하한 숫자이고, 주석이 그 목적을 적는다 —
"A floor below the truth would let a scanner that read only console.go's first half go on
passing." a075는 라우트 셋을 더하므로 하한도 셋 올린다.

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `registeredRoutes(t)` | 소스에서 파싱된 라우트 | `console.go` 등 패키지 전체 | 리터럴이 아닌 경로는 보이지 않는다 |
| `public` | 2개 (`/healthz`, `/login`) | 이 함수 | 그 밖은 전부 session0 필요 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (278) | 라우트 순회 | 없음 | — | 자기 자신 |
| B2 (279) | 세션 게이트 없음 & public 아님 | 없음 | 실패 보고 | 자기 자신 |
| B3 (315) | `len(routes) < 30` | 없음 | 실패 보고 | 자기 자신 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `registeredRoutes` | 소스에서 라우트 표를 읽는다 | 파싱 실패 = 적게 읽힘 | AST |

## State mutations and fallbacks

- 아무것도 바꾸지 않는다. a075의 편집은 하한 상수 `27 → 30`과 그 이유를 적은 주석뿐이다.
- 하한은 실제 등록 수(49)보다 여전히 낮다 — a075 이전부터 그랬고, 이 change는 자기가 더한 만큼만 올려 격차를 넓히지 않는다 (issues I4).

## Safety conclusion

- Safe edit boundary: 상수 하나와 주석.
- High-risk impact: **no** — 검사 자체이며 제품 코드가 아니다.
- 이 함수를 약화시키지 않았다: 하한은 올라갔고 세션 게이트 규칙은 그대로다.
