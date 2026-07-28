# Function Logic Map: `TestTheRequestedCountIsCappedAtTheDocumentedMaximum`

- Source: `internal/candidatesrc/candidatesrc_test.go`
- Change: `fix-chase-veto-measurement`
- AST evidence: `ast.json` (revision `current`, L77–89, 분기 3개)
- Risk scan: `risk-pattern-report.md`

API가 count를 100에서 자르고 조용히 truncate한다. 150을 요청해 100을 받은 호출자는 **존재한
적 없는 목록 길이**에 대해 백분위를 계산한다 — 산술이 아니라 요청이 만든 정규화 오류다.

이 경고는 `candidatesrc.go`가 쓰인 날부터 있었고, 이 change가 그것을 **검사 가능한 것으로**
만들었다: 상한 적용 후의 수가 `Row.RankRequested`로 실려 나가고 저장소가 비교한다.
그 짝은 `TestTheRequestedCountIsTheCappedOneRatherThanTheOneTheCallerAsked`다.


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
| 요청 | 500 | 이 테스트 | 엔드포인트에는 100이 가야 한다 |
| `f.gotN` | fake가 받은 수 | `fakeRankings` | `t.Errorf` |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `mkErr != nil` | 없음 | `t.Fatalf` | 자체 실행 |
| B2 | `Read` 오류 | 없음 | `t.Fatalf` | 자체 실행 |
| B3 | `f.gotN != 100` | 없음 | `t.Errorf` | 자체 실행 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `OfficialRanking(f, nil, RankingTopGainers, 500, nil)` | 상한 대상 | — | ast.json calls |

## State mutations and fallbacks

- 테스트 전용. fake reader와 프로세스 내 상태만 쓴다.
- 실계좌·실브로커·네트워크 무접촉. 주문 side effect 0.

## Safety conclusion

- Safe edit boundary: 테스트 수정 — 호출부 인자 1개(`nil`). 단언 무변경.
- High-risk impact: no (테스트 전용).
