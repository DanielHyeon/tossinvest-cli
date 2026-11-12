# Function Logic Map: `strictMinuteCheckCandleSymbol`

- Source: `internal/official/strict_minute_candles.go`(212~217행, base 6dbb752f6)
- AST evidence: `ast.json`. 편집 전 사본은 `../../pre-edit-ast/strictminutecheckcandlesymbol.ast.json` 이며 내용이 같음(tasks 1.2 에서 diff 0 확인).
- Risk scan: `risk-pattern-report.md`(일치 0건)
- 작성 시점: 편집 전(tasks 1.2). 아래 분기 열거는 `ast.json` 을 옮긴 것임.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `market` | `"KR"` 또는 `"US"`(시장 값 검사는 호출자 `Client.StrictMinuteCandles` 가 먼저 함 — B3 140행 앞) | 호출자 | 공용 검사로 넘겨 거절 |
| `symbol` | KR: `^[A-Z0-9]{6}$`(`strictMinuteCandleKRSymbol`), US: `^[A-Z][A-Z0-9.]{0,9}$` | 패키지 정규식 변수(51~53행) | `StrictReasonSymbol` 거절 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (213행) | `market == "KR" && strictMinuteCandleKRSymbol.MatchString(symbol)` | 없음 | `nil`(214행) | `TestStrictMinuteCandlesAcceptsDomesticAlphanumericAndPreservesRequestSymbol`(0101N0) |
| fallthrough (216행) | B1 거짓 | 없음 | `strictMinuteCheckSymbol(market, symbol)` 의 결과 | `TestStrictMinuteCandlesRefusesBadArgumentsBeforeAnyRequest`(KR 소문자·구두점·공백·비 ASCII·길이 5·7, US 사례), 숫자 KR 수락 시험(005930) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `strictMinuteCandleKRSymbol.MatchString` | 분봉 전용 KR 영숫자 문법 | 없음(불리언) | AST 213:23 |
| `strictMinuteCheckSymbol` | 공용 문법(KR 숫자 전용·US) | `StrictReasonSymbol` 오류 | AST 216:9 |

호출자: `Client.StrictMinuteCandles` B3(140행) 한 곳(시험 제외, `grep` 과 CodeGraph — tasks 1.3).

## State mutations and fallbacks

- 상태 변경 없음. KR 영숫자가 분봉 문법에 안 맞으면 공용 검사로 「대체」되지만 공용 검사는 그보다 좁으므로(숫자 전용) 결과는 거절뿐임.

## 편집 계획 (design D1·D2)

- 공용 `strictMinuteKRSymbol` 을 `^[A-Z0-9]{6}$` 로 넓히고 `strictMinuteCandleKRSymbol` 을 지움.
- 이 함수 본문은 `return strictMinuteCheckSymbol(market, symbol)` 한 줄로 줄임. B1 이 사라짐. 분봉 리더가 받는 집합은 바뀌지 않음: 편집 전 B1 의 수락 집합(KR `^[A-Z0-9]{6}$`)과 편집 뒤 공용 검사의 KR 수락 집합이 같고, US 는 원래부터 공용 검사로 갔음.
- 주석 「The shared validator remains numeric-only for strict quote readers」는 사실이 아니게 되므로 고침.

## Safety conclusion

- Safe edit boundary: 이 함수 본문과 정규식 변수 선언 2줄, 주석. `strictMinuteCheckSymbol`·`strictQuoteIdentity`·두 호가 리더·`Client.StrictMinuteCandles` 본문은 고치지 않음.
- High-risk impact: 엔진 증거로 쓰이는 엄격 리더의 입력 문법(a143 design 결정 2 분류). 주문 경로(`internal/orderintent`)는 건드리지 않음.
