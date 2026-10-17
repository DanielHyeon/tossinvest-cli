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
> **5판(2026-09-29)**: codex 3라운드(REJECT, P0 0) 반영 — a092 입구는 **구현 하드 의존**(입구 밖 대안 삭제, 설계 freeze 는 독립) · a090 이 넘기는 값은
> 구성상 무계좌(전용 정화 공지자 · 실패 주입 카나리) · tasks 의 옛 key 기대값 수리.
>
> **4판(2026-09-29)**: codex 2라운드(REJECT, P0 1 — 계좌 정보) + a092 교차 반영 — 새 알림·공지·로그는 **계좌 ref 무탑재**, 관측자 전체 로거 대신
> **새 줄만 내는 전용 로거** · 커밋된 강화의 공지 적재 재시도 · `go/parser` 이탈 전수 핀 · 기록은 a092 **단일 입구**로. 계좌번호가 기존 경보·로그 19 자리에 원문으로 실리는
> **기존 관행**은 범위 밖 — 사용자 결정(design D12).
>
> **3판(2026-09-29)**: codex 1라운드(REJECT, P0 2) 반영 — 완료 정책은 표시 해제로 명시 제외(구조 핀) · **모드 강화 공지도 enqueue-only**(다음 주기를
> 늦추지 않는다) · 생산 로거 배선 한 줄 · 단조 시계·연속 id key · 실패 전이 · 보장의 조건. `review.md` 「3판」.
>
> **2판(2026-09-29)**: 1라운드 적대 보이스 1(REJECT, P1 4) 반영 — 세는 단위를 두 자리에서 **보유 대상 집합**으로, 기점을 **마지막 판정**으로,
> 알림을 **순회 뒤 enqueue-only** 로, 임계 아래도 **로그·주기 수**로 남긴다. `review.md` 「1라운드」·「2판」.
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
they happen." 브로커측 손절이 없는 지금, 판정되지 않는 포지션은 보호되지 않는 포지션이다 — 같은 파일 `:40-41`: "we chose not to
look" and "we could not look" leave a position equally unprotected. a089 2차 리뷰가 이것을 C2(P0)로 짚었고
("보유 포지션 하나가 시세 응답을 못 받으면 무기한 조용히 청산 루프에서 빠진다 … 453-460은 아무것도 올리지 않는다"),
HEAD 에서 코드는 그대로이며 a111 이 같은 모양의 무음 `continue` 를 하나 더했다(`:459-462`).

## What changes

| | 자리 | 성격 |
| --- | --- | --- |
| 계수(2판) | `workingSet` 이 본 **보유·대상 포지션 집합** 중 그 주기에 판정에 닿지 않은 것 — `ObserveOnce` B6·B7 과 `workingSet` 의 탈락 다섯 자리를 한 규칙으로 | **포지션 단위 기록 · 연속 시작/해제 로그 · `ExitCycle.Unobserved`** |
| 판정(2판) | 새 파일의 새 메서드, **순회 뒤** | **마지막 판정 시각**부터 계정 사다리와 같은 임계(`outageAfter()`, 기본 60초)를 넘으면 critical 1회 |
| 알림(2판 · 5판) | **a092 단일 입구 `RecordAlert`(창 0)** → `obs.EventExitObservationOutage` | **enqueue-only**(뒤 포지션의 손절 판정을 기다리게 하지 않는다) · key 에 **미관측 연속** 신원 · 새 이벤트 타입 없음 |
| 모드 강화 | 기존 `EscalateOperatingMode(…ModeTriggerExitObservationOutage…)` | **정본 준수(Q1 확정)** — 정본 exit-policy 문장(`spec.md:62`·`:65`)이 두절에 ENTRY_BLOCKED 까지 요구한다(design D5) |

**바꾸지 않는 것.** 판정·기준선·워터마크·발의·주문·B1~B4 의 동작·계정 사다리. 새 브로커 호출 0(§0.4). 토글 없음.

## 선후 관계

| change | 관계 |
| --- | --- |
| a089 (아카이브) | 이 change 를 선행으로 권고한 곳. 아카이브 기록이 이 change 로 승계를 예고했다 |
| a092 (알림이 손절을 잡지 않는다) | 이 change 의 RED 목록(R1~R6)의 출처. **2판: 새 critical 은 enqueue-only 라 동기 전달을 타지 않는다**(design D4). 기존 동기 경보는 a092 소관 |
| a092 단일 입구(K6) | **구현 하드 의존(5판)** — 새 알림·공지는 a092 의 critical 기록 단일 입구로 기록한다. 구현은 입구 착지 뒤에만. **설계 freeze 는 독립** |
| a094 (손절은 자기를 막는 것을 치운다) | 같은 파일(`exitloop.go`)의 다른 함수(`record`·`clearTheSymbol`)를 편집한다. a090 의 편집은 `ObserveOnce`·`workingSet` + 새 파일. **알림 형태(enqueue-only · 에피소드 key · 적재 실패 잠금)를 a094 D−4.6·D−5.2·D−5.3 과 같게 쓴다** — a094 의 에피소드는 park attempt, a090 은 미관측 연속 |
| a112 결정 46 | 장 마감 시 `/prices` 가 두 시장 모두 행 1개를 돌려준 실측(n=1/시장) — fail-closed 정상 입력 열거의 근거(design D6) |
| (미측정) 정지·0가격 종목 | **[미측정 · 사전 승인된 실측 대기]** — 구현 로트가 장중에 읽기 전용 시세 GET 1회(Manager 사전 승인 2026-09-29, design Q2) |
| (후속 후보) 하류 무음 5자리 | 판정 진입 뒤 임대 재검사 5자리(`:859` `:956` `:1027` `:1050` `:1180`)의 무음 — 명명된 잔여(design Q3) |

## Non-goals

- 미관측의 **원인**을 브로커 응답에서 가르는 것(부재·0가격·요청 밖 종목). `observe`(`:743-814`)를 바꿔야 하고 계수의 판정에는
  필요 없다 — B6 와 B7 만 가른다.
- 시세 **나이** 게이트(장 마감 시세가 FetchedAt=지금으로 들어오는 문제, a112 잔여 (k)와 같은 계열). 별도.
- 형제 한 종목의 성공이 진입 게이트의 가격 신선도를 찍는 문제(`observe` `:779-783`) — 이 change 는 그것을 포지션 단위 경보 + Q1 강화로만 다룬다(F16).
- `workingSet` 의 미관리(`:512`)·완료 정책(`:533`) 포지션 — 명명 잔여(design D1).
- 사람이 푸는 모드 완화 경로 — 기존 운영자 절차 그대로.
