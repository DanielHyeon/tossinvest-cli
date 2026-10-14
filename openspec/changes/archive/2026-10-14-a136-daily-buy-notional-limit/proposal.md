# a136 — 하루 매수 총액에 상한을 둔다 (daily-buy-notional-limit)

> **상태(2026-10-14):** 사용자 결정 수신 — design 「사람 결정」 원문(1차: D1=D · D2=c · D4=≥ · D5=기동 거부 · D9=b, 2차: 「별도 change 로 나눌 필요
> 없음」 — 수동 매수 합산과 콘솔 편집 칸도 a136 안). proposal-freeze 적대 리뷰 APPROVE-WITH-FIXES(review.md) 반영. 핵심(원장 집계 · 설정 ·
> 인터록 · 발급 문 셋) 착지 후 수동 매수 refresher → 콘솔 순서.

## Why

사용자 지시(2026-10-14, 최종): **하루 매수 총액 한도**(매수만), 사용자 설정 가능, 기본 10,000,000 KRW. 세는 금액은
「체결된 매수 + 미체결 · 예약된 매수」(동시 주문이 한도를 넘지 못하게), 오늘 합계에 새 주문을 더해 한도를 넘으면 거부,
오늘 합계를 계산할 수 없으면 매수 거부(fail-closed). 매도 · 손절 · 비상 청산은 이 한도로 막히거나 늦어지지 않는다(§0.3).

지금 HEAD 대조(base 7b073f55):

- `risk-management` 「Guardian 판정 체인」이 `DAILY_TURNOVER` 를 P3 **제외** 로 열거한다(spec.md:13). 그 밖에 하루 매수 총액을 세는 칸은 없다.
- Guardian 의 계좌 총계 한도는 셋 — 총 개방 노출 · 일일 실현 손실(금액 · 자본비) — 이고, 진입 발급 문 셋
  (`IssueEntry` · `IssuePrecheckedQFinalEntry` · `IssuePrecheckedQFinalCampaignFirstLeg`)은 모두 **OPEN_EXPOSURE 한 종류만**
  원자 예약한다(AST: `analysis/ast-evidence/internal-execgw--riskguardian.issueentry/`, q_final 두 문은 riskguardian_qfinal.go:244-246 ·
  riskguardian_first_leg.go:194-196). 체결되면 그 보유가 해제되므로(`TestDerivedTerminalFillReleasesTheHold`) 「오늘 이미 산 금액」은
  어떤 한도에도 남지 않는다.
- 일일 경계 관례는 **시장별 거래일 하나뿐**이다: `clock.Market.TradingDay`(Asia/Seoul · America/New_York 달력 날짜),
  DAILY_LOSS 예약의 `trading_day`(DDL 주석 「market-local」, 해제는 예약별 시장의 오늘과 비교 — AST
  `analysis/ast-evidence/internal-journal--sweeptradingdayboundary/` B9), `riskcalc.DailyLossInputs.TradingDay`(「market-local day」),
  주간 첫 다리 예약의 `MarketWeek`. **두 시장을 합친 「오늘」은 어디에도 없다** — 그래서 D1 이 사용자 결정이다.

## What Changes (사용자 결정 반영)

1. **설정** — `engine.automation_gate.max_daily_buy_notional`(선택, `limit_currency` = account base 통화). 없음 · 0 은 기존 관례
   (`omitempty` 0 = 미설정)대로 **기본 10,000,000 KRW** 를 적용. 음수 · NaN · Inf 는 기동 거부(다른 다섯 한도와 같음). base 가 KRW 가 아니고
   값이 없으면 → design D5(사용자 결정).
2. **판정 위치** — 체인 rung 이 아니라 **결정 · 예약 원자 트랜잭션 안**(design D3). 오늘 합계 = 오늘 발급된 진입 결정의 OPEN_EXPOSURE
   예약 금액 중 「체결이 없었음이 증명된 것」을 뺀 합 + 지금 HELD 인 예약 + 새 진입 금액. **≥ 한도이면 LIMIT_REACHED**(기존 총계 경계,
   design D4). 같은 트랜잭션이므로 동시 발급이 함께 넘을 수 없다. 합계를 계산할 수 없으면(경계 · 행 해석 실패) 발급 거부.
3. **금액 · 환산** — 세는 금액은 예약이 이미 쓴 `risk.EntryExposureValue`(지정가 × 수량 + 과대 추정 비용, account base, US 는
   요청 단위 frozen FX 로 환산 · ceil)를 그대로 쓴다. **새 FX 경로 0**, 새 예약 종류 0, 원장 스키마 변경 0(design D2-c).
4. **적용 문** — EXPOSURE_RAISING 진입 발급 문 셋 전부. `IssueReduction`(매도) · 보호 · 비상 청산 · 대사 · 체결 감지는 이 판정을 지나지 않는다
   (매도는 `IssueEntry` 에 못 들어오고(AST B3), `risk.Evaluate` 는 SELL 을 reduction 분기로 보냄(AST B2)).
