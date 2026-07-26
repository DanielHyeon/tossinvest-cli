# Guardian 판정 체인 — 순서와 StockOS 대응

> 참고 산출물(add-core-domain task 2.1). **규범은 `openspec/specs`의 risk-management 요구사항**이고, 이 문서는 그 순서를 코드·원본과 나란히 놓아 읽을 수 있게 한 표다.
> 구현: `internal/risk`(체인·사유 코드), `internal/costs`(실질 본전). 원본: `/mnt/D/project/axipient/stockos/packages/trading/stockos_trading/guardian.py`.
> 원 체인의 상대 순서 보존은 **규범이 아니다** — 원본에는 TossOS가 이식하지 않는 단계가 섞여 있고, 최소 RR은 원본에 아예 없는 신규 검사다.

## 1. TossOS 순서

`internal/risk.EntryChainSteps()`가 반환하는 순서가 정본이고, `TestChainOrderIsFixed`가 그것을 고정한다. 첫 실패에서 정지한다.

| # | 단계 | 사유 코드 | 원본 대응 |
|---|---|---|---|
| 0 | preflight (입력 가용성) | `INPUT_UNAVAILABLE` | a090 `inputs_unavailable` |
| 1 | kill switch | `KILL_SWITCH_ACTIVE` | guardian.py:378-379 |
| 2 | 운영 모드 | `OPERATING_MODE_BLOCKED` | guardian.py:420-433 (BUY_PAUSED·SELL_ONLY_MODE) — **구조 대체** |
| 3 | 진입 latch | `ENTRY_GATE_BLOCKED` | 없음 — TossOS 실장(execgw EntryGate). 배선은 task 2.4 |
| 4 | 심볼 allowlist | `SYMBOL_NOT_ALLOWED` | guardian.py:399-412 (레버리지·인버스 클래스 차단) — **구조 대체** |
| 5 | 손절 계약 | `STOP_MISSING` · `STOP_NOT_BELOW_ENTRY` · `TARGET_NOT_ABOVE_ENTRY` · `INVALID_TARGET_STOP` · `TARGET_BELOW_BREAK_EVEN` | guardian.py:659-714 `_verify_target_stop_contract` |
| 6 | 주문 크기 | `INVALID_ORDER_SIZE` · `MAX_ORDER_EXCEEDED` | guardian.py:448-449, :460-461 + a090 `size_zero` |
| 7 | 최소 RR | `MIN_RR_NOT_MET` | **신규** (원 체인에 카운터파트 없음 — 전략 계층 소관이던 것을 의도 산술로 승격) |
| 8 | 현금(비용 포함) | `INSUFFICIENT_CASH` | guardian.py:462-466 (두 코드를 하나로 — issues I4) |
| 9 | 당일 재진입 | `PENDING_BUY_ORDER_BLOCKED` · `SAME_DAY_REENTRY_BLOCKED` · `SAME_DAY_REENTRY_COOLDOWN_ACTIVE` | guardian.py:469-486, :510-553 |
| 10 | 총 개방 노출 | `OPEN_EXPOSURE_EXCEEDED` | guardian.py:487-489 |
| 11 | 일일 손실 | `DAILY_LOSS_LIMIT_REACHED` | guardian.py:499-502 |
| 12 | 중복 주문 | `DUPLICATE_ORDER` | guardian.py:503-504 |
| — | 통과 | `ALLOWED` | guardian.py:507 |

위험 감소(SELL) 의도는 이 체인을 타지 않는다. `INVALID_ORDER_SIZE`(수량 ≤ 0)와 `SELL_EXCEEDS_HOLDINGS`(보유 초과 = short) 두 가지만 판정하고 나머지는 전부 통과한다 — 근거는 §4.

### 순서에 대한 두 가지 결정

**손절 계약이 크기보다 먼저다.** risk-management: 손절가가 없거나 보호적이지 않은 진입 의도는 **수량 계산 이전 단계에서** 거부되어야 한다(SHALL). 위험 기반 수량이 손절폭의 함수이므로, 손절이 유효하지 않으면 계산할 수량 자체가 없다. 원본도 같다(guardian.py:457이 :460보다 앞).

**크기가 최소 RR보다 먼저다.** 스펙 표의 순서다("주문 크기 한도 → 최소 RR → 현금 검증"). 원본 a090은 반대(RR → sizing)였고, 그래서 a090의 `test_us_market_stop_rr_rungs_still_enforced`처럼 "사이징이 막혀도 RR이 먼저 보인다"에 의존하는 케이스는 순서 그대로는 이식되지 않는다(이식 시 통화가 맞는 정책으로 재작성 — `TestForeignCurrencyIntentIsNotSizedAgainstADomesticBudget`).

