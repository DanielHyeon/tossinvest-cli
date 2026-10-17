## Context

우선순위 4/9: 손절 정책 오프라인 비교. 선행 change: a150.
이 문서는 목표 동작을 명세한다. 소스 구현·배포·실거래 검증 완료를 뜻하지 않는다.
공통 기준 증거: `../a148-clarify-parker-risk-reward-contract/analysis/code-context/`; 구현 전에는 현 HEAD의 호출 사슬·설정 binding과 수정 함수 AST를 새로 확인한다.

## Goals / Non-Goals

**Goals:** 0.7% 고정과 구조·변동성 손절을 비교할 때 진입·익절까지 동시에 바꾸면 원인을 구분할 수 없다. 손절만 바꾸는 사전 고정 비교부터 수행한다.

**Non-Goals:** 기존 frozen v1, r1, a128/a147 증거, 인증·토큰 공유, 원장과 위험 통제는 보존한다. 이 문서는 운영 설정 변경·실주문·재시작·추가 세션 수집 승인이 아니다. Manager(Astra)가 설계를 작성·검토하고, Terra Teammate가 코드·테스트를 담당한다. 구현 writer와 독립 reviewer는 분리한다. gstack 리뷰, 해당 change 전체 gate, Manager 최종 확인 후에만 Terra가 범위 한정 커밋·푸시·정식 archive/PM sync를 수행한다.

## Decisions

동일 후보·진입 시점·비용 모델·청산 정책을 고정하고 세 초기 손절 arm만 비교한다: fixed 0.7%; confirmed signal-bar low minus one tick; entry minus 1.5*ATR14 of closed 5m bars. ATR uses Wilder smoothing seeded by 14 true ranges, first TR needs previous confirmed close; missing warmup excludes the paired observation from all arms. ATR/구조 arm은 실거래에 채택하지 않는다.
기본 체결 모델 v1은 신호 확정 시점 이후 첫 유효 1분봉 open에 불리한 slippage를 더한 long entry다. 신호봉 close 체결은 허용하지 않는다. long stop gap-through는 다음 관측 open이 stop 아래일 때 그 open에서 exit slippage를 차감하고, 그 외 stop touch는 stop에서 exit slippage를 차감한다. 수수료·세금은 별도 차감한다. buy와 per-share risk는 불리한 tick 방향(ceil), protective sell price는 floor로 반올림한 후 수량을 정수 floor한다. 필요한 next bar가 없으면 unfilled/censored로 보고하며 없는 가격을 만들지 않는다.
초기 손절 비교의 익절가는 모든 arm에서 baseline entry*(1+3*0.007)의 동일 절대 가격으로 고정한다. long limit target은 보수적으로 target 가격에서 비용을 차감하며 유리한 gap을 수익에 더하지 않는다. stop 변경으로 목표가격이나 target RR을 재설정하지 않는다. trailing/partial exits는 이 첫 비교에서 끄며 기존 runtime의 익절 설정을 끄는 뜻은 아니다. 이후 common exit 비교는 별도 고정 manifest로 분리한다. 세션 종료까지 미청산이면 종가를 체결가로 꾸미지 않고 잔존 평가손익과 censored count를 별도 보고하며 realized-R 쌍 비교에서는 양 arm을 제외한다.
정규장 long-only baseline으로 제한한다. 모든 arm은 동일 account risk budget으로 수량을 다시 산정하고 기존 caps를 적용한다. fill mismatch/거래비용, stop span, risk usage를 별도 기록한다. 장중 touch protective stop과 EMA9 closed-bar strategic exit를 구분한다. EMA 비교는 초기 손절 비교가 끝난 후 같은 최초 보호 stop을 유지한 추가 paired experiment로만 수행한다.
OHLC 한 봉에서 target/stop 모두 닿으면 순서를 모른다. 기본은 conservative stop-first, ambiguous count와 대체 경계 결과를 같이 보고한다. 다음 시점 executable price 모델 없이 신호봉 종가에 즉시 체결했다고 가정하지 않는다. 입력 hash와 파라미터를 결과 전 고정하며 수익률이 나쁘거나 표본 부족이면 그대로 inconclusive를 낸다.

## Risks / Trade-offs

- 과적합·표본 부족 → 수익성 주장과 기능 정확성 판정을 분리하고 unknown/inconclusive를 보존한다.
- 기존 계약 회귀 → frozen fixtures, OFF 동작, 기존 포지션 snapshot을 보존한다.
- 구현 표면 확대 → Offline research tool consuming immutable fixtures and a149 features/a150 contract; no live strategy imports with side effects. 범위부터 조사하고 기존 기능을 재사용한다. 별도 strategy framework를 만들지 않는다.
- 위험도: Offline research only; no trading code invocation. high-risk 변경은 FLM/BTM·Pre-Edit Gate와 적대적 리뷰가 선행된다.

## Migration Plan

설계→선행 change 검증→Terra 구현/시험→독립·gstack 리뷰→Manager 검증→전체 gate→범위 한정 커밋/푸시/정식 archive. DB 변경이 없으면 migration은 not-applicable로 기록한다. 변경이 필요하면 기존 reader 호환성과 additive-nullable rollback을 먼저 리뷰한다. 운영 반영은 별도 승인 전 실행하지 않는다.

## Open Questions

선행 산출물의 실제 가용성, 기존 구현 재사용 범위와 검증 자료의 충분성은 구현 착수 시 확인한다. 자료가 없으면 승인·측정을 추측하지 않고 해당 기능을 unsupported/not_ready로 남긴다. 후속 정책 성능이나 live 승인을 이 설계로 대체하지 않는다.
