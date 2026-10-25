# Function Logic Map: `optimizationCategoryViews`

- Source: `internal/console/optimization_view.go` (203-230) — 편집 뒤(a173 rebase 후 main).
- Qualified: `optimizationCategoryViews`
- AST evidence: `ast.json` (`source_sha256` 1f54437e13fb7b62…) — AST branches 10 · returns 1
- Risk scan: `risk-pattern-report.md`
- 작성 시점: 완료 게이트 직전(2026-10-25, 종결 에이전트). tasks 3.1 의 비례 원칙 면제는 required 7 이라 gate ⑤ 가 받지 않아 사후에 만든 산출물임(FLM 이 구현 판단에 쓰이지 않음).

**역할.** 최적화 화면 왼쪽 카테고리 6 개의 표시 뷰(활성 · 사용 가능 · 상태 문구)를 만든다.

a173 편집: 상태 문구 넷(B7~B10)에서 change ID(`a044` · `a046` · `a047` · `a049`)를 빼 새 필드 `Tech` 로 옮김 — 본문에는 운영자 문장만, ID 는 접힌 기술 정보로 렌더. Available 판정 무변경.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `selected` | strategyopt.Category | 요청 쿼리(핸들러가 검증) | 미지 값은 핸들러가 overview 로 |
| `writable` | bool | exit-protection command seam 배선 여부 | 거짓 → 읽기 전용 문구 |

## Branches and early returns

| Branch | Source line | AST kind | Condition / 효과 | a173 |
|---|---|---|---|---|
| B1 | `205:2` | range | 카테고리 순회 | 무변경 |
| B2 | `207:3` | switch | 카테고리 ID switch | 무변경 |
| B3 | `208:3` | case | overview | 무변경 |
| B4 | `210:3` | case | exit-protection → Available=writable | 무변경 |
| B5 | `212:4` | if | writable → 연결됨 | 무변경 |
| B6 | `214:11` | else | !writable → 미배선 · 읽기 전용 | 무변경 |
| B7 | `217:3` | case | position-management | a173: Status 에서 ID 분리 → Tech |
| B8 | `219:3` | case | candidate-filters | a173: Status 에서 ID 분리 → Tech |
| B9 | `221:3` | case | strategy-runtime | a173: Status 에서 ID 분리 → Tech |
| B10 | `223:3` | case | performance-history | a173: Status 에서 ID 분리 → Tech |

Exact AST return positions: 229:2

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `strategyopt.Categories` | 고정 카테고리 목록 | 순수 | AST calls 204 · 205 |
| `make` · `len` · `append` | 슬라이스 | 순수 | AST calls 204 · 227 |

## State mutations and fallbacks

- 지역 슬라이스만 채워 반환. 쓰기 0.

## Safety conclusion

- Safe edit boundary: 문구 문자열 분리와 필드 하나 추가. 분기 조건 · Available 무변경.
- High-risk impact: no — 콘솔 읽기 전용 렌더(주문 · 손절 · 사이징 · 원장 · 인증 경로 미접촉).