크기 단계 **안**의 순서는 (1) 수량 자체 → (2) 위험예산 상한 → (3) 설정 주문당 상한이다. 위험예산 상한이 설정 상한보다 앞인 이유: 그것은 방금 검증한 손절에서 파생된 **사이징 규칙**이고, 설정 상한은 그 바깥의 봉투다. 자기 손절에 비해 큰 주문은 한도를 올려도 여전히 틀렸다.

## 2. 경계값 규칙

| 대상 | 규칙 | 근거 |
|---|---|---|
| 주문당 수량·notional | **포함 상한** — 같으면 통과, 초과면 차단 | design D2, issues.md 판정 유지 |
| 총 개방 노출·일일 손실 | **도달 시 차단(≥)** | 예약 트랜잭션 실장과 일치 — `riskcalc.WithinLimit`(동률 fail-closed)을 그대로 소비 |
| 현금 | 포함 상한 — notional+비용이 현금과 같으면 통과 | 원본 `>` 비교와 동일 |
| 최소 RR | **미달만 거부** — 정확히 2.0은 통과 | "미달"의 문언. 유리수 정확 비교라 경계가 부동소수점 산물이 아니다 |
| 재진입 쿨다운 | 경과 시각 도달이면 통과 | 초 해상도 시계에서 쿨다운이 무한정 늘어나지 않게 |
| 계좌자본 | 0 이하면 즉시 차단 | risk-management 명문 |
| 집계 입력(총 노출·일손실) | **음수는 비교하지 않고 거부**(`INPUT_UNAVAILABLE`) | 둘 다 생산자 계약상 크기다(`riskcalc.DailyLoss`는 `max(0, −net)`). 부호 있는 손익을 넘기면 손실 난 날이 "한도 여유"로 읽힌다 — 이 파일에서 유일하게 게이트를 여는 방향의 입력 오류라 이름으로 막는다 |

## 3. 이식 분류표

risk-management의 열거를 코드 위치와 함께 옮긴다.

### 이식

| 원본 | 원본 위치 | TossOS |
|---|---|---|
| kill switch | guardian.py:378-379 | `checkKillSwitch` |
| 손절/목표 계약 | guardian.py:659-714 | `checkStopContract` |
| 실질 본전 | costs.py:240-266 | `costs.Model.BreakEvenSellPrice` |
| 주문 크기 | guardian.py:448-449, :460-461 | `checkOrderSize` |
| 현금(비용 포함) | guardian.py:462-466 | `checkCash` |
| 미체결 매수 차단 | guardian.py:469-478 | `checkSameDayReentry` |
| 당일 재진입 횟수·쿨다운 | guardian.py:510-553 | `checkSameDayReentry` |
| 총 개방 노출 | guardian.py:487-489 | `checkOpenExposure` |
| 일일 손실(절대·자본비) | guardian.py:499-502 | `checkDailyLoss` |
| 중복 주문 | guardian.py:503-504 | `checkDuplicateOrder` |
| 비용 모델 구조·검증 게이트 | costs.py:69, :72-116, :179-206 | `internal/costs` |

### 신규 (원본에 카운터파트 없음)

