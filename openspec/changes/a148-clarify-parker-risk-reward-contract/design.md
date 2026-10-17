## Context

우선순위 1/9: Parker RR 계약 분리. 선행 change: 없음.
이 문서는 목표 동작을 명세한다. 소스 구현·배포·실거래 검증 완료를 뜻하지 않는다.
공통 기준 증거: `../a148-clarify-parker-risk-reward-contract/analysis/code-context/`; 구현 전에는 현 HEAD의 호출 사슬·설정 binding과 수정 함수 AST를 새로 확인한다.

## Goals / Non-Goals

**Goals:** Parker의 공간 RR, 목표 RR, 실제 실현 RR를 같은 숫자로 해석하면 손절·수량·익절의 검증 기준이 달라진다. frozen v1 동작을 보존하면서 세 의미를 먼저 구분한다.

**Non-Goals:** 기존 frozen v1, r1, a128/a147 증거, 인증·토큰 공유, 원장과 위험 통제는 보존한다. 이 문서는 운영 설정 변경·실주문·재시작·추가 세션 수집 승인이 아니다. Manager(Astra)가 설계를 작성·검토하고, Terra Teammate가 코드·테스트를 담당한다. 구현 writer와 독립 reviewer는 분리한다. gstack 리뷰, 해당 change 전체 gate, Manager 최종 확인 후에만 Terra가 범위 한정 커밋·푸시·정식 archive/PM sync를 수행한다.

## Decisions

신규 진단 계약은 initial_risk=entry-initial_stop, space_rr=(forward_obstacle-entry)/initial_risk, target_rr=(target-entry)/initial_risk, realized_rr=net_realized_pnl/initial_account_risk로 분리한다. 분모가 0 이하이거나 비용·체결이 없으면 invalid/unknown이지 0이 아니다. realized_rr의 initial_account_risk는 실제 진입 체결 수량·평균가와 최초 보호 손절, 진입 시점 환율로 고정한다. 비교 보고에서는 계획 R과 체결 R을 별도 표시한다.
기존 v1 ID, predicate, stop/target, reason code와 fixture는 그대로 둔다. RR_GE_2 레거시 표시는 gate 증명으로 사용하지 않고 raw compatibility field로 남긴다. 새 진단 필드/문구만 명확히 한다. 강화된 공간 RR 필터나 손절 변경은 새 정책 버전과 후속 실험·승인 사항이며 이번 change에서 추가하지 않는다.
100 가격·99.3 손절·1.2 전방 공간이면 space_rr=12/7, target=102.1이면 target_rr=3이다. 3R target은 3R 이익 보장도, target까지 무장애 공간 확보도 아니다. 순수 계산을 우선 재사용하고 별도 RR 프레임워크는 만들지 않는다.

## Risks / Trade-offs

- 과적합·표본 부족 → 수익성 주장과 기능 정확성 판정을 분리하고 unknown/inconclusive를 보존한다.
- 기존 계약 회귀 → frozen fixtures, OFF 동작, 기존 포지션 snapshot을 보존한다.
- 구현 표면 확대 → internal/strategyengine/{lane.go,contracts.go}, existing Parker fixtures; no activation or dispatch changes. 범위부터 조사하고 기존 기능을 재사용한다. 별도 strategy framework를 만들지 않는다.
- 위험도: Low for documentation/diagnostics; High if any entry predicate or RiskIntent changes. high-risk 변경은 FLM/BTM·Pre-Edit Gate와 적대적 리뷰가 선행된다.

## Migration Plan

설계→선행 change 검증→Terra 구현/시험→독립·gstack 리뷰→Manager 검증→전체 gate→범위 한정 커밋/푸시/정식 archive. DB 변경이 없으면 migration은 not-applicable로 기록한다. 변경이 필요하면 기존 reader 호환성과 additive-nullable rollback을 먼저 리뷰한다. 운영 반영은 별도 승인 전 실행하지 않는다.

## Open Questions

선행 산출물의 실제 가용성, 기존 구현 재사용 범위와 검증 자료의 충분성은 구현 착수 시 확인한다. 자료가 없으면 승인·측정을 추측하지 않고 해당 기능을 unsupported/not_ready로 남긴다. 후속 정책 성능이나 live 승인을 이 설계로 대체하지 않는다.
