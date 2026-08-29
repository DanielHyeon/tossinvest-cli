# Branch Test Map: `AllReasonCodes`

**GREEN 칸은 호출 지점을 실측해서 채운다** — 덮이지 않은 것을 덮였다고 적지 않는다.
`ast.json`의 열거가 정본이다: AST 분기 0 · 이탈 2.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | **조건 분기가 없다.** 이 함수의 유일한 행동은 리터럴의 내용이고, 새 ReasonCode 하나가 거기 들어간다 | `TestReasonCodeEnumIsStable` (`internal/execgw/failclosed_test.go:33`) | **예정 — 코드를 더하면 golden 불일치로 즉시 실패한다** | no |

> **이 자리의 RED는 공짜다 — 그리고 그것이 함정이다.**
> 새 상수를 이 리터럴에 넣는 순간 `TestReasonCodeEnumIsStable`이 깨진다.
> 그 빨강은 **문자열이 바뀌었다**만 말하고 **차단이 실제로 걸리는지**는 아무것도 말하지 않는다.
> a098의 진짜 요구(배달 실행자가 죽으면 진입이 막힌다)는 **다른 테스트**가 져야 한다 —
> tasks §3의 R로 등록한다. 이 빨강을 그 요구의 증거로 쓰면 안 된다.

> **반대 방향의 함정도 적는다.** 새 상수를 `reason.go`에만 선언하고 이 함수에 안 넣으면
> **아무 테스트도 안 깨진다.** 실제로 그 일이 있었고 `failclosed.go:283-288`이 기록한다.
> 그래서 §5의 VERIFY는 `AllReasonCodes()`의 **길이 증가**를 직접 잰다.

## 이 함수에 **테스트를 새로 쓰지 않는 분기**

없다 — 분기가 0이다.

## 덮이지 않은 것을 이름으로 적는다

- **`latchOrder`(`retry.go:584`)의 append** — 이 파일의 테스트가 안 본다.
  `latchOrder`에 안 넣어도 차단은 걸리고(`retry.go:545-547`의 폴백) 테스트도 안 깨진다.
  덮이는 것은 **우선순위 결정론뿐**이므로 §3이 그 단언을 R로 따로 등록한다.
- **`internal/risk`의 동명 함수**(`risk/reason.go:184`) — 다른 열거이고 a098이 안 건드린다.
  `not-applicable`.
