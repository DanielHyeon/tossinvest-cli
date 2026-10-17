# a146 review

## 사람 결정 기록 (2026-10-15)

Manager 전달(원문): 「USER DECISION (2026-10-15, quoted): Guardian 상한 0 → 레인 잠금 대신 「이번 진입만 건너뜀」. Budget stays 1/3 (a139). Explicitly REJECTED
by the user-facing analysis: moving latch state out of the ledger / in-memory control / async ledger writes (breaks a112 5.3.3 durable latch and the a136/a138
reservation-tx serialization). Out of scope: scaling quantity down to the remaining daily loss.」

## 특성 시험 (편집 전 HEAD, tasks 0.4)

실제 경로(a112 거래 fixture → 감독자 `runMarket`)로 a139 잔여 #1 을 잼 — `analysis/red/characterization-head.txt` 4/4 PASS(편집 전 HEAD 7ce7e411).
결과가 a139 서술을 둘 고침(design F6 · F7): 감독자 잠금은 원장이 아니라 메모리(프로세스 재시작까지), 오늘 생산 감독자는 권한 갱신 전용이라 잠금이 아니라
삼킴 + 같은 주기 뒤 범위 굶음.

## proposal-freeze 리뷰 (2026-10-17, tasks 0.5)

- 보이스: 분리된 컨텍스트의 적대 리뷰어 1(code-reviewer — 적대 Eng + CEO 범위), 자기 격리 worktree(7264a920 — Go 는 befea6ee 이후 특성 시험만 다름)에서
  특성 시험 4 + 기존 시험 재실행 PASS, 탐침 1(주간 예약 결속 결함 주입 + 예산 1 KRW → 상한 0 삼중이 결함을 가림). 판정 **APPROVE-WITH-FIXES** — P0 0 · P1 1 · P2 3 · P3 7.
  F1~F9 전부 참으로 확인(F3 정밀화: 원인은 nil 이 아니라 `decision.Refusal` — sentinel 없음). 안전 불변식 위반 없음, D7(범위 거절 밖) 근거 맞음.
- 수용 · 반영(이 커밋, production 편집 전):
  - **P1-1 상한 0 이 그 뒤의 판정을 가림** — design F10 신설(가려지는 판정 열거: first-leg B2 정책 버전 · B8 주간 예약 · B9 투영, entry B15~B17 버킷 · 체인 · 노출 —
    엔진 `validateStrategyFirstLegAuthority` 가 다시 막는 B3~B7 은 제외), D5 정정(「새로 가려지는 것 없음」 철회 → **명시 한계**), proposal · spec 에 한계 문장 ·
    시나리오, RED 시험 `TestA146ACapZeroSkipHidesTheChecksAfterSizingUntilTheCapRises`(체인 입력 불가를 양쪽으로). 리뷰의 (a) 안 — 결속 판정을 sizing 앞으로 —
    은 execgw High-risk 편집이고 다른 동시 발생(버킷 고갈 · 일일 손실 강화 + 결속 결함)의 표면 사유까지 바꿔 **채택하지 않고 사람 결정 잔여**로 둠((b) 안 채택).
  - P2-1 「서명 활성화」와 감독자 `Effective` 혼동 — spec · proposal · design D2 를 worker 종류로 고쳐 씀(effective 쪽은 방어 불변식, 생산 효과는 refreshOnly 셋).
  - P2-2 notional 상한 0 포함 — 결정 문면(「Guardian 상한 0」)대로 포함하되 proposal 결정 블록에 「해석 — 확인 대기」로 올리고 Manager 보고의 질문으로 넘김
    (답이 「위험 예산만」이면 `risk` 에 별도 sentinel 후속).
  - P2-3 가시성 — 빈 일 키에서도 알림(a137 의 침묵을 따르지 않음), 스냅숏 칸은 진단용임을 명시(생산 투영은 읽지 않음).
  - P3-1 판정식 강화 — sentinel 은 q_final 거절 **자신의** 원인 사슬(`errors.Is(qFinal, …)`), 곁에 Join 된 모양은 결함(음성 시험 행). 양성 표본은 실제 Guardian 거절.
    `Field` 조건은 넣지 않음(이 코드의 두 생산 자리가 같은 Field — 판별력 0).
  - P3-2 알림 줄에 진입가 · 손절가(시장 가격). P3-3 D2 표에 「발급 뒤 스킵 → 공유 버킷 STALE 잠금(사유만 바뀜)」 행. P3-4 기록 자리는 admit 유지 — 실제 경로 시험이
    「주기마다 스킵 수 = precheck 수」로 두 자리의 일치를 잼. P3-5 「a137 갈래 앞으로」 변이는 동등(입력 서로소)으로 미리 선언. P3-6 문구(브로커 호출 · 「프로세스 안에서」
    한 번 · AST 기준 sha 주석). P3-7 비례 — 지적 없음.
