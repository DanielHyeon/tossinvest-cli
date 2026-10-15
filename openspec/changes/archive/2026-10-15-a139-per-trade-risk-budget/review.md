# a139 review

## 사람 결정 기록 (2026-10-15)

Manager 전달: 「a139-per-trade-risk-budget (user decision 2026-10-15: RiskBudget = MaxDailyLoss / 3)」 — 보수 사이징만(불변식 6), 내림은 보수(아래),
RiskBudget 을 정의하는 요구를 provenance 와 함께 MODIFIED, 모든 소비자를 CodeGraph + AST 로 확인하고 시험으로 고정.

## proposal-freeze 리뷰 (2026-10-15)

- 보이스: 분리된 컨텍스트의 적대 리뷰어 1(code-reviewer — 적대 Eng + CEO 범위), 읽기 전용, HEAD 044b7b94. 자기 사본에서 DefaultPolicy · 유도를 바꿔 시험 영향 실측.
- 판정: **APPROVE-WITH-FIXES** — P0 0 · P1 2 · P2 4 · P3 4. 보수 방향 확인(세 상한 min 단조 · q_final = min(…), 예산 축소로 넓어지는 분기 없음, 매도는 사이징 전 분기).
- 수용 · 반영(이 커밋, production 편집 전):
  - P1-1 Guardian 상한 0 의 출력 소비자 — `EXISTING_GUARDIAN_CAP` 은 범위 거절이 아니어서 첫 다리 `AUTHORITY_MISMATCH` → 레인 첫 실패 잠금(원장). design D0 보강(AST
    `engine--strategyfirstlegadmissionbridge.admit` 추가) · D5 · 특성 시험. **분류 변경은 a139 밖 — 사람 결정 잔여**.
  - P1-2 유도 불가일 때 audit 줄 — 빈 값 + 사유, 기동을 막지 않음(게이트 OFF = upstream). spec 문장 · 시나리오, RED 둘(게이트 OFF 일일 손실 0 기동 · JPY/0.02 USD 게이트 ON
    은 factory `ErrLimitsRequired`).
  - P2-1 「reason-code 무변경」 정정(입력 단위로는 MAX_ORDER_EXCEEDED → INVALID_ORDER_SIZE 가능). P2-2 기존 시험 5 건 목록 · 고치는 방법(tasks 1.1).
  - P2-3 정책 버전 v2 로 올림(D4 뒤집음 — 비용 실측 0, 휴면 어댑터만 fail-closed). audit detail 에 규칙. P2-4 「정책 수치의 provenance」 MODIFIED.
  - P3 digest 서술 갱신(a138 의 개수 상한 포함) · guardian_wiring 주석 · config/limits.go 낡은 줄 참조 · factory 사전 검사 · audit 값 = 생산 유도의 답 대조 시험.
- 리뷰어가 본 미커밋 a139 RED 시험 · 번들 · 하네스는 이 저자의 것(리뷰와 병행 — production 편집은 리뷰 뒤).

## Pre-Edit Gate

```text
Pre-Edit Gate:
- change id / task id: a139-per-trade-risk-budget / 1.1~1.3
- 대상 심볼(편집하는 기존 함수 — 번들 analysis/function-logic/, ast-base.json = base 소스):
    internal/risk/input.go                  DefaultPolicy (값 · 주석)
    internal/app/engine/guardian_wiring.go  riskPolicyFromAutomationGate (유도 분기 하나) · newProductionRiskGuardian (사전 검사 분기 하나) · 버전 상수 · 주석
                                            (wholeOpenPositions 는 이웃 편집만 — revision: base)
    internal/app/engine/interlock.go        newAutomationStatus · acceptanceDetail · recordGateSettings (칸 · 줄 하나씩)
    시험 fixture 5: guardian_wiring_internal_test · riskguardian_account_base_testseam_test 둘 · chain_test · contract_test
  새 파일(새 leaf): internal/risk/risk_budget.go (PerTradeRiskBudget) · guardian_wiring.go 끝의 perTradeRiskBudgetAudit
  주석만: internal/config/limits.go (var 블록 안 낡은 줄 참조 — 함수 아님)
- CodeGraph: analysis/code-context/ (DefaultPolicy 호출자 누락 1 — HEAD 로 해소) · RiskBudget 소비자 전수 design D0
- 기존 동작 근거: HEAD 읽기 + analysis/ast-evidence/ 13 · freeze 리뷰 실측(시험 5 건)
- upstream 상속 시험 영향: no — 게이트 OFF 는 Guardian 을 만들지 않음, 유도 불가 audit 는 빈 값으로 기동을 막지 않음(시험)
- 실패 시험 선행: yes (analysis/red/)
- 설정 · DB · journal 변경: 없음(새 키 0 · 스키마 0). rollback = 이전 이미지. 정책 버전 문자열 v1 → v2(결정 preimage 기록만 — 기대 버전은 Guardian 자기 유도)
- §0 검토: 통과 — 사이징만 줄임(§0.9), 매도 · 손절은 사이징 단 전에 분기(AST Evaluate B2 · preflight B6)
```

## 구현 독립 적대 리뷰 (2026-10-15, tasks 2.3)

- 보이스: 분리된 컨텍스트 적대 리뷰어 1(code-reviewer — 적대 Eng · 안전 불변식 · 시험 반증), 자기 격리 worktree(dadc386a)에서 전체 시험 · 자체 변이 · 산술 속성 시험(무작위 · 경계 float 약 20 만, 0 실패).
- 판정: **APPROVE-WITH-FIXES** — P0 0 · P1 1 · P2 2 · P3 4. 생산 산술 · 배선 정확, 방향은 사이징 축소뿐.
- 수용 · 수리(이 커밋):
  - P1-1 생산 Guardian **객체**의 예산을 재는 시험 0 — 생성자에서 예산을 되돌리는 변이가 스위트를 통과(audit 는 여전히 1/3 을 적음) → 실제 journal 로 생산 Guardian 을 만들어
    IssueEntry 위험 상한 66 · 67 을 재는 시험(변이 N16).
  - P2-1 RED 목록의 빈칸 — 체인 경계 한 단위 · account-base(FX) 경로 KR 10 / US 4 · 매도 무관 시험 추가(변이 N18 · N19), tasks 문구를 실제 처리(짝 시험의 예산 고정)에 맞춤.
  - P2-2 잔여 서술 정정 — 재분류만으로 잠금을 피하는지는 미검증(범위 거절도 주기 오류 → latchMarket, 코드 읽기). design D0 보강.
  - P3-1 게이트 OFF 시험이 줄 부재를 통과 → 줄 존재 단언(변이 N17). P3-2 spec 「한 함수」 문구 정정(기본 정책은 같은 규칙의 값을 시험이 대조). P3-3 「한 거래가 그날 손실 한도 전부」를
    「금액 상한」으로 한정(비율 상한은 이 규칙이 보지 않음). P3-4 audit detail 에 옛 규칙(「previously = max_daily_loss_amount, risk-policy-v1」).
- 잔여(사람 결정 후보, a139 밖):
  1. Guardian 상한 0 → `EXISTING_GUARDIAN_CAP` → 레인 잠금(design D0 보강) — 재분류 전에 `latchMarket` 까지의 특성 시험 필요. 현재 생산 레인 DORMANT.
  2. 예산을 min(금액, 비율 × 자본)/3 로 — 보수 방향 후속 후보(런타임 자본 필요).
