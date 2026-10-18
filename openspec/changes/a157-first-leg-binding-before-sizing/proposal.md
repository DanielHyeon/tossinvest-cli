# a157 — 전략 첫 다리 결속 판정을 Guardian 사이징 앞으로 (first-leg-binding-before-sizing)

> **사람 결정(2026-10-18, a146 design 「사람 결정」 2):** F10(상한 0 이 뒤의 Guardian 판정을 가림)은 a146 에서 알려진 한계로 수용하되,
> execgw 결속 판정을 사이징 앞으로 옮기는 **별도 후속 change 가 어떤 전략 레인을 켜기 전(P2 S5/S6)에 착지해야 한다.**
> **사용자 지시(2026-10-18, 인수인계 §1 · §3):** 작업 A(A1 결속 판정 이동 + A2 조립 감독자 알림 배선)를 시작한다. A1 · A2 는 한 change 로 묶어도 된다.

## Why

`PrecheckQFinalCampaignFirstLeg` 는 첫 줄에서 `PrecheckQFinalEntry`(사이징 · 버킷 · 체인 · 노출)를 부르고, 그 오류를 B1(51)에서 바로 돌려준다
(AST `analysis/ast-evidence/execgw--riskguardian.precheckqfinalcampaignfirstleg.ast.json`). 그래서 Guardian 상한 0 (위험 예산 또는 주문 금액 상한으로
1주도 안 되는 진입)이면 첫 다리 고유의 결속 판정 — B2 정책 버전 · B3 소유자 prospective 세대 · B4 가격 단위 · B5 계보 결속 · B6 포지션 세대 ·
B7 attempt/활성화 · B8 주간 예약 결속 — 이 하나도 돌지 않는다. a146 이전에는 그 결과가 잠금/삼킴이었고, a146 뒤에는 **조용한 스킵**이다.
그중 B2 · B8 은 엔진의 `validateStrategyFirstLegAuthority` 도 다시 막지 않는다(a146 design F10). 레인을 켜면 이 결함은 상한이 1 이상이 되는
주기에야 드러난다.

B2~B8 은 q_final 을 쓰지 않는다. 그 입력은 전부 요청 값이거나 `PrecheckQFinalEntry` 가 바꾸지 않고 통과시키는 값이다(design F2~F4).
그러니 순서만 바꿔 결속 결함을 결속 결함으로 먼저 거절할 수 있다.

조립 감독자(`(*Context).NewPairedStrategyEntryProductionAssembly` 안의 `NewStrategyEntrySupervisor`, strategy_entry_supervisor.go:392)에는
생산 감독자(`NewRefreshingPairedStrategyEntrySupervisor`, :433-434)에 있는 `DailyLossNotice` · `EntrySkipNotice` 가 없다(a146 구현 리뷰 1 P3-5).
오늘은 투영에만 쓰이고 Run 되지 않아 해가 없다. 그러나 그 감독자를 Run 하게 되면 일일 손실 차단과 상한 0 스킵이 스냅숏 계수로만 남고
경고 줄이 나가지 않는다.

## What Changes

1. **A1 — 순서만 바꿈.** `PrecheckQFinalCampaignFirstLeg` 가 B2~B8 의 판정을 `PrecheckQFinalEntry` 호출 **앞**에서 한다.
   판정식 · 거절 코드 · field · 문구는 그대로 둔다. 바뀌는 것은 입력 출처 둘뿐이고, 둘 다 같은 값임을 AST 로 보였다(design F3 · F4).
   - B2 · B3 · B5 가 읽던 `entry.admission.{TransactionID, Owner}` → `request.Entry.Admission.{TransactionID, Owner}`
   - B5 가 읽던 `entry.decision.QCandidate` → `request.Entry.QCandidate`
   - nil Guardian 가드 하나를 맨 앞에 둔다. 옮긴 판정이 `g.accountRef` · `g.policyVersion` 을 읽기 때문이다. 거절 코드와 문구는
     `PrecheckQFinalEntry` B1 과 같다.
   B9 계보 투영은 q_final 이 필요해서 사이징 뒤에 남는다.
