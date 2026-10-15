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
