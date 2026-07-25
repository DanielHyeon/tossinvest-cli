# StockOS 재사용 자산 인벤토리

> 조사일: 2026-07-26 · 대상: `/mnt/D/project/axipient/stockos` · 이식 언어: Go (UI는 React/TS)
> 원칙: (A) React UI 자산, (B) 거래 불변조건, (C) 순수 로직만 선별 이식. KIS·shadow/canary·runtime flag 체계는 이식 금지.

## 이식 판정 요약

| 자산 | 판정 |
|---|---|
| `apps/web/src/style.css` (10,451 LOC 단일 파일 디자인 시스템) | **그대로 이식 — 최대 가치 자산** |
| 안전 UX 패턴 (typed-confirmation, 사유 필수 입력, 차단사유 칩, 드로어 진단, ⌘K 팔레트) | 그대로 이식 |
| `apps/web/src/lib/useDashboardStream.ts` (SSE + 시퀀스 가드 + 폴링 폴백, 132줄) | 그대로 이식 |
| 정책 상수표 (RISK_PROFILES, min_rr=1.5, 재진입 한도, 슬롯 30% 등) | 수치 그대로, env 결합 제거 |
| 순수 로직 ~25개 모듈 | Go 재구현 + 테스트 케이스 이식 |
| API client `api.ts` (4,652 LOC, 261 타입) | 폐기 — `readJson`/`errorMessage`/`ApiError` 프리미티브와 nonce 갱신 패턴만 채택 |
| 차트 | **자산 없음(라이브러리 0개) — 그린필드** |
| MFE/MAE | 정리된 구현 없음 — 신규 설계 |
| React Query/Redux | 미사용 (Context + useState 폴링). TanStack Query 도입은 Phase 5에서 결정 |
| StockOS 백엔드(FastAPI·SQLAlchemy·KIS) | 이식 금지 |

## (A) React UI 주요 경로

- 셸·라우팅: `apps/web/src/App.tsx`(KIS 필드 재작성 필요), `src/router/routes.ts`(레거시 shim 제거), `src/router/useDrawerSync.ts`(그대로)
- 화면: 계좌 요약 `components/aboveTheFold/SummaryBento.tsx` · 포지션/주문 `components/OrdersPanel.tsx` + `orderTicket/UnifiedOrderTicket.tsx` · 체결 `dashboard/RecentFillsStrip.tsx` · 종목 상세 드로어 `components/drawer/` · 레인 `components/optimization/laneBoard·lanePerformance·lanePromotionChecklist` · kill switch `KillSwitchDisengageDialog.tsx`, `aboveTheFold/AutomationReadinessPanel.tsx`
- 공통: `confirmDialog/`, `palette/CommandPalette.tsx`, `settings/BannerStack.tsx`, `common/CurrencyBreakdownLines.tsx`, `lib/format.ts`, `lib/currencyTotals.ts`
- 폐기: `components/optimization/sections/`(canary/shadow 19종), `components/a149/`·`a152/`, `pages/LegacyTabPage.tsx`, KIS 카드류(`KisStatusCard` 등), `pages/ExternalIntelligencePage.tsx`

## (B) 거래 불변조건 — 계약과 위치

- **Guardian 단일 관문**: `packages/trading/stockos_trading/guardian.py:370 evaluate_guardian` — 판정 순서 고정(KILL_SWITCH → … → DAILY_LOSS → DUPLICATE → ALLOWED), 첫 실패 정지. 판정 순서와 수치를 보존해 Go로 이식, `os.getenv` 결합 제거
- **kill switch = BLOCK-ONLY**: `tests/test_a150_kill_switch_property.py` — 진입만 차단, 어떤 소비자도 강제청산하지 않음. 청산 평가는 별도 축
- **구조적 손절 (No Stop = No Trade)**: `tradeplan/contract.py:203`(stop 부재/비보호 거부), `guardian.py:659`(target/stop 계약), `orders/common_admission.py`(GEOMETRY_INVALID: stop<entry<target)
- **위험 기반 수량**: `tradeplan/contract.py:302`(risk×배수/(entry−stop), 마지막 1회 내림), `surge_dip_buy_canary/sizing.py`(fail-closed), `auto_order_execution.py:8727 _risk_capped_quantity`(fail-safe — 두 함수의 fail 방향이 의도적으로 다름)
- **최소 RR**: `tradeplan/contract.py` min_rr=1.5, rr 부재 시 거부(0 대체 금지); `structural_rr.py` measured-move, cap 6.0, 계산 불가 시 None(스탬프 금지)
- **한도**: RISK_PROFILES(`guardian.py:66`) — smoke 주문50만/노출500만/회전5,000만, small_live 100만/1,000만/5,000만; 일일 손실 = 절대액 OR equity% 중 먼저 도달(equity≤0 → 즉시 차단); 초과 시 신규만 차단, 강제청산 금지
- **중복 진입**: 당일 재진입 심볼당 1회(최대 3), 쿨다운 30분(최대 120), 미체결 BUY 보유 심볼 차단
- **멱등성**: `orders/common_admission.py`(idempotency_key 필수 + 23필드 fingerprint permit), `strategies/parker_vwap/signal_idempotency.py`(SHA-256, NUL 구분자, naive datetime 거부)
- **부분체결 상태기계**: `exit/in_flight_lifecycle.py` — 11상태(CREATED…PARTIALLY_FILLED…FAILED) 전이표, 순수. 강등 규칙 포함
- **합성 OCO**: `orders/synthetic_oco.py` — stop-first, 5상태, 브로커/DB 접근 0
- **슬롯 예산**: `slot_budget.py` — FAST 30% cap-only(총량 상향 불가)
- **레인 성과 표시 계약**: `lanePerformance.ts` — 결정적 PnL 링크 없으면 수치 렌더 금지(추정치 표시 금지)

