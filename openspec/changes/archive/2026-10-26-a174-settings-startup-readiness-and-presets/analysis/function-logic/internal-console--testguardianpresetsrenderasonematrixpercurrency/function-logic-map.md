# Function Logic Map: `TestGuardianPresetsRenderAsOneMatrixPerCurrency`

- Source: `internal/console/readiness_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 비교표 수 | 등록 통화 수와 동일 | `LimitTierCards()` | 불일치 시 시험 실패 |
| 현재 통화 결과 | notice 1개, `current` 속성 | `LimitCurrencyConsequence()` | 프리셋 결과와 혼합되면 시험 실패 |
| tier 셀 | Preview From/To·실제 Axis와 동일 | `LimitTierCards().Preview` | 불일치 시 시험 실패 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | 카드에서 통화 집합 수집 | 없음 | 이후 기대값 설정 | 통화별 matrix |
| B2-B4 | matrix·행·current notice 수 불일치 | 없음 | `t.Fatalf`/`t.Errorf` | 표 구조·notice 1회 |
| B5-B9 | tier form·가정 결과·hidden 값 불일치 | 없음 | `t.Errorf` | POST 계약·가정 문맥 보존 |
| B10-B12 | Preview From/To·Axis 불일치 | 없음 | `t.Errorf` | 실제 값·방향 표시 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `fullSettingsHarness`, `h.get` | 정상 USD fixture의 daily GET 렌더 | 시험 fixture, 재시도 없음 | AST calls |
| `settingsPage.LimitTierCards` | 프리셋 값·방향의 정본 | 순수 계산, 오류 반환 없음 | `readiness.go` |
| `strings.Count`, `strings.Contains` | HTML 구조·값 직접 비교 | 불일치 시 시험 실패 | AST calls |

## State mutations and fallbacks

- GET 렌더만 검사하고 preset POST를 보내지 않음.
- 현재 통화 notice와 가정적 preset 결과 문장을 속성으로 분리함.

## Safety conclusion

- Safe edit boundary: 비교표·notice 렌더 계약 단언만 수정함.
- High-risk impact: no. Guardian 값과 저장 handler는 바꾸지 않음.
