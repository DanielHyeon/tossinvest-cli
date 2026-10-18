## Context

우선순위 7/9: 검증된 손절 정책 사용자 선택. 선행 change: a150, a153.
이 문서는 목표 동작을 명세한다. 소스 구현·배포·실거래 검증 완료를 뜻하지 않는다.
공통 기준 증거: `../archive/2026-10-18-a148-clarify-parker-risk-reward-contract/analysis/code-context/`; 구현 전에는 현 HEAD의 호출 사슬·설정 binding과 수정 함수 AST를 새로 확인한다.

## Goals / Non-Goals

**Goals:** 사용자는 초기 손절과 실제 청산 정책을 구분해서 선택해야 한다. 기존 설정 lifecycle을 확장하고 검증되지 않은 옵션이 운영 기본값이 되지 않게 한다.

**Non-Goals:** 기존 frozen v1, r1, a128/a147 증거, 인증·토큰 공유, 원장과 위험 통제는 보존한다. 이 문서는 운영 설정 변경·실주문·재시작·추가 세션 수집 승인이 아니다. Manager(Astra)가 설계를 작성·검토하고, Terra Teammate가 코드·테스트를 담당한다. 구현 writer와 독립 reviewer는 분리한다. gstack 리뷰, 해당 change 전체 gate, Manager 최종 확인 후에만 Terra가 범위 한정 커밋·푸시·정식 archive/PM sync를 수행한다.

## Decisions

화면은 초기 보호 손절 정책, 수량 산정 기준, 전략적 종료 조건, 공통 이익보호 정책을 구분한다. strategy/market별 검증된 finite option ID만 descriptor에 싣고 임의 숫자·자유 입력은 허용하지 않는다. initial stop selection과 기존 common exit selector를 혼합하거나 새 설정 저장소를 만들지 않는다.
기본값은 legacy/unapproved 그대로다. a153 추천과 사람의 정책 승인 없이는 후보가 disabled/read-only다. synthetic fixture로 UI를 개발할 수 있지만 mock validation은 운영 승인 자료가 아니다. preview에는 stop 가격·거리, 계획·체결 기준가, 산정 수량·예상 risk, 비용 가정, source/as_of, apply timing, policy version을 표시한다.
CAS preview/apply/audit/rollback 계약을 재사용한다. 이 작업에서는 실제 apply 요청을 실행하지 않는다. approved apply도 새로운 campaign부터만 적용되고 기존 포지션 stop 완화, 레인 ON, scheduler 변경, kill switch 해제 권한을 갖지 않는다. read-only 화면에서 unknown은 0%로 나타내지 않는다.

## Risks / Trade-offs

- 과적합·표본 부족 → 수익성 주장과 기능 정확성 판정을 분리하고 unknown/inconclusive를 보존한다.
- 기존 계약 회귀 → frozen fixtures, OFF 동작, 기존 포지션 snapshot을 보존한다.
- 구현 표면 확대 → Existing optimization descriptor registry, preview/CAS lifecycle and console; reuse common exit policy picker. 범위부터 조사하고 기존 기능을 재사용한다. 별도 strategy framework를 만들지 않는다.
- 위험도: High at settings apply boundary; UI preview itself is read-only. high-risk 변경은 FLM/BTM·Pre-Edit Gate와 적대적 리뷰가 선행된다.

## Migration Plan

설계→선행 change 검증→Terra 구현/시험→독립·gstack 리뷰→Manager 검증→전체 gate→범위 한정 커밋/푸시/정식 archive. DB 변경이 없으면 migration은 not-applicable로 기록한다. 변경이 필요하면 기존 reader 호환성과 additive-nullable rollback을 먼저 리뷰한다. 운영 반영은 별도 승인 전 실행하지 않는다.

## Open Questions

선행 산출물의 실제 가용성, 기존 구현 재사용 범위와 검증 자료의 충분성은 구현 착수 시 확인한다. 자료가 없으면 승인·측정을 추측하지 않고 해당 기능을 unsupported/not_ready로 남긴다. 후속 정책 성능이나 live 승인을 이 설계로 대체하지 않는다.
