# Function Logic Map: `TestA136TheManualBuySnapshotCountsAndMustBeFresh`

- Source: `internal/journal/a136_daily_buy_test.go` (371-402, 편집 뒤)
- Qualified: `TestA136TheManualBuySnapshotCountsAndMustBeFresh`
- AST evidence: `ast.json` (revision: base, 새 base 6cd36e6d) — 본문 무편집, 같은 파일의 이웃 편집(뒤에 덧붙인 시험 · 위의 문서 주석)으로 검사기가 교체된 base 함수로 분류해 편집 전 AST 를 요구함.
- Risk scan: `risk-pattern-report.md`

**역할.** a136 시험(생산 코드 아님).

**a136 편집.** 844a042c: 픽스처를 실제 OpenAPI 세부 상태 열거로 바꾸거나(OPEN/CLOSED 그룹 라벨 제거) 같은 파일 뒤에 시험을 덧붙임 — 단언 무변경.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| — | — | 시험 | — |

## Branches and early returns

| Branch | Condition (소스 줄 인용) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if err := reserveWith(j, "d-manual", counted); !errors.Is(err, ErrReservationLimitExceeded) {` (if 383) | — | 분기 본문 참조 | `TestA136TheManualBuySnapshotCountsAndMustBeFresh` |
| B2 | `if err := reserveWith(j, "d-stale", stale); !errors.Is(err, ErrSnapshotStale) {` (if 388) | — | 분기 본문 참조 | `TestA136TheManualBuySnapshotCountsAndMustBeFresh` |
| B3 | `if err := reserveWith(j, "d-day", other); !errors.Is(err, ErrSnapshotStale) {` (if 393) | — | 분기 본문 참조 | `TestA136TheManualBuySnapshotCountsAndMustBeFresh` |
| B4 | `if err := reserveWith(j, "d-nil", nil); !errors.Is(err, ErrInvalidRequest) {` (if 396) | — | 분기 본문 참조 | `TestA136TheManualBuySnapshotCountsAndMustBeFresh` |
| B5 | `if err := reserveWith(j, "d-ok", a136ZeroManualBuy(j.Now())); err != nil {` (if 399) | — | 분기 본문 참조 | `TestA136TheManualBuySnapshotCountsAndMustBeFresh` |

Exact AST return positions: 378

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| — | fixture | — | AST |

## State mutations and fallbacks

- 시험 데이터만.

## Safety conclusion

- Safe edit boundary: 픽스처 문자열.
- High-risk impact: no(시험 전용)
