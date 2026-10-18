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

### 구현 착수 검토 (2026-10-18, Manager)

선행 a148·a149는 main 반영 및 정식 archive를 완료했다. 신규 admission과 첫 exit 초기화 사이의 정책 snapshot 연결을 우선 조사한다. 기존 `RiskIntent.PolicyVersion`은 Guardian 버전이므로 exit 정책 식별자로 재해석하지 않는다. 이미 저장된 `ExecutionTerms`, 전략 lineage payload와 `ExitStateSeed.PolicyIdentity`를 먼저 재사용한다. 실제로 누락된 정보만 추가하며, 과거 campaign·adopted 경로와 미연결 breakout/후속 leg는 자동 전환·활성화하지 않는다. 기존 함수를 수정하기 전 fresh AST/FLM/BTM과 atomic 영속·복구 경로를 확인하고 Manager가 구체 변경안을 검토한다.

a149 리뷰에서 발견된 a148 진단 입력의 숫자 파싱 제한 누락도 초기 위험 입력 계약의 후속 항목으로 처리한다. 대상은 `rr_diagnostic.go`의 호출자 제공 숫자뿐이다. 길이가 제한된 십진 표기만 허용해 `p/P` 지수 등 과도한 파싱을 거부하며, frozen `lane.go`의 공용 파서나 기존 판단·봉인 형식은 변경하지 않는다. 기존 a148 결과값과 unknown/invalid 의미를 유지하고 입력 경계 회귀 시험을 추가한다. 이 발견은 배포된 호출 경로나 실거래 악용의 증거를 뜻하지 않는다.

### 승인된 영속·복구 경계 (2026-10-18, Manager)

1. 신규 지원 first-leg admission은 exit policy kind·ID/version/digest를 v4 전략 risk-binding payload에 고정한다. 기존 v2/v3 직렬화·identity 및 RiskIntent는 바꾸지 않는다. optional lineage가 없어졌을 때 v4를 legacy로 오인하지 않도록 decision 행에 독립적인 보호 계약 버전 표지를 additive migration으로 추가한다. 기존/명시적 legacy 경로는 0, 신규 v4 first-leg은 4를 동일 admission transaction에 기록한다. v4 binding 누락·손상은 legacy fallback 사유가 아니다. 기존 payload나 보호 snapshot을 backfill하지 않는다.
2. 신규 production first-leg writer는 v4를 필수로 전달하며 replay도 같은 snapshot과 identity여야 한다. Guardian뿐 아니라 `RecordQFinalCampaignFirstLeg` journal 경계에서도 신규 지원 strategyflow admission은 v4를 요구한다. 과거 marker 0/v2/v3의 정확한 replay·읽기·복구는 보존하되, 기존 시험 fixture 편의를 위한 신규 legacy admission 우회 옵션은 만들지 않는다. 신규 admission fixture는 v4로 전환하고 과거 호환성은 historical DB fixture로 검증한다. 이 writer 밖의 명시적 Parker frozen 및 adopted 보호 경로, 미연결 breakout/후속 leg는 자동 전환하지 않는다. 독립 버전 표지는 admission 이후 변경할 수 없으며, 지원 경로 목록과 실제 writer를 시험으로 대응한다.
3. post-fill 복구는 account·campaign·entry decision·position generation을 대조하는 좁은 journal API를 사용한다. 같은 transaction에서 campaign의 추가 진입 차단과 idempotent reconciliation 기록을 영속한 뒤 기존 OpenExitState로 승인 snapshot을 복원한다. 종료 campaign을 활성 상태로 되살리지 않는다. 차단 commit 이후 초기화 전 중단은 재시작 시 복구하며, 이미 유효한 보호 평가는 유지한다. 손절 평가까지 제외하는 quarantine은 이 경로에서 사용하지 않는다.
4. 최종 transport authority 검사는 하나의 일관된 DB transaction snapshot에서 독립 버전 표지, v4 payload, campaign 차단, position generation 및 관련 exit binding을 다시 검증한다. read 실패나 재현 가능한 영속 불일치는 신규 승인을 거부한다. 차단 저장 실패에도 이 검사는 별도로 작동하며, 보호 평가는 중단하지 않고 차단 재시도·실패 경고를 남긴다. 유효한 admitted snapshot이 없으면 정책을 추측하지 않고 unprotected/unknown을 명시하며 이미 존재하는 보호는 유지한다.
5. 동시성 보장은 **차단 이후 새 전송 승인 거부**다. 최종 authority의 일관된 읽기 시점을 이미 지난 명령은 전송 진행 중일 수 있다. 이후 차단으로 실제 bytes 미전송·취소·체결 거절을 단정하지 않는다. 기존 SUBMITTING은 보수적으로 possibly-in-flight로 취급하며 예약·미확정·대사를 유지한다. 재사용된 승인으로 새 전송을 허용하지 않는다. 신규 네트워크 구간 잠금이나 손절 스레드가 기다리는 EntryGate 잠금은 추가하지 않는다.
6. 이 보장은 영속 권위에서 다시 검증할 수 있는 불일치에 적용한다. 일시적 표시/옵션 차이만으로 저장 실패 이후의 영속 차단을 주장하지 않는다. 현재 옵션이 달라도 유효한 승인 snapshot이 우선한다. fresh AST/FLM/BTM, 차단 commit 전후 및 latch-write 실패의 결정적 경쟁 시험을 구현 전후 증거로 남긴다.
7. 복구에 사용할 admission은 봉인된 projection의 position generation까지 해당 포지션과 일치해야 한다. 다른 세대의 policy를 오류와 함께 반환해 새 exit seed에 재사용하지 않는다. 동일 세대의 유효 admission과 기존 보호가 있으나 기존 손절이 더 느슨한 경우, 평가 baseline에 `max(existing baseline, admitted stop)`을 적용하고 기존 evaluator·journal의 단조 snapshot 갱신 경로를 사용한다. 기존 initial stop/R·entry·policy·pending order의 출처를 덮어쓰지 않으며 이미 더 타이트한 보호는 유지한다. latch 저장 실패에도 유효한 보호 floor 평가는 계속한다. snapshot selector 검증을 우회하거나 high-water를 조작하지 않는다. 복구 tuple이 거부되면 기존 보호를 유지하면서 복구 미완료를 명시하며 정상 복구로 보고하지 않는다. 새 자동 flatten 경로는 추가하지 않는다.

