# Function Logic Map: `TestTheUSPanelDoesNotIncludeTheKoreanPopularityRanking`

- Source: `internal/candidatesrc/candidatesrc_test.go`
- Change: `fix-chase-veto-measurement`
- AST evidence: `ast.json` (revision `current`, L243–260, 분기 3개)
- Risk scan: `risk-pattern-report.md`

시장별 패널 멤버십. 정리가 아니라 안전이다 — 스캔은 "이 소스가 답했고 심볼을 싣지 않았다"를
심볼이 목록을 떠난 증거로 읽고 냉각 시계를 시작하며, 그 끝은 만료이고 만료는 `first_seen_at`을
버린다. 구조적으로 그 시장을 볼 수 없는 소스는 다른 시장의 행으로 답하므로, 모든 US 후보가
한 번도 보지 않은 소스에 의해 매 스캔 냉각된다.


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
| KR 패널 | 공식 셋 + WTS | `Panel` | — |
| US 패널 | 공식 셋만 | `Panel` | WTS가 있으면 실패 |
| 기대 | KR에 WTS 있고 US에 없고 US가 비어 있지 않다 | 이 테스트 | `t.Error`/`t.Fatal` |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | KR 패널에 WTS가 없다 | 없음 | `t.Error` | 자체 실행 |
| B2 | US 패널에 WTS가 있다 | 없음 | `t.Error` | 자체 실행 |
| B3 | US 패널이 비었다 | 없음 | `t.Fatal` — 공식 순위는 두 시장을 다 본다 | 자체 실행 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `Panel(market, official, nil, wts, nil)` | 패널 구성 — **`nil` clock** | — | ast.json calls |
| `hasSource(panel, id)` | 멤버십 확인 | 순수 | ast.json calls |

## State mutations and fallbacks

- 테스트 전용. fake reader와 프로세스 내 상태만 쓴다.
- 실계좌·실브로커·네트워크 무접촉. 주문 side effect 0.

## Safety conclusion

- Safe edit boundary: 테스트 수정 — 호출부 인자 1개(`nil`). 단언 무변경.
- High-risk impact: no (테스트 전용).
