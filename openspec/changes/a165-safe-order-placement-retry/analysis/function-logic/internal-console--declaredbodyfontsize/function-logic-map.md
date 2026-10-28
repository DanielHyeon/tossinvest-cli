# Function Logic Map: `declaredBodyFontSize`

- Source: `internal/console/display_primitives_test.go` (250-252)
- AST evidence: `ast.json` — AST branches 0 · returns 1
- Risk scan: `risk-pattern-report.md`
- 출처: `archive/2026-10-27-a172-console-anti-ai-redesign` 번들 복사, source sha 대조 일치 — a170 종결 주석(2026-10-28): a170 base `0d168e0a3` 가 그 change 착지보다 앞이라 창(base→워킹트리)에 들어온 함수. `sha256sum internal/console/display_primitives_test.go` = `a6099f61c085…` 이 번들의 `source_sha256` 과 같아 번들을 그대로 복사함.
- 작성 시점: a172 종결(2026-10-28, 종결 에이전트). 구현(2026-10-25) 뒤 사후 작성한 증거임 — FLM 이 구현 판단에 쓰이지 않음(review.md 「gate ⑤」).

**역할.** 시험 헬퍼. `declaredFontSize(sheet, "body")` 한 줄 위임. a172 3단계에서 본문 크기가 `body` 규칙의 토큰으로 옮겨져 전용 파싱을 위임으로 줄임.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `sheet` | CSS 문자열 | `pageTemplates` | `declaredFontSize` 의 `ok=false` 그대로 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| (없음) | 분기 0 — 위임 한 줄 | — | `declaredFontSize` 결과 | `TestTheHeadingStepsAreDistinguishable` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `declaredFontSize` | body 규칙 크기 | 그대로 반환 | AST calls 251 |

## State mutations and fallbacks

- 순수 함수. 상태 변경 0.

## Safety conclusion

- Safe edit boundary: 시험 헬퍼.
- High-risk impact: no — 시험 파일.