- 과적합·표본 부족 → 수익성 주장과 기능 정확성 판정을 분리하고 unknown/inconclusive를 보존한다.
- 기존 계약 회귀 → frozen fixtures, OFF 동작, 기존 포지션 snapshot을 보존한다.
- 구현 표면 확대 → internal/strategyproposal, strategyflow, continuationlane, reversallane, weeklyvaluelane, breakoutlane, risk, riskbucket, execgw, app/engine exit initialization and journal bindings. 범위부터 조사하고 기존 기능을 재사용한다. 별도 strategy framework를 만들지 않는다.
- 위험도: High: sizing, protection, RiskIntent, reservations, journal and exit initialization. high-risk 변경은 FLM/BTM·Pre-Edit Gate와 적대적 리뷰가 선행된다.

### 보호선 갱신과 주문 처리의 순서

승인 손절 floor를 처음 높이는 관찰에서는 기존 effective snapshot ID(없음 상태 포함)를 journal transaction에서 대조한 뒤 호환되는 평가 snapshot만 영속한다. 오래된 관찰은 selector 실행 전에 충돌로 거절하며, 최신 보호를 quarantine하지 않는다. 이 단계는 명시적인 floor-reconciliation 사유로 주문 제안을 보류하고 기존 주문 취소·신규 주문을 실행하지 않는다. 기존 working-order suppression 사유를 거짓으로 재사용하지 않는다. 다음 정상 관찰 주기가 저장된 floor를 읽어 기존 보호 주문 경로로 진행한다. 따라서 최초 floor 복구에는 한 관찰 주기의 주문 지연이 있으며 즉시 체결 보장을 주장하지 않는다. 기존 보호·대기 주문과 다른 포지션 평가는 계속 유지한다. 갱신 후 반복 보류가 발생하지 않는지, 재시작과 충돌·실패 시 보호 보존 및 다음 주기의 sellable 제한 주문을 검증한다. 동일 관찰에 가짜 ID를 붙이거나 새 re-arm 권한·네트워크 잠금을 추가하지 않는다.

평가 snapshot이 아직 없거나 legacy writer가 scalar 상태만 갱신하는 구간도 동일하게 보호한다. floor 전용 CAS는 snapshot ID뿐 아니라 관찰 당시 evaluator 입력인 baseline·high-water·level/rung·수량·taken ratio·pending action 등을 대조한다. 기존 lifecycle/position generation 검사는 유지한다. 최신 watermark나 체결·대기 상태가 바뀌면 오래된 floor 갱신을 거절하며 최신 상태를 덮어쓰지 않는다.

## Migration Plan

설계→선행 change 검증→Terra 구현/시험→독립·gstack 리뷰→Manager 검증→전체 gate→범위 한정 커밋/푸시/정식 archive. DB 변경이 없으면 migration은 not-applicable로 기록한다. 변경이 필요하면 기존 reader 호환성과 additive-nullable rollback을 먼저 리뷰한다. 운영 반영은 별도 승인 전 실행하지 않는다.

## Open Questions

선행 산출물의 실제 가용성, 기존 구현 재사용 범위와 검증 자료의 충분성은 구현 착수 시 확인한다. 자료가 없으면 승인·측정을 추측하지 않고 해당 기능을 unsupported/not_ready로 남긴다. 후속 정책 성능이나 live 승인을 이 설계로 대체하지 않는다.
