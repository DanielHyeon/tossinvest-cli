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
| 3 | 진입 latch | `ENTRY_GATE_BLOCKED` | 없음 — TossOS 실장(execgw EntryGate). 매핑은 §5 |
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
| 최소 RR | **미달만 거부** — 정확히 2.0은 통과 | "미달"의 문언. 유리수 정확 비교라 경계가 부동소수점 산물이 아니다. **이 정밀도 근거는 총 RR 게이트에만 성립한다** — 순 RR 관측값은 `BreakEvenSellPrice`의 float64 산술에서 온 본전을 물려받으므로 마지막 유효자리에 상대오차가 있다(관측값이므로 결함이 아니라 기록 대상). 순 기준으로 게이트를 옮기는 change는 본전을 위로 올림하거나 유리수 end-to-end로 바꾸는 것을 **선행**해야 한다 — `internal/risk/netrr_test.go`가 반례 둘(요율 0에서 등호, 고정밀 진입가에서 순 > 총)을 고정한다 |
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
| 최소 RR 2.0 | 원 체인에 없다. 전략 계층의 `expected_rr` 게이트를 의도 산술로 승격 — 신호 산출물이 아니라 의도 필드 세 개의 순수 산술이다. **기준은 총 RR이다.** 기본값 provenance: parker_vwap `default_lock.py:35-38`(§22 #2 lock, 초기값 2.0 — Plan 044가 1.3으로 완화한 것은 §0.9 역방향이라 미추종), `live_entry_contract.py:53`(`_DEFAULT_US_MIN_RR = 2.0`, **미국 시장 한정·설정 가능 범위 2.0~4.0의 구조 RR 플로어**). 1.5는 live 게이트 최저 티어 값이라 기각 — 이 근거는 기준(총·순)에 의존하지 않으므로 기준을 순으로 바꾸는 것만으로 해소되지 않는다. **인용된 두 출처는 모두 총·구조 RR 게이트이며 순 기준 선례가 아니다** — 순 기준으로 거론되는 값은 StockOS `early_entry_geometry.py`의 `NET_RR_INSUFFICIENT`(KRX 1.5 / US 2.0)과 058 사후 분석 처방 1.3이고, 둘 다 파일·문서 경로와 검증 상태를 병기하지 않으면 provenance 없는 수치다. 순 기준 임계값을 정하는 change는 총 2.0을 승계 근거로 인용할 수 없고 관측 분포에서 도출해야 한다(risk-management SHALL NOT) |
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

## 5. 진입 latch 매핑 (task 2.4)

체인 3단계의 입력은 `AccountState.EntryBlockedLatch`·`EntryBlockedReason` 두 값이고, 그 값의 유일한 생산자는 `execgw.(*EntryGate).EntryLatchFor(market, symbol)`다(`internal/execgw/entrylatch.go`). 체인은 조건을 **재유도하지 않는다** — 게이트가 이미 답한 질문(`CheckEntryFor`, 봉인된 제출 시퀀스가 부르는 그 호출)을 같은 값으로 옮겨 받는다.

| risk-management이 부르는 조건 | 게이트 사유 코드 | 해제 방법 |
|---|---|---|
| 자격증명 실패(401/403) | `broker_auth_rejected` | 운영자 `Clear` — 조회 성공은 해제하지 않는다 |
| 체결 감지 SLO 위반 | `fill_detection_slo_violated` | 측정 회복 |
| RECONCILE | `reconciliation_mismatch` / `reconciliation_mismatch_permanent` | 재대조 일치(전자) / 운영자(후자) — journal `reconcile_states` 투영 |
| recovery 미완료 | `recovery_incomplete` | 복구 완료 |
| (열거 밖) 조회 신선도 | `required_query_stale` | 다음 성공 폴에서 자동 |
| (열거 밖) flatten 진행·미해소 IN_DOUBT·알림 미전달·UNKNOWN_BROKER_STATE | `flatten_in_progress` · `unresolved_in_doubt` · `critical_alert_undelivered` · `unknown_broker_state` | 각 사유의 규칙 |
| 운영 모드 투영 (task 3.1) | `operating_mode_blocked` | 모드 완화(사람 승인) |

열거 밖 사유도 그대로 통과시키는 이유: 체인이 게이트 사유의 **부분집합**만 본다면, Guardian이 허용한 의도를 Gateway가 곧바로 거부하게 된다 — 결정과 예약을 쓰고 되돌리는 왕복이 그 사이에 낀다. 체인의 latch 단계는 게이트 판정의 사본이 아니라 같은 판정의 **이른 소비**다.

**중복 차단이 아니다.** 한 조건은 물어본 지점마다 한 번씩만 거부를 낸다:

- 체인이 먼저 거부하면 결정 행이 없다 → Gateway에 도달하는 것이 없으므로 Gateway는 아무것도 거부하지 않는다.
- Guardian ALLOW 이후 제출 전에 latch가 서면 Gateway가 자기 코드로 거부한다(예약 롤백은 4.1의 원자 발급이 담당). 같은 사실을 **두 시점**에 관측한 것이지 한 시점을 두 번 판정한 것이 아니다.

모드는 체인 2단계(`OPERATING_MODE_BLOCKED`)와 3단계(`operating_mode_blocked` latch) 양쪽에 나타난다. 둘은 같은 journal 행을 읽으므로 답이 갈릴 수 없고, 순서가 정하는 것은 **운영자가 보는 이름**뿐이다 — 구체적이고 조치 가능한 쪽(모드)이 먼저다.

## 6. 운영 모드 (task 3.1–3.3)

권위는 journal `operating_modes`이고, 현재 모드는 **최신 행**이다 — `ORDER BY created_at DESC, rowid DESC LIMIT 1`. journal 타임스탬프가 초 해상도라 한 초 안의 두 전환은 `created_at`만으로 순서가 없고, 그 모호성이 걸리는 유일한 방향이 "보수 강화 직후의 완화가 먼저 읽힌다"이기 때문이다(issues.md 0.1 → 3.1).

### 모드×클래스 표 — `journal.ModeAllows`

| 모드 | EXPOSURE_RAISING | RISK_REDUCING | PROTECTION_WEAKENING\* |
|---|---|---|---|
| NORMAL | 허용 | 허용 | 허용(audit) |
| ENTRY_BLOCKED | 거부 | 허용 | 허용(audit) |
| HALT_ALL | 거부 | 허용 | **거부** |

\* 열 예약. 이 빌드는 PROTECTION_WEAKENING을 **발급하지 않는다** — landed `RecordDecision`이 전 모드에서 거부하므로 허용(audit) 셀은 2c 발급 도입 후에 효력을 갖는다. 즉 지금은 journal의 거부가 이 표보다 엄격하고, 그게 안전한 방향이다. 미지 모드·미지 클래스는 아무것도 허용하지 않는다.

### 강제 지점 — 전환은 하나의 흐름

`journal.TransitionOperatingMode`가 **영속 → 커밋 → EntryGate 투영 → 알림** 순으로 한 번에 한다. "투영 없이 append"도 "append 없이 투영"도 공개 API에 없다. 커밋 후 투영인 이유는 완화 방향이다(투영이 먼저면 근거 행이 durable해지기 전에 latch가 풀린다). 강화 쪽 창(커밋됐지만 아직 투영 안 됨)은 기동 시 `RestoreOperatingModeProjection`이 닫는다.

봉인된 제출 시퀀스는 변경되지 않는다 — 모드는 `operating_mode_blocked` latch로 도착하고, Gateway가 늘 부르던 `CheckEntryFor`가 그것을 소비한다.

### 자동 강화 트리거 (전부 → ENTRY_BLOCKED)

| 트리거 상수 | 조건 | 생산자 |
|---|---|---|
| `DAILY_LOSS_LIMIT_REACHED` | 일일 손실 한도 도달 | 발급자 (task 4.x) |
| `BROKER_AUTH_REJECTED` | 자격증명 실패(401/403) | execgw Retrier 배선 (task 4.x) |
| `CRITICAL_ALERT_UNDELIVERED` | critical 알림 outbox 전달 실패 지속 | obs Notifier 배선 (task 4.x) |
| `EXIT_OBSERVATION_OUTAGE` | exit 관측 두절 임계 초과 | 판정 루프 (task 7.4) |

**HALT_ALL로 가는 트리거는 없다**(SHALL NOT — 운영자 결정). 분석·성과 작업 실패는 열거 밖이므로 구조적으로 강화할 수 없다(SHALL NOT). 트리거 목표가 현재 모드보다 느슨하면 no-op이다(보수 우선) — 이것이 반복 트리거의 멱등성이자, 알림 실패 → 강화 → 알림 실패 되먹임 고리가 한 바퀴에 닫히는 이유다.

완화는 `actor=OPERATOR` + 승인 참조 + audit 줄을 요구하고, audit 쓰기가 실패하면 전환 자체가 롤백된다(2a의 운영자 예약 해제와 같은 순서).

### 스냅샷 읽기 표면

`journal.ModeSnapshot`(`CurrentOperatingMode`) 하나를 세 소비자가 읽는다: Gateway는 `Allows(EXPOSURE_RAISING)`, Guardian 체인은 `Mode`(발급자가 `risk.AccountState.Mode`로 매핑), flatten은 `Allows(RISK_REDUCING)` — 세 모드 전부 참이다(§0.3). 행이 없는 계좌는 NORMAL(`Recorded=false`)이다.

## 7. 후속 task가 채울 자리

| task | 채울 것 |
|---|---|
| 4.1 | 체인 ALLOW → `RecordDecisionAndReserve` → Gateway. `Policy`를 감사된 설정 한도에, `AccountState`의 총계·현금·재진입·allowlist 값을 실제 출처(riskcalc 집계, journal 당일 이력, 설정 allowlist)에 배선 |
| 4.2 | `Policy`를 `execgw.Limits`(Set 비트)로 진술하는 `ExposureLimiter` |
