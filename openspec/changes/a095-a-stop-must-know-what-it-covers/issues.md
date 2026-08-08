# a095 · issues

> 이 파일은 a095가 **고치지 않고 남기는 것**과 **다른 change의 선행 조건이 되는 것**을
> 기록한다. 침묵한 생략을 만들지 않기 위한 자리다.

## I1. 손절 상향 래칫의 선행 조건 셋 (tasks 4.3)

**사용자가 지목한 문제의 나머지 절반이다.**

> **정정(1라운드).** 이 절의 초판은 *"475150은 평단 58,000, 손절 56,163, 즉 −3.17%"*라고
> 적었다. **`initial_stop`을 유효 손절로 착각한 것이다.** 실제 유효 손절은
> `baseline_price = 57,900`(본전 승격)이고 평단 대비 **−0.17%**다. **래칫이 이미 올렸다.**

정정 후에도 이 항목이 남는 이유: 래칫은 **가격이 오를 때** 손절을 올린다.
**가격이 오르지 않은 채 평단만 올라가는 경우**(불타기)는 그 경로로 덮이지 않는다.
그때 커지는 것은 주당 위험이며, 그 방향의 손절 **상향**은 §6에 부합한다.

그 방향의 손절 **상향**은 위험을 줄이므로 §6에 부합한다. **그럼에도 a095가 하지 않는
이유는 셋이고, 셋 다 측정에서 나왔다.**

**(1) 유효 손절가를 쓰는 경로는 하나뿐이고, 그것은 운영자 행동이다.**

`resetExitStateForReadoptTx`(`internal/journal/apply_hook.go:684`)의 주석이 선언한다 —
*"the only reset writer for the four guarded columns."* 호출자는
`positionpolicy.ActionReadopt`(`position_policy.go:145`) 하나다.

**그 함수는 분기 6개 중 4개가 미진입이고 넷 다 오류 처리다**(a095 실측).
손절가를 덮어쓰는 유일한 자리의 오류 경로가 거의 시험되지 않았다. 여기에 자동 호출자를
더하는 것은 **시험되지 않은 쓰기 자리에 자동 트리거를 붙이는 일**이다.

StockOS도 같은 자리에 SHALL을 둔다 —
`openspec/specs/position-campaign-core/spec.md` *"손절 갱신 경로는 하나다 … 비후퇴 검사를
수행하는 단일 저장 경로만이 이 컬럼을 갱신한다."*

**(2) 진입가를 옮기면 한 값만 옮겨지지 않는다.**

`EvaluateLadder`(`internal/exitpolicy/ladder.go:307`, 분기 32 · 미진입 11)에서
`entry`는 세 곳에 쓰인다:

| 자리 | 쓰임 |
| --- | --- |
| `:358` | `percentOf(probe, entry)` — **수익률의 분모** |
| `:387` | `lockPrice(entry, Rungs[i].StopPct)` — **rung 잠금가** |
| `:503-509` | `lockPrice` = `entry × (1 + pct/100)` |

정책 표의 `StopPct` 자체가 *"the protected stop relative to the **entry price**"*
(`:99-100`)다. **레벨 판정·rung·high-water 대비가 전부 같이 움직인다.**

**(3) 하향은 반드시 거부하고 기록해야 한다.**

평단이 **내려간** 포지션(실측 066570·080220·272210)에서 평단 기준 손절은 지금보다
**낮다.** 래칫 없이 재계산을 도입하면 그 셋의 손절이 **내려간다** — §6 위반이다.

StockOS의 SHALL이 그 형태를 준다
(`position-campaign-core/spec.md:53-56`, 원문 그대로):

> The system SHALL **reject** any update that would move a long campaign's effective
> stop price **below** its previously stored value, SHALL **record the rejected attempt**
> instead of silently ignoring it, and SHALL keep the prior value in effect.

