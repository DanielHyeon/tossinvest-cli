## Context

우선순위 2/9: 세션 VWAP와 확정봉 EMA9 계산. 선행 change: a148.
이 문서는 목표 동작을 명세한다. 소스 구현·배포·실거래 검증 완료를 뜻하지 않는다.
공통 기준 증거: `../a148-clarify-parker-risk-reward-contract/analysis/code-context/`; 구현 전에는 현 HEAD의 호출 사슬·설정 binding과 수정 함수 AST를 새로 확인한다.

## Goals / Non-Goals

**Goals:** 세션 누적 VWAP와 기간·시드가 명확한 EMA9를 계산해야 손절·진입 실험을 같은 입력으로 비교할 수 있다. rolling window 값을 세션 지표로 대체 사용하지 않는다.

**Non-Goals:** 기존 frozen v1, r1, a128/a147 증거, 인증·토큰 공유, 원장과 위험 통제는 보존한다. 이 문서는 운영 설정 변경·실주문·재시작·추가 세션 수집 승인이 아니다. Manager(Astra)가 설계를 작성·검토하고, Terra Teammate가 코드·테스트를 담당한다. 구현 writer와 독립 reviewer는 분리한다. gstack 리뷰, 해당 change 전체 gate, Manager 최종 확인 후에만 Terra가 범위 한정 커밋·푸시·정식 archive/PM sync를 수행한다.

## Decisions

정규장 세션 identity=(market,symbol,exchange trading date,calendar version)로 구분한다. v1 범위는 KR 정규장으로 제한하며 US는 검증된 거래소 캘린더·DST·단축장 증거가 없으면 unsupported로 반환한다. 1분봉 close label t는 [t-1m,t), 5분봉 close label t는 [t-5m,t)다. 계산시점 이하에서 닫히고 수신된 봉만 쓴다.
VWAP v1은 공식 unadjusted 1분봉 TP=(H+L+C)/3, sum(TP*volume)/sum(volume)의 세션 누적 근사값이다. 정확한 체결 VWAP로 이름 붙이지 않는다(method=session_typical_price_1m_v1). 개장부터 누락 없는 prefix가 필수이며 중간 접속은 prefix를 보충하기 전 not_ready다. 중복·역순·불일치 조정가격·거래일 혼합은 거절한다. 0거래량 봉은 누적 금액/분모를 늘리지 않고 전체 분모0은 not_ready다. negative volume, invalid OHLC, overflow는 거절한다.
EMA9는 정규장 확정 5분봉 종가 9개 SMA를 seed로, alpha=2/(9+1)=0.2로 다음 봉부터 계산한다. 세션마다 초기화하고 9개 미만 not_ready다. 따라서 이 v1 연구 계산기는 첫 45분 동안 EMA 신호를 내지 않는다. 이는 frozen Parker v1 입력을 대체하지 않는다. 연속 세션 EMA는 혼용하지 않고 별도 실험 버전 사항으로 남긴다.
VWAP slope는 인접한 두 확정 5분 close 시점 VWAP의 (current/previous-1)*100으로 고정한다. as_of, received_at, source, calendar/method version, bar range/count, input digest와 readiness 이유를 반환한다. 내부 계산은 checked rational/decimal, 표시만 반올림한다.

## Risks / Trade-offs

- 과적합·표본 부족 → 수익성 주장과 기능 정확성 판정을 분리하고 unknown/inconclusive를 보존한다.
- 기존 계약 회귀 → frozen fixtures, OFF 동작, 기존 포지션 snapshot을 보존한다.
- 구현 표면 확대 → internal/strategycandle, internal/strategymarket, internal/strategyengine/contracts.go; new pure calculation leaf, no production caller wiring. 범위부터 조사하고 기존 기능을 재사용한다. 별도 strategy framework를 만들지 않는다.
- 위험도: High before wiring into production; pure calculations first. high-risk 변경은 FLM/BTM·Pre-Edit Gate와 적대적 리뷰가 선행된다.

## Migration Plan

설계→선행 change 검증→Terra 구현/시험→독립·gstack 리뷰→Manager 검증→전체 gate→범위 한정 커밋/푸시/정식 archive. DB 변경이 없으면 migration은 not-applicable로 기록한다. 변경이 필요하면 기존 reader 호환성과 additive-nullable rollback을 먼저 리뷰한다. 운영 반영은 별도 승인 전 실행하지 않는다.

## Open Questions

선행 산출물의 실제 가용성, 기존 구현 재사용 범위와 검증 자료의 충분성은 구현 착수 시 확인한다. 자료가 없으면 승인·측정을 추측하지 않고 해당 기능을 unsupported/not_ready로 남긴다. 후속 정책 성능이나 live 승인을 이 설계로 대체하지 않는다.
