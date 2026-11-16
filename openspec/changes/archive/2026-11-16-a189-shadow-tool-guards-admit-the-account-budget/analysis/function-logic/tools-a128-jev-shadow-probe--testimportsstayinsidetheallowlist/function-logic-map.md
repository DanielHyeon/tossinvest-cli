# Function Logic Map: `TestImportsStayInsideTheAllowlist` (tools/a128-jev-shadow-probe)

- Source: `tools/a128-jev-shadow-probe/static_test.go`(편집 뒤 198~231행). 편집 전 logic-map 산출물: base `../../pre-edit-ast/a128--testimportsstayinsidetheallowlist.ast.json`(162~181행, 분기 수 3개), Manager 결정 1 직전 `../../pre-edit-ast/a128--testimportsstayinsidetheallowlist.before-decision1.ast.json`(198~220행, 분기 수 3개).
- AST evidence: `ast.json`(편집 뒤, 분기 4, return 0)
- Risk scan: `risk-pattern-report.md`(일치 0)

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 비시험 소스의 import | 표준 라이브러리 + `allowedModule` | 이 함수 | 경로에 `.` 이 있고 목록 밖이면 Errorf |
| `official`·`accountbudget`·`accountbudgetruntime` import 의 이름 | `spec.Name == nil`(패키지 이름 그대로) | 이 함수(Manager 결정 1, 2026-11-16) | 별칭·dot·blank 면 Errorf |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (220) | `range` 파일 | — | — | 자신 |
| B2 (221) | `range` import | — | — | 자신 |
| B3 (223) | 모듈 경로이고 목록 밖 | — | Errorf | 자신(편집 전 두 예산 경로로 RED, 변이 a128-e-import) |
| B4 (226) | 세 패키지 중 하나를 이름 붙여 import | — | Errorf | 자신 — 변이 d1a-a128(별칭+Critical)·d1b-a128(dot+Critical)·d1c-a128(blank) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `nonTestSources`·`strconv.Unquote`·`strings.Contains` | 대상 열거·경로 해독 | — | AST |

## State mutations and fallbacks

- 없음.

## 편집

- 2.2: `allowedModule` 에 `internal/accountbudget`·`internal/accountbudgetruntime` 두 경로와 한글 주석 한 줄. 실패 문구의 괄호 목록에 두 이름을 더해 사실에 맞춤.
- Manager 결정 1(2026-11-16): `ownNameOnly` 와 B4 를 더함. 이유: 이름 시험(`TestNoMutatingOrAccountSurfaceIsNamed`)은 선택자의 X 철자로 패키지를 가리므로 별칭·dot import 로 `WithPriority(ctx, Critical)` 을 달면 두 시험이 다 통과했음(gstack 2·3회차 red team 실측, 변이 d1a/d1b-a128-PRE1). 별칭을 쫓지 않고 금지함.

## Safety conclusion

- Safe edit boundary: 허용 경로 두 개와 이름 붙인 import 금지 갈래.
- High-risk impact: 없음(시험). 측정 도구가 엔진 보호 주문 몫(Critical)을 달 길 하나를 닫음(안전 불변식 4).
