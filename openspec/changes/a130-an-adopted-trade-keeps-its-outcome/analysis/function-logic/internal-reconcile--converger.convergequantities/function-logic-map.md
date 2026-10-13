# Function Logic Map: `Converger.ConvergeQuantities`

- Source: `internal/reconcile/converge.go` (편집 전 `142`–`273`, base `233d5d9c`)
- AST evidence: `ast.json` (`source_sha256` 00a784b3f6b3a3ef…) — branches 15 · returns 10
- Risk scan: `risk-pattern-report.md`
- 작성 시점: **편집 전**(2판, 리뷰 F-5 — 설계가 이 함수 내부를 근거로 씀).

**역할.** 대사 수량 불일치를 계좌 값으로 수렴 — 인스턴스마다 ApplyPositionAdjustment.
호출자: 생산 대사 루프.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 함수 인자 · 같은 트랜잭션의 원장 | — | 원장 | 각 분기 표 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `:144` `if len(diff.Quantities) == 0 {` | — | — | 무편집 갈래 |
| B2 | `:147` `if c == nil \|\| c.Journal == nil {` | — | — | 무편집 갈래 |
| B3 | `:152` `if account == "" {` | — | — | 무편집 갈래 |
| B4 | `:156` `if asOf == "" {` | — | — | 무편집 갈래 |
| B5 | `:173` `if len(credited) == 0 {` | — | — | 무편집 갈래 |
| B6 | `:178` `if c.Credit != nil {` | — | — | 무편집 갈래 |
| B7 | `:194` `if err != nil {` | — | — | 무편집 갈래 |
| B8 | `:197` `for _, mismatch := range diff.Quantities {` | — | — | 무편집 갈래 |
| B9 | `:200` `if reason != "" {` | — | — | 무편집 갈래 |
| B10 | `:206` `if err != nil {` | — | — | 무편집 갈래 |
| B11 | `:228` `if errors.Is(err, journal.ErrAdjustmentStale) {` | — | — | 무편집 갈래 |
| B12 | `:233` `if err != nil {` | — | — | 무편집 갈래 |
| B13 | `:244` `if result.ClosedExitState {` | — | — | 무편집 갈래 |
| B14 | `:251` `if c.Alert == nil {` | — | — | 무편집 갈래 |
| B15 | `:254` `if err := c.Alert.ManagedPositionClosedExternally(ctx, ManagedCloseAlert{` | — | — | 무편집 갈래 |

## Calls and live bindings

AST `calls` 목록 참조. Live config binding: 없음.

## State mutations and fallbacks

- 아래 「a130 편집 계획」 의 근거 자리 참조.

## a130 편집 계획 (2판)

**인용 근거만 · 무편집(a132 의 대상).** 근거 자리: B10~B12 사이 `:209-226` 요청 리터럴의 `NewAvgPrice: ""`(:221).

## Safety conclusion

- Safe edit boundary: 위 계획만.
- High-risk impact: no(무편집) — 인용 근거.
