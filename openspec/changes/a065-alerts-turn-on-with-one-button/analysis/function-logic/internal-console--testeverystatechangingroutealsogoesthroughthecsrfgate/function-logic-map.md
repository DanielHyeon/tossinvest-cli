# Function Logic Map: `TestEveryStateChangingRouteAlsoGoesThroughTheCSRFGate`

- Source: `internal/console/static_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`
- Base commit: `openspec/changes/a065-alerts-turn-on-with-one-button/base-commit.txt`
- 위험 등급: High (guard)

## Inputs and invariants

양방향 검사다: 상태변경 목록에 있는데 CSRF 게이트가 없으면 실패하고, 게이트가
있는데 목록에 없어도 실패한다. 후자가 이 change에서 실제로 발동했다 — 라우트 셋을
등록하고 목록을 갱신하기 전에 정확히 그 셋이 "read route behind the CSRF gate"로 실패했다.

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `stateChanging` | 이 함수 안의 닫힌 map | operator-console 스펙 문장 | 목록과 스펙은 같은 커밋에서 움직인다 |
| `registeredRoutes(t)` | 소스에서 파싱 | 패키지 전체 | 리터럴 경로만 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (419) | 라우트 순회 | 없음 | — | 자기 자신 |
| B2 (421) | 분류 | 없음 | — | 자기 자신 |
| B3 (422) | 목록에 있는데 게이트 없음 | 없음 | 실패 보고 | 자기 자신 |
| B4 (424) | 목록에 없는데 게이트 있음 | 없음 | 실패 보고 | **자기 자신 — a065에서 실제 RED** |
| B5 (428) | 목록 순회 | 없음 | — | 자기 자신 |
| B6 (429) | 목록에 있는데 라우트가 없음 | 없음 | 실패 보고 | 자기 자신 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `registeredRoutes` | 소스에서 라우트 표를 읽는다 | 파싱 실패 = 적게 읽힘 | AST |

## State mutations and fallbacks

- 아무것도 바꾸지 않는다. a065의 편집은 `stateChanging` map에 세 항목과 그 이유를 적은 주석이다.
- 세 항목 모두 계좌·원장·브로커에 닿지 않는다: `on`/`off`는 `engine.notifications`의 키를 옮기고 `test`는 아무것도 쓰지 않고 한 통 publish한다.

## Safety conclusion

- Safe edit boundary: map 항목 셋과 주석.
- High-risk impact: **no** — 검사이며 제품 코드가 아니다.
- 약화 없음: 목록에 더한 셋은 전부 실제로 `c.mutating` 뒤에 등록되어 있고, B4가 그것을 강제한다.
