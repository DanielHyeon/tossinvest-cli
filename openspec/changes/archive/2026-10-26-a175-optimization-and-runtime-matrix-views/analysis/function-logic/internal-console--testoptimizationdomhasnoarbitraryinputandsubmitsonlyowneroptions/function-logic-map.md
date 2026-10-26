# Function Logic Map: `TestOptimizationDOMHasNoArbitraryInputAndSubmitsOnlyOwnerOptions`

- Source: `internal/console/optimization_ui_contract_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 인증된 overview HTML | lifecycle/evidence 상태는 최상위 overview에만 존재 | `h.page("/optimization")`와 `optimizationTemplates` | marker 누락은 시험 실패 |
| 인증된 exit-protection HTML | owner 등록 option만 hidden input으로 제출 | `commander.view.Registry`와 route 렌더 | parse/DOM 계약 위반은 시험 실패 |
| readonly category DOM | 입력·편집 affordance 없음 | native HTML과 DOM walk | 금지 element·attribute·문구는 시험 실패 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1/B2 | overview marker 반복·누락(19:2,20:3) | 없음 | `t.Errorf` | 같은 시험 |
| B3 | exit-protection HTML parse 실패(27:2) | 없음 | `t.Fatal` | 같은 시험 |
| B4/B5 | DOM node walk와 child 반복(31:2,32:3) | 없음 | 없음 | 같은 시험 |
| B6 | non-element node(37:3) | 없음 | return | 같은 시험 |
| B7/B8 | textarea 발견(40:3,41:3) | 없음 | `t.Errorf` | 같은 시험 |
| B9 | input 발견(43:3) | 없음 | 아래 입력 속성 검사 | 같은 시험 |
| B10 | hidden/checkbox 외 input type(45:4) | 없음 | `t.Errorf` | 같은 시험 |
| B11 | registry 밖 option_id 제출(50:4) | 없음 | `t.Errorf` | 같은 시험 |
| B12 | contenteditable 발견(54:3) | 없음 | `t.Errorf` | 같은 시험 |
| B13/B14 | 금지 text affordance 반복·발견(58:2,59:3) | 없음 | `t.Errorf` | 같은 시험 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `newDashboardHarness`·`h.page` | 인증된 로컬 route HTML을 수집 | in-process, retry 없음 | `ast.json` calls |
| `optimizationPolicyOnly` | unrelated hypothetical markup를 제거 | 순수 test helper | `ast.json` calls |
| `html.Parse`·`walkHTML` | native DOM controls를 실제 element로 검사 | parse 실패는 즉시 중단 | `ast.json` calls |
| `Registry.Fields` | owner 등록 option 집합을 결정 | fixture registry only | `ast.json` calls |
| `strings.Contains` | marker와 입력 affordance 문자열을 검사 | 불일치는 `t.Errorf` | `ast.json` calls |

## State mutations and fallbacks

- 테스트 fixture와 HTML DOM만 읽음. 주문·설정·레지스트리 mutation 없음.

## Safety conclusion

- Safe edit boundary: a175의 lifecycle/evidence marker를 비개요 exit route가 아니라 overview에 한정하고, owner option·임의 입력 금지 검사를 유지함.
- High-risk impact: no — production 동작을 바꾸지 않고 mutation form의 DOM 계약 회귀만 검출함.
