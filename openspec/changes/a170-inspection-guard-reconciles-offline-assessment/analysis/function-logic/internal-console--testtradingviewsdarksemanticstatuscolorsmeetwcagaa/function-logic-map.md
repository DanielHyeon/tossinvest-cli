# Function Logic Map: TestTradingViewsDarkSemanticStatusColorsMeetWCAGAA (삭제된 base 판)

- Source: `internal/console/trading_views_test.go` (170-195) — base `fd6afe23b` 판. 현재 트리에는 없음.
- AST evidence: `ast.json` (`revision: base`, `source_sha256` 95187c81f864… = `git show fd6afe23b:internal/console/trading_views_test.go | sha256sum`) — AST branches 5 · returns 0
- Risk scan: `risk-pattern-report.md`(base 판 파일 대상)
- 출처: `archive/2026-10-27-a172-console-anti-ai-redesign` 번들 복사, source sha 대조 일치 — a170 종결 주석(2026-10-28): a170 base `0d168e0a3` 가 그 change 착지보다 앞이라 창(base→워킹트리)에 들어온 함수. `git show 0d168e0a3:internal/console/trading_views_test.go | sha256sum`(revision base) = `95187c81f864…` 이 번들의 `source_sha256` 과 같아 번들을 그대로 복사함.
- 작성 시점: a172 종결(2026-10-28, 종결 에이전트). 사후 작성한 증거임 — FLM 이 구현 판단에 쓰이지 않음(review.md 「gate ⑤」).
- a043 아카이브 번들(sha `443091c0c299…`)은 a172 base 판과 sha 가 달라 쓰지 않음.

**역할(삭제 전).** 다크 미디어 블록 안 `.ok`·`.bad` 규칙 리터럴(`#22c55e`·`#f43f5e`)이 있고 각 색이 면 `#1d1d22` 위에서 4.5:1 이상인지 단언.

**a172 처분(2단계 afe962bed · 3단계 15b4ec00a).** 다크 색이 개별 규칙에서 `:root` 변수로 옮겨져 규칙 리터럴이 사라졌으므로, 이 시험을 `TestTradingViewsThemeColorsMeetWCAGInBothThemes` 로 대체·삭제함(tasks 「고정 리터럴 시험 수정 기록」). 대체 시험은 라이트·다크 두 테마에서 텍스트 8종 × 세 면 4.5:1, 의미색 fg × 자기 bg 4.5:1, UI 경계 × 세 면 3:1, 잉크 단조성을 잼 — 대상 색 2 → 8, 면 1 → 3(약화 아님).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `pageTemplates` 다크 블록 | 두 `@media (prefers-color-scheme: dark)` 사이 | `templates.go`(base) | 블록 없음 → Fatal |
| 대비 | ≥ 4.5:1 | WCAG AA | 미만 → Errorf |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | 첫 다크 블록 없음 | — | Fatal | 대체 시험 |
| B2 | 둘째 다크 블록 없음 | — | Fatal | 대체 시험 |
| B3 | 계약 2건 순회 | — | — | 대체 시험 |
| B4 | 규칙 리터럴 없음 | — | Errorf | 대체 시험(규칙 리터럴 대신 변수 값을 읽음) |
| B5 | 대비 < 4.5 | — | Errorf | 대체 시험 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `strings.Index` · `strings.Contains` | 블록·규칙 탐색 | 순수 | AST calls 172·176·188 |
| `wcagContrastRatio` | 대비 계산 | 순수 | AST calls 191 |

## State mutations and fallbacks

- 순수 시험. 상태 변경 0.

## Safety conclusion

- Safe edit boundary: 시험 전용 — 삭제·대체.
- High-risk impact: no — 시험 파일.
