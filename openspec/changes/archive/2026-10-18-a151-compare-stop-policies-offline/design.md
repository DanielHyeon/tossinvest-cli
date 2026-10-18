## Context

우선순위 4/9. 선행 a150은 main `cf9260f3f5e89a1d8382f2d0437a70a088d36704`에서 검증·아카이브됐다. 2026-10-18 사용자 결정: 권장 최소 비교 범위에 US 시장도 포함한다. 이 계약은 오프라인 계산 도구의 명세이며 실거래 활성화나 실제 수익성 검증을 뜻하지 않는다.

## Goals / Non-Goals

KR·US 정규장 long-only에서 동일 후보의 세 초기 손절을 비교한다: fixed 0.7%, 확정 5분 신호봉 저점 아래 한 tick, entry minus 1.5 × ATR14. 진입·목표·비용·사전 예산은 동일하다. 합성 fixture로 계산 정확성을 검증한다.

기존 frozen v1, r1, a128/a147 증거, 인증·토큰 공유, 원장·위험 통제, 운영 설정은 변경하지 않는다. API 호출·네트워크·실주문·재시작·추가 세션 수집·런타임 연결·DB migration은 없다. a149 KR 기능을 US용으로 확장하지 않는다. 별도 전략 프레임워크, UI, 최적 파라미터 탐색, 실제 계좌 시뮬레이터, 포트폴리오 복리·동시보유·환전 모델은 만들지 않는다. EMA9 전략 청산은 후속 별도 실험으로 남기고 이 도구는 요청 시 unsupported로 거절한다.

## SDD skill decisions

Manager가 grill-me 및 grilling을 읽고 독립·gstack 사전 검토를 수행했다. 사용자가 US 포함을 선택했다. 나머지 최소 범위는 세 초기 손절 비교 도구·합성 정확성 검증, 당일 ATR warmup, 세 arm 공통 실현 코호트, 성과 inconclusive 및 EMA 후속이다. 아래 시간·단위·반올림은 이 범위를 구현 가능한 계약으로 구체화한 결정이다. 이미 승인된 범위를 반복 승인받지 않는다.

impeccable: N/A — CLI 오프라인 연구 도구이며 웹·앱 UI 변경이 없다. 독립 리뷰, gstack 리뷰, 전체 gate와 Manager 최종 검증은 그대로 필수다. Manager는 설계, Terra는 단일 구현·시험 writer다.

## Input and trust boundary

단일 버전 JSON manifest와 로컬 immutable OHLCV 자료만 받는다. 입력 bytes SHA-256, manifest/도구 버전과 모델 파라미터를 결과에 기록한다. 해시는 자료의 진실성이나 거래소 인증을 입증하지 않는다. provenance는 caller-supplied/unverified이며 synthetic 여부를 명시한다. 출력은 stdout으로 제공하거나 새 출력 파일만 생성하고 기존 파일·입력·symlink 대상을 덮어쓰지 않는다.

입력 최대 16 MiB, 최대 후보 1000개, 후보당 1분봉 최대 390개·5분봉 최대 78개. 입력 숫자는 최대 64자 unsigned plain decimal, 과학표기·NaN·Inf·분수·진법 표기를 거절한다. 유리수 연산을 사용한다. JSON 알 수 없는 필드·중복 키·중복 후보 ID는 거절한다. 구조/한도 오류는 비정상 종료, 후보별 결측·지원불가·실행불가 사유는 결과에 남긴다.

각 후보는 ID, KR/KRW 또는 US/USD, 종목, 거래일, 가격 기준(unadjusted), 공급자·자료 버전, 신호 확정 시각과 신호 5분봉, 1분/5분 OHLCV, 세션·tick·비용·risk/cap snapshot을 포함한다. OHLC 양수와 low≤open/close≤high, 거래량≥0을 확인한다. 봉 시간은 close_at, received_at으로 구분한다. 신호 정의는 후보로 주어지며 미래 데이터를 보고 후보를 생성하지 않는다.

## Session and information timing

KR zone=Asia/Seoul, US zone=America/New_York. 날짜·IANA zone·offset을 명시한 open/close, 정규/조기폐장/휴장 구분, calendar source/version/digest를 요구한다. 정규 세션은 현지 KR 09:00–15:30, US 09:30–16:00와 일치해야 한다. early-close는 같은 개장시각, 더 이른 폐장시각, 5분 배수 길이를 명시한다. 휴장 표시는 거래 후보를 허용하지 않는다. DST offset은 IANA와 교차 확인한다. 공급자 달력의 실제 휴일 정확성은 검증했다고 주장하지 않는다. 범위 밖/다른 날짜/잘못된 offset/불완전한 세션은 거절한다.

