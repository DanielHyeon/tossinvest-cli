# Function Logic Map: `TestARateLimitedRankingIsReportedAsOne`

- Source: `internal/candidatesrc/candidatesrc_test.go`
- Change: `fix-chase-veto-measurement`
- AST evidence: `ast.json` (revision `current`, L122–132, 분기 2개)
- Risk scan: `risk-pattern-report.md`

429가 `candidate.ErrRateLimited`로 사상된다는 것. 이 소스에 대해 그 집계는 **측정 그 자체**다 —
공식 RANKING 그룹에는 공표된 한도가 없고 hybrid가 이 엔드포인트에 WTS fallback을 주지 않으므로
429는 소스를 저하시키는 것이 아니라 **제거**한다.

이 change에서 이 경로가 새로운 의미를 하나 얻었다: 읽기 실패는 `rememberRead`보다 **앞에서**
반환하므로 기억이 손대지 않은 채 장애를 살아남는다. 그것이 F1이 시간 상한을 붙인 이유이고,
그 시나리오는 `TestTheMemoryOfAReadingBeforeAnOutageIsNotAnAnswer`가 잡는다.


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
| `f.err` | `official.ErrRateLimited`를 감싼 오류 | 이 테스트 | — |
| 기대 | `errors.Is(err, candidate.ErrRateLimited)` | 이 테스트 | `t.Fatalf` |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `mkErr != nil` | 없음 | `t.Fatalf` | 자체 실행 |
| B2 | 429가 sentinel로 오지 않았다 | 없음 | `t.Fatalf` | 자체 실행 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `OfficialRanking(f, nil, RankingTradingAmount, 100, nil)` | 소스 | — | ast.json calls |
| `errors.Is(err, candidate.ErrRateLimited)` | sentinel 확인 | 순수 | ast.json calls |

## State mutations and fallbacks

- 테스트 전용. fake reader와 프로세스 내 상태만 쓴다.
- 실계좌·실브로커·네트워크 무접촉. 주문 side effect 0.

## Safety conclusion

- Safe edit boundary: 테스트 수정 — 호출부 인자 1개(`nil`). 단언 무변경.
- High-risk impact: no (테스트 전용).