**단, 그 모듈은 라이브에 배선돼 있지 않다.** 같은 spec `:123-126`이 그것을 요구사항으로
고정한다 — *"SHALL keep the campaign core free of any call site in the live order, entry,
or exit paths … and SHALL prove the absence of such wiring by test."*

**따라서 인용하는 것은 작동 중인 구현이 아니라 검토를 마친 계약이다.**
StockOS의 실제 라이브 경로(parker)는 TossOS와 **똑같이 동결**돼 있다.

**선행 조건은 `specs/exit-policy/spec.md`의 「손절가는 평균단가 변화를 따라 내려가서는
안 된다」에 SHALL로 있다.**

## I2. `positions.avg_price`는 stale일 수 있다

`ApplyPositionAdjustment`(`internal/journal/position_adjustments.go:312`):

```go
NewAvgPrice: firstNonEmpty(req.NewAvgPrice, target.AvgPrice),
```

**브로커가 원가를 주지 않으면 직전 평단이 그대로 이어진다.**

원장이 그 가능성을 보인다 — 475150의 조정 이력은 수량을 3→32로 옮기는 동안
`new_avg_price`가 58,000에서 움직이지 않았다. 그것이 「실제로 평단이 안 변했다」인지
「브로커가 원가를 안 줘서 이어붙였다」인지 **원장만으로는 가를 수 없다.**

a095는 그 값을 **읽기만** 하고 고치지 않으며, 총위험 보고에 **불확실성을 함께 담는다**
(`specs/exit-policy` SHALL). **평단 자체의 정확성은 별도 change의 주제다** — 그것을
확정하려면 브로커의 원가 응답 유무를 조정 행에 남기는 스키마 변경이 필요하다.

## I3. 발신 자리 넷의 등급을 함께 올리는 것 (tasks 1.11)

`EventExitPositionUnmanaged`는 네 곳에서 나온다:

| 자리 | 사실 |
| --- | --- |
| `adoption.go:418` | 편입도 진입 결정도 없는 보유 |
| `adoption.go:456` | 편입 후 수량 증가 |
| `exitloop.go:1501` | 관측 루프가 본 무보호 |
| `exitwiring.go:104` | 배선 시점의 무보호 |

**등급은 이벤트 종류로 매겨지므로 넷은 같이 올라간다.** 분리하려면 새 이벤트 종류를
만들어야 하고, 그것은 표를 늘리는 대신 의미를 쪼개는 일이다.

**a095의 판단**: 넷 다 「가진 것에 손절이 안 걸려 있다」는 같은 사실이므로 함께 올린다.
tasks 1.11이 네 문구를 실제로 대조해 그 판단을 확인한다. **대조 결과 다른 사실이면
이 판단이 틀린 것이고, 그때는 이벤트 종류를 쪼갠다.**

## I4. 배포 재생 결과 (tasks 5.3 — 미실행)

`[미실행]` — 5.1·5.2가 재생을 만든 뒤 여기에 결과를 적는다.
a091·a094와의 상호작용도 같은 자리에 적는다.

## I5. 현재 열린 포지션은 이 change가 보호하지 않는다

**a095는 보고만 바꾼다.** 배포 후에도:

- **010170 `instance_seq=2`(30주, 평단 11,630)는 무보호다.** `exit_states` 0행,
  `exit_events` 0건, `adoption_id`·`entry_decision_id` 둘 다 없음.
  배포가 하는 일은 그것을 **critical로 보이게** 하는 것뿐이며, 편입은 운영자 행동이다.
- **475150(32주, 유효손절 57,900 기준 총위험 3,200원)·080220(12주, 19,128원)의
  손절가는 바뀌지 않는다.** (1판은 475150을 58,784로 적었다 — `initial_stop`을 쓴
  오류이며 1라운드가 잡았다.)
  바뀌는 것은 그 수가 보고된다는 사실뿐이다.

**배포 전까지 사람이 처리한다**(tasks 7.3).