신호는 당일 연속한 확정 5분봉이다. 신호 계산에 쓰는 봉은 close_at 및 received_at ≤ signal_confirmed_at이어야 한다. 동일 interval의 중복·역순은 invalid. 신호 계산용 5분봉은 같은 1분봉 5개의 OHLCV 집계와 일치해야 한다. 사후 제공 5분봉도 구성 1분봉이 완전하면 일치를 검사한다. 사후 구성 1분봉이 빠졌다면 해당 집계는 미검증 사유를 기록하고 진입 사전 제외로 바꾸지 않는다. 실제 1분 실행 경로가 그 결측에 도달할 때 censored로 처리한다. 선행 결측은 not_ready로 세 arm 모두 제외한다. 봉은 세션 경계에 정렬한다.

진입은 open_at이 signal_confirmed_at보다 엄격히 뒤인 첫 예정 1분봉 open이다. 이는 신호 확정 시 미리 예약된 가상 시장가 진입 모델이며 실제 broker 체결 증거가 아니다. 여기서 시점 제한은 신호 입력의 received_at에 적용한다. 진입·후속 1분봉 전체 OHLCV의 received_at은 봉 close 이후여야 하지만 성과 관측용이므로 해당 open 이후 수신되어도 허용한다. 미래 봉 전체가 수신될 때까지 기다린 뒤 과거 open으로 실제 주문했다는 주장을 하지 않는다. 다음 예정 봉이 없으면 나중 봉을 건너뛰어 선택하지 않고 unfilled로 분류한다. 진입 이후 경로에 결측이 있으면 그 시점에서 censored이며 이후 가격으로 청산을 꾸미지 않는다. 신호 확정 시에는 구조 저점·ATR·calendar/tick/cost/budget/caps snapshot을 고정하며 각 snapshot의 as_of는 confirmation 이하여야 한다. 미리 예약된 진입 open만 실제 모의 entry와 fixed/ATR 절대 stop·공통 target·수량 산정에 사용한다. 진입봉의 high/low/close와 이후 봉은 이 계산에 사용하지 않고 성과 관측용으로만 사용한다. 이미 산정한 stop/수량은 이후 수정하지 않는다.

## Stops, ticks and quantity

ATR14는 당일 연속 15개 확정 5분봉이 필요하다. 첫 봉 close를 anchor로, 다음 14개 TR=max(high−low, abs(high−previous_close), abs(low−previous_close))의 산술평균으로 seed한다. 이후 확정봉마다 `(previousATR×13+TR)/14`. 전일 warmup은 이 버전에서 금지한다.

v1 tick manifest는 후보에 적용되는 양수 tick 하나와 [min_price,max_price] 유효 범위, 시장·종목·source/version/digest·유효 시작/종료를 명시하고 전체 세션을 포함하는지(start≤open, end≥close) 확인한다. 범위 내 가격은 0 기준 tick 배수다. 가격 구간별 tick 변경은 unsupported로 거절하며, 모든 관측 가격·계산 stop/target/fill이 이 범위를 벗어나면 해당 후보를 unsupported로 남긴다. 이는 caller-supplied bounded grid이며 최신 KR/US 호가 규칙을 내장·인증했다는 주장이 아니다. US를 무조건 0.01로 가정하지 않는다.

entry = ceil_to_tick(next_open + entry_slippage_ticks×tick). fixed stop=floor_to_tick(entry×0.993), structural stop=floor_to_tick(signal_low)−tick, ATR stop=floor_to_tick(entry−1.5×ATR). target=ceil_to_tick(entry×1.021)로 세 arm에 같은 절대 가격을 쓴다. target 반올림은 최소 2.1% 목표를 보존한다. 모든 stop은 0<stop<entry<target이어야 하며 하나라도 불가능하면 세 arm 모두 사전 제외한다.

후보별 동일한 사전 위험 예산·최대 정수 수량·최대 명목금액을 각 arm에 적용한다. KR 입력은 KRW 예산, US 입력은 USD 예산만 허용한다. account_currency가 다르거나 FX가 필요한 입력은 unsupported다. 시장 간 금액 합산이나 계좌 수익률은 제공하지 않는다. risk span=entry−rounded_stop (tick 정렬), quantity=min(floor(budget/span), max_quantity, floor(max_notional/entry)). 기존 순수 수량 계산을 안전하게 재사용할 수 있으면 재사용하되 production 계약은 변경하지 않는다. caps 0을 unlimited로 보지 않는다. 한 arm이라도 수량 0이면 사전 공통 제외한다. 비용·gap은 초기 가격 위험보다 실제 모의손실을 크게 만들 수 있으며 budget 준수를 보장했다고 주장하지 않는다.

