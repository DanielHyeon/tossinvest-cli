# Function Logic Map: `TestTheCurrencyConsequenceIsRenderedAndNotFolded`

- Source: `internal/console/settings_card_test.go` (267-283) — base `fd6afe23b` 판, 현재 트리 275-291 에 같은 본문
- AST evidence: `ast.json` (`revision: base`, `source_sha256` 962152c64007…) — AST branches 4 · returns 0
- Risk scan: `risk-pattern-report.md`
- 출처: `archive/2026-10-27-a172-console-anti-ai-redesign` 번들 복사, source sha 대조 일치 — a170 종결 주석(2026-10-28): a170 base `0d168e0a3` 가 그 change 착지보다 앞이라 창(base→워킹트리)에 들어온 함수. `git show 0d168e0a3:internal/console/settings_card_test.go | sha256sum`(revision base) = `962152c64007…` 이 번들의 `source_sha256` 과 같아 번들을 그대로 복사함.
- 작성 시점: a172 종결(2026-10-28, 종결 에이전트). 사후 작성한 증거임 — FLM 이 구현 판단에 쓰이지 않음(review.md 「gate ⑤」).
- 같은 이름의 a174 아카이브 번들도 `revision: base` 지만 a174 base 의 파일 sha(`cf628bf01a4f…`)라 a172 base 판(`962152c64007…`)과 달라 새로 만듦.
- 출처: `archive/2026-10-28-a170-inspection-guard-reconciles-offline-assessment/analysis/function-logic/internal-console--testthecurrencyconsequenceisrenderedandnotfolded` 번들 복사, source sha 대조 일치 — a164 종결 주석(2026-10-28): a164 base `4cd7af19d` 가 그 함수를 고친 형제 착지보다 앞이라 창(base→워킹트리)에 들어온 함수. `git show 4cd7af19d:internal/console/settings_card_test.go | sha256sum` = `962152c64007…` 이 번들의 `source_sha256`(revision base)과 같아 번들을 그대로 복사함.

**역할.** 일일 설정 화면(`pathSettingsDaily`)이 한도 통화의 결과를 접히지 않는 `.notice` 로 말하고, 짝 준비 항목(계좌 기준 통화·KR·US·FX·같은 구현 웨이브)을 이름으로 대는지 단언.

**본문 무변경.** base `fd6afe23b` 판(267-283)과 줄 단위로 같음. a174 `aa2864abe` 가 이 함수 바로 뒤에 새 시험을 붙여 `git diff --unified=0` hunk `@@ -284,0 +293,51 @@` 의 base 쪽 앵커가 함수 끝에 닿아 gate ⑤ 가 수정 함수로 셈. a172 의 편집은 0. gate ⑤ 는 이 hunk 가 base 쪽에만 닿으므로 `revision: base` 증거를 요구함 — ast.json 은 base 판 파일(`git show fd6afe23b:<file>`)에서 뽑음. 현재 판 본문도 같음(줄 이동만).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 렌더 HTML | `data-currency-consequence=` 있음 | 설정 템플릿 | 없음 → Fatal |
| 고지 모양 | 앞 60자 안 `class="notice"` | 설정 템플릿 | 없음 → Error |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | 고지 표식 없음 | — | Fatal | 이 시험 자신(실패 갈래) |
| B2 | `.notice` 아님 | — | Error | 이 시험 자신(실패 갈래) |
| B3 | 요구 단어 순회 | — | — | 이 시험 자신 |
| B4 | 단어 빠짐 | — | Errorf | 이 시험 자신(실패 갈래) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `limitsHarness` · `fullLimits` · `h.authenticate` · `h.page` | 하네스 렌더 | 실패 시 하네스가 Fatal | AST calls 268·269·270 |
| `strings.Index` · `strings.Contains` · `max` | HTML 검사 | 순수 | AST calls 271·275·279 |

## State mutations and fallbacks

- fake 한도 저장소만 씀. 실 설정·토글 변경 0.

## Safety conclusion

- Safe edit boundary: 시험 전용, 본문 무변경.
- High-risk impact: no — 시험 파일.
