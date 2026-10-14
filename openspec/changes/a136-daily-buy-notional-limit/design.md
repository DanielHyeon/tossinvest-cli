# a136 design — 하루 매수 총액 한도

근거 표기: 「AST:」는 `analysis/ast-evidence/<번들>/` 의 `tools/logic-map` 산출물(base 7b073f55, 편집 전)이다. 그 밖의 줄 번호는
같은 base 의 HEAD 를 읽은 것이다.

## D0. 저장소가 이미 가진 것

| 사실 | 근거 |
|---|---|
| 계좌 총계 한도의 최종 권위는 예약 트랜잭션(BEGIN IMMEDIATE, 단일 커넥션) | risk-management 「발급 절차와 예약의 권위」, reservations.go 머리 주석 |
| 총계 경계는 `usage + held + new ≥ limit` → 거부 | AST `reserveplan.checklimits` B7(704) |
| 진입 발급 문 셋 모두 OPEN_EXPOSURE 한 종류만 예약 | AST `riskguardian.issueentry` collectIssue(481-505) · riskguardian_qfinal.go:244-246 · riskguardian_first_leg.go:194-196 |
| 예약 금액 = `risk.EntryExposureValue`(지정가×수량+과대 비용, account base, US 는 frozen FX · ceil) | chain.go `EntryExposureValue` · riskguardian_qfinal.go:233 |
| 체결 · 취소 모두 BROKER_TERMINAL 로 보유 해제, 행은 지우지 않음(RELEASED) | reservation_release.go 머리 주석 · `TestDerivedTerminalFillReleasesTheHold` |
| 예약 종류는 닫힌 집합 + DDL CHECK, `trading_day` 칸은 DAILY_LOSS 전용 | AST `reserverequest.build` B10 · B12 · B13, execution_contract.go:169 · 172 |
| 일일 경계는 **시장별 거래일**뿐 | `clock.Market.TradingDay`(market.go:158), AST `sweeptradingdayboundary` B9, riskcalc `DailyLossInputs.TradingDay` |
| 생산 코드에 DAILY_LOSS 예약을 만드는 곳은 없음 | `ReservationKindDailyLoss` 사용처 = journal 패키지 안뿐(grep, 비시험) |
| 매도는 진입 체인과 진입 발급 문에 닿지 않음 | AST `risk--evaluate` B2(135) · `riskguardian.issueentry` B3(416) |
| 한도 변경 seam: 콘솔 편집기 → 키 단위 쓰기 → 필드별 audit | `cmd/tossctl/limitsettings.go` · `config/limits_io.go` 머리 주석 |

## D1. 「하루」의 경계 — **사용자 결정**

요구: 「1일 구매 한도」 = 두 시장을 합친 하나의 총액. 그러나 D0 의 유일한 관례는 **시장별** 거래일이다. 한 UTC 순간이 서울과 뉴욕에서
다른 날이므로, 시장별 관례로는 「계좌의 오늘」 하나가 정의되지 않는다. KST 기준 정규장: KR 09:00–15:30, US 22:30–05:00(서머타임) ·
23:30–06:00(표준시). 두 시장이 다 닫힌 창은 매일 06:00–09:00 KST.

| 안 | 정의 | 새 관례? | 장점 | 단점 |
|---|---|---|---|---|
| D1-A | KST 달력 날짜(`clock.MarketKR.TradingDay(now)`)를 두 시장 모두에 | 함수는 재사용, 「US 매수도 서울 날짜」는 새 해석 | 가장 단순 · 글자 그대로 「1일」 | **자정이 US 정규장 한가운데** — 한 밤 US 세션이 자정 전후로 한도의 2배까지 살 수 있음 |
| D1-B | 매수마다 자기 시장의 거래일, 합계 = 「각 매수의 시장에서 본 오늘」에 속한 것의 합(DAILY_LOSS 해제 규칙 그대로) | 아니오(관례 그대로) | 기존 sweep 의미 승계 | 「오늘」이 두 창의 합집합이라 예산이 시장별로 다른 시각에 풀림 — KR 자정에 KR 분이 빠져 US 세션 후반에 다시 살 수 있고, 다음 날 KR 아침에는 아직 US 전날 분이 남아 KR 이 막힘. 운영자가 예측 못 함 |
| D1-C | 시장별 별도 한도 | 아니오 | 관례 그대로 | 요구(시장 합산 하나)와 다름 |
| **D1-D (추천)** | **KST 08:00 에 넘어가는 하루**: 일 라벨 = `clock.MarketKR.TradingDay(now − 8h)` | 예 — 상수 하나(8h) | KR 정규장과 **그날 밤** US 정규장이 한 바구니, 어느 정규장도 쪼개지지 않음. 경계가 두 시장 모두 닫힌 창 안 | 새 상수. US 장후 시간외(KST 05:00/06:00–09:00)를 TossOS 가 언젠가 거래하면 그 창은 08:00 에서 갈린다 `[미측정 — 현재 US 진입 시간창은 spec 상 미이식]` |