2. **A2 — 알림 배선.** 조립 감독자에 생산 감독자와 같은 두 옵션 `DailyLossNotice: c.strategyDailyLossNotice()` ·
   `EntrySkipNotice: c.strategyEntrySkipNotice()` 를 넘긴다. 판정은 바뀌지 않는다.
3. **spec** — `risk-management`:
   - ADDED 「전략 첫 다리 결속 판정은 Guardian 사이징보다 먼저 한다」.
   - MODIFIED 「Guardian 상한 0 인 전략 진입은 그 진입만 건너뛴다」 — 명시 한계 문장에서 결속 판정을 빼고, 조립 감독자 알림 시나리오를 더한다.

## What Does NOT Change

- **Guardian 이 받아들이는 첫 다리 요청의 집합.** B1~B9 은 순수한 판정의 논리곱이고(clock 은 B9 앞의 `issuedAt` 에서만 읽음), 순서를 바꿔도
  통과 조건은 같다. 바뀌는 것은 **결함이 둘 이상인 요청이 어느 거절로 보고되는가**뿐이다(design D2 에 전부 열거).
- `PrecheckQFinalEntry` 는 편집하지 않는다. 단독 발급 `IssueQFinalEntry`(:322)도 그 함수를 쓴다(CodeGraph callers).
- 사이징 규칙(위험 예산 1/3 · min 상한), 체인 · 버킷 · 노출 판정과 그 순서, a146 의 상한 0 분류, a137 의 일일 손실 분류.
- 발급 `IssuePrecheckedQFinalCampaignFirstLeg` · 원장 · Gateway · 브로커 호출.
- 매도 · 손절 · 익절 · 비상 청산 · 대사 · 체결 감지. 편집하는 두 함수는 모두 첫 다리(매수) 진입 경로에만 있다.
- 원장 스키마 · 설정 키 · 서명 매니페스트.
- **남는 한계(명시):** 진입 판정 안의 버킷(B15) · 체인(B16 — kill switch · 일일 손실 · 손절 계약 · 재진입 · 입력 불가) · 노출(B17)은
  q_final 위에서 돌기 때문에 상한 0 뒤에 계속 가려진다. B9 계보 투영도 같다. 이 한계는 spec 문장에 남긴다.

## 결정이 필요한 것 (proposal-freeze 리뷰에서 함께 판정)

- **D2-c 일일 손실 강화와 겹칠 때:** 「결속 결함 + 일일 손실 한도 도달」이 겹친 요청은 지금은 강화(`EscalateDailyLossLimit`)로 가고 주기가 오류 없이 끝난다.
  A1 뒤에는 결속 결함(`AUTHORITY_MISMATCH`)이 되고 강화는 그 주기에 일어나지 않는다. 권고는 **수용**이다. 근거는 design D2-c 에 적었다.
- **D5 조립 감독자의 `EntryGate`:** 조립 감독자에는 알림 둘뿐 아니라 `EntryGate` 도 없다. 진입 게이트가 없는 조립에서는 중앙 무결성 고장이
  진입이 아니라 프로세스를 닫는다(:417-419 주석). 그리고 엔진 정지 = 손절 없음이다. 이번 범위는 사람이 정한 두 옵션뿐이다.
  `EntryGate` 를 같이 배선할지는 사람이 정한다(design D5).

## Impact

- Affected specs: `risk-management`(ADDED 1 · MODIFIED 1).
- Affected code:
  - `internal/execgw/riskguardian_first_leg.go` — `PrecheckQFinalCampaignFirstLeg` 순서 이동.
  - `internal/app/engine/strategy_entry_supervisor.go` — `(*Context).NewPairedStrategyEntryProductionAssembly` 옵션 두 칸.
  - 시험: 새 시험 파일 둘(execgw · engine).
- 위험 등급: **High-risk**(Guardian · execgw 첫 다리 판정 순서). FLM/BTM · Pre-Edit · RED 선행(현재 순서를 특성 시험으로 고정) · 변이 원장 ·
  독립 적대 리뷰가 의무다. A2 는 Medium(알림 배선, 판정 불변)이지만 같은 change 라 같은 절차를 따른다.
- 배포: 엔진 이미지 교체. 설정 · 서명 변경은 필요 없다. 현재 생산 레인은 DORMANT(서명 매니페스트 0)라 동작 변화는 레인 활성화 뒤에 나타난다.
