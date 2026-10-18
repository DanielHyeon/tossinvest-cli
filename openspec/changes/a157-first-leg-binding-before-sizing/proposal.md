# a157 — 전략 첫 다리 결속 판정을 Guardian 사이징 앞으로 (first-leg-binding-before-sizing)

> **사람 결정(2026-10-18, a146 design 「사람 결정」 2):** F10(상한 0 이 뒤의 Guardian 판정을 가림)은 a146 에서 알려진 한계로 수용하되,
> execgw 결속 판정을 사이징 앞으로 옮기는 **별도 후속 change 가 어떤 전략 레인을 켜기 전(P2 S5/S6)에 착지해야 한다.**
> **사용자 지시(2026-10-18, 인수인계 §1 · §3):** 작업 A(A1 결속 판정 이동 + A2 조립 감독자 알림 배선)를 시작한다. 한 change 로 묶는다.
> **사용자 결정(2026-10-18, freeze 리뷰 뒤 「1-1, 2-1」):** 일일 손실 거절은 결속 결함보다 먼저 보고한다(a137 강화 보존).
> 조립 감독자에 `EntryGate` 도 nil 가드와 함께 배선한다.

## Why

`PrecheckQFinalCampaignFirstLeg` 는 첫 줄에서 `PrecheckQFinalEntry`(사이징 · 버킷 · 체인 · 노출)를 부르고, 그 오류를 B1(51)에서 바로 돌려준다
(AST `analysis/ast-evidence/execgw--riskguardian.precheckqfinalcampaignfirstleg.ast.json`).

그래서 Guardian 상한 0 (위험 예산 또는 주문 금액 상한으로 1주도 안 되는 진입)이면, 첫 다리 고유의 결속 판정 일곱이 하나도 돌지 않는다.
일곱은 B2 정책 버전 · B3 소유자 prospective 세대 · B4 가격 단위 · B5 계보 결속 · B6 포지션 세대 · B7 attempt/활성화 · B8 주간 예약 결속이다.
a146 이전에는 그 결과가 잠금/삼킴이었고, a146 뒤에는 **조용한 스킵**이다.

엔진 경로에서 이 가림이 실제로 닿는 것은 B2 · B8 이다. B3~B7 꼴은 엔진의 `validateStrategyFirstLegAuthority` 가 precheck 앞에서 이미 막는다(design F8).
레인을 켜면 B2 · B8 의 결함은 상한이 1 이상이 되는 주기에야 드러난다.

B2~B8 은 q_final 을 쓰지 않는다. 입력은 전부 요청 값이거나 `PrecheckQFinalEntry` 가 바꾸지 않고 통과시키는 값이다(design F2~F4).
그래서 판정을 앞으로 옮겨 결속 결함을 먼저 보고할 수 있다.

조립 감독자(`(*Context).NewPairedStrategyEntryProductionAssembly` 의 `NewStrategyEntrySupervisor`, strategy_entry_supervisor.go:394)에는
생산 감독자(`NewRefreshingPairedStrategyEntrySupervisor`, :433-434)에 있는 `DailyLossNotice` · `EntrySkipNotice` · `EntryGate` 가 없다
(a146 구현 리뷰 1 P3-5, a157 freeze 리뷰 P1-2). 오늘은 투영에만 쓰이고 Run 되지 않아 해가 없다.

## What Changes

1. **A1 — 결속 판정을 앞에, 일일 손실을 그보다 먼저 보고.** `PrecheckQFinalCampaignFirstLeg` 를 다음 순서로 바꾼다(design D1).
   1. nil 가드.
   2. B2~B8 판정(편집 전 순서, 첫 결함을 잡아 둠).
   3. `PrecheckQFinalEntry` 호출 한 번.
   4. 보고 우선순위: **일일 손실 체인 거절 > 결속 결함 > 그 밖의 진입 거절.**
   5. B9 계보 투영.

   판정식 · 거절 코드 · field · 문구는 그대로다. 입력 출처 둘은 같은 값으로 바뀐다(`request.Entry.Admission` · `request.Entry.QCandidate`).