추천 이유: 한도의 목적은 「하루에 살 수 있는 돈」을 묶는 것이고, D1-A 는 가장 위험한 지점(한 세션 안)에서 그 묶음을 푼다. D1-D 는 같은
함수를 재사용하고 상수 하나만 더한다(KISS). 경계 시각은 06:00–09:00 KST 안이면 의미가 같으며 08:00 은 KR 장전 시간외(08:30~)보다 앞이다.

## D2. 「체결된 매수 + 미체결 · 예약된 매수」를 어디서 세나 — **사용자 결정(추천 있음)**

| 안 | 방법 | 스키마 | 비고 |
|---|---|---|---|
| D2-a | 새 예약 종류 `DAILY_BUY_NOTIONAL`(거래일 칸), 체결돼도 해제 안 하고 경계에서 소멸 | **v36** — DDL CHECK 재작성(테이블 재생성) + 백업 복원 롤백 | 해제 규칙(체결 경로 fills.go)까지 편집 — High-risk 표면 최대 |
| D2-b | 트랜잭션 안에서 체결 기록(fill_snapshots 등) × 평균가 합 + HELD OPEN_EXPOSURE | 없음 | US 체결의 base 환산에 그 순간의 FX 가 필요 — KR 요청의 identity FX 로는 못 함 → **새 FX 경로**가 생김(요구 위반) |
| **D2-c (추천)** | 트랜잭션 안에서 **예약 원장 자체**를 센다: 오늘(D1 라벨) 발급된 EXPOSURE_RAISING 결정의 OPEN_EXPOSURE 예약 금액 합 − 「체결 없음이 증명된」 것 | 없음 | 금액은 발급 때 이미 base · frozen FX · ceil · 비용 포함 → 새 FX 0. 행은 지워지지 않으므로 체결 뒤에도 남는다 |

D2-c 의 「체결 없음이 증명됨」(합계에서 빼는 것): 결정이 미전송으로 끝남(NOT_DISPATCHED 계열) · `EXPIRED_UNCONSUMED` 해제 · 체결 수량 0 으로
BROKER_TERMINAL 해제. 그 밖은 전부 센다(부분 체결 · 상태 불명 · in-doubt 포함 — 보수 방향). 부분 체결 뒤 취소를 체결분 비율로 줄일지는
attempt → 주문 → 체결 수량 결합이 믿을 만한지 **구현 전 실측** 후 정한다(기본: 줄이지 않음 = 과대 집계, 안전 방향).
HELD 인 예약은 발급 날짜와 무관하게 센다(어제 걸린 미체결 매수가 아직 살아 있으면 그것도 「미체결 매수」).

## D3. 판정 위치 — 체인 rung 이 아니라 원자 트랜잭션

체인은 발급 시점에 한 번, 호출자가 모은 값으로 돈다(AST `riskguardian.issueentry` B4). 그 값으로 「오늘 합계」를 넘기면 동시 발급 둘이
같은 숫자를 보고 함께 통과한다. 예약 트랜잭션은 단일 커넥션 BEGIN IMMEDIATE 라 직렬이다. 그래서 판정은 journal 의 결정 · 예약 원자
트랜잭션 안(`reserveRows` 의 총계 검사 옆)에 둔다. 체인 12단계 · reason-code 는 바뀌지 않는다.

