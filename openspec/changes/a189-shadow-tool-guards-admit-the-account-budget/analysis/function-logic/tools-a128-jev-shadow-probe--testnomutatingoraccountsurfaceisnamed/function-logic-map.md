# Function Logic Map: `TestNoMutatingOrAccountSurfaceIsNamed` (tools/a128-jev-shadow-probe)

- Source: `tools/a128-jev-shadow-probe/static_test.go`(편집 뒤 79~162행). 편집 전 AST `../../pre-edit-ast/a128--testnomutatingoraccountsurfaceisnamed.ast.json`(78~126행; 편집 전 갈래 수는 그 파일 참조).
- AST evidence: `ast.json`(편집 뒤, 분기 15·return 2 — 둘 다 `ast.Inspect` 콜백의 return)
- Risk scan: `risk-pattern-report.md`(일치 0)
- 편집 두 차례: 2.2(예산 이름 허용·`Option`), 보강(Manager 결정 2026-11-15, gstack 발견 1 — 우선순위 인자 고정·예산 핸들 메서드 금지).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `nonTestSources(t)` | 도구 디렉터리의 비시험 `.go` 12개 | 디스크 | 헬퍼가 Fatal |
| `forbidden` | 주문·계좌·config 쓰기 이름 + 예산 핸들 메서드 7개(`Budget`·`Acquire`·`EnginePollKey`·`NewObservationFence`·`ObservationSnapshot`·`PollRemaining`·`VerifyIdentity`) | 이 함수 | 어떤 선택자든 이름이 맞으면 Errorf |
| `allowedOfficial`·`allowedConfig`·`allowedBudget` | 패키지별 허용 이름 | 이 함수 | 목록 밖이면 Errorf |
| `accountbudget.WithPriority` 호출 | 직접 호출, 인자 2개, 둘째가 `accountbudget.Auxiliary` | 이 함수 | 어긋나면 Errorf, 호출 0 이거나 참조 수 ≠ 호출 수면 Fatalf |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (113) | `range` 소스 파일 | — | — | 자신 |
| B2 (115) **새** | 노드가 `accountbudget.WithPriority(…)` 직접 호출 | `priorityCalls++` | — | 자신 |
| B3 (117) **새** | 인자 수 ≠ 2 이거나 둘째 인자가 `accountbudget.Auxiliary` 가 아님 | — | Errorf | 변이 a128-h·a128-i |
| B4 (122) | 노드가 `SelectorExpr` 아님 | — | 콜백 true(123) | 자신 |
| B5 (125) | `Sel` 이름이 `forbidden` | — | Errorf | 변이 a128-d·a128-j |
| B6 (128) **새** | 선택자가 `accountbudget.WithPriority` | `priorityRefs++` | — | 자신 |
| B7 (131) | X 가 `official` | `sawOfficial=true` | — | 자신 |
| B8 (133) | `official` 이름이 목록 밖 | — | Errorf | 변이 a128-c(RED 에서 `Option`) |
| B9 (137) | X 가 `tossconfig` | `sawConfig=true` | — | 자신 |
| B10 (139) | `tossconfig` 이름이 목록 밖 | — | Errorf | 무편집 갈래, 이번에 재지 않음 |
| B11 (143) | X 가 예산 패키지 | `sawBudget[pkg]=true` | — | 자신 |
| B12 (145) | 예산 패키지 이름이 목록 밖 | — | Errorf | 변이 a128-a·a128-h(Critical) |
| B13 (152) | official·tossconfig 못 봄 | — | Fatalf | 무편집 갈래 |
| B14 (155) | 예산 패키지 못 봄 | — | Fatalf | 변이 a128-f 두 종 |
| B15 (159) **새** | `WithPriority` 직접 호출 0 이거나 참조 수 ≠ 호출 수 | — | Fatalf | 변이 a128-k(호출 0)·a128-l(함수 값) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `nonTestSources` | 검사 대상 파일 | 실패 시 Fatal | AST·CodeGraph |
| `ast.Inspect` | 노드 순회 | — | AST |
| `types.ExprString` | 호출·인자·선택자를 문자열로 비교(표준 라이브러리 `go/types`) | — | AST |
| `t.Errorf`·`t.Fatalf` | 판정 | — | AST |

## State mutations and fallbacks

- 지역 변수만(`sawOfficial`·`sawConfig`·`sawBudget`·`priorityRefs`·`priorityCalls`).

## 편집

- 2.2: `allowedOfficial` 에 `Option`, `allowedBudget`·`sawBudget`, 갈래 B11·B12, 루프 뒤 B14.
- 보강: `forbidden` 에 예산 핸들 메서드 7개, `priorityRefs`·`priorityCalls`, 갈래 B2·B3·B6, 루프 뒤 B15, import `go/types`. 기존 금지 이름·허용 이름·문구 무편집.
- 알려진 한계: 패키지 추적은 식별자 철자 기준이라 별칭 import(`ab.WithPriority`)는 B2·B6 에 안 잡힘(R1, 범위 밖).

## Safety conclusion

- Safe edit boundary: 허용 이름·금지 이름 추가와 우선순위 인자 고정.
- High-risk impact: 없음(시험). 금지 검사 B5 는 그대로임을 변이 a128-d 로 확인.
