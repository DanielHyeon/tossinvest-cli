# Function Logic Map: `TestOptimizationUsesOnePresetPreviewFlowWithoutClientDraft`

- Source: `internal/console/optimization_ui_contract_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`
- 출처: `archive/2026-10-27-a172-console-anti-ai-redesign` 번들 복사, source sha 대조 일치 — a170 종결 주석(2026-10-28): a170 base `0d168e0a3` 가 그 change 착지보다 앞이라 창(base→워킹트리)에 들어온 함수. `sha256sum internal/console/optimization_ui_contract_test.go` = `1c0503c53d86…` 이 번들의 `source_sha256` 과 같아 번들을 그대로 복사함.
- 출처: `archive/2026-10-26-a175-optimization-and-runtime-matrix-views` 번들 복사, source sha 대조 일치 — a172 종결 주석(2026-10-28): a172 base `fd6afe23b` 가 그 change 착지보다 앞이라 창(base→워킹트리)에 들어온 함수. `sha256sum internal/console/optimization_ui_contract_test.go` = `1c0503c53d86…` 이 번들의 `source_sha256` 과 같아 번들을 그대로 복사함.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 인증된 exit-protection HTML | 서버 정의 preset 3개와 preview 진입 경로 | `newDashboardHarness`·`newFakeOptimizationCommander` | DOM·계약 누락은 시험 실패 |
| 화면별 lifecycle/evidence 상태 | 최상위 상태 속성은 overview에만, 비개요는 concise summary | a175 design §2.1와 `optimizationTemplates` | 다른 화면에 가짜 속성을 되살리지 않고 overview 렌더로 검증 |
| preset preview 흐름 | native form/button 3개, client draft 없음 | 현재 route와 DOM | 개수·금지 요소·문구 누락은 시험 실패 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `html.Parse` 실패(120:2) | 없음 | `t.Fatal` | 같은 시험 |
| B2 | form 수가 3개가 아님(123:2) | 없음 | `t.Fatalf` | 같은 시험 |
| B3 | button 수가 3개가 아님(126:2) | 없음 | `t.Fatalf` | 같은 시험 |
| B4/B5 | 금지 client control 존재(129:2,130:3) | 없음 | `t.Errorf` | 같은 시험 |
| B6/B7 | preview flow·개요 lifecycle/evidence 계약 누락(134:2,140:3) | 없음 | `t.Errorf` | 같은 시험 |
| B8/B9 | client draft marker 존재(144:2,145:3) | 없음 | `t.Errorf` | 같은 시험 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `newDashboardHarness`·`h.page` | 인증된 로컬 렌더를 수집 | in-process, retry 없음 | `ast.json` calls |
| `optimizationPolicyOnly` | unrelated hypothetical markup를 제거 | 순수 test helper | `ast.json` calls |
| `html.Parse`·DOM helper | native form/button 개수를 실제 DOM에서 확인 | parse 실패는 즉시 중단 | `ast.json` calls |
| `strings.Contains` | route별 문구·금지 client state를 확인 | 문자열 부재/존재는 `t.Errorf` | `ast.json` calls |

## State mutations and fallbacks

- 테스트 fixture와 HTML DOM만 읽음. 주문·설정·레지스트리 mutation 없음.

## Safety conclusion

- Safe edit boundary: a175가 KPI를 overview에 한정한 뒤, 비개요 preset test가 사라진 `data-*` 속성을 요구하지 않도록 overview와 preset 흐름 검증을 분리함.
- High-risk impact: no — 화면 계약 회귀를 검출할 뿐 production 동작을 바꾸지 않음.
