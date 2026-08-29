# Function Logic Map: `TestEveryPanelSourceHasItsOwnID`

- Source: `internal/candidatesrc/candidatesrc_test.go`
- Change: `fix-chase-veto-measurement`
- AST evidence: `ast.json` (revision `current`, L296–311, 분기 4개)
- Risk scan: `risk-pattern-report.md`

§2 리뷰의 P0을 그것이 나온 자리에서 잡는다. 세 공식 순위가 하나의 source id로 출하됐었다.
정합성 문제가 아니다 — 스캔은 그 후보를 올린 **모든** 소스가 답했을 때만 냉각할 수 있고 그
검사는 id로 키잉되므로, id가 셋이 하나면 답한 둘이 rate-limited된 하나를 보증하고, 없는
목록만 올린 후보가 한 번도 보지 않은 스캔에 의해 냉각된다. 거기서 냉각 시계가 만료시키고
`first_seen_at`이 사라진다.

이 테스트가 `Panel`의 B3(`err == nil` arm)이 **하나도 버리지 않는다**는 것도 함께 지킨다 —
셋 중 하나가 빠지면 빈 자리가 생기고, 빈 패널이면 B4가 실패한다.


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
| 두 시장 | KR·US | 이 테스트 | 각각 검사 |
| 기대 | id 중복 없음, 패널 비지 않음 | 이 테스트 | `t.Errorf` |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `for _, market := range {KR, US}` | 없음 | — | 자체 실행 |
| B2 | `for _, src := range panel` | `seen` 구성 | — | 자체 실행 |
| B3 | id 중복 | 없음 | `t.Errorf` | 자체 실행 |
| B4 | 패널이 비었다 | 없음 | `t.Errorf` | 자체 실행 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `Panel(market, &fakeRankings{}, nil, &fakePopular{}, nil)` | 패널 — **`nil` clock** | — | ast.json calls |
| `src.ID()` | 소스 id | 순수 | ast.json calls |

## State mutations and fallbacks

- 테스트 전용. fake reader와 프로세스 내 상태만 쓴다.
- 실계좌·실브로커·네트워크 무접촉. 주문 side effect 0.

## Safety conclusion

- Safe edit boundary: 테스트 수정 — 호출부 인자 1개(`nil`). 단언 무변경.
- High-risk impact: no (테스트 전용).
