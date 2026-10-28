# Function Logic Map: `declaredFontSize`

- Source: `internal/console/display_primitives_test.go` (230-248)
- AST evidence: `ast.json` — AST branches 4 · returns 5
- Risk scan: `risk-pattern-report.md`
- 출처: `archive/2026-10-27-a172-console-anti-ai-redesign` 번들 복사, source sha 대조 일치 — a170 종결 주석(2026-10-28): a170 base `0d168e0a3` 가 그 change 착지보다 앞이라 창(base→워킹트리)에 들어온 함수. `sha256sum internal/console/display_primitives_test.go` = `a6099f61c085…` 이 번들의 `source_sha256` 과 같아 번들을 그대로 복사함.
- 작성 시점: a172 종결(2026-10-28, 종결 에이전트). 구현(2026-10-25) 뒤 사후 작성한 증거임 — FLM 이 구현 판단에 쓰이지 않음(review.md 「gate ⑤」).

**역할.** 시험 헬퍼. 스타일시트에서 맨 요소 규칙(`\n<element> {`)의 font-size 범위(px, lo·hi)를 읽음.

a172 편집(2·3단계): 단일 리터럴 대신 `clamp()` 최소·최대와 `font-size: var(--text-N)` 토큰(→ `tokenSize`)을 읽도록 확장. 평범한 `Nrem` 도 그대로 읽음.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `sheet` | CSS 문자열 | `pageTemplates` | 규칙 없음 → `ok=false` |
| `element` | `h1`·`h2`·`h3`·`body` | 호출자 | 규칙 없음 → `ok=false` |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | 요소 규칙 시작 없음 | — | `0,0,false` | 호출 시험 |
| B2 | 규칙 닫는 `}` 없음 | — | `0,0,false` | 호출 시험 |
| B3 | `var(--text-N)` 매치 | — | `tokenSize` 결과 | 호출 시험 |
| B4 | `Nrem` 매치 | — | `remRange(m,m)` | 호출 시험 |
| (끝) | 둘 다 아님 | — | `0,0,false` | 호출 시험 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `strings.Index` | 규칙 경계 | -1 → B1/B2 | AST calls 231·236 |
| `varSizePattern.FindStringSubmatch` · `tokenSize` | 토큰 크기 | 매치 없음 → 다음 갈래 | AST calls 241·242 |
| `remSizePattern.FindStringSubmatch` · `remRange` | rem 크기 | 매치 없음 → 끝 | AST calls 244·245 |

## State mutations and fallbacks

- 순수 함수. 상태 변경 0.

## Safety conclusion

- Safe edit boundary: 시험 헬퍼 — 문자열 파싱만.
- High-risk impact: no — 시험 파일.
