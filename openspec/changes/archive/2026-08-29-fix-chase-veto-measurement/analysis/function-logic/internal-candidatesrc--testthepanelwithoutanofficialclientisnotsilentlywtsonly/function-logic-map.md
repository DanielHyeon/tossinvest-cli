# Function Logic Map: `TestThePanelWithoutAnOfficialClientIsNotSilentlyWTSOnly`

- Source: `internal/candidatesrc/candidatesrc_test.go`
- Change: `fix-chase-veto-measurement`
- AST evidence: `ast.json` (revision `current`, L268–276, 분기 2개)
- Risk scan: `risk-pattern-report.md`

spec Requirement 5는 공식 소스만으로 충분해야 한다고 말한다. 그 역 — WTS 단독 — 은 세션이
만료되는 날 멈추는 구성이다. 공식 클라이언트 없이 만든 패널은 존재해도 되지만, 호출자가
그것을 **볼 수 있어야** 하고 소스 목록이 그것을 드러낸다.


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
| `only` | `Panel(KR, nil, nil, wts, nil)` | 이 테스트 | WTS 하나여야 한다 |
| 클라이언트 없음 | `Panel(KR, nil, nil, nil, nil)` | 이 테스트 | 빈 패널 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | 패널이 WTS 하나가 아니다 | 없음 | `t.Fatalf` | 자체 실행 |
| B2 | 클라이언트 없는 패널이 비어 있지 않다 | 없음 | `t.Error` | 자체 실행 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `Panel(...)` | 두 구성 — **`nil` clock** | — | ast.json calls |

## State mutations and fallbacks

- 테스트 전용. fake reader와 프로세스 내 상태만 쓴다.
- 실계좌·실브로커·네트워크 무접촉. 주문 side effect 0.

## Safety conclusion

- Safe edit boundary: 테스트 수정 — 호출부 인자 1개(`nil`). 단언 무변경.
- High-risk impact: no (테스트 전용).
