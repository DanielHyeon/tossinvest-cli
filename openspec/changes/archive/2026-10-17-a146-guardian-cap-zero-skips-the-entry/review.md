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
- §0 검토: 통과 — 건너뛴 진입의 브로커 0. 다만 굶음이 풀려 같은 주기의 뒤 범위가 주문까지 갈 수 있음(design D5 — 각 범위는 전체 Guardian 판정 그대로, 구현 리뷰 2 P3-4 정정), 손절 · 청산 경로 무관(F9), 토글 OFF 무관,
  사이징 무편집(§0.9), 잠금 저장소 무편집(사용자 기각 안 회피). 명시 한계 F10(상한 0 이 결함을 가림 — 주문 0)
```

## 사람 결정 (2026-10-18 — Manager 전달, 사용자 「권장대로」)

1. 주문 금액(notional) 상한 0 포함 — 구현 유지, 문구 「Guardian 상한 0 (위험 예산 또는 주문 금액 상한으로 1주도 안 되는 진입)」(freeze P2-2 종결).
2. F10 은 알려진 한계로 수용 — execgw 결속 판정을 사이징 앞으로 옮기는 별도 후속 change 가 전략 레인 활성화(P2 S5/S6) 전에 착지해야 함(freeze P1-1 종결).
3. 범위 거절의 effective 잠금 유지(a112 J4 · a137 spec) — 특성 시험 유지(design D7 종결).

## 변이 원장 (tasks 2.2)

round1(7395cc02) 21 중 CAUGHT 17 · build 2 · SURVIVED 2 → round2(01ca7ff1, 저장소 밖 pid 사본, 무변이 대조군 GREEN): **비동등 20/20 CAUGHT**, M1 동등(사전 선언).
`analysis/mutation/round1.md` · `round2.md`.

## 구현 독립 적대 리뷰 (Manager 가 띄운 분리 컨텍스트 둘, tasks 2.3)

- 리뷰 1(7395cc02): **APPROVE-WITH-FIXES** — P0 0 · P1 0 · P2 1 · P3 6. 리뷰 2(01ca7ff1 + 936d7e0d): **APPROVE-WITH-FIXES** — P0 0 · P1 0 · P2 1(1 의 P2-1 과 같은 자리) · P3.
- 수용 · 수리(이 커밋):
  - P2-1(둘 다) 전달 루프가 타입만 보고 모양을 못 박지 않음 — 문구 흉내 · `|| errors.As(QFinalRefusal)` · 「원인 무관 EXISTING_GUARDIAN_CAP」 변이가 생존 →
    `strategyEntrySkipOf` 로 **두 모양만** 인정(스킵 자체 · dispatch 감싸개 하나), 음성 시험 `TestA146OnlyAPureSkipIsSkippedEveryLookalikeStops`(감싼 상한 거절 ·
    문구 흉내 · Join 두 순서 · 결함이 감싼 스킵 · 다른 코드 감싸개 · 이중 감싸개).
  - 승격 P3-1(리뷰 2) `errors.As` 가 Join/감싸개 트리를 뒤져 결함이 사라짐 — 위 판정이 다중 오류를 어느 자리든 거절, 탐침 두 모양을 시험으로.
  - P3-1(리뷰 1) 시장 귀속 — US 주기도 기록, US 알림 · 스냅숏에만 남음을 단언. P3-2 표 행 이름 정정(「sentinel beside another q_final code」), M1 동등 유지(grep 근거).
  - P3-3 문서 표류 — design D1 판정식 · proposal F 번호. P3-4 알림에 FX 권위(통화 쌍 · digest — 봉인 권위는 율을 내보이지 않아 율 대신), 값 결함 → 조용한 스킵을 알려진 한계로.
  - P3-6 일 키 seam 은 패키지 변수 유지 + t.Parallel 금지 · t.Cleanup 주석.
  - 리뷰 2 P3-2 경합 — 스킵 계수 · refreshOnly 시험을 `a146_entry_skip_race_test.go` 로 옮겨 `RACE_ENGINE_FILES/TESTS` 에 등록(완전성 가드 통과).
  - 리뷰 2 P3-3 notional 상한 0 — spec 시나리오 + `TestA146ANotionalCapBelowOneUnitIsTheSameZeroSentinel`(risk, KR · US FX, 경계 한 단위) +
    `TestA146ANotionalCapZeroIsTheSameSkipOnTheRealPath`(엔진 실제 경로, KR · US, 한 주 대조군은 주문까지).
  - 리뷰 2 P3-4 문서 — F7 · proposal 에 굶음 조건(검증된 4-가족 활성화), Pre-Edit §0 문장 정정. shadow 단계 시작은 design D5 · proposal Impact 에 이미 적힘(확인).
  - 리뷰 1 P3-5 조립 감독자(strategy_entry_supervisor.go:379 — Effective 가능, 오늘 Run 안 됨)에 알림 옵션 없음 — tasks 후속 메모.
- 변이 round3(35db12b7): 비동등 8/8 CAUGHT(R10 · R12 · M35 · M36 · M31 · P2 · P3 · P4). P1(다중 오류 거름)은 `errors.Unwrap` 이 다중 오류에서 nil 이라 중복 → 지움(시험 행이 같은 모양을 계속 잡음).
  `make test-race` rc 0, a146/a137 집중 시험 `-race` · 태그 통과. `round3.md`.