## (C) 순수 로직 — Go 이식 우선순위와 테스트 소스

| 순위 | 모듈 | 테스트 (케이스 수) |
|---|---|---|
| 1 | `costs.py` + `strategies/parker_vwap/backtest/cost_model.py` (수수료 1.5bps·거래세 18bps·half-spread 3bps, 호가 라운딩) | `test_costs.py`(4), `test_a043_scalp_exit_math.py`(9) |
| 2 | `structural_rr.py` | `test_structural_rr.py`(14) |
| 3 | `tradeplan/contract.py` (사다리: 손절→보호→RR→등급→수량) | `test_a090_tradeplan_entry_contract.py`(36) |
| 4 | `orders/common_admission.py` | `test_a130_common_order_admission.py`(36, 순수 부분만) |
| 5 | `slot_budget.py` + `capital_stage.py` | `test_slot_budget.py`(14), `test_capital_stage.py`(24) |
| 6 | `guardian.py::evaluate_guardian` (env 제거) | `test_guardian.py`(20), `test_target_stop_contract.py`(29) |
| 7 | `exit/in_flight_lifecycle.py` + `orders/synthetic_oco.py` | `test_pm12_in_flight_lifecycle.py`(19), `test_a128_synthetic_oco.py`(12) |
| 8 | `backtest/metrics.py` + `ladder/r_multiple.py` (PF·MDD·승률·avg_r 등) | `test_backtests.py` |
| 참조 | `profit_ladder.py`, `abc_grade.py`, `setup_classification.py`, `exit_strategy.py`(상수표만), `indicators.py` | `test_profit_ladder.py`(19), `test_exit_strategy.py`(54), `test_abc_grade.py`(10) |

- 진입 후 markout: `toss_entry_comparison.py::A152_COMPARISON_PROTOCOL` — 윈도우 **5/15/30분**(1/3/5분 아님), 공식 = sell_fill − buy_fill − round_trip_cost. `_fixed_horizon_markout` 함수만 추출
- 오염 주의(순수해 보이지만 재작성): `orders/protective.py`(KIS 타입), `correlation_gate.py`(httpx+KIS), `execution_sync.py`(SQLAlchemy+KIS WS), `risk_envelope.py`(runtime_config 176필드)

## 회피 대상 (이식 금지)

- KIS: `packages/trading/stockos_trading/kis/` 전체, Python 파일 48%(237/498) 오염, 테스트 31개, env 키 36개
- 롤아웃 장치: shadow(139파일)/observe(119)/canary(53), `ExitEvaluatorMode` 4모드 축, A-넘버 shim 100+개
- runtime flag: `runtime_config.py` 2,887 LOC·176필드, 최적화 메뉴 스택 ~45만 자 — 같은 함정에 빠지지 말 것
- 거대 파일: `auto_order_execution.py` 17,830 LOC(순수 함수만 발췌), `persistence/models.py` 175K
- LLM 승인 게이트(`LLM_NOT_APPROVED`, vLLM 결박) — TossOS 미채택
- 루트에 커밋된 DB 파일(`stockos.db` 등) — 반면교사: .gitignore 강제

## 참고

- `apps/toss-intelligence/`(Go, 2,151 LOC)는 tossinvest-cli `d456a79`(v0.30.0)에 고정된 읽기 전용 사이드카 — TossOS와 동일 계보. Go 테스트 관례 참조용(`main_test.go` 84.6K)
- 프런트 테스트: vitest 112개(`useDashboardStream.test.ts` 등), Playwright E2E 21개(`rad-2-fold-1-invariant.spec.ts` 등) — UI 이식 시 계약 소스
