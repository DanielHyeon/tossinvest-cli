## Context

우선순위 3/9: 전략별 초기 손절 수량 원장 청산 일관성. 선행 change: a148, a149.
이 문서는 목표 동작을 명세한다. 소스 구현·배포·실거래 검증 완료를 뜻하지 않는다.
공통 기준 증거: `../archive/2026-10-18-a148-clarify-parker-risk-reward-contract/analysis/code-context/`; 구현 전에는 현 HEAD의 호출 사슬·설정 binding과 수정 함수 AST를 새로 확인한다.

## Goals / Non-Goals

**Goals:** 전략마다 초기 손절의 출처가 다르더라도 승인된 가격·수량·정책이 주문·원장·청산까지 같은 계약으로 이어져야 한다. 계산과 전달 경계의 불일치부터 검사한다.

**Non-Goals:** 기존 frozen v1, r1, a128/a147 증거, 인증·토큰 공유, 원장과 위험 통제는 보존한다. 이 문서는 운영 설정 변경·실주문·재시작·추가 세션 수집 승인이 아니다. Manager(Astra)가 설계를 작성·검토하고, Terra Teammate가 코드·테스트를 담당한다. 구현 writer와 독립 reviewer는 분리한다. gstack 리뷰, 해당 change 전체 gate, Manager 최종 확인 후에만 Terra가 범위 한정 커밋·푸시·정식 archive/PM sync를 수행한다.

## Decisions

전략별 stop geometry producer와 consumer를 먼저 표로 고정한다. Parker frozen v1, continuation 8:4:2, reversal 2:4:8, weekly 7분할, breakout, adopted position을 별도 행으로 둔다. 공급 stop을 읽는 것과 local indicator로 stop을 만드는 것을 구분하며, 미연결 경로는 unsupported로 기록한다.
새 계약 대상은 policy_id/version, initial_stop, entry_price_basis, computed_at, source_digest, campaign/lane/market identity, risk_budget/account_currency/FX provenance, per_share_risk, final_quantity와 exit_policy_snapshot이다. 기존 ExecutionTerms/RiskIntent/예약 값을 재사용하고 동일 필드를 복제한 중앙 policy framework는 만들지 않는다. 새 필드가 필요할 때만 additive-nullable migration을 만들고 과거 row를 덮어쓰지 않는다.
계획 entry 기준 R과 실제 fill 평균가 기준 realized risk를 동시에 표시하되 기존 exit snapshot의 기준가를 자동 변경하지 않는다. 정책 변경은 신규 campaign만 대상으로 한다. final quantity는 모든 기존 Guardian·notional·exposure·bucket 제한의 min 이하여야 하고 zero cap을 무제한으로 해석하지 않는다. 비용은 어디서 적용되는지 명시하고 이중 계상하지 않는다.
주문 전 계약 누락·불일치는 신규 진입을 fail-closed한다. 주문 전 판단과 주문 후 보호 복구를 분리한다. 체결 이후 exit 초기화에서 불일치가 발견되면 체결을 소급 거절했다고 표시하지 않는다. 이미 승인된 stop/policy snapshot과 현재 authoritative sellable을 사용해 기존 보호 복구 경로로 넘기고, 해당 campaign 추가 진입을 차단하며 reconciliation/운영 경고를 남긴다. 복구 실패는 unprotected 상태로 명시하고 기존 비상 보호 절차를 유지한다. 이 변경은 자동 flatten 새 경로를 만들거나 기존 stop을 완화하지 않는다. 이미 보유한 포지션의 기존 보호 exit는 유지한다. 후속 leg·재시작·partial fill의 stop 동기화는 재현 fixture로 검증하고 미연결 live 후속 leg를 활성화하지 않는다. 기존 손절을 넓히는 수정은 이 change에서 금지한다.

## Risks / Trade-offs

- 과적합·표본 부족 → 수익성 주장과 기능 정확성 판정을 분리하고 unknown/inconclusive를 보존한다.
- 기존 계약 회귀 → frozen fixtures, OFF 동작, 기존 포지션 snapshot을 보존한다.
- 구현 표면 확대 → internal/strategyproposal, strategyflow, continuationlane, reversallane, weeklyvaluelane, breakoutlane, risk, riskbucket, execgw, app/engine exit initialization and journal bindings. 범위부터 조사하고 기존 기능을 재사용한다. 별도 strategy framework를 만들지 않는다.
- 위험도: High: sizing, protection, RiskIntent, reservations, journal and exit initialization. high-risk 변경은 FLM/BTM·Pre-Edit Gate와 적대적 리뷰가 선행된다.

## Migration Plan

설계→선행 change 검증→Terra 구현/시험→독립·gstack 리뷰→Manager 검증→전체 gate→범위 한정 커밋/푸시/정식 archive. DB 변경이 없으면 migration은 not-applicable로 기록한다. 변경이 필요하면 기존 reader 호환성과 additive-nullable rollback을 먼저 리뷰한다. 운영 반영은 별도 승인 전 실행하지 않는다.

## Open Questions

선행 산출물의 실제 가용성, 기존 구현 재사용 범위와 검증 자료의 충분성은 구현 착수 시 확인한다. 자료가 없으면 승인·측정을 추측하지 않고 해당 기능을 unsupported/not_ready로 남긴다. 후속 정책 성능이나 live 승인을 이 설계로 대체하지 않는다.