5. **단일 출처 · 감사** — `execgw.Limits`(결정 스냅숏 5필드)는 그대로 두고(design D6), Guardian 이 상한을 진술하고 인터록이 config 해석 값과
   대조한다. `LimitsDigest` 는 상한을 포함해 바인딩을 결속한다. 변경은 기존 콘솔 한도 편집기 seam(`cmd/tossctl/limitsettings.go` →
   `config.SaveEngineGateLimits` 키 단위 쓰기 + 필드별 audit), 수동 편집은 기동 audit 가 추적 — 확인 문구 마찰 없음. 적용은 엔진 재시작 때.
6. **spec** — `risk-management`: 「Guardian 판정 체인」 제외 목록에서 DAILY_TURNOVER 를 빼고 신규로 옮김(MODIFIED),
   「Automation gate limits …」 상한 진술 · 인터록 대조(MODIFIED), 「정책 수치의 provenance」 기본값 집합에 일일 매수 추가(MODIFIED),
   「하루 매수 총액 한도」 요구 추가(ADDED).
7. **수동 매수 합산(D9=b)** — 엔진 보조 실행자가 기존 공식 주문 조회로 오늘 매수 주문을 읽어 시각 찍힌 스냅숏을 만들고, 엔진 주문(attempt 의
   broker_order_id)을 빼고 합계에 더한다. US 금액은 paired FX 조항. 스냅숏 없음 · 낡음 → 매수만 거부.
8. **콘솔 편집 칸** — 현재 한도 통화의 자유 숫자 입력(USD 프리셋 없음), KRW 미설정 → 기본값 표시, USD 미설정 → 기동 거부 안내, 통화 전환 시 지움.

## What Does NOT Change

- 매도 · 손절 · 익절 · 비상 청산 · flatten · 보호 주문 · 대사 · 체결 감지의 경로와 즉시성. reduce-only 발급(`IssueReduction`)은 무편집.
- 체인 12단계의 순서 · 이름 · reason-code. 재수집 루프 · 결정 TTL · 관측 기록 seam.
- 원장 스키마(v35 유지), 예약 종류 집합(OPEN_EXPOSURE · DAILY_LOSS · CASH), 예약 해제 규칙, 결정 `limits_json` 5필드.
- 수동 주문(`tossctl order place` · 토스 앱 등)을 **막지 않는다**(OFF = upstream). 합계에만 들어간다(design D9).
- 다른 다섯 한도의 값 · 의미. 주문당 notional 한도 · 그 FX 환산.

## Impact

- Affected specs: `risk-management`(MODIFIED 4 — FX 조항 포함, a137 사본은 a137 이 뺌 · ADDED 1), `engine-safety`(MODIFIED 1 — 인터록 조항 1·4). 병행 a137(일일 손실 입력 생산자)이 같은 capability 를 건드릴 수 있음 — 같은 요구를
  둘이 MODIFIED 하면 archive 충돌이므로 Manager 조정 필요.
- Affected code(예정): `internal/config/engine.go` · `limits.go` · `limits_io.go`(필드 하나), `internal/risk/input.go`(Policy · DefaultPolicy ·
  Validate), `riskguardian.go`(발급 요청 · LimitsDigest) · `riskguardian_qfinal.go` ·
  `riskguardian_first_leg.go`, `internal/journal/issuance.go` 계열(트랜잭션 안 합산 — 새 파일), `internal/app/engine/guardian_wiring.go` ·
  `interlock.go`, `cmd/tossctl/limitsettings.go`, `internal/console/settings_limits.go`(필드 하나), `internal/measure/counterfactual.go`(격자 정책), 수동 매수 refresher(새 파일).
  **공유 파일 위험**: `guardian_wiring.go` · `risk.Policy` · risk-management spec 은 a137 과 겹친다.
- 위험 등급: **High-risk**(Guardian · 사이징 인접 · 예약 트랜잭션). Pre-Edit 선언 · 편집 함수 FLM/BTM · RED 선행 · 변이 원장 · 독립 적대 리뷰 의무.
- 배포: 엔진 이미지 교체(엔진 코드 편집). `LimitsDigest` 값이 한 번 바뀐다 — 생산 전략 바인딩은 Guardian 자신에게서 유도하므로
  (`strategy_account_first_leg_authority.go:391`) 재서명 불요.

## 사용자 결정 (2026-10-14 — design 「사람 결정」 원문)

| ID | 결정 |
|---|---|
| D1 | D — KST 08:00 경계, `clock.AccountTradingDay` 한 곳(a137 공유, 82edde52) |
| D2 | c — 예약 원장 집계(0.7 실측: 증명 가능) |
| D4 | ≥ 도달 거부 |
| D5 | USD base 무값 → 기동 거부 |
| D9 | b — 수동 매수는 막지 않되 합계에 넣음. 분리 없이 a136 안(2차 결정) |
| 콘솔 | a136 안 — 자유 숫자 입력, USD 프리셋 없음, 통화 전환 시 지움 |
