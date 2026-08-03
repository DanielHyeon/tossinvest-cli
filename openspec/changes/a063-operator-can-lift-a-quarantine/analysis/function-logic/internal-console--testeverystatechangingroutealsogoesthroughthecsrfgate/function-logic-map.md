# Function Logic Map: `TestEveryStateChangingRouteAlsoGoesThroughTheCSRFGate`

- Source: `internal/console/static_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`
- Base commit: `openspec/changes/a063-operator-can-lift-a-quarantine/base-commit.txt`
- 위험 등급: Normal (테스트). 다만 이 테스트가 **콘솔 전체의 CSRF 계약을 소유**하므로
  편집 시 검사가 느슨해지지 않았음을 보여야 한다. Pre-Edit 선언은 `review.md`.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `stateChanging` | 상태를 바꾸는 라우트 경로의 **명시 allowlist** | 이 함수 안의 리터럴 | 목록에 없으면 read route로 간주 |
| `registeredRoutes(t)` | 실제로 mux에 등록된 라우트와 각각의 CSRF 게이트 여부 | `Console.routes` | — |

**불변식(양방향)**: allowlist에 있으면 CSRF 게이트 아래여야 하고, 없으면 게이트
아래여선 안 된다. 두 방향을 함께 검사하는 것이 이 테스트의 설계다 — 한 방향만
검사하면 "모든 라우트를 게이트에 넣기"로 통과시킬 수 있고, 그러면 대시보드를 아무도
열 수 없다.

**a063이 바꾸는 것**: allowlist에 `/position-management/quarantine/preview`와
`/position-management/quarantine/apply` 두 항목을 추가한다. 검사 로직·분기·비교
방향은 한 글자도 바뀌지 않는다. 리터럴 목록에 두 줄이 늘어날 뿐이다.

**왜 목록에 넣는가**: 두 라우트는 원장의 `released_at`을 쓰고 포지션을 자동 판정으로
되돌린다. 위조된 요청이 그것을 할 수 있어서는 안 된다 — 이 목록이 존재하는 바로 그
의미의 "act"다.

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (404) | 등록된 라우트 순회 | `seen[path] = true` | 계속 | 자기 자신 |
| B2 (406) | 등록 라우트 분류 switch | 없음 | 아래 두 갈래 | 자기 자신 |
| B3 (407) | allowlist에 있는데 CSRF 게이트가 없음 | 없음 | `t.Errorf` | 자기 자신 |
| B4 (409) | allowlist에 없는데 CSRF 게이트가 있음 | 없음 | `t.Errorf` | 자기 자신 |
| B5 (413) | allowlist 순회 | 없음 | 계속 | 자기 자신 |
| B6 (414) | allowlist 항목이 아예 등록되지 않음 | 없음 | `t.Errorf` | 자기 자신 |

a063은 분기를 추가·제거·재배치하지 않는다. B3·B4·B6이 새 두 경로에 대해서도 그대로
적용되는 것이 이 편집의 전부다.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `registeredRoutes(t)` | 실제 mux 등록 상태를 읽는다 | 실패 시 테스트 실패 | AST |
| `t.Errorf` | 계약 위반 보고 | — | AST |

프로덕션 코드를 호출하지 않는다. I/O·네트워크·원장 접근이 없다.

## State mutations and fallbacks

- 테스트 로컬 `seen` 맵만 채운다. fallback 없음.
- 이 함수는 **fail-closed 방향으로만 틀릴 수 있다**: 목록에 넣지 않은 mutation
  라우트는 B4에서 실패한다. 실제로 a063 구현 중 두 라우트를 넣기 전에 실패했다.

## Safety conclusion

- Safe edit boundary: `stateChanging` 리터럴의 항목 두 개.
- High-risk impact: **no** — 테스트이고 프로덕션 동작을 바꾸지 않는다. 다만 이 목록이
  느슨해지면 콘솔 전체의 CSRF 계약이 느슨해지므로, 이 편집이 **항목 추가일 뿐 검사
  완화가 아님**을 위 분기표가 보인다.
- §0.7: 새 두 경로를 목록에 넣는 것은 "사람 승인 경로다"라고 선언하는 것과 같다.
