# Function Logic Map: `TestNoMutatingAccountOrWriteSurfaceIsNamed` (tools/a140-breakout-signal-shadow)

- Source: `tools/a140-breakout-signal-shadow/static_test.go`(편집 뒤 71~140행). 편집 전 AST `../../pre-edit-ast/a140--testnomutatingaccountorwritesurfaceisnamed.ast.json`(70~116행; 편집 전 갈래 수는 그 파일 참조).
- AST evidence: `ast.json`(편집 뒤, 분기 12·return 2 — 콜백의 return)
- Risk scan: `risk-pattern-report.md`(일치 0)
- 편집 두 차례: 2.2(맵 항목), 보강(Manager 결정 2026-11-15, gstack 발견 1).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `nonTestSources(t)` | 비시험 `.go` 13개 | 디스크 | 헬퍼가 Fatal |
| `forbidden` | 주문·계좌·config 쓰기 이름 + 예산 핸들 메서드 7개 | 이 함수 | 선택자 이름이 맞으면 Errorf |
| `allowed` | 패키지 → 허용 이름(official·tossconfig·journal·accountbudget·accountbudgetruntime) | 이 함수 | 목록 밖 Errorf, 키마다 못 보면 Errorf |
| `accountbudget.WithPriority` 호출 | 직접 호출, 인자 2개, 둘째가 `accountbudget.Auxiliary` | 이 함수 | 어긋나면 Errorf, 호출 0 이거나 참조 수 ≠ 호출 수면 Fatalf |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (102) | `range` 파일 | — | — | 자신 |
| B2 (104) **새** | `accountbudget.WithPriority(…)` 직접 호출 | `priorityCalls++` | — | 자신 |
| B3 (106) **새** | 인자가 Auxiliary 아님 | — | Errorf | 변이 a140-h·a140-i |
| B4 (111) | 노드가 `SelectorExpr` 아님 | — | 콜백 true(112) | 자신 |
| B5 (114) | `Sel` 이름이 `forbidden` | — | Errorf | 변이 a140-d·a140-j |
| B6 (117) **새** | 선택자가 `accountbudget.WithPriority` | `priorityRefs++` | — | 자신 |
| B7 (120) | X 가 식별자 | — | — | 자신 |
| B8 (121) | 그 식별자가 `allowed` 키 | `saw[pkg]=true` | — | 자신 |
| B9 (123) | 그 패키지 목록 밖 이름 | — | Errorf | 변이 a140-a·a140-c·a140-h(RED 에서 `Option`) |
| B10 (131) | `range allowed` | — | — | 자신 |
| B11 (132) | 키를 못 봄 | — | Errorf | 변이 a140-f 두 종 |
| B12 (137) **새** | 호출 0 이거나 참조 ≠ 호출 | — | Fatalf | 변이 a140-k·a140-l |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `nonTestSources`·`ast.Inspect` | 노드 순회 | — | AST·CodeGraph |
| `types.ExprString` | 호출·인자·선택자 문자열 비교 | — | AST |

## State mutations and fallbacks

- 지역 `saw`·`priorityRefs`·`priorityCalls` 만.

## 편집

- 2.2: `allowed["official"]` 에 `Option`, 새 키 `accountbudget`·`accountbudgetruntime`.
- 보강: `forbidden` 에 예산 핸들 메서드 7개, 갈래 B2·B3·B6·B12, import `go/types`. 기존 이름·문구 무편집.
- 알려진 한계: 별칭 import 는 B2·B6 에 안 잡힘(R1, 범위 밖).

## Safety conclusion

- Safe edit boundary: 맵 항목·금지 이름 추가와 우선순위 인자 고정.
- High-risk impact: 없음(시험).
