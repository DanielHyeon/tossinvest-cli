## Context

우선순위 5/9: Volume Profile 데이터 적합성 평가. 선행 change: a149.
이 문서는 목표 동작을 명세한다. 소스 구현·배포·실거래 검증 완료를 뜻하지 않는다.
공통 기준 증거: `../archive/2026-10-18-a148-clarify-parker-risk-reward-contract/analysis/code-context/`; 구현 전에는 현 HEAD의 호출 사슬·설정 binding과 수정 함수 AST를 새로 확인한다.

## Goals / Non-Goals

**Goals:** LVN/HVN을 계산하기 전에 데이터가 가격대별 실제 체결량을 재현할 수 있는지 확인해야 한다. 5분 OHLCV 분배 근사로 진짜 Volume Profile이라고 주장하지 않는다.

**Non-Goals:** 기존 frozen v1, r1, a128/a147 증거, 인증·토큰 공유, 원장과 위험 통제는 보존한다. 이 문서는 운영 설정 변경·실주문·재시작·추가 세션 수집 승인이 아니다. Manager(Astra)가 설계를 작성·검토하고, Terra Teammate가 코드·테스트를 담당한다. 구현 writer와 독립 reviewer는 분리한다. gstack 리뷰, 해당 change 전체 gate, Manager 최종 확인 후에만 Terra가 범위 한정 커밋·푸시·정식 archive/PM sync를 수행한다.

## Decisions

기존에 제공된 파일만 조사하며 신규 API 수집·구독·운영 예약은 하지 않는다. 분류는 exact_trades, documented_price_bins, ohlcv_proxy, unavailable 네 가지다. trade 자료는 가격/거래량/체결시각/수신시각/정정·취소 식별자, 세션 coverage를 요구한다. aggregate bin 자료는 bin 폭·경계·작성 방법·갱신시각을 요구한다. 호가 잔량은 체결량으로 대체하지 않는다.
산출물은 coverage·중복/정정 처리·틱/bin 단위·as-of 가능 여부·비용/접근 제약·지원 판정표다. 지원 가능한 경우 전일 고정 profile과 당일 prefix profile의 입력 계약을 각각 설계한다. LVN cutoff/bins의 최적화나 생산용 계산 엔진은 이 change에서 구현하지 않는다.
데이터 부재도 적합성 평가의 정상 완료 결과다. unsupported는 기능 승인이 아니며 a153에서 VP arm 제외 이유로 이어진다. 자료 확보와 계산 프로파일 구현이 필요하면 별도 범위를 승인받는다.

## Risks / Trade-offs

- 과적합·표본 부족 → 수익성 주장과 기능 정확성 판정을 분리하고 unknown/inconclusive를 보존한다.
- 기존 계약 회귀 → frozen fixtures, OFF 동작, 기존 포지션 snapshot을 보존한다.
- 구현 표면 확대 → Offline input schema/quality report; immutable provided trade/price-bin data; existing LVN feature input boundary. 범위부터 조사하고 기존 기능을 재사용한다. 별도 strategy framework를 만들지 않는다.
- 위험도: Read-only offline data feasibility. high-risk 변경은 FLM/BTM·Pre-Edit Gate와 적대적 리뷰가 선행된다.

## Migration Plan

설계→선행 change 검증→Terra 구현/시험→독립·gstack 리뷰→Manager 검증→전체 gate→범위 한정 커밋/푸시/정식 archive. DB 변경이 없으면 migration은 not-applicable로 기록한다. 변경이 필요하면 기존 reader 호환성과 additive-nullable rollback을 먼저 리뷰한다. 운영 반영은 별도 승인 전 실행하지 않는다.

## Open Questions

선행 산출물의 실제 가용성, 기존 구현 재사용 범위와 검증 자료의 충분성은 구현 착수 시 확인한다. 자료가 없으면 승인·측정을 추측하지 않고 해당 기능을 unsupported/not_ready로 남긴다. 후속 정책 성능이나 live 승인을 이 설계로 대체하지 않는다.
