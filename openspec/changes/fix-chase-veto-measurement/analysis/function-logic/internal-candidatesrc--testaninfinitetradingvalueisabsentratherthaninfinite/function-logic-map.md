# Function Logic Map: `TestAnInfiniteTradingValueIsAbsentRatherThanInfinite`

- Source: `internal/candidatesrc/candidatesrc_test.go`
- Change: `fix-chase-veto-measurement`
- AST evidence: `ast.json` (revision `current`, L328–351, 분기 5개)
- Risk scan: `risk-pattern-report.md`

`internal/official`의 `parseDecimal`은 err가 nil일 때 `ParseFloat`가 만든 것을 그대로
돌려주고, `ParseFloat`는 `"NaN"`·`"Inf"`·`"Infinity"`를 불평 없이 받는다. 순진하게
포맷하면 그대로 왕복하고, **무한한 거래대금은 비교되는 모든 임계를 통과한다** — 조작된
최대 신호이고, 관측 검증기가 순위에 대해 이미 거부하는 실패 형태다.


## 이 테스트에 대한 이 change의 편집

**`nil` 인자 하나**다. `OfficialRanking`·`WTSPopular`·`Panel`이 `clock.Clock`을 받게 되었고,
이 파일의 호출부가 전부 `nil`을 넘긴다 — 시스템 시계를 뜻한다. 단언·fixture·기대값은 하나도
바뀌지 않았다.

`nil`을 넘기는 것이 옳은 이유는 **이 테스트들이 시간에 대한 것이 아니기 때문**이다. 기억의
나이 상한을 재는 것은 `reading_validity_test.go`이고 그쪽은 `clock.NewFake`로 시각을 직접
몬다. 여기서 fake clock을 쓰면 이 테스트들이 무엇을 재는지가 흐려진다.

부수적으로 이 편집은 `clk == nil` → `clock.System()` 대체 경로를 **모든 생성자 호출에서**
지나가게 만든다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 응답 | `TradingAmount: +Inf`, `TradingVolume: NaN`, `LastPrice: 70000` | fake | — |
| 기대 | 앞 둘은 빈 문자열, 가격은 보존 | 이 테스트 | `t.Errorf` |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `mkErr != nil` | 없음 | `t.Fatalf` | 자체 실행 |
| B2 | `Read` 오류 | 없음 | `t.Fatalf` | 자체 실행 |
| B3 | 무한 거래대금이 통과했다 | 없음 | `t.Errorf` | 자체 실행 |
| B4 | NaN 거래량이 통과했다 | 없음 | `t.Errorf` | 자체 실행 |
| B5 | 유한한 가격이 손상됐다 | 없음 | `t.Errorf` | 자체 실행 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `OfficialRanking(f, nil, RankingTradingAmount, 100, nil)` | 소스 | — | ast.json calls |
| `decimal(...)`(간접) | NaN/Inf → 빈 문자열 | 순수 | ast.json calls |

## State mutations and fallbacks

- 테스트 전용. fake reader와 프로세스 내 상태만 쓴다.
- 실계좌·실브로커·네트워크 무접촉. 주문 side effect 0.

## Safety conclusion

- Safe edit boundary: 테스트 수정 — 호출부 인자 1개(`nil`). 단언 무변경.
- High-risk impact: no (테스트 전용).
