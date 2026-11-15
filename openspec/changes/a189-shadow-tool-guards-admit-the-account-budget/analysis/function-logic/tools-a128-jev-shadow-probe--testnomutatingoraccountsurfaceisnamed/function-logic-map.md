# Function Logic Map: `TestNoMutatingOrAccountSurfaceIsNamed` (tools/a128-jev-shadow-probe)

- Source: `tools/a128-jev-shadow-probe/static_test.go`(편집 뒤 78~143행). 편집 전 AST `../../pre-edit-ast/a128--testnomutatingoraccountsurfaceisnamed.ast.json`(78~126행; 편집 전 갈래 수는 그 파일 참조 — B8·B9·B11 이 새로 생김).
- AST evidence: `ast.json`(편집 뒤, 분기 11·return 2 — 둘 다 `ast.Inspect` 콜백의 return)
- Risk scan: `risk-pattern-report.md`(일치 0)
- 시험 함수지만 본문을 고쳐 gate ⑤ 대상임.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `nonTestSources(t)` | 도구 디렉터리의 비시험 `.go` 12개(편집 시점) | 디스크 | 파싱 실패·0개면 Fatal(헬퍼) |
| `forbidden` | 주문·계좌·config 쓰기 이름 | 이 함수(무편집) | 어떤 선택자든 이름이 맞으면 Errorf |
| `allowedOfficial`·`allowedConfig`·`allowedBudget` | 패키지별 허용 이름 | 이 함수 | 목록 밖이면 Errorf |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (107) | `range` 소스 파일 | — | — | 자신 |
| B2 (110) | 노드가 `SelectorExpr` 아님 | — | 콜백 true(111) | 자신 |
| B3 (113) | `Sel` 이름이 `forbidden` | — | Errorf | 자신(변이 a128-d) |
| B4 (116) | X 가 `official` | `sawOfficial=true` | — | 자신 |
| B5 (118) | `official` 이름이 목록 밖 | — | Errorf | 자신(변이 a128-c; 편집 전 `Option` 이 여기서 RED) |
| B6 (122) | X 가 `tossconfig` | `sawConfig=true` | — | 자신 |
| B7 (124) | `tossconfig` 이름이 목록 밖 | — | Errorf | 이번 범위 밖(무편집, 변이 없음) |
| B8 (128) **새** | X 가 `accountbudget`·`accountbudgetruntime` | `sawBudget[pkg]=true` | — | 자신 |
| B9 (130) **새** | 예산 패키지 이름이 목록 밖 | — | Errorf | 자신(변이 a128-a-budget-name·a128-a-runtime-name) |
| B10 (137) | official·tossconfig 중 못 본 쪽 | — | Fatalf | 이번 범위 밖(무편집) |
| B11 (140) **새** | 예산 패키지 중 못 본 쪽 | — | Fatalf | 자신(변이 a128-f-no-accountbudget·a128-f-alias-runtime) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `nonTestSources` | 검사 대상 파일 | 실패 시 Fatal | AST·CodeGraph |
| `ast.Inspect` | 선택자 순회 | — | AST |
| `t.Errorf`·`t.Fatalf` | 판정 | — | AST |

## State mutations and fallbacks

- 지역 변수만(`sawOfficial`·`sawConfig`·`sawBudget`).

## 편집

- `allowedOfficial` 에 `Option`. 새 `allowedBudget`(accountbudget: WithPriority·Auxiliary / accountbudgetruntime: ParseLaunchConfig·Open)과 `sawBudget`, 콜백 안 갈래 B8·B9, 루프 뒤 B11. 같은 파일 `TestWTSSourceIsTheThreeReads` 의 `allowedPackage`+`saw` 모양. `forbidden`·기존 문구 무편집.

## Safety conclusion

- Safe edit boundary: 허용 이름 넷 + `Option` 과 그 판정 갈래.
- High-risk impact: 없음(시험). 금지 검사 B3 는 그대로임을 변이 a128-d 로 확인.