거부 사유: 도달은 기존 `LIMIT_REACHED`(detail 에 `DAILY_BUY_NOTIONAL` · 합계 · 한도), 계산 불가는 분류되지 않은 오류로 발급 거부
(결정 · 예약 롤백 — 기존 「DB 오류는 위험 거부로 꾸미지 않는다」 규칙, riskguardian.go `issuanceRefusal`).

## D4. 경계 — 도달하면 거부(≥)

사용자 표현은 「넘으면」이지만 저장소의 총계 관례는 도달 시 차단(≥, spec 「총계 한도의 경계값은 도달 시 차단」)이다. 같은 트랜잭션 안에서
두 경계가 다르면 운영자가 읽는 규칙이 둘이 된다. ≥ 는 정확히 한도와 같은 합계 하나만 더 막는다(보수 방향). **사용자 확인 요청**.

## D5. 설정 의미

- 키: `engine.automation_gate.max_daily_buy_notional`(float64, `omitempty`), 통화는 `limit_currency`.
- 없음 · 0 → 기본 10,000,000 KRW(기존 관례: 0 = 미설정). 끄는 스위치는 만들지 않는다(요구에 없음). 음수 · NaN · Inf → 기동 거부.
- base 가 USD 이고 값이 없을 때: 10,000,000 을 USD 로 읽으면 큰 완화이고, KRW→USD 환율을 코드에 넣는 것은 spec 이 금한다. **추천: 기동 거부**
  (운영자가 USD 값을 명시). 대안: 사용자가 USD 기본값을 명시 승인(provenance 요구상 사람 승인 + 기록 필요).
- 값의 provenance: 사용자 결정 2026-10-14(이 change). 이전에는 상한이 없었으므로 어떤 값도 완화가 아니다(새 검사).

## D6. 단일 출처 — 한도 스냅샷 여섯째 칸

`execgw.Limits` 에 `max_daily_buy_notional` 을 더하고 `ExposureLimitsFor` · 인터록 감사가 같은 값을 진술하게 한다(engine-safety 인터록 단일 출처).
결과: 결정 `limits_json` 과 `LimitsDigest` 가 바뀐다. 기동 때 in-flight 결정은 없고(TTL 60초), 생산 전략 바인딩은 Guardian 에서 유도한다
(`strategy_account_first_leg_authority.go:391`) — 구현 단계에서 서명 매니페스트가 digest 를 묶는지 실측.

## D7. 변경 경로 · 감사 · 콘솔

- `config.GuardianLimits` 에 필드 하나, `SaveEngineGateLimits` 의 키 단위 쓰기 목록에 하나(`enabled` 는 계속 목록 밖), `recordLimitSave` 가 필드별
  이전 → 새 값 audit. CLI `tossctl` 쪽 쓰기 명령은 새로 만들지 않는다(기존 seam 재사용).
- 콘솔 한도 편집기에 숫자 칸 하나. 확인 문구 · 추가 승인 마찰 없음(기억 규칙). 규칙 12(ui-skills 1~3개) · 13(렌더 확인, 버튼은 누르지 않음).
  편집기 작업이 커지면 콘솔 칸을 후속 change 로 분리한다.
- 적용: 다른 한도처럼 엔진 재시작 때. 콘솔의 기존 안내(settings_card.go:67 「저장은 즉시 파일에 기록되지만 가동 중인 엔진은 기동 시점
  설정으로 계속 동작한다 — 반영하려면 재시작한다.」)를 그대로 따른다.

## D8. fail-closed 범위

계산 불가(경계 계산 실패 · 금액 해석 실패 · 통화 불일치) → **그 매수만** 거부. 매도 · 보호 · 청산 · 대사 · 체결 감지는 이 코드를 지나지 않는다(D0).

## D9. 수동 주문

Guardian 밖의 수동 주문(upstream `tossctl order place` 등 `mutating: true`)은 막지 않는다(OFF = upstream). D2-c 는 Guardian 발급분만 센다 —
수동 매수는 합계에 들어가지 않는다. 수동분까지 세려면 브로커 주문 이력을 읽어야 하고 그것은 D2-b 의 FX 문제를 다시 연다. **사용자 확인 요청**.

## 잔재 · 비목표

- 실현 손익 · 일일 손실 생산자(a137 소관). 체인 단계 추가. US 진입 시간창. MAX_POSITIONS · CANCEL_RATE(P3 유지).
