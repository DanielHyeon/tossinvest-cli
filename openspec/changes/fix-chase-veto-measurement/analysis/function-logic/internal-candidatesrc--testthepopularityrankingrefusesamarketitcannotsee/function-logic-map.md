# Function Logic Map: `TestThePopularityRankingRefusesAMarketItCannotSee`

- Source: `internal/candidatesrc/candidatesrc_test.go`
- Change: `fix-chase-veto-measurement`
- AST evidence: `ast.json` (revision `current`, L357–367, 분기 2개)
- Risk scan: `risk-pattern-report.md`

가드가 `Panel`뿐 아니라 **소스 자신**에 있다는 것. 손으로 패널을 만들거나 KR 패널을 US
스캔에 재사용하는 호출자는 그렇지 않으면 한국 행을 US로 파일링하고, 스캔은 응답한 소스를
**요청받은 시장에 대한 증거**로 취급한다.

`f.gotSize != 0` 단언이 요점이다 — 거부가 클라이언트를 부르기 **전에** 일어나야 한다.


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
| 시장 | `MarketUS` | 이 테스트 | 이 소스는 KR 전용 |
| 기대 | 오류이고 클라이언트를 부르지 않았다 | 이 테스트 | `t.Fatal`/`t.Error` |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | US 스캔이 답을 받았다 | 없음 | `t.Fatal` | 자체 실행 |
| B2 | `f.gotSize != 0` — 거부 전에 클라이언트를 불렀다 | 없음 | `t.Error` | 자체 실행 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `WTSPopular(f, 30, nil).Read(ctx, MarketUS)` | 거부 대상 — **`nil` clock** | — | ast.json calls |

## State mutations and fallbacks

- 테스트 전용. fake reader와 프로세스 내 상태만 쓴다.
- 실계좌·실브로커·네트워크 무접촉. 주문 side effect 0.

## Safety conclusion

- Safe edit boundary: 테스트 수정 — 호출부 인자 1개(`nil`). 단언 무변경.
- High-risk impact: no (테스트 전용).
