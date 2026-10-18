## Context

우선순위 6/9: 비용 반영 통합 비교 실험. 선행 change: a151, a152.
이 문서는 목표 동작을 명세한다. 소스 구현·배포·실거래 검증 완료를 뜻하지 않는다.
공통 기준 증거: `../archive/2026-10-18-a148-clarify-parker-risk-reward-contract/analysis/code-context/`; 구현 전에는 현 HEAD의 호출 사슬·설정 binding과 수정 함수 AST를 새로 확인한다.

## Goals / Non-Goals

**Goals:** 개별 계산 정확성이 수익성 증명은 아니다. 비용과 의존 표본을 반영한 비교로 채택·보류를 결정한다.

**Non-Goals:** 기존 frozen v1, r1, a128/a147 증거, 인증·토큰 공유, 원장과 위험 통제는 보존한다. 이 문서는 운영 설정 변경·실주문·재시작·추가 세션 수집 승인이 아니다. Manager(Astra)가 설계를 작성·검토하고, Terra Teammate가 코드·테스트를 담당한다. 구현 writer와 독립 reviewer는 분리한다. gstack 리뷰, 해당 change 전체 gate, Manager 최종 확인 후에만 Terra가 범위 한정 커밋·푸시·정식 archive/PM sync를 수행한다.

## Decisions

실험 manifest에 data hash, date-based chronological train/validation/untouched holdout, candidate universe, exclusions, parameters, fees/taxes/spread/slippage/FX versions를 결과 보기 전에 고정한다. 테스트 결과를 본 뒤 다시 튜닝한 경우 그 holdout은 consumed로 표시하고 새로운 미사용 자료 없이는 confirmatory 판정을 내리지 않는다.
train/validation에서 primary variant 하나를 선택하여 holdout 공개 전에 고정한다. 여러 variant의 holdout 중 최고를 고르는 행위는 금지한다. 다른 ablation 결과는 exploratory이며 primary 승급 판정을 대신하지 않는다.
baseline→feature correctness→stop-only→EMA strategic exit 순서의 ablation을 유지한다. VP는 a152 supported이고 사전 고정 계산 계약·구현·검증이 별도로 완료된 경우만 추가한다. a147 MFI와 r1은 독립 실험이며 이 change가 r1 질문·선별을 바꾸지 않는다.
net return, realized R, MDD, MAE/MFE, trade/session counts, turnover/exposure, refusal and missingness, cost sensitivity를 시장·국면별로 보고한다. 동일 세션/동일 종목 관측은 독립 표본으로 세지 않는다. 시장별로 동일 후보 집합에서 initial filled account risk로 정규화한 net realized R을 계산하고, 세션 안에서는 후보 평균, 세션 간에는 동일 가중 평균의 variant-minus-baseline 차이를 사용한다. zero risk/invalid fill은 paired exclusion으로 양 arm에서 제외한다. KR/US를 혼합하지 않는다. 세션을 replacement resampling하는 paired percentile bootstrap 95% CI(2.5/97.5 percentile, seed=153, 2000회)를 사용한다. overlapping campaign이 여러 세션을 잇는 경우 독립 session 가정이 깨지므로 해당 실험은 unsupported로 분리한다. MDD는 동일 초기 자본·날짜순 체결·평가시점·환율 기준의 mark-to-market equity series로 구하며, 평가 가격 누락 시 promotion을 차단한다. 20개 미만 independent holdout sessions는 자동 inconclusive; 20개 이상도 충분성 보장이 아니다.
승급 추천의 필요조건: 누락 provenance 없음, paired net expectancy 개선 CI lower>0, holdout MDD 비악화, 2배 slip stress에서도 net expectancy>0. 이 조건은 추천 가능성만 열며 live 승인이 아니다. 표본/비용 원천 부족은 결과를 만들지 않고 inconclusive 사유를 남긴다.

## Risks / Trade-offs

- 과적합·표본 부족 → 수익성 주장과 기능 정확성 판정을 분리하고 unknown/inconclusive를 보존한다.
- 기존 계약 회귀 → frozen fixtures, OFF 동작, 기존 포지션 snapshot을 보존한다.
- 구현 표면 확대 → Reuse offline evaluation artifacts from a151; immutable datasets and reports; no additional collection. 범위부터 조사하고 기존 기능을 재사용한다. 별도 strategy framework를 만들지 않는다.
- 위험도: Offline evaluation; empirical claims require independent holdout evidence. high-risk 변경은 FLM/BTM·Pre-Edit Gate와 적대적 리뷰가 선행된다.

## Migration Plan

설계→선행 change 검증→Terra 구현/시험→독립·gstack 리뷰→Manager 검증→전체 gate→범위 한정 커밋/푸시/정식 archive. DB 변경이 없으면 migration은 not-applicable로 기록한다. 변경이 필요하면 기존 reader 호환성과 additive-nullable rollback을 먼저 리뷰한다. 운영 반영은 별도 승인 전 실행하지 않는다.

## Open Questions

선행 산출물의 실제 가용성, 기존 구현 재사용 범위와 검증 자료의 충분성은 구현 착수 시 확인한다. 자료가 없으면 승인·측정을 추측하지 않고 해당 기능을 unsupported/not_ready로 남긴다. 후속 정책 성능이나 live 승인을 이 설계로 대체하지 않는다.