- 거절: 없음(P1-1 의 (a) 안은 거절이 아니라 범위 밖 사람 결정으로 보냄).

## Pre-Edit Gate

```text
Pre-Edit Gate:
- change id / task id: a146-guardian-cap-zero-skips-the-entry / 1.1~1.3
- 대상 심볼(편집하는 기존 함수 — 번들 analysis/function-logic/, ast-base.json = base 7ce7e411 소스):
    internal/app/engine/strategy_first_leg_admission.go   strategyFirstLegAdmissionBridge.admit (분류 갈래 하나) · 코드 상수 하나
    internal/app/engine/strategy_market_handoff_delivery.go deliverEachStrategyHandoff (스킵 continue 갈래 하나)
    internal/app/engine/strategy_entry_supervisor.go      StrategyEntrySupervisor.runMarket (기록자 문장 하나) · Snapshot (칸 둘) ·
                                                          NewStrategyEntrySupervisor (옵션 칸 하나) · Context.NewRefreshingPairedStrategyEntrySupervisor (배선 한 줄) ·
                                                          옵션 · 스냅숏 · 런타임 · 감독자 구조체 칸(함수 아님)
    시험: internal/app/engine/a112_qfinal_family_test.go TestAnExistingGuardianCapRefusalIsNotAScopeRefusal (새 동작) ·
          internal/strategyworker/policy_receipt_test.go TestTheProductionThresholdMatchesAnEngineThatKeepsNoFailureCounter (칸 목록 + 근거)
  새 파일(새 leaf): internal/app/engine/a146_entry_skip.go · internal/obs/event.go 상수 하나(함수 아님)
  무편집: internal/risk · internal/execgw · internal/riskbucket · journal · 설정 · 매도/손절/청산 경로 전부
- CodeGraph: analysis/code-context/ (admit 생산 호출자 누락 1 — HEAD grep 으로 해소) · callers/impact raw
- 기존 동작 근거: HEAD 읽기 + analysis/ast-evidence/ 19 + 특성 시험 4(편집 전 통과) + freeze 리뷰 실측
- upstream 상속 시험 영향: no — 게이트 OFF 는 Guardian · 전략 감독자 경로를 만들지 않음, 엔진 패키지 밖 동작 무변경
- 실패 시험 선행: yes (analysis/red/ — 컴파일 RED + census 행동 RED + 특성 시험의 편집 전 반대 동작)
- 설정 · DB · journal 변경: 없음(새 키 0 · 스키마 0 · 원장 쓰기 0). rollback = 이전 이미지
- §0 검토: 통과 — 주문을 늘리는 판정 없음(건너뛴 진입 브로커 0, 뒤 범위는 전체 Guardian 판정 그대로), 손절 · 청산 경로 무관(F9), 토글 OFF 무관,
  사이징 무편집(§0.9), 잠금 저장소 무편집(사용자 기각 안 회피). 명시 한계 F10(상한 0 이 결함을 가림 — 주문 0)
```