2. **일일 손실 판별 하나로.** `admit` 이 오늘 쓰는 「일일 손실 측정 불가 / 한도 도달」 조건을 execgw 의 판별 함수 하나로 옮긴다.
   precheck 의 우선순위와 `admit` 의 분류가 그 함수를 같이 쓴다. 조건을 글자 그대로 옮기므로 `admit` 의 분기 결과는 같다.
3. **A2 — 조립 감독자 옵션.** `DailyLossNotice: c.strategyDailyLossNotice()` · `EntrySkipNotice: c.strategyEntrySkipNotice()` 를 넘기고,
   `c.Entry != nil` 일 때만 `EntryGate: c.Entry` 를 넘긴다. 타입 있는 nil 은 인터페이스를 nil 이 아니게 만들어 panic 을 낸다(design F10).
4. **spec** — `risk-management`:
   - ADDED 「전략 첫 다리 결속 판정은 Guardian 사이징보다 먼저 한다」.
   - MODIFIED 「Guardian 상한 0 인 전략 진입은 그 진입만 건너뛴다」 — 명시 한계에서 결속 판정을 빼고, 결속 결함이 겹친 요청의 보고를 새 요구로 넘기며,
     두 `Context` 조립의 알림 · 게이트 동등을 더한다.

## What Does NOT Change

- **같은 입력 · 같은 시계 관측에서 Guardian 이 받아들이는 첫 다리 요청의 집합.** 바뀌는 것은 둘 이상의 판정이 걸리는 요청의 보고뿐이다.
  design D2 에 (a)~(f) 로 전부 열거했다. 일일 손실 강화 · 계좌 단위 진입 차단은 겹친 요청에서도 같다(1-1).
- `PrecheckQFinalEntry` · `IssueQFinalEntry` · `IssuePrecheckedQFinalCampaignFirstLeg` 의 본문.
- 사이징 규칙(위험 예산 1/3 · min 상한), 체인 · 버킷 · 노출 판정과 그 순서, a146 상한 0 분류.
- 원장 · Gateway · 브로커 호출, 원장 스키마 · 설정 키 · 서명 매니페스트.
- 매도 · 손절 · 익절 · 비상 청산 · 대사 · 체결 감지. 편집하는 함수는 모두 첫 다리(매수) 진입 경로 또는 그 감독자 조립에만 있다.
- **남는 한계:**
  - 진입 판정 안의 버킷 · 체인(kill switch · 손절 계약 · 재진입 · 입력 불가) · 노출과 B9 계보 투영은 q_final 위에서 돈다.
    그래서 상한 0 뒤에 계속 가려진다(spec 문장에 남김).
  - 조립 감독자의 알림 dedup 은 감독자마다 있고, 그 감독자는 파도마다 새로 만들어진다. 이것을 Run 하게 되는 change 가 dedup 을 `Context` 로 올려야 한다(tasks 후속).

## Impact

- Affected specs: `risk-management`(ADDED 1 · MODIFIED 1).
- Affected code:
  - `internal/execgw/riskguardian_first_leg.go` — `PrecheckQFinalCampaignFirstLeg`.
  - `internal/execgw/a137_daily_loss.go`(또는 새 파일) — 판별 함수.
  - `internal/app/engine/strategy_first_leg_admission.go` — `admit` 의 일일 손실 두 조건이 판별 함수를 부름.
  - `internal/app/engine/strategy_entry_supervisor.go` — `(*Context).NewPairedStrategyEntryProductionAssembly` 옵션.
  - 새 시험 파일(execgw · engine).
- 위험 등급: **High-risk**(Guardian · execgw 첫 다리 판정 순서 · 일일 손실 분류 경로). FLM/BTM · Pre-Edit · RED 선행 · 변이 원장 · 독립 적대 리뷰.
- 배포: 엔진 이미지 교체. 설정 · 서명 변경은 필요 없다. 현재 생산 레인은 DORMANT(서명 매니페스트 0)라 동작 변화는 레인 활성화 뒤에 나타난다.
