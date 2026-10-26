# Function Logic Map: `TestAmbientMotionContract`

- Source: `internal/console/display_primitives_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `pageTemplates` CSS | 콘솔 공통 style define 전체 | `internal/console/templates.go` | animation·hover·focus 계약 누락은 시험 실패 |
| 렌더 화면 DOM | POST form을 품은 요소에는 hover glow 금지 | `consoleScreens`, 로컬 harness | 실제 form 조상 선택 시 시험 실패 |
| 신규 프리셋 비교표 | table 안 POST form을 포함 | a175 exit preset template | 기존 `.choice-tile` 문자열 검사가 아닌 실제 `.preset-compare` 역할 검사 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1–B17 | 공통 motion·color·reduced-motion CSS 계약 누락 | 없음 | `t.Error` | `TestAmbientMotionContract` |
| B18–B22 | hover target·제외 selector가 바뀜 | 없음 | `t.Error`/`t.Fatal` | `TestAmbientMotionContract` |
| B23–B31 | 렌더 DOM에서 hover 대상이 POST form을 품음 | 없음 | `t.Error` | `TestAmbientMotionContract` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| regexp·strings 검사 | CSS 계약을 구조적으로 추출 | 불일치는 시험 실패 | `ast.json` calls |
| `newSettingsWiredHarness`·`h.page` | 실제 POST form 조상 확인 | in-process, retry 없음 | `ast.json` calls |
| `walkHTML` | glow target 아래의 form 존재 판정 | 없음 | `ast.json` calls |

## State mutations and fallbacks

- 테스트 fixture·HTML·상수 CSS만 읽음. 실제 설정·주문·네트워크 mutation 없음.

## Safety conclusion

- Safe edit boundary: 새 form container `.preset-compare`를 hover glow 금지 목록에 추가하고, CSS 문자열 존재가 아닌 실제 렌더 조상 검사를 유지함.
- High-risk impact: no — 접근성·동작 회귀를 막는 테스트 변경임.
