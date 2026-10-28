# Function Logic Map: `TestTheHeadingStepsAreDistinguishable`

- Source: `internal/console/display_primitives_test.go` (184-209)
- AST evidence: `ast.json` — AST branches 6 · returns 0
- Risk scan: `risk-pattern-report.md`
- 출처: `archive/2026-10-27-a172-console-anti-ai-redesign` 번들 복사, source sha 대조 일치 — a170 종결 주석(2026-10-28): a170 base `0d168e0a3` 가 그 change 착지보다 앞이라 창(base→워킹트리)에 들어온 함수. `sha256sum internal/console/display_primitives_test.go` = `a6099f61c085…` 이 번들의 `source_sha256` 과 같아 번들을 그대로 복사함.
- 작성 시점: a172 종결(2026-10-28, 종결 에이전트). 구현(2026-10-25) 뒤 사후 작성한 증거임 — FLM 이 구현 판단에 쓰이지 않음(review.md 「gate ⑤」).
- 출처: `archive/2026-10-28-a170-inspection-guard-reconciles-offline-assessment/analysis/function-logic/internal-console--testtheheadingstepsaredistinguishable` 번들 복사, source sha 대조 일치 — a164 종결 주석(2026-10-28): a164 base `4cd7af19d` 가 그 함수를 고친 형제 착지보다 앞이라 창(base→워킹트리)에 들어온 함수. `sha256sum internal/console/display_primitives_test.go` = `a6099f61c085…` 이 번들의 `source_sha256`(revision current)과 같아 번들을 그대로 복사함.

**역할.** 콘솔 스타일시트(`pageTemplates`)의 제목 단계(h1>h2>h3>본문)가 가장 좁은·넓은 뷰포트 양쪽에서 구분되는지, 본문 최소 크기가 15px 이상인지 단언하는 시험.

a172 편집(2단계 afe962bed·3단계 15b4ec00a): 크기가 `clamp()`·`--text-N` 토큰으로 바뀌어 리터럴 하나를 읽던 비교를 최소(lo)·최대(hi) 두 지점 비교로 바꿈(tasks 「고정 리터럴 시험 수정 기록」 — 검사 지점 1 → 2, 약화 아님).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `pageTemplates` | 콘솔 CSS 문자열 | `internal/console/templates.go` | 크기 선언 없음 → Errorf/Fatal |
| 본문 최소 크기 | ≥ 15px | 시험 상수 | 미만 → Errorf |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `range` h1·h2·h3 | lo/hi 맵 채움 | — | 이 시험 자신 |
| B2 | 단계 크기 선언 없음(`!ok`) | — | Errorf + continue | 이 시험 자신(실패 갈래) |
| B3 | 본문 크기 선언 없음 | — | Fatal | 이 시험 자신(실패 갈래) |
| B4 | `range` narrowest·widest | — | — | 이 시험 자신 |
| B5 | 단계 순서 깨짐 | — | Errorf | 이 시험 자신(실패 갈래) |
| B6 | 본문 최소 < 15 | — | Errorf | 이 시험 자신(실패 갈래) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `declaredFontSize` | 단계별 크기 범위 | `ok=false` → B2 | AST calls 187 |
| `declaredBodyFontSize` | 본문 크기 범위 | `ok=false` → B3 | AST calls 194 |
| `t.Errorf` · `t.Fatal` | 계약 위반 보고 | 시험 실패 | AST calls 189·196·202·206 |

## State mutations and fallbacks

- 지역 맵만 채움. 제품 상태·파일·원장 변경 0.

## Safety conclusion

- Safe edit boundary: 시험 전용 — CSS 문자열 파싱과 비교만.
- High-risk impact: no — 시험 파일, 주문·손절·사이징·원장·인증 경로 미접촉.
