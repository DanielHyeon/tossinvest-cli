# a090 · 관측되지 않은 보유 포지션은 세어지고 알려진다

- **Feature**: `FEAT-TOS-009` — Exit line truth and position policy lifecycle
- **Story**: `STORY-TOS-a090`
- **Spec**: `exit-policy`
- **위험 등급**: **High-risk** — 손절 관측 경로(`ExitObserver.ObserveOnce`). 판정·발의·주문은 바꾸지 않는다.

> **계보.** 이 이름은 a089 2차 리뷰(2026-08-06)가 지었다(`openspec/changes/archive/2026-09-28-a089-an-unserved-stop-is-counted/
> review.md` 「판정과 다음 (2라운드)」):
>
> > **a090(신설, 선행)** — P0 계측. "이번 주기에 관측되지 않은 보유 포지션"을 종목 단위로 세고, 연속 미관측이 한계를 넘으면
> > critical. 이것이 없으면 8/5가 무엇이었는지 다음에도 알 수 없다
>
> 그 뒤 **어느 change 도 이 역할을 맡지 않았다.** a092 는 이것을 a090 몫으로 명시하고 구현하지 않았다(a092 의 `ObserveOnce` FLM
> 「필요한 RED (a090 후보)」). a089 는 2026-09-28 사용자 승인으로 불구현 아카이브됐고(`64a1b2b3`), 이 P0 를 신설 change 로 넘겼다.

## Why

**보유 포지션 하나가 한 번의 가격 읽기에서 답을 못 받으면, 그 포지션은 무기한 조용히 손절 판정에서 빠진다.**

`ExitObserver.ObserveOnce`(`internal/app/engine/exitloop.go:413-470`)는 한 주기에 보유 종목 전부를 한 번에 읽고, 포지션마다 판정한다.
응답에서 빠진 종목(`:453` `if !ok`)과 앞 포지션을 처리하는 동안 사용 임대가 끝난 시세(`:459` `if !o.quoteUsable(quote)`)는 **둘 다
아무것도 남기지 않고 `continue`** 한다 — 로그도 알림도 원장 행도 없다.

계정 단위 두절 시계는 이것을 보지 못한다. **한 종목이라도 응답하면** `:447-448` 이 `lastObserved` 를 지금으로, `outageRaised` 를
거짓으로 되돌린다. 계정 사다리(60초 → critical + ENTRY_BLOCKED)는 **전 종목이 답하지 않을 때**(`observe` `:779-781`)만 탄다.
주기 실패 로그(`reportCycle` `:382-388`)도 `cycle.Err != nil` 일 때만 쓴다. 진입 게이트의 가격 신선도도 형제 한 종목의 성공으로
찍힌다 — 기존 시험 `TestA111ValidSiblingIsJudgedWithoutLendingFreshnessToInvalidSymbol`(`a111_flat_exit_observation_test.go:759`)이
바로 그것(무효 형제는 Seed 로 남고 게이트는 열림)을 단언한다.

그래서 **보유 포지션 하나가 판정 없이 얼마나 오래 있었는지 아무도 모른다.** 이 파일의 머리 주석은 반대를 약속한다
(`exitloop.go:377-378`): "The conditions that mean a position is actually unprotected raise their own critical events from where
they happen." 브로커측 손절이 없는 지금, 판정되지 않는 포지션은 보호되지 않는 포지션이다 — 같은 파일 `:38-39`: "we chose not to
look" and "we could not look" leave a position equally unprotected. a089 2차 리뷰가 이것을 C2(P0)로 짚었고
("보유 포지션 하나가 시세 응답을 못 받으면 무기한 조용히 청산 루프에서 빠진다 … 453-460은 아무것도 올리지 않는다"),
HEAD 에서 코드는 그대로이며 a111 이 같은 모양의 무음 `continue` 를 하나 더했다(`:459-462`).

## What changes

| | 자리 | 성격 |
| --- | --- | --- |
| 계수 | `ObserveOnce` B6·B7(`:453`·`:459`)의 두 `continue` 직전 | **포지션 단위 미관측 시각을 적는다.** 판정에 닿으면(`:464`) 지운다. 보유에서 사라지면 정리한다 |
| 판정 | 새 파일의 새 메서드 | 미관측이 **계정 사다리와 같은 임계**(`outageAfter()`, 기본 60초)를 넘으면 critical 1회 |
| 알림 | 기존 `o.alert` → `obs.EventExitObservationOutage` | **새 이벤트 타입·전송 경로 없음.** 같은 사실("관측되지 않는 포지션은 보호되지 않는다")을 포지션 key 로 |
| 모드 강화 | 기존 `EscalateOperatingMode(…ModeTriggerExitObservationOutage…)` | **Q1 — 사용자 결정.** 정본 exit-policy 문장은 두절에 ENTRY_BLOCKED 까지 요구한다(design D5) |

**바꾸지 않는 것.** 판정·기준선·워터마크·발의·주문·B1~B4 의 동작·계정 사다리. 새 브로커 호출 0(§0.4). 토글 없음.

## 선후 관계

| change | 관계 |
| --- | --- |
| a089 (아카이브) | 이 change 를 선행으로 권고한 곳. 아카이브 기록이 이 change 로 승계를 예고했다 |
| a092 (알림이 손절을 잡지 않는다) | 이 change 의 RED 목록(R1~R6)의 출처. 새 critical 은 오늘의 동기 전달을 탄다 — **이름 붙인 잔여**(design D7) |
| a094 (손절은 자기를 막는 것을 치운다) | 같은 파일(`exitloop.go`)의 다른 함수(`record`·`clearTheSymbol`)를 편집한다. a090 의 편집은 `ObserveOnce` + 새 파일 — 충돌 면 최소 |
| a112 결정 46 | 장 마감 시 `/prices` 가 두 시장 모두 행 1개를 돌려준 실측(n=1/시장) — fail-closed 정상 입력 열거의 근거(design D6) |

## Non-goals

- 미관측의 **원인**을 브로커 응답에서 가르는 것(부재·0가격·요청 밖 종목). `observe`(`:743-814`)를 바꿔야 하고 계수의 판정에는
  필요 없다 — B6 와 B7 만 가른다.
- 시세 **나이** 게이트(장 마감 시세가 FetchedAt=지금으로 들어오는 문제, a112 잔여 (k)와 같은 계열). 별도.
- 사람이 푸는 모드 완화 경로 — 기존 운영자 절차 그대로.
