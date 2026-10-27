# Function Logic Map: `TestTheSummaryCostsNoBrokerCall`

- Source: `internal/console/display_primitives_test.go` (359-371) — base `fd6afe23b` 판, 현재 트리 380-392 에 같은 본문
- AST evidence: `ast.json` (`revision: base`, `source_sha256` 2bcf341060f8…) — AST branches 2 · returns 0
- Risk scan: `risk-pattern-report.md`
- 작성 시점: a172 종결(2026-10-28, 종결 에이전트). 구현(2026-10-25) 뒤 사후 작성한 증거임 — FLM 이 구현 판단에 쓰이지 않음(review.md 「gate ⑤」).

**역할.** 개요(`/dashboard`)를 3회 렌더해도 브로커 보유 조회 호출 수가 늘지 않음을 단언.

**본문 무변경.** base `fd6afe23b` 판(359-371)과 줄 단위로 같음. a172 가 이 함수 바로 뒤에 새 시험들(`TestNoSideTabAccentBorders`·`TestMotionContract` 등)을 붙여 `git diff --unified=0` hunk `@@ -371,0 +393,386 @@` 의 base 쪽 앵커가 이 함수의 끝 줄(371)에 닿아 gate ⑤ 가 수정 함수로 셈. gate ⑤ 는 이 hunk 가 base 쪽에만 닿으므로 `revision: base` 증거를 요구함 — ast.json 은 base 판 파일(`git show fd6afe23b:<file>`)에서 뽑음. 현재 판 본문도 같음(줄 이동만).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 하네스 보유 조회 카운터 | 렌더 전후 동일 | `newOverviewHarness` fake | 증가 → Errorf |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `range 3` | `/dashboard` 렌더 3회 | — | 이 시험 자신 |
| B2 | 호출 수 증가 | — | Errorf | 이 시험 자신(실패 갈래) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `newOverviewHarness` · `seedJournal` · `h.authenticate` | 하네스 준비 | 실패 시 하네스가 Fatal | AST calls 360·361·362 |
| `h.holdings.count` | 호출 수 측정 | 순수 카운터 | AST calls 363·367 |
| `h.page` | 렌더 | 실패 시 하네스가 Fatal | AST calls 365 |

## State mutations and fallbacks

- 임시 하네스 저널만 씀. 실 브로커·실 계좌 접촉 0.

## Safety conclusion

- Safe edit boundary: 시험 전용.
- High-risk impact: no — 시험 파일, 본문 무변경.
