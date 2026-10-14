# Function Logic Map: `ReserveRequest.build`

- Source: `internal/journal/reservations.go` (565-672, base 7b073f55)
- Qualified: `ReserveRequest.build`
- 위치: `analysis/ast-evidence/` — **proposal 단계 근거(게이트 번들 아님)**. 편집하게 되는 함수는 구현 착수 때 `analysis/function-logic/` 에 정식 번들을 만든다.
- AST evidence: `ast.json` (`source_sha256` 933d37c95a6212fe…) — AST branches 26 · returns 21
- Risk scan: `risk-pattern-report.md`
- 용도: a136 design D2·D3 의 주장 「예약 종류는 닫힌 집합(OPEN_EXPOSURE · DAILY_LOSS · CASH)이고 거래일 칸은 DAILY_LOSS 전용이다」의 근거. 새 예약 종류를 택하면(D2-안 a) 이 함수의 B10 · B12 · B13 을 편집하게 되고 이 번들을 편집 뒤로 재추출한다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `DecisionID` · `AccountRef` | 비어 있지 않음 | 발급자 | `ErrInvalidRequest` |
| `SnapshotAsOf` · `ObservedVersion` | 0 아님 · >0 | 발급자 수집 | `ErrInvalidRequest` |
| `Reservations[].Kind` | `validReservationKind` (OPEN_EXPOSURE · DAILY_LOSS · CASH) | 코드 + DDL CHECK | `ErrInvalidRequest` |
| `Reservations[].TradingDay` | DAILY_LOSS 는 필수, 나머지는 금지 | 이 함수 | `ErrInvalidRequest` |
| `SnapshotUsage` · `Limits` | 예약하는 종류마다 하나씩, 같은 통화, limit > 0 | 발급자 | `ErrInvalidRequest` (부재 = 미지, 0 아님) |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | 요청 머리 `switch` (576) | — | — | B2~B6 |
| B2 | `decisionID == ""` (577) | — | invalid | `TestReservationRequiresARecordedUnexpiredDecision` |
| B3 | `accountRef == ""` (579) | — | invalid | 기존 계좌 범위 시험 |
| B4 | `SnapshotAsOf.IsZero()` (581) | — | invalid | `TestStaleSnapshotIsRefused` 계열 |
| B5 | `ObservedVersion <= 0` (583) | — | invalid | `TestReservationVersionStartsAboveZero` |
| B6 | `len(Reservations) == 0` (586) | — | invalid | 기존 |
| B7 | `range Reservations` (590) | 행 정규화 | — | B8~B16 |
| B8 | 행별 `switch` (595) | — | — | B9~B13 |
| B9 | `id == ""` (596) | — | invalid | 기존 |
| B10 | `!validReservationKind(kind)` (598) | — | invalid — **닫힌 종류 집합** | `TestUnknownAggregateIsRefusedRatherThanTreatedAsZero` |
| B11 | `currency == ""` (600) | — | invalid | `TestMixedCurrenciesAreRefused` |
| B12 | `kind == DAILY_LOSS && day == ""` (602) | — | invalid — 일일 종류는 거래일 필수 | `TestDailyLossReservationNeedsItsTradingDay` |
| B13 | `kind != DAILY_LOSS && day != ""` (605) | — | invalid — **거래일 칸은 DAILY_LOSS 전용** | `TestDailyLossReservationNeedsItsTradingDay` |
| B14 | 금액 `nonNegativeDecimal` 실패 (610) | — | 오류 | `TestReservationArithmeticIsExact` 계열 |
| B15 | 종류별 누계 첫 행 (618) | running="0" | — | `TestReservationArithmeticIsExact` |
| B16 | 누계 덧셈 실패 (622) | — | 오류 | 없음(정규화 뒤 도달 어려움) |
| B17 | `range SnapshotUsage` (629) | usage 맵 | — | B18 |
| B18 | usage 정규화 실패 (631) | — | 오류 | `TestUnknownAggregateIsRefusedRatherThanTreatedAsZero` |
| B19 | `range Limits` (636) | limits 맵 | — | B20 |
| B20 | limit 정규화 실패 (638) | — | 오류 | 기존 |
| B21 | 예약 행마다 usage·limit 대조 (647) | — | — | B22~B26 |
| B22 | usage 부재 (649) | — | invalid — 부재는 0 이 아님 | `TestUnknownAggregateIsRefusedRatherThanTreatedAsZero` |
| B23 | limit 부재 (654) | — | invalid — 부재는 무제한이 아님 | 기존 |
| B24 | 통화 불일치 (657) | — | invalid — journal 은 환산하지 않음 | `TestMixedCurrenciesAreRefused` |
| B25 | limit 비교 실패 (663) | — | 오류 | 없음 |
| B26 | `limit <= 0` (666) | — | invalid | 기존 |

Exact AST return positions: 578 · 580 · 582 · 584 · 587 · 597 · 599 · 601 · 603 · 606 · 611 · 623 · 632 · 639 · 650 · 655 · 658 · 664 · 667 · 671 외(AST 21개)

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `validReservationKind` | 닫힌 종류 집합 | bool | reservations.go:851-858 |
| `nonNegativeDecimal` · `normaliseAggregate` | 금액 정규화 | `ErrInvalidRequest` | AST calls |
| `riskcalc.AddDecimal` · `CompareDecimal` | 정확 decimal | 오류 반환 | AST calls |

## State mutations and fallbacks

- 지역 `plan` 만. DB 쓰기 없음. fallback 없음(전부 fail-closed).

## Safety conclusion

- Safe edit boundary: 새 종류를 더하면 B10(집합) · B12 · B13(거래일 전용 규칙)을 함께 바꿔야 하고 DDL `CHECK (kind IN (...))` 도 바뀌어야 함(execution_contract.go:169) — 스키마 v36 테이블 재작성. 이것이 design D2 가 「새 예약 종류」 대신 「트랜잭션 안 합산」을 추천하는 이유.
- High-risk impact: yes (편집 시)
