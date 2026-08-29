# Function Logic Map: `TestAnUnreportedBudgetStaysUnreported`

- Source: `internal/candidatesrc/candidatesrc_test.go`
- Change: `fix-chase-veto-measurement`
- AST evidence: `ast.json` (revision `current`, L163–189, 분기 6개)
- Risk scan: `risk-pattern-report.md`

부재와 0을 가르는 규칙이 패키지 경계를 건너간다는 것. 접근자 없는 클라이언트와 헤더 없는
응답은 **둘 다** "아무 말도 없었다"로 와야 하고, "남은 게 없다"로 오면 안 된다.

미보고를 0으로 읽는 스케줄러는 한 번도 물러나라고 하지 않은 엔드포인트에서 물러난다.


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
| 접근자 없음 | `OfficialRanking(f, nil, ...)` | 이 테스트 | `Reported == false` |
| 헤더 없음 | `fakeBudget{}` | 이 테스트 | 동상 |
| 기대 | 둘 다 `!Reported && !Tight(5)` | 이 테스트 | `t.Errorf` |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | 접근자 없는 소스 생성 실패 | 없음 | `t.Fatalf` | 자체 실행 |
| B2 | 읽기 실패 | 없음 | `t.Fatalf` | 자체 실행 |
| B3 | 접근자 없는 클라이언트가 예산을 보고했다 | 없음 | `t.Errorf` | 자체 실행 |
| B4 | 침묵 소스 생성 실패 | 없음 | `t.Fatalf` | 자체 실행 |
| B5 | 읽기 실패 | 없음 | `t.Fatalf` | 자체 실행 |
| B6 | 헤더 없는 응답이 예산을 보고했다 | 없음 | `t.Errorf` | 자체 실행 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `OfficialRanking(f, nil, ...)` / `OfficialRanking(f, fakeBudget{}, ...)` | 두 미보고 형태 | — | ast.json calls |

## State mutations and fallbacks

- 테스트 전용. fake reader와 프로세스 내 상태만 쓴다.
- 실계좌·실브로커·네트워크 무접촉. 주문 side effect 0.

## Safety conclusion

- Safe edit boundary: 테스트 수정 — 호출부 인자 1개(`nil`). 단언 무변경.
- High-risk impact: no (테스트 전용).
