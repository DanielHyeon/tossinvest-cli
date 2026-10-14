# Function Logic Map: `RiskGuardian.dailyLossFloor`

- Source: `internal/execgw/a137_daily_loss.go` (145-165, 편집 뒤)
- Qualified: `RiskGuardian.dailyLossFloor`
- AST evidence: `ast.json`(편집 뒤, `source_sha256` d0869728c1f8f72d…) · `ast-base.json`(편집 전, base `63fd0c92`, `source_sha256` 5ebfb6c220c7568b…, 편집 전 분기수 3) — 현재 branches 4 · returns 4
- Risk scan: `risk-pattern-report.md`

**역할.** 측정값과 호출자 값으로 체인에 넣을 계좌 상태를 만듦(a137 바닥 규칙).

**a137 편집.** 구현 리뷰 P3: `CompareDecimal` 오류면 측정 실패로 fail-closed(B3 추가) — 전에는 호출자 값을 유지했음. 분기 3 → 4.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| account.DailyRealizedLoss | 기준 통화 · 음수 아님 | 호출자 | 무효면 대체 없음(B2) |
| m | 이 Guardian 의 토큰 | MeasureDailyLoss | 실패 → B1 |

## Branches and early returns

| Branch | Condition (소스 줄 인용) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if err != nil {` (if 147) | — | 분기 본문 참조 | `TestA137AMeasurementFailureNeverFallsBackToTheCaller` · `TestA137APrecheckMeasurementMustBeThisGuardiansFreshTodaysOne` |
| B2 | `if !ok {` (if 152) | — | 분기 본문 참조 | `TestA137AnInvalidCallerValueIsNotReplaced` |
| B3 | `if err != nil {` (if 156) | — | 분기 본문 참조 | 분기 몸체 도달 불가(측정값 · 호출자 값 모두 정규 decimal 로 이미 검증됨) — 방어 갈래, 변이 원장에 동등으로 기록 |
| B4 | `if cmp > 0 {` (if 161) | — | 분기 본문 참조 | `TestA137TheLedgerLossRefusesAZeroCallerAndEscalates` · `TestA137ALargerValidCallerValueWins` |

Exact AST return positions: 149 · 153 · 159 · 164

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `g.measuredDailyLoss` | 토큰 확인 · 환산 | 실패 → B1 | AST calls |
| `validDailyLoss` | 호출자 값 모양 | 무효 → B2 | AST calls |
| `riskcalc.CompareDecimal` | 큰 쪽 | 오류 → B3 | AST calls |

## State mutations and fallbacks

- 반환하는 계좌 사본의 DailyRealizedLoss 만.

## Safety conclusion

- Safe edit boundary: 바닥 규칙 하나.
- High-risk impact: yes — 진입을 더 막는 방향만.