| 항목 | 이유 |
|---|---|
| 최소 RR 2.0 | 원 체인에 없다. 전략 계층의 `expected_rr` 게이트를 의도 산술로 승격 — 신호 산출물이 아니라 의도 필드 세 개의 순수 산술이다. 기본값 provenance: parker_vwap `default_lock.py:35-38`(§22 #2 lock, 초기값 2.0 — Plan 044가 1.3으로 완화한 것은 §0.9 역방향이라 미추종), `live_entry_contract.py:53`(`_DEFAULT_US_MIN_RR = 2.0`). 1.5는 live 게이트 최저 티어 값이라 기각 |
| 심볼 allowlist | 아래 구조 대체 참조 |
| `STOP_MISSING` | 원본은 `OrderCandidate.__post_init__` 생성자가 강제해 Guardian 사유 코드가 없다. TossOS는 생성자가 없으므로 체인이 사유를 낸다 |
| 진입 latch 단계 | TossOS 실장(execgw EntryGate — 401/403·SLO·RECONCILE·recovery) |
| 위험예산 상한 재계산 | 발급자가 계산한 수량을 체인이 다시 유도한다. 발급 후 변조를 게이트가 잡는다는 계약(engine-safety)의 사이징 판 |
| 입력 가용성 preflight | 원본은 Python 예외로 죽는다. TossOS는 사유 코드로 거부한다 |

### 제외

| 원본 | 원본 위치 | 제외 사유 |
|---|---|---|
| `CASH_ONLY_REQUIRED` (신용·미수) | guardian.py:442-443 | KIS 고유. Toss 공식 API에 대응 개념 없음 |
| `LLM_NOT_APPROVED` · `LLM_CONFIDENCE_LOW` | guardian.py:444-447 | LLM 게이트 미이식 |
| `LIVE_DISABLED` | guardian.py:434-435 | 기동 인터록이 대체(자동화 게이트 + 조항 1–6) |
| `ARM_EXPIRED` · `DASHBOARD_CONFIRMATION_REQUIRED` | guardian.py:436-441 | 운영 UI 소관 — 게이트 flip 사람 승인(§0.7)이 대체 |
| `BUY_PAUSED` · `SELL_ONLY_MODE` | guardian.py:420-433 | 운영 모드가 대체(구조 대체, §3 아래) |
| `TIME_WINDOW_BLOCKED` (미국장 진입 시간창) | guardian.py:616-656 | 미이식. 세션 판정은 `internal/clock`에 있으나 진입 시간창 정책은 이 change 범위 밖 |
| `DAILY_TURNOVER_EXCEEDED` | guardian.py:490-492 | P3 |
| `MAX_POSITIONS_EXCEEDED` | guardian.py:493-495 | P3 |
| `CANCEL_RATE_EXCEEDED` | guardian.py:505-506 | P3 |
| `SELL_COST_BUFFER_EXCEEDED` | guardian.py:467-468 | **§0.3** — 비용은 청산 게이트가 아니다. 매도가 비싸 보인다는 이유로 청산을 막는 것은 위험 통제가 실패해서는 안 되는 유일한 방향이다. `internal/costs/sellgate_test.go`가 이 식별자가 트리에 없음을 AST 스캔으로 고정한다 |
| capital stage | risk_profile_defaults.py | P3 |

### 구조 대체

| 원본 | TossOS | 비고 |
|---|---|---|
| 레버리지/인버스·ETF/ETN 클래스 차단 (`LEVERAGED_INVERSE_ETF_BLOCKED`, guardian.py:380-412) | 심볼 allowlist (`SYMBOL_NOT_ALLOWED`) | 분류 소스 `[미측정]` — Toss 공식 API에 상품 클래스 필드가 있는지 미확인. allowlist는 그 분류 없이도 같은 결과를 낸다(P3에서 재설계) |
| `BUY_PAUSED`/`SELL_ONLY_MODE` 운영자 토글 | 운영 모드 3종(NORMAL·ENTRY_BLOCKED·HALT_ALL) | design D3. EXIT_ONLY는 두지 않는다 — ENTRY_BLOCKED와 행동이 같아 §0.7 사다리의 무의미 단계 |
| 등급 배수 사이징 (a090) | 배수 **1.0 고정** | risk-management: 등급배수는 P3까지 1.0 고정(보수 하한). 등급은 신호 계층 산물이고 이 체인의 입력이 아니다 |
| `INSUFFICIENT_CASH` + `INSUFFICIENT_CASH_AFTER_COSTS` | `INSUFFICIENT_CASH` 단일 | 비용 포함이 항상 더 엄격 — issues.md I4 |

## 4. 위험 감소 경로가 통과하는 이유

각 생략은 규칙이지 누락이 아니다.

| 생략한 검사 | 근거 |
|---|---|
| kill switch | BLOCK-ONLY(SHALL NOT — 어떤 소비자도 강제청산을 유발하지 않는다). 신규 노출만 막는다. 청산까지 막으면 안전장치가 덫이 된다 |
| 운영 모드 | 모드×클래스 표: RISK_REDUCING은 세 모드 전부 허용. 수동 flatten-all도 모든 모드 통과(§0.3) |
| 진입 latch | 이름 그대로 진입용 |
| allowlist | 진입 통제다. allowlist에서 뺀 심볼을 청산할 수 없다면, 그 제거가 정확히 막으려던 포지션을 가둔다 |
| 비용 | §0.3 — 청산 게이트 미적용 |
| 총계 | 청산은 노출을 낮춘다 |
| 손절 계약·RR | 청산에는 목표도 손절도 없다 |

남는 것은 하나다: 보유보다 많이 파는 것 = short. `SELL_EXCEEDS_HOLDINGS`.

## 5. 후속 task가 채울 자리

| task | 채울 것 |
|---|---|
| 4.1 | 체인 ALLOW → `RecordDecisionAndReserve` → Gateway. `Policy`를 감사된 설정 한도에, `AccountState`의 총계·현금·재진입·allowlist 값을 실제 출처(riskcalc 집계, journal 당일 이력, 설정 allowlist)에 배선 |
| 4.2 | `Policy`를 `execgw.Limits`(Set 비트)로 진술하는 `ExposureLimiter` |
