# Function Logic Map: `TestThePopularityRankingFallsBackToTheProductCode`

- Source: `internal/candidatesrc/candidatesrc_test.go`
- Change: `fix-chase-veto-measurement`
- AST evidence: `ast.json` (revision `current`, L222–233, 분기 2개)
- Risk scan: `risk-pattern-report.md`

일부 WTS 행은 product code만 싣고 symbol이 없다. 그 행들을 버리면 목록이 조용히 줄어들고,
**목록 길이는 백분위의 분모**다.

이 change가 그 fallback을 `wtsSymbol`로 뽑아냈다 — 직전 읽기의 집합이 행과 **같은 문자열로**
키잉되어야 하기 때문이다. 그 대칭이 깨지면 fallback 행은 매 읽기마다 신규 진입으로
보고된다(`TestAWTSRowIdentifiedByItsProductCodeIsNotANewEntrantEveryTime`).


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
| 응답 | `ProductCode: "A005930"`만 있는 1행 | fake | — |
| 기대 | `Symbol == "A005930"`인 행 1개 | 이 테스트 | `t.Fatalf` |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `Read` 오류 | 없음 | `t.Fatalf` | 자체 실행 |
| B2 | 행이 없거나 심볼이 product code가 아니다 | 없음 | `t.Fatalf` | 자체 실행 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `WTSPopular(f, 30, nil).Read(...)` | 소스와 읽기 | — | ast.json calls |

## State mutations and fallbacks

- 테스트 전용. fake reader와 프로세스 내 상태만 쓴다.
- 실계좌·실브로커·네트워크 무접촉. 주문 side effect 0.

## Safety conclusion

- Safe edit boundary: 테스트 수정 — 호출부 인자 1개(`nil`). 단언 무변경.
- High-risk impact: no (테스트 전용).
