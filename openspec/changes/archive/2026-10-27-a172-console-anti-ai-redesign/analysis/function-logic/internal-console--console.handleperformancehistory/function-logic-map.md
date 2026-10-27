# Function Logic Map: `Console.handlePerformanceHistory`

- Source: `internal/console/performance_history.go` (73-95) — 편집 뒤(a177). base(`b27a1ea87`) 대비 한 문장 추가(81행).
- Qualified: `Console.handlePerformanceHistory`
- AST evidence: `ast.json` (편집 뒤 추출, `source_sha256` a917f3c70a0d7df6…) — AST branches 5 · returns 1
- Risk scan: `risk-pattern-report.md` (configured risk pattern 0)
- 출처: `archive/2026-10-25-a173-console-ux-common-cleanup` 번들 복사, source sha 대조 일치 — a172 종결 주석(2026-10-28): a172 base `fd6afe23b` 가 그 change 착지보다 앞이라 창(base→워킹트리)에 들어온 함수. `sha256sum internal/console/performance_history.go` = `a917f3c70a0d…` 이 번들의 `source_sha256` 과 같아 번들을 그대로 복사함.
- 작성 시점: 완료 게이트 직전(2026-10-25). review.md 의 `not-applicable` 면제는 required 3 이라 gate ⑤ 가 받지 않아 사후에 만든 산출물임.
- a173 종결 주석(2026-10-25): a173 base `b27a1ea87` 가 a177 착지보다 앞이라 a173 창(base→HEAD)에 들어온 함수. a173 의 편집은 0 이고 소스 sha 가 a177 번들과 같아(`sha256sum` 대조) a177 아카이브 번들(`openspec/changes/archive/2026-10-25-a177-performance-window-single-source/analysis/function-logic/`)을 그대로 복사함.

**역할.** `GET/HEAD /performance-history` 읽기 전용 화면. 근본 원인(a177 1.1): 칩 `<section>` 이 `{{if .Unwired}}` 분기 **앞**에서
`.View.Query.PeriodDays` 를 찍었는데 B2 가 거짓(미배선)이거나 B3(LoadErr)이면 `page.View` 가 영값이라 `최근 0일` 이 렌더됨.
a177 편집은 `page.Window = newPerformanceWindow(c.opts.Performance != nil, performance.DefaultQuery(c.now()))` 한 문장(81행) —
창 표기를 View 가 아니라 배선 여부 + 요청 상수에서 만들어, 미배선은 `미측정 · seam_unwired`, LoadErr 는 요청 창(30일)으로 표기. 분기 추가 0.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `r.Method` | GET · HEAD | HTTP 요청 | 그 외 405 + `Allow`, 즉시 반환 |
| `c.opts.Performance` | nil 또는 리더 | 콘솔 Options | nil → Unwired, 창 `미측정 · seam_unwired` |
| `Dashboard` 결과 | DashboardView 또는 오류 | 성과 DB 리더 | 오류 → LoadErr 한 줄, View 영값(창은 영향 없음 — a177) |

## Branches and early returns

| Branch | Source line | AST kind | Condition / 효과 | a177 |
|---|---|---|---|---|
| B1 | `74:2` | if | 비 GET/HEAD → 405 refuse, return | 무변경 |
| B2 | `82:2` | if | Performance 배선 → Dashboard 호출 | 무변경 |
| B3 | `84:3` | if | Dashboard 오류 → LoadErr | 무변경(창 표기는 이 분기와 독립이 됨) |
| B4 | `86:10` | else | 성공 → View · Attributions | 무변경 |
| B5 | `89:4` | if | attribution 읽기 오류 → AttributionErr | 무변경 |

Exact AST return positions: 78:3

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `newPerformanceWindow`(a177 추가) | 창 표기 단일 출처 | 순수 함수, I/O 0 | AST calls 81 |
| `performance.DefaultQuery` | 81행(a177 추가) 표기용 · 83행 조회용 — 같은 상수 질의 | 순수 함수 | AST calls 81 · 83 |
| `c.opts.Performance.Dashboard` | 성과 요약 조회 | 오류 → LoadErr, 재시도 없음 — 무변경 | AST calls 83 |
| `c.readPerformanceAttributions` | attribution 행 | 오류 → AttributionErr — 무변경 | AST calls 88 |
| `c.render` | 렌더 | 무변경 | AST calls 94 |

## State mutations and fallbacks

- 지역 `page` 만 채움. a177 은 `page.Window` 를 B2 앞에서 한 번 채움 — 이 값은 B2~B5 어느 갈래로 가도 바뀌지 않음. 쓰기 경로 0.

## Safety conclusion

- Safe edit boundary: 분기 밖 직선 대입 한 문장. 5 분기의 조건 · 순서 · 반환 무변경(AST 분기 수 5 동일).
- High-risk impact: no — 읽기 전용 렌더, 성과 DB 질의 · 주문 · 설정 미접촉.
