## Context

우선순위 9/9: 실거래 연결과 후속 진입 마감 복구 승인. 선행 change: a155.
이 문서는 목표 동작을 명세한다. 소스 구현·배포·실거래 검증 완료를 뜻하지 않는다.
공통 기준 증거: `../a148-clarify-parker-risk-reward-contract/analysis/code-context/`; 구현 전에는 현 HEAD의 호출 사슬·설정 binding과 수정 함수 AST를 새로 확인한다.

## Goals / Non-Goals

**Goals:** 새 정책을 실거래에 연결하려면 최초 진입뿐 아니라 후속 leg·부분 체결·마감·복구까지 같은 안전 계약을 충족해야 한다. 개발 완료와 운영 승인을 분리한다.

**Non-Goals:** 기존 frozen v1, r1, a128/a147 증거, 인증·토큰 공유, 원장과 위험 통제는 보존한다. 이 문서는 운영 설정 변경·실주문·재시작·추가 세션 수집 승인이 아니다. Manager(Astra)가 설계를 작성·검토하고, Terra Teammate가 코드·테스트를 담당한다. 구현 writer와 독립 reviewer는 분리한다. gstack 리뷰, 해당 change 전체 gate, Manager 최종 확인 후에만 Terra가 범위 한정 커밋·푸시·정식 archive/PM sync를 수행한다.

## Decisions

이 change는 마지막 단계의 승인·검증 계약이다. 위험을 확대하거나 기존 손절을 늦추는 runtime 변경은 현재 승인 범위 밖이다. 첫 8단계의 검증·관측 결과, 정확한 신규 version/digest, 대상 market/symbol/lane, 최대 손실·노출·기간, 중단/복구 계획과 사람의 명시적 승인이 없으면 live-ready=false다.
후속 leg 구현은 예정 가중치만으로 완료 판정하지 않는다. 실제 filled quantity·pending reservation·sellable·새 위험예산·stop monotonicity를 재검증하고 신규 qFinal을 계산한다. broker accepted, partial filled, canceled, rejected, unknown을 분리한다. unknown 또는 cancel404는 terminal proof가 아니다. 중복 delivery/restart에서 주문 중복과 보호 수량 부족이 없어야 한다.
당일 마감 정책은 intraday campaign에만 적용한다. weekly/adopted 보유분을 계좌 전체 청산으로 묶지 않는다. 거래소 달력의 auction/deadline을 기준으로 충분한 취소·청산 시간을 설계하고 15:30 시장가 한 번으로 무조건 flat이라고 판정하지 않는다. final fill·잔량·authoritative reconciliation이 있어야 flat이다.
운영 rollback은 신규 entry 차단과 기존 protective exit 지속을 함께 보장한다. 새 policy snapshot으로 기존 row를 덮어쓰지 않는다. 필요한 migration은 additive-nullable, 역호환 reader부터 전개한다. 이번 작성 요청은 배포·재시작·live주문·토글 변경을 승인하지 않는다. 실제 live 실증과 승인 task가 남아 있으면 archive하지 않는다.

## Risks / Trade-offs

- 과적합·표본 부족 → 수익성 주장과 기능 정확성 판정을 분리하고 unknown/inconclusive를 보존한다.
- 기존 계약 회귀 → frozen fixtures, OFF 동작, 기존 포지션 snapshot을 보존한다.
- 구현 표면 확대 → Existing activation manifests, official gateway, Guardian, campaign/reservations, fill reconciliation and protective exit; scope refreshed before implementation. 범위부터 조사하고 기존 기능을 재사용한다. 별도 strategy framework를 만들지 않는다.
- 위험도: High; NOT READY for live activation or weakening protection. high-risk 변경은 FLM/BTM·Pre-Edit Gate와 적대적 리뷰가 선행된다.

## Migration Plan

설계→선행 change 검증→Terra 구현/시험→독립·gstack 리뷰→Manager 검증→전체 gate→범위 한정 커밋/푸시/정식 archive. DB 변경이 없으면 migration은 not-applicable로 기록한다. 변경이 필요하면 기존 reader 호환성과 additive-nullable rollback을 먼저 리뷰한다. 운영 반영은 별도 승인 전 실행하지 않는다.

## Open Questions

선행 산출물의 실제 가용성, 기존 구현 재사용 범위와 검증 자료의 충분성은 구현 착수 시 확인한다. 자료가 없으면 승인·측정을 추측하지 않고 해당 기능을 unsupported/not_ready로 남긴다. 후속 정책 성능이나 live 승인을 이 설계로 대체하지 않는다.
