# Function Logic Map: `TestImportsStayInsideTheAllowlist` (tools/a140-breakout-signal-shadow)

- Source: `tools/a140-breakout-signal-shadow/static_test.go`(편집 뒤 142~168행). 편집 전 logic-map 산출물: base `../../pre-edit-ast/a140--testimportsstayinsidetheallowlist.ast.json`(118~136행, 분기 수 4개), Manager 결정 1 직전 `../../pre-edit-ast/a140--testimportsstayinsidetheallowlist.before-decision1.ast.json`(142~162행, 분기 수 4개).
- AST evidence: `ast.json`(편집 뒤, 분기 5, return 0)
- Risk scan: `risk-pattern-report.md`(일치 0)

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 비시험 소스의 import 경로 | 표준 라이브러리 + `module` 접두사 + `allowed` 접미사 | 이 함수 | 목록 밖 Errorf |
| `official`·`accountbudget`·`accountbudgetruntime` import 의 이름 | `spec.Name == nil`(패키지 이름 그대로) | 이 함수(Manager 결정 1, 2026-11-16) | 별칭·dot·blank 면 Errorf |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (154) | `range` 파일 | — | — | 자신 |
| B2 (155) | `range` import | — | — | 자신 |
| B3 (157) | 표준 라이브러리 | — | continue | 자신 |
| B4 (160) | 모듈 밖이거나 접미사 목록 밖 | — | Errorf | 자신(편집 전 두 예산 경로 RED, 변이 a140-e-import) |
| B5 (163) | 세 패키지 중 하나를 이름 붙여 import | — | Errorf | 자신 — 변이 d1a-a140(별칭+Critical)·d1b-a140(dot+Critical)·d1c-a140(official blank) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `nonTestSources`·`strconv.Unquote`·`strings.*` | 열거·판정 | — | AST |

## State mutations and fallbacks

- 없음.

## 편집

- 2.2: `allowed` 에 `accountbudget`·`accountbudgetruntime` 와 한글 주석 한 줄.
- Manager 결정 1(2026-11-16): `ownNameOnly` 와 B5 를 더함. 이유: 이름 시험(`TestNoMutatingAccountOrWriteSurfaceIsNamed`)은 선택자의 X 철자로 패키지를 가리므로 별칭·dot import 로 `WithPriority(ctx, Critical)` 을 달면 이름 시험·이 시험이 둘 다 통과했음(gstack 2·3회차 red team 실측, 변이 d1a/d1b-a140-PRE1 이 결정 전 판본에서 PASS). 별칭을 쫓지 않고 금지함.

## Safety conclusion

- Safe edit boundary: 허용 접미사 둘과 이름 붙인 import 금지 갈래.
- High-risk impact: 없음(시험). 측정 도구가 엔진 보호 주문 몫(Critical)을 달 길 하나를 닫음(안전 불변식 4).
