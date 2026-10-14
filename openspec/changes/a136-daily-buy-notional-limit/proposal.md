# a136 — 하루 매수 총액에 상한을 둔다 (daily-buy-notional-limit)

> **상태: 사용자 결정 대기(2026-10-14).** 일일 경계(design D1)는 저장소에 「시장 합산 하루」를 정의하는 관례가 없어서
> 지시대로 proposal 에서 멈춘다. 아래 spec 델타 · tasks 는 **추천안(D1-D · D2-c)** 으로 적었고, 사용자가 다른 안을 고르면
> 해당 문장만 바꾼다. 결정 전에는 production 코드를 편집하지 않는다.

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

## What Changes (추천안 기준 — D1 · D2 · D5 · D9 승인 뒤 확정)

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
5. **단일 출처 · 감사** — 새 한도를 Guardian 한도 스냅샷(`execgw.Limits`)과 인터록 감사 대상에 여섯째로 넣는다. 변경은 기존 콘솔
   한도 편집기 seam(`cmd/tossctl/limitsettings.go` → `config.SaveEngineGateLimits` 키 단위 쓰기 + 필드별 audit)으로만 — 확인 문구 마찰 없음.
   적용은 다른 한도와 같이 엔진 재시작 때.
6. **spec** — `risk-management`: 「Guardian 판정 체인」 제외 목록에서 DAILY_TURNOVER 를 빼고 신규로 옮김(MODIFIED),
   「Automation gate limits …」 다섯 → 여섯(MODIFIED), 「정책 수치의 provenance」 기본값 집합에 일일 매수 추가(MODIFIED),
   「하루 매수 총액 한도」 요구 추가(ADDED).

## What Does NOT Change

- 매도 · 손절 · 익절 · 비상 청산 · flatten · 보호 주문 · 대사 · 체결 감지의 경로와 즉시성. reduce-only 발급(`IssueReduction`)은 무편집.
- 체인 12단계의 순서 · 이름 · reason-code. 재수집 루프 · 결정 TTL · 관측 기록 seam.
- 원장 스키마(v35 유지), 예약 종류 집합(OPEN_EXPOSURE · DAILY_LOSS · CASH), 예약 해제 규칙.
- 수동 주문(`tossctl order place` 등 upstream `mutating: true` 명령) — Guardian 밖이므로 막지 않는다(OFF = upstream, design D9).
- 다른 다섯 한도의 값 · 의미. 주문당 notional 한도 · 그 FX 환산.

## Impact

- Affected specs: `risk-management`(MODIFIED 3 · ADDED 1). 병행 a137(일일 손실 입력 생산자)이 같은 capability 를 건드릴 수 있음 — 같은 요구를
  둘이 MODIFIED 하면 archive 충돌이므로 Manager 조정 필요.
- Affected code(예정): `internal/config/engine.go` · `limits.go` · `limits_io.go`(필드 하나), `internal/risk/input.go`(Policy · DefaultPolicy ·
  Validate), `internal/execgw/guardian.go`(Limits) · `riskguardian.go`(ExposureLimitsFor · 발급 요청) · `riskguardian_qfinal.go` ·
  `riskguardian_first_leg.go`, `internal/journal/issuance.go` 계열(트랜잭션 안 합산 — 새 파일), `internal/app/engine/guardian_wiring.go` ·
  `interlock.go`, `cmd/tossctl/limitsettings.go`, `internal/console/settings_limits.go`(필드 하나).
  **공유 파일 위험**: `guardian_wiring.go` · `risk.Policy` · risk-management spec 은 a137 과 겹친다.
- 위험 등급: **High-risk**(Guardian · 사이징 인접 · 예약 트랜잭션). Pre-Edit 선언 · 편집 함수 FLM/BTM · RED 선행 · 변이 원장 · 독립 적대 리뷰 의무.
- 배포: 엔진 이미지 교체(엔진 코드 편집). 한도 스냅샷이 바뀌므로 `LimitsDigest` 가 바뀐다 — 생산 전략 바인딩은 Guardian 자신에게서 유도하므로
  (`strategy_account_first_leg_authority.go:391`) 재서명은 불요로 보이나 구현 단계에서 실측.

## 사용자 결정 대기 (design 에 선택지 · 추천)

| ID | 질문 | 추천 |
|---|---|---|
| D1 | 시장 합산 「하루」의 경계 | **D1-D**: KST 08:00 에 넘어가는 하루(KR 정규장 + 그날 밤 US 정규장이 한 바구니) |
| D2 | 「체결된 매수」를 어디서 세나 | **D2-c**: 예약 원장(오늘 발급 · 체결 없음이 증명된 것 제외), 스키마 변경 없음 |
| D4 | 경계: 「넘으면」 vs 「도달하면」 거부 | **≥ 도달 거부**(기존 총계 관례, 1원 더 엄격) |
| D5 | base 통화가 USD 인데 값이 없을 때 | **기동 거부**(USD 값은 승인된 적 없음 · 코드 환율 금지) |
| D9 | 수동 주문을 합계에 넣나 | **넣지 않음**(Guardian 발급분만 · 수동 주문은 막지도 세지도 않음) |
