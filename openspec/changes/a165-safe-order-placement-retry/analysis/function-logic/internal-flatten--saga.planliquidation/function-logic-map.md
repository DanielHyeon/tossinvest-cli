# Function Logic Map: `Saga.planLiquidation`

- Source: `internal/flatten/liquidate.go` (279-327)
- Qualified function: `Saga.planLiquidation`
- Revision: `current` (a165 base `34e12967e`, `source_sha256` ef9ad2603edf9c0b…)
- AST evidence: `ast.json` — AST branches 6, 반환 3, 호출 14
- Risk scan: `risk-pattern-report.md`
- 실행 칸 측정: 편집 전 기준선 `go test -count=1 -coverprofile` (execgw·official·trading·flatten 패키지 전체, journal `-run 'Classify|SendTracker|Dispatch'`, app/engine `-run 'Clear|Cancel'` — journal·engine 은 부분 실행이라 count=0 이 「미실행」을 뜻하지 않을 수 있음, cmd/tossctl 은 미측정) — 분기 줄에서 시작하는 본문 블록의 실행 수. 시험 이름별 귀속은 GREEN 재추출에서 함.

## 역할 (a165)

청산 매도 계획. B5(307) `UnsettledCancelSymbols` 가 준 held 종목은 계획에서 뺀다 — 취소가 정착 안 된 종목의 oversell 방지 가드. FlattenStepFailed 는 held 가 아니라서 매도가 계획된다(1.3 재리뷰 F1 의 사슬).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| snapshot | reconcile.Snapshot | Saga.Run | B1 오류 |

## Branches and early returns

AST 열거 순서 그대로. Condition 칸은 AST 좌표 줄의 원문.

Exact AST return positions: 284:3, 325:44, 326:2

| Branch | Coordinate | Condition | Mutation/side effect | Return/error | Executed (baseline) |
|---|---|---|---|---|---|
| B1 | `if at 283:2` | if err != nil { | 아래 「State mutations」 | 소스 줄 283 이하 | 미실행 — 패키지 스위트 count=0 |
| B2 | `range at 288:2` | for _, holding := range snapshot.Holdings { | 아래 「State mutations」 | 소스 줄 288 이하 | 실행됨 — 패키지 스위트 count=1 |
| B3 | `if at 290:3` | if quantity <= 0 { | 아래 「State mutations」 | 소스 줄 290 이하 | 미실행 — 패키지 스위트 count=0 |
| B4 | `if at 297:3` | if market == "" { | 아래 「State mutations」 | 소스 줄 297 이하 | 미실행 — 패키지 스위트 count=0 |
| B5 | `if at 307:3` | if state, blocked := held[symbol]; blocked { | 아래 「State mutations」 | 소스 줄 307 이하 | 실행됨 — 패키지 스위트 count=1 |
| B6 | `if at 317:3` | if err := s.priceAndSize(ctx, &target); err != nil { | 아래 「State mutations」 | 소스 줄 317 이하 | 실행됨 — 패키지 스위트 count=1 |

## Calls and live bindings

`UnsettledCancelSymbols`, `priceAndSize`.

AST 호출 이름 전체: `MarketOf`, `append`, `currencyFor`, `err.Error`, `fmt.Sprintf`, `parseDecimal`, `s.Journal.UnsettledCancelSymbols`, `s.priceAndSize`, `sort.Slice`, `string`, `strings.ToUpper`, `strings.TrimSpace`.

## State mutations and fallbacks

- 없음.

## Safety conclusion

- 편집 계획 없음.
