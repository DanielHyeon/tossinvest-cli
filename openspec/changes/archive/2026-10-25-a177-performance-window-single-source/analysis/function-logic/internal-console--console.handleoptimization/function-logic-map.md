# Function Logic Map: `Console.handleOptimization`

- Source: `internal/console/optimization.go` (66-127) — 편집 뒤(a177). base(`b27a1ea87`) 대비 한 문장 추가(85행).
- Qualified: `Console.handleOptimization`
- AST evidence: `ast.json` (편집 뒤 추출, `source_sha256` 387f5d445be08e92…) — AST branches 14 · returns 1
- Risk scan: `risk-pattern-report.md` (configured risk pattern 0)
- 작성 시점: 완료 게이트 직전(2026-10-25). review.md 의 `not-applicable` 면제는 required 3 이라 gate ⑤ 가 받지 않아(`check_analysis._verdict` — 면제는 required 0 일 때만) 사후에 만든 산출물임. a177 편집은 분기를 더하지 않아 구현 판단을 바꾸지 않음.

**역할.** `GET/HEAD /optimization` 읽기 전용 화면의 뷰 모델을 만들어 렌더함. a177 편집은
`page.PerformanceWindow = newPerformanceWindow(c.opts.Performance != nil, performance.DefaultQuery(c.now()))` 한 문장(85행) —
성과 이력 화면과 같은 값 객체로 조회 창 문구를 만들게 함. 조건문 · early return · 쓰기 경로 추가 0.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `r.Method` | GET · HEAD | HTTP 요청 | 그 외 405 + `Allow` 헤더, 즉시 반환 |
| `category` 질의 | `strategyopt.ParseCategory` 가 아는 값 또는 빈 값 | 요청 URL | 모르는 값 → 개요 + 경고, 동작 없음 |
| `c.opts.Performance` | nil(미배선) 또는 리더 | 콘솔 Options | nil → 창 `미측정 · seam_unwired`(a177), 리더 호출 0 |
| `performance.DefaultQuery(c.now())` | PeriodDays 30 · 전체 시장 · 전체 lane · complete only | `internal/performance/query.go` | 상수 질의 — 실패 없음 |

## Branches and early returns

| Branch | Source line | AST kind | Condition / 효과 | a177 |
|---|---|---|---|---|
| B1 | `67:2` | if | 비 GET/HEAD → 405 refuse, return | 무변경 |
| B2 | `74:2` | if | category 빈 값 → known=true | 무변경 |
| B3 | `81:2` | if | !known → Warning | 무변경 |
| B4 | `87:2` | if | ExitPolicies 배선 → Load | 무변경 |
| B5 | `89:3` | if | Load 오류 → LegacyLoadErr | 무변경 |
| B6 | `91:10` | else | Load 성공 → LegacyCurrent | 무변경 |
| B7 | `95:2` | if | Optimization 배선 → Read | 무변경 |
| B8 | `97:3` | if | Read 오류 → LifecycleErr | 무변경 |
| B9 | `99:10` | else | Read 성공 → Snapshot · Evidence · Fields | 무변경 |
| B10 | `108:2` | range | 등록 공통 정책 순회 | 무변경 |
| B11 | `110:3` | if | Desired 빈 값 → legacy 정책으로 대체 | 무변경 |
| B12 | `117:2` | if | Protections 배선 → List | 무변경 |
| B13 | `120:3` | if | List 오류 → ProtectionLoadErr | 무변경 |
| B14 | `122:10` | else | List 성공 → Protections | 무변경 |

Exact AST return positions: 71:3

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `newPerformanceWindow`(a177 추가) | 조회 창 표시 값 — 성과 이력 화면과 같은 출처 | 순수 함수, 오류 없음, I/O 0 | AST calls 85 · `performance_history.go` |
| `performance.DefaultQuery`(a177 추가) | 서버 조회 기본값(30일) | 순수 함수 | AST calls 85 |
| `c.opts.ExitPolicies.Load` · `c.opts.Optimization.Read` · `c.opts.Protections.List` | 기존 읽기 seam | 오류는 화면 오류 줄로, 재시도 없음 — 무변경 | AST calls |
| `c.render` | 템플릿 렌더 | 무변경 | AST calls |

## State mutations and fallbacks

- 지역 `page` 만 채움. 저장 · 주문 · 설정 쓰기 0. a177 은 `page.PerformanceWindow` 필드 하나를 채울 뿐 fallback 추가 없음.

## Safety conclusion

- Safe edit boundary: 분기 밖 직선 대입 한 문장. 기존 14 분기의 조건 · 순서 · 반환 무변경(편집 전후 AST 분기 수 동일 14).
- High-risk impact: no — 읽기 전용 렌더, 주문 · 손절 · 사이징 · 원장 · 인증 경로 미접촉.
