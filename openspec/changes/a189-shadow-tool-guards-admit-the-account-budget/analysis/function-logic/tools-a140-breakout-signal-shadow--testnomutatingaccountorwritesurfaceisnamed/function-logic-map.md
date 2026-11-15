# Function Logic Map: `TestNoMutatingAccountOrWriteSurfaceIsNamed` (tools/a140-breakout-signal-shadow)

- Source: `tools/a140-breakout-signal-shadow/static_test.go`(편집 뒤 70~121행). 편집 전 AST `../../pre-edit-ast/a140--testnomutatingaccountorwritesurfaceisnamed.ast.json`(70~116행, 분기 8).
- AST evidence: `ast.json`(편집 뒤, 분기 8 그대로·return 2 — 콜백의 return)
- Risk scan: `risk-pattern-report.md`(일치 0)

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `nonTestSources(t)` | 비시험 `.go` 13개(편집 시점) | 디스크 | 헬퍼가 Fatal |
| `forbidden` | 주문·계좌·config 쓰기 이름 | 이 함수(무편집) | 선택자 이름이 맞으면 Errorf |
| `allowed` | 패키지 → 허용 이름(official·tossconfig·journal·accountbudget·accountbudgetruntime) | 이 함수 | 목록 밖 Errorf, 키마다 못 보면 Errorf |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (96) | `range` 파일 | — | — | 자신 |
| B2 (99) | 노드가 `SelectorExpr` 아님 | — | 콜백 true(100) | 자신 |
| B3 (102) | `Sel` 이름이 `forbidden` | — | Errorf | 자신(변이 a140-d) |
| B4 (105) | X 가 식별자 | — | — | 자신 |
| B5 (106) | 그 식별자가 `allowed` 키 | `saw[pkg]=true` | — | 자신 |
| B6 (108) | 그 패키지 목록 밖 이름 | — | Errorf | 자신(편집 전 `official.Option` RED, 변이 a140-a·a140-c) |
| B7 (116) | `range allowed` | — | — | 자신 |
| B8 (117) | 키를 못 봄 | — | Errorf | 자신(변이 a140-f-no-accountbudget·a140-f-alias-runtime) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `nonTestSources`·`ast.Inspect` | 선택자 순회 | — | AST·CodeGraph |

## State mutations and fallbacks

- 지역 `saw` 맵만.

## 편집

- `allowed["official"]` 에 `Option`, 새 키 `accountbudget`(WithPriority·Auxiliary)·`accountbudgetruntime`(ParseLaunchConfig·Open)와 한글 주석. 분기 코드 무편집 — 새 키는 B5·B6·B8 이 그대로 잼. `forbidden` 무편집.

## Safety conclusion

- Safe edit boundary: 맵 항목 셋.
- High-risk impact: 없음(시험).