## Fill and cost model

세 arm 진입은 동일 open 및 slippage. 보유량은 정수, partial/trailing은 없다. 진입봉부터 stop/target을 판정하며 진입 전 봉의 가격을 사용하지 않는다. 진입봉을 포함한 각 봉에서 open≤stop이면 floor_to_tick(open−exit_slippage_ticks×tick), 그렇지 않고 open≥target이면 target에 청산한다. 나머지 경우 low≤stop/high≥target touch로 판정한다. stop touch fill=floor_to_tick(stop−exit_slippage_ticks×tick); target fill=target (유리한 gap 이익 없음). 계산 fill≤0 또는 유효 tick 범위 밖 fill은 해당 arm의 사후 outcome invalid이며 0으로 보정하지 않는다. 이미 성립한 eligible 상태와 다른 arm의 결과·수량·가격은 보존하고 공통 complete 비교에서만 제외한다. 기본 결과와 sensitivity의 invalid도 서로 덮어쓰지 않는다.

비용 manifest는 quote 통화, buy/sell fee rate 및 sell tax rate(각 0..1), side별 fixed fee≥0, 비용 rounding quantum>0, source/version/digest 및 신호 이전 snapshot 시각을 명시한다. 입력 0은 명시된 모델값이며 실제 비용없음의 증거가 아니다. 각 체결의 fee=ceil_quantum(notional×side_rate+fixed_fee), sell tax=ceil_quantum(exit_notional×tax_rate). net simulated PnL=(exit−entry)×quantity−buy_fee−sell_fee−sell_tax. simulated R=net PnL/(quantity×initial_span). a148 실제 확정 체결용 realized-R 증거 타입으로 가장하지 않는다. slippage ticks는 비음수 정수로 고정한다.

## Ambiguity, censoring and comparison

open 순서로 결과가 결정되지 않고 동일봉에 stop·target 모두 touch하면 stop-first 기본 결과를 쓰고 ambiguous로 집계한다. 별도 target-first 재생 결과를 arm별 민감도 끝점으로 제공한다. 이 두 가격 경로의 범위는 각 arm별일 뿐, 서로 다른 arm의 유리한 끝점을 하나의 공유 시장 경로 또는 paired 차이의 실현가능 경계로 주장하지 않는다. 해당 경로에 의존한 우열은 inconclusive다.

사전 eligible cohort와 사후 complete cohort를 분리한다. 세 arm 모두 실제 모의 청산된 동일 후보 ID만 주 paired simulated-R 차이에 포함한다. 한 arm이라도 unfilled/censored/invalid면 세 arm 전체를 주 비교에서 제외한다. 각 시장별 전체 후보·사전 제외 사유·eligible 수·arm별 청산/미체결/censored/ambiguity·complete 수를 표시한다. 세션 끝 미청산은 마지막 연속 관측 close의 평가손익과 시각을 별도 표시하며 실현 체결로 만들지 않는다. 분모가 0이면 평균은 null/not_ready이며 0으로 바꾸지 않는다.

기술통계는 시장별·동일 후보별 차이이며 독립 거래의 합성/관측 비교다. 수익성·통계적 우월성 판정은 항상 inconclusive: 실측 표본·사전 가설·통계 계획을 이 change가 확보하지 않는다. 표본을 늘렸다는 이유만으로 자동 채택하지 않는다.

## Verification and lifecycle

KR·US 양시장 골든 산술, USD 소수 경계, US EST/EDT·조기폐장·휴장·offset 오류, 잘못된 통화·tick/범위·비용, ATR warmup/수신 지연, zero caps, entry gap, stop gap, 동일봉 모호성, 미래 누수, 결측/censoring/분모를 시험한다. 원본 hash 불변·네트워크/운영 호출 없음·출력 덮어쓰기 거절도 검증한다. 소스와 입력을 실제 CLI로 실행해 산출물을 확인한다. 기존 함수 편집 시 해당 AST/FLM/BTM 절차를 따른다.

Terra 구현·대상 시험 → 독립 리뷰 → gstack 리뷰 → 동결 및 sdd-sync/check·전체 gate → Manager 최종 검증 → Terra 정식 archive/spec sync·PM/strict·범위 commit/push. 기존 증거 및 실제 운영은 유지한다. EMA9 비교는 후속 별도 manifest/change 없이는 지원하지 않는다.
