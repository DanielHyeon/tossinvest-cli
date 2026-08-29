# Function Logic Map: `TestThePopularityRankingReportsNoTradingFigures`

- Source: `internal/candidatesrc/candidatesrc_test.go`
- Change: `fix-chase-veto-measurement`
- AST evidence: `ast.json` (revision `current`, L199–217, 분기 4개)
- Risk scan: `risk-pattern-report.md`

WTS 인기 순위가 **싣지 않는 것**을 우리 코드가 지어내지 않는다는 것. `domain.RankedStock`에는
거래대금 자체가 없으므로 이 소스는 후보를 올릴 수는 있어도 rate·acceleration에는 절대
기여할 수 없다. 그 칸들은 "0"이 아니라 **비어** 있어야 한다 — 0은 per-source rate 시리즈가
진짜 값과 차분하게 될 조작된 데이터 포인트다.


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
| 응답 | 심볼·이름만 있는 1행 | fake | — |
| 기대 | `TradingValue`/`TradingVolume`/`Price`가 전부 빈 문자열, `rank 1/1` | 이 테스트 | `t.Errorf` |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `Read` 오류 | 없음 | `t.Fatalf` | 자체 실행 |
| B2 | 행 수가 1이 아니다 | 없음 | `t.Fatalf` | 자체 실행 |
| B3 | 세 수치 중 하나라도 채워졌다 | 없음 | `t.Errorf` | 자체 실행 |
| B4 | rank가 1/1이 아니다 | 없음 | `t.Errorf` | 자체 실행 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `WTSPopular(f, 30, nil).Read(ctx, MarketKR)` | 소스 생성과 읽기 — **`nil` clock** | — | ast.json calls |

## State mutations and fallbacks

- 테스트 전용. fake reader와 프로세스 내 상태만 쓴다.
- 실계좌·실브로커·네트워크 무접촉. 주문 side effect 0.

## Safety conclusion

- Safe edit boundary: 테스트 수정 — 호출부 인자 1개(`nil`). 단언 무변경.
- High-risk impact: no (테스트 전용).
