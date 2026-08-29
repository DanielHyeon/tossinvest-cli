# risk-management Specification

## Purpose
Guardian 판정 체인(고정 순서·reason-code)·위험 기반 수량·발급 절차와 예약 권위·정책 수치 provenance·kill switch·운영 모드(모드×클래스 표) 요구사항.
## Requirements
### Requirement: Guardian 판정 체인
모든 자동 진입 의도는 Guardian 판정 체인을 통과해야 하며(SHALL), 체인은 고정 순서로 평가되고 첫 실패에서 정지한다: kill switch/운영 모드 → 게이트 상태(진입 차단 latch: 401/403·SLO 위반·RECONCILE·recovery 미완료) → 심볼 allowlist → 구조적 손절 계약(No Stop=No Trade·보호적 stop·실질 본전 미달 target 거부) → 주문 크기 한도 → 최소 RR(신규 검사) → 현금 검증(비용 포함) → 당일 재진입 규칙 → 총 개방 노출 한도 → 일일 손실 한도 → 중복 주문 검사 → ALLOW. 각 거부는 안정적 reason-code로 기록된다(SHALL).

**순서의 권위는 이 표다**(SHALL — StockOS `evaluate_guardian`과의 대응은 `docs/guardian-chain.md` 참고 산출물로 남기되 순서 보존을 규범으로 주장하지 않는다: 원 체인에는 미이식 단계가 섞여 있고 최소 RR은 원 체인에 없는 신규 검사다). 이식 분류의 열거(SHALL):

- **이식**: kill switch·모드, 진입 latch, 손절 계약(reason: STOP_MISSING·STOP_NOT_BELOW_ENTRY·TARGET_NOT_ABOVE_ENTRY·INVALID_TARGET_STOP·TARGET_BELOW_BREAK_EVEN), 주문 크기(INVALID_ORDER_SIZE·MAX_ORDER_EXCEEDED), 현금(INSUFFICIENT_CASH·비용 포함), 당일 재진입(쿨다운 기본 30분 `[미검증]`·당일 심볼당 최대 진입 수 기본 2회 `[미검증]`·**미체결 매수 존재 시 진입 차단 PENDING_BUY_ORDER_BLOCKED**), 총 노출, 일손실, 중복 주문
- **신규**: 최소 RR, 심볼 allowlist, STOP_MISSING(원본은 생성자 강제라 카운터파트 없음)
- **제외**: KIS 고유(CASH_ONLY 등), LLM 게이트, capital stage, ARM/DASHBOARD 확인(운영 UI 소관 — 게이트 flip 승인이 대체), LIVE_DISABLED(인터록이 대체), BUY_PAUSED/SELL_ONLY(운영 모드가 대체), 미국장 진입 시간창, DAILY_TURNOVER·MAX_POSITIONS·CANCEL_RATE(P3), SELL_COST_BUFFER(§0.3)
- **구조 대체**: 레버리지/인버스·ETF/ETN 클래스 차단 → 심볼 allowlist(분류 소스 `[미측정]`, P3 재설계)

체인의 모든 입력은 의도 필드·브로커 스냅샷·journal 상태에서 온다(SHALL — 신호 계층 산물 없음: 구조적 RR 계산·등급배수는 P3). 총계 한도의 경계값은 도달 시 차단(≥)이며(SHALL — 예약 트랜잭션의 실장과 일치), 주문 단위 한도는 포함 상한(초과 시 차단)을 유지한다.

#### Scenario: 체인 첫 실패 정지
- **WHEN** kill switch가 활성인 상태에서 진입 의도가 평가되면
- **THEN** 후속 판정 없이 KILL_SWITCH_ACTIVE로 즉시 거부된다

#### Scenario: allowlist 밖 심볼
- **WHEN** allowlist에 없는 심볼의 진입 의도가 평가되면
- **THEN** SYMBOL_NOT_ALLOWED로 거부된다

#### Scenario: 총계 경계값 도달
- **WHEN** 당일 실현 손실이 한도에 정확히 도달한 상태에서 진입 의도가 평가되면
- **THEN** DAILY_LOSS_LIMIT_REACHED로 거부된다 (계좌자본 0 이하이면 즉시 차단)

### Requirement: No Stop = No Trade와 위험 기반 수량
손절가가 없거나 보호적이지 않은 진입 의도는 수량 계산 이전 단계에서 거부되어야 한다(SHALL). TossOS는 long-only이므로 진입은 매수이고 보호적 손절은 `stop < entry`이며, 매도는 보유수량 이하 reduce-only, short 노출은 구조적으로 금지된다(SHALL NOT). 위험 기반 수량은 `floor(위험예산 / (entry − stop))`이고(등급배수는 P3까지 1.0 고정 — 보수 하한), stop 폭 0 이하는 수량 0(fail-closed)이다(SHALL). 최소 RR은 의도 필드의 순수 산술 `(target − entry) / (entry − stop)`로 검사하며(신규 검사 — 세션 구조 기반 산출은 P3), 기본값 2.0 미달·계산 불가는 거부한다(SHALL — 0 대체 금지).

**임계값 2.0의 provenance는 총·구조 RR 기준에 한정된다**(SHALL — 순 기준으로 이전 적용해서는 안 된다): 인용 출처는 StockOS 라이브 게이트 `live_entry_contract.py:53`의 2.0(시장 한정·설정 범위 1.0~5.0의 구조 RR 플로어)과 `default_lock` 초기값(Plan 044가 1.3으로 완화 — §0.9상 미추종)이다. **1.5를 기각한 근거는 설정 범위의 최저 티어 값이라는 사실이며 이 근거는 기준(총·순)에 의존하지 않는다**(SHALL — 기준을 순으로 바꾸는 것만으로 이 기각이 해소되지 않는다).

순 기준 선례로 거론되는 시장별 KRX 1.5 / US 2.0과 사후 분석 처방 1.3은 **파일·문서 경로와 검증 상태를 병기하지 않는 한 provenance 없는 수치다**(SHALL NOT — 출처 없는 수치를 선례로 인용해서는 안 된다). 순 기준 임계값을 정하는 change는 총 기준 2.0을 승계 근거로 인용할 수 없고(SHALL NOT), 관측된 분포에서 도출해야 한다.

**provenance 규율이 완화 경로가 되어서는 안 된다**(SHALL NOT). 순·총 기준의 엄격성 순서는 **손절 폭에 따라 뒤집힌다**: 실질본전 비율을 `c`, 순 임계값을 `r`, 총 임계값을 `R`(`R > r`)이라 하면 두 요구 target이 같아지는 손절 폭은

```
s* = (1 + r)(c − 1) / (R − r)
```

이다. `r=1.5·R=2.0`에서 `s* = 5(c − 1) ≈ 5 × (매수측 요율 + 매도측 요율)`이며, 현행 `DefaultModel`에서 KR(`c` = 1.0050201) `s*` = **2.51%**, US(`c` = 1.0212336) `s*` = **10.62%**다. 손절 폭이 `s*`를 넘으면 순 1.5가 총 2.0보다 느슨하다 — KR 손절 −5%에서 총 2.0은 target +10.000%를 요구하는데 순 1.5는 +8.755%만 요구한다(그 target의 총 RR은 1.751로 현행 거부 대상이다).

**이 두 교차점은 `[미검증]` 과대추정 요율에서 나온 상한이다**(SHALL — 2b 실측이 요율을 낮추면 `s*`도 낮아져 완화 구간이 넓어진다: 수수료 0.015%·거래세 0.18%면 `c` = 1.0021041, `s*` = **1.05%**). 교차점 수치는 비용모델 지문과 함께만 인용될 수 있다(SHALL). US 교차점이 10.62%라는 사실이 증명 의무를 면제하지 않는다(SHALL NOT — TossOS는 손절 폭 상한을 갖지 않으므로 그보다 넓은 손절도 허용 입력이다).

순 기준으로 전환하는 change는 임계값의 출처와 무관하게 **지원되는 모든 시장·허용 입력에서 새 게이트의 허용 집합이 현행 총 RR 2.0의 허용 집합을 초과하지 않음을 증명해야 한다**(SHALL — §0.9). 단일 임계값 수치 비교로는 불충분하고, **반례가 하나라도 있으면 그 임계값을 채택해서는 안 된다**(SHALL NOT). 이 요구의 귀결을 명시한다: 손절 폭이 커지면 `r·s`와 `R·s`가 지배하므로 **`r < 2.0`인 어떤 순 임계값도 충분히 넓은 손절에서는 반드시 완화된다.** 따라서 §0.9를 만족하는 형태는 둘뿐이다 — 순 임계값을 2.0 이상으로 두거나, 현행 총 2.0을 **유지한 채 순 검사를 논리곱으로 추가**하는 것(후자는 구성상 단조 강화다).

진입 손절가는 사전 검사로 끝나지 않고 포지션 개시 시 exit 정책의 t0 기준선이 된다(SHALL — exit-policy).

#### Scenario: 손절 없는 진입
- **WHEN** stop이 없는 진입 의도가 평가되면
- **THEN** STOP_MISSING으로 거부되고 수량 계산은 수행되지 않는다

#### Scenario: RR 미달
- **WHEN** (target−entry)/(entry−stop)이 2.0 미만인 의도가 평가되면
- **THEN** MIN_RR_NOT_MET으로 거부된다

#### Scenario: 보유수량 초과 매도
- **WHEN** 보유수량을 초과하는 매도 의도가 평가되면
- **THEN** short 노출이 되므로 거부된다

#### Scenario: 순 기준 임계값의 근거 제한
- **WHEN** 순 기준 최소 RR 임계값을 정하는 변경이 제안되면
- **THEN** 총 기준 2.0을 승계 근거로 인용할 수 없고 관측된 분포를 근거로 제시해야 한다

#### Scenario: 손절 폭에 따른 엄격성 역전
- **WHEN** 순 기준 임계값이 어떤 손절 폭에서는 현행 총 기준보다 느슨해지는 것으로 확인되면
- **THEN** 그 값은 채택될 수 없고, 운용 범위 전체에서 느슨해지지 않는 값이 선택되어야 한다

### Requirement: 발급 절차와 예약의 권위
발급 절차는 실장 가능한 순서를 따른다(SHALL): 체인 ALLOW → **결정 영속과 예약 삽입을 하나의 journal 트랜잭션으로 수행**(신규 원자 API — 예약의 결정 FK가 만족되면서 결정과 예약 사이 크래시 창이 없다) → Gateway 제출. 총계 한도의 최종 권위는 예약 트랜잭션이며(SHALL), 예약이 거부되면 결정도 같은 트랜잭션에서 함께 롤백되어 제출 가능한 결정이 남지 않는다(SHALL — Gateway의 HELD 예약 검증이 이를 이중으로 강제한다: engine-safety delta). 예약 실패는 원인별 안정 reason-code로 기록된다(SHALL): LIMIT_REACHED(한도 도달·즉시), SNAPSHOT_RECOLLECTION_EXHAUSTED(재수집 상한·데드라인 소진), VERSION_CONFLICT(원장 버전 경합), DECISION_EXPIRED(재수집 중 결정 만료). 재수집 시 체인은 재실행하지 않는다(SHALL — 모드·latch는 Gateway 제출 시 EntryGate 재검사가 커버하고, 결정 만료(60초 — 실장 `DefaultDecisionTTL`)가 나머지 입력의 신선도 상한이다. 재수집 예산 10초 < TTL).

**관측 기록은 이 원자 트랜잭션의 구성원이 아니다**(SHALL — 진입 판정의 관측 기록 요구사항이 정의하는 seam): 관측은 트랜잭션 커밋 이후에 시작되며, 그 실패는 예약 실패 reason-code 집합에 속하지 않고 위험 거부로 분류되지 않는다(SHALL NOT).

발급 순서 자체는 변경되지 않는다(SHALL — 체인 ALLOW → 결정·예약 원자 트랜잭션 커밋 → Gateway 제출). **관측 기록은 Gateway 제출의 선행조건이거나 동기 대기 지점이어서는 안 되며, 그 실패·지연이 제출이나 HELD 예약 검증을 지연·취소해서는 안 된다**(SHALL NOT — 관측 쓰기가 제출 앞을 막으면 원장 지연이 결정 TTL을 소진해 정상 진입을 잃는다). 이 순서가 만드는 크래시 창은 관측 결손의 부분 복원 가능성으로 상계된다.

#### Scenario: 예약 거부 시 결정 부재
- **WHEN** 체인 ALLOW 후 동시 결정이 한도 잔여를 소진해 예약이 거부되면
- **THEN** LIMIT_REACHED가 기록되고, 롤백으로 제출 가능한 결정이 존재하지 않는다

#### Scenario: 발급과 제출 사이 모드 강화
- **WHEN** 발급 직후 자동 강화로 ENTRY_BLOCKED가 되면
- **THEN** Gateway 제출 시 EntryGate 재검사가 진입을 거부한다

#### Scenario: 관측 실패가 발급을 되돌리지 않는다
- **WHEN** 원자 트랜잭션이 커밋된 뒤 관측 기록이 실패하면
- **THEN** 발급된 결정과 예약은 유지되고, 제출은 정상 진행된다

### Requirement: GuardianDecision 발급자
Guardian은 engine-safety 메인 스펙의 결정 계약을 구현하는 발급자다(SHALL): EXPOSURE_RAISING 결정(RiskIntent preimage·`f(decision_id, generation)` 멱등키·인터록이 감사한 설정 한도 스냅샷·만료 60초·nonce)과 위험 감소의 ReductionIntent 결정(한도 없음)을 발급하고, `ExposureLimiter`를 구현해 자기 한도를 진술한다(SHALL — 인터록 단일 출처 검증 통과). 위험 입력은 발급 요청의 원 의도에서 오고, 발급 후 변조는 Gateway의 journal 기반 재검증이 차단한다.

#### Scenario: 발급자 한도 진술
- **WHEN** 인터록이 Guardian의 ExposureLimits를 감사된 설정 한도와 대조하면
- **THEN** 필드·Set 비트 단위로 일치한다

### Requirement: 정책 수치의 provenance

모든 한도·정책 수치는 코드에 출처와 함께 기록되어야 하며(SHALL — 출처는 ① StockOS 파일·심볼과 그 검증 상태, 또는 ② TossOS 실측이다; ②는 `verify-*` change의 `measurements.md`에 남은 관측을 **식별자로 인용**해야 한다(SHALL — 관측 번호·날짜·시장; "관측했다"는 서술은 출처가 아니다)), 사용자 미확정 시 보수 기본값 전체 집합을 사용한다(SHALL — 인터록 5필드를 전부 충족): 주문당 notional 1,000,000 KRW·주문당 수량 100주·총 노출 10,000,000 KRW·일일 손실 100,000 KRW·일일 손실 자본비 1%·통화 KRW. USD 축의 승인 집합은 주문당 notional 500 USD·주문당 수량 100주·총 노출 1,500 USD·일일 손실 50 USD·일일 손실 자본비 1%이며, 이 집합은 기본값이 아니라 **운영자가 명시적으로 고를 때만** 적용된다(SHALL NOT — 통화가 하나인 게이트에서 USD를 기본값으로 두면 국내 자동 진입이 조용히 닫힌다). "Toss 검증됨" 표시는 실계좌 검증 결과에만 근거해 전환한다(SHALL NOT — 미검증 상태에서 검증됨 표기 금지). 수치 변경은 §0.9(보수 방향만) 검토와 audit 기록을 요구한다(SHALL). 완화 방향의 수치 등재는 **새 통화 축을 등재하는 경우에 한해**, 각 수치가 이미 승인된 다른 통화의 선호 안쪽에 있고(SHALL — 등가 논증과 **그 논증이 깨지는 환율**을 change 문서에 기록한다), 사람이 그 수치를 명시적으로 승인했으며, 근거를 실측 식별자로 인용할 때만 허용한다(SHALL — 셋 중 하나라도 없으면 등재하지 않는다). 환산 환율은 코드에 넣지 않는다(SHALL NOT — 등가는 등재 시점의 정책 논증이며 런타임이 계산할 값이 아니다; 통화 간 정규화는 FX staleness 경계를 가진 riskcalc의 자리다).

#### Scenario: 한도 완화 시도

- **WHEN** 일일 손실 한도를 높이는 설정 변경이 적용되면
- **THEN** audit 로그에 이전·새 값이 기록되고 변경은 사람 승인 절차를 거친다

#### Scenario: 기본값 게이트 ON 기동

- **WHEN** 사용자 미확정 기본값 집합으로 게이트 ON 기동하면
- **THEN** 인터록 한도 검증(5필드·통화)이 통과한다

#### Scenario: 실측 출처를 가진 수치의 등재

- **WHEN** StockOS에 대응 심볼이 없는 수치를 레지스트리에 등재하면
- **THEN** 코드가 그 수치의 근거를 `measurements.md`의 관측 식별자·날짜·시장으로 인용하고, 등가 논증과 그것이 깨지는 지점이 change 문서에 남는다

#### Scenario: 승인 봉투를 벗어나는 완화 거절

- **WHEN** 새 통화 축의 어떤 수치가 이미 승인된 다른 통화의 선호를 넘으면
- **THEN** 그 수치는 등재되지 않고 기존 값이 유지된다

### Requirement: Kill switch와 운영 모드 — 모드×클래스 표

kill switch는 신규 진입 차단 전용(BLOCK-ONLY)이며 어떤 소비자도 강제청산을 유발하지 않는다(SHALL NOT). 운영 모드는 mutation safety class 허용 표로 정의된다(SHALL):

| 모드 | EXPOSURE_RAISING | RISK_REDUCING | PROTECTION_WEAKENING* |
|---|---|---|---|
| NORMAL | 허용 | 허용 | 허용(audit) |
| ENTRY_BLOCKED | 거부 | 허용 | 허용(audit) |
| HALT_ALL | 거부 | 허용 | **거부** |

*PROTECTION_WEAKENING의 발급·소비는 보호주문 도입 change가 정의하며 이 표는 열을 예약한다 — **현재는 landed `RecordDecision`이 전 모드에서 이 class를 거부하므로 허용(audit) 셀은 2c 발급 도입 후에 효력을 갖는다**. HALT_ALL은 이 change 안에서는 ENTRY_BLOCKED와 행동이 같지만, 운영자 전용 진입(자동 강화 없음)이라는 승인 의미와 2c의 PROTECTION_WEAKENING 거부로 구별되므로 유지한다. EXIT_ONLY는 두지 않는다 — ENTRY_BLOCKED와 행동이 동일해지므로(구별 근거 없는 모드는 §0.7 승인 사다리의 무의미 단계다), 실제 행동 차이가 생기는 change가 재도입한다. 수동 flatten-all은 모든 모드에서 통과한다(§0.3).

모드의 강제 지점은 EntryGate 투영이다(SHALL — 모드 전환은 journal 영속과 동시에 EntryGate 계좌 latch로 투영되고, Gateway의 기존 제출 시 재검사가 이를 소비한다; 봉인된 제출 시퀀스는 변경되지 않는다). 전환 승인은 방향 비대칭이다(SHALL): 보수 방향(NORMAL→ENTRY_BLOCKED→HALT_ALL)은 자동·즉시·durable, 완화는 사람 승인(§0.7)+audit. 자동 강화 트리거와 목적 상태의 열거(SHALL): 일일 손실 한도 도달 → ENTRY_BLOCKED, 자격증명 실패(401/403) → ENTRY_BLOCKED, critical 알림 outbox 전달 실패 지속 → ENTRY_BLOCKED, exit 관측 두절 임계 초과 → ENTRY_BLOCKED, **reconcile 사이클 지속 실패(연속 5주기) → ENTRY_BLOCKED, 체결 감지 사이클 지속 실패(연속 5주기) → ENTRY_BLOCKED**(엔진 런타임 change — 대사·체결에 눈이 먼 엔진은 새 진입을 받으면 안 된다) — 전부 메인 스펙의 "신규 진입 차단"과 정합하며 HALT_ALL 자동 진입은 없다(SHALL NOT — HALT_ALL은 운영자 결정). 분석·성과 작업 실패는 트리거가 아니다(SHALL NOT — 대사·체결 감지는 실행 경로이지 분석이 아니다). 모드·kill switch·이력은 journal 영속·재시작 유지(SHALL), 동시 적용 시 보수 우선(SHALL).

#### Scenario: 손실 한도 도달 시 자동 강화
- **WHEN** 일일 손실 한도 도달로 자동 강화가 발동하면
- **THEN** 사람 승인 없이 ENTRY_BLOCKED로 즉시 전환·영속되고 EntryGate에 투영되며 알림이 발송된다

#### Scenario: HALT_ALL 중 청산
- **WHEN** HALT_ALL 상태에서 RISK_REDUCING 청산이 요청되면
- **THEN** 허용된다 (수동 flatten-all 포함)

#### Scenario: 분석 작업 실패는 비트리거
- **WHEN** 성과 집계 작업이 반복 실패하면
- **THEN** 운영 모드는 변하지 않고 분석 재시도만 수행된다

#### Scenario: 재시작 후 모드 유지
- **WHEN** ENTRY_BLOCKED에서 프로세스가 재시작되면
- **THEN** journal에서 복원되어 EntryGate 투영과 함께 유지된다

#### Scenario: 모드 완화 시도
- **WHEN** ENTRY_BLOCKED에서 NORMAL로 되돌리려 하면
- **THEN** 사람 승인 절차와 audit 기록을 거쳐야 한다

#### Scenario: 대사 지속 실패 시 자동 강화
- **WHEN** reconcile 사이클이 연속 5회 실패하면
- **THEN** 사람 승인 없이 ENTRY_BLOCKED로 전환·영속되고 critical 알림이 발송되며 루프는 재시도를 계속한다

### Requirement: Automation gate limits are the production Guardian policy source
The production `RiskGuardian` policy SHALL be derived from the automation gate's
five configured limits and one normalized currency. The Guardian's maximum
quantity, maximum order notional, maximum open exposure, maximum daily loss, and
maximum daily loss ratio SHALL produce a limit snapshot byte-for-byte equivalent
to the interlock's audited snapshot, including configured/set bits and currency.
The per-trade risk budget SHALL equal the configured maximum daily-loss amount in
the same currency.

#### Scenario: USD policy construction
- **WHEN** the gate contains valid USD limits
- **THEN** every Guardian money field and its audited limit snapshot use USD and the configured values without falling back to KRW defaults

#### Scenario: One configured value differs
- **WHEN** any one of the five policy values or the currency differs from the gate
- **THEN** the startup interlock refuses the Guardian as a single-source mismatch

#### Scenario: Risk budget derivation
- **WHEN** the Guardian policy is created from a valid gate
- **THEN** its per-trade risk budget equals the configured daily-loss amount and grants no larger loss budget

### Requirement: Production Guardian uses the engine journal
All decisions and reservations issued by the production Guardian SHALL be
written through the same `journal.Journal` instance owned and closed by the
engine context. Production assembly SHALL NOT open a second Guardian-only
journal.

#### Scenario: Guardian issuance storage
- **WHEN** the production Guardian issues an allowed decision
- **THEN** the decision and reservations are visible through the engine context's journal and share its account scope

#### Scenario: Context closes shared ownership
- **WHEN** the command-assembled context is closed after production Guardian issuance
- **THEN** the one engine-owned journal handle is closed and no Guardian-only journal remains open

### Requirement: 진입 판정의 관측 기록
진입 체인이 판정을 내릴 때마다 그 판정의 관측 기록이 영속되어야 한다(SHALL) — **ALLOW과 거부 양쪽**. 거부는 현재 어떤 영속 경로도 갖지 않으므로(체인 판정이 메모리 상 발급 거부로만 반환된다) 이 요구사항이 거부 모집단의 첫 기록이다. 범위는 **진입(EXPOSURE_RAISING)에 한정된다**(SHALL — 청산 판정은 stop·target을 갖지 않으므로 전 항목이 결측인 행만 만든다).

기록은 자기완결이어야 한다(SHALL — 분석 쿼리가 결정 계약 테이블을 조인하지 않아도 읽힐 것: trade-analytics "분석 경로의 격리"). 항목의 열거(SHALL): 계좌·시장·심볼, 진입가·손절가·목표가, 실질본전, **총 RR**, **순 RR**, 비용 기준 표기, 비용모델 지문, **정지한 체인 단계**와 reason code, **결과 구분**, 관측 시각, 재구성 여부.

**결과 구분은 필수 열거다**(SHALL): `REFUSED_CHAIN`(체인이 거부) · `ALLOWED_ISSUED`(체인 ALLOW·발급 성공) · `ALLOWED_ISSUANCE_REFUSED`(체인 ALLOW·발급 단계 거부, 그 안정 reason-code 병기). 체인 ALLOW과 발급 성공은 같은 사실이 아니므로(발급은 한도·버전·만료로 별도 거부될 수 있다) 결정 참조의 부재만으로는 세 경우를 구별할 수 없다(SHALL NOT — 참조 null을 거부의 표지로 읽어서는 안 된다).

**정지한 체인 단계는 판정이 직접 보고해야 한다**(SHALL): 하나의 reason code가 여러 단계에서 발생하므로 reason에서 단계를 역산해서는 안 된다(SHALL NOT — 입력 결측 사유는 체인 전역에서 발생한다). 판정 값에 단계를 담는 additive 확장이 이 요구사항의 일부다.

결정 참조는 발급된 판정에만 채워지며 **외래키 제약이어서는 안 된다**(SHALL NOT): 계약 테이블의 정리·만료가 분석 행에 막히거나 전파되어서는 안 되고(같은 이유로 소비된 nonce 기록도 제약을 두지 않는다), 제약은 자기완결 요구와도 충돌한다.

기록은 결정 계약을 건드리지 않는다(SHALL NOT): `decisions`의 preimage와 그 해시, 봉인된 제출 시퀀스는 변경되지 않는다. 스키마는 additive 신규 테이블이며 기존 테이블·열은 변경되지 않는다(SHALL — §0.6). 보존 기간 정책을 갖는다(SHALL — 매 평가마다 쓰이므로 무한 증가해서는 안 된다).

#### Scenario: 거부가 처음으로 영속된다
- **WHEN** 진입 의도가 체인의 어느 단계에서 거부되면
- **THEN** 그 거부의 관측 기록이 reason code·정지 단계·결과 구분 `REFUSED_CHAIN`으로 영속된다

#### Scenario: 체인 ALLOW 후 발급이 거부된다
- **WHEN** 체인이 ALLOW를 냈으나 발급 단계가 한도·버전·만료로 거부하면
- **THEN** 결과 구분 `ALLOWED_ISSUANCE_REFUSED`와 그 안정 reason-code가 기록되고, 거부 판정으로 오기록되지 않는다

#### Scenario: 청산 판정은 기록되지 않는다
- **WHEN** 위험 축소 판정이 평가되면
- **THEN** 진입 관측 기록은 생성되지 않는다

### Requirement: 관측 기록의 결손은 복원 가능해야 하고 거래를 막아서는 안 된다
관측 기록은 결정·예약의 원자 트랜잭션 **밖**에서 수행된다(SHALL). 트랜잭션 안에 두면 관측 쓰기 실패가 결정을 함께 롤백해 체인 ALLOW를 사실상 거부로 바꾸므로, 측정 결함이 거래를 멈추게 된다.

그 대가인 크래시 창은 숨기지 않고 **부분 복원 가능성으로 상계한다**(SHALL): 관측이 없는 발급된 결정은 탐지 가능해야 한다(계약 테이블과의 안티조인). 복원되는 항목과 되지 않는 항목의 열거(SHALL): 가격 3개·시장·계좌·심볼·수량·정책 버전은 결정 preimage에서, 발급 시각은 결정 행에서 복원한다. 결과 구분은 커밋된 진입 결정이 존재한다는 사실로 `ALLOWED_ISSUED`로 정한다. **순 RR·실질본전·비용모델 지문은 복원이 아니라 재구성 시점 모델로 새로 산출하는 값이며 원래 관측값의 복원이라고 표현해서는 안 된다**(SHALL NOT — 당시 요율의 지문이 preimage에 없으므로 원래 값은 원리적으로 복원 불가다). 재구성된 행은 재구성 표지와 **발급 시각·재구성 시각 둘 다**, 그리고 재구성 시점 지문을 담아야 한다(SHALL — 한 행의 시계와 다른 행의 지문을 섞어서는 안 된다). 목표가의 복원 가능성은 최소 RR rung이 목표가 없는 진입을 거부한다는 사실에 의존하므로(preimage 계약 자체는 목표가를 선택 항목으로 둔다) 그 의존을 명시한다(SHALL).

**탐지·재구성 작업은 엔진 런타임의 감독 루프로 등록되어서는 안 된다**(SHALL NOT): 감독 루프의 반환은 전 루프 종료이므로 측정 작업의 오류가 청산 관측·체결 감지를 함께 죽이고(§0.3 위반), 연속 실패 임계는 ENTRY_BLOCKED 강화로 이어지며, 감독 루프의 트리거는 폐쇄 열거라 측정 작업이 대사 실패 트리거를 **차용**해 실패 원인을 오귀속하게 된다. 이 작업은 별개 프로세스·스케줄 경계에서 수행되고(SHALL), 그 실패는 어떤 운영 모드 트리거에도 사상되어서는 안 되며 기존 트리거 이름을 재사용해서도 안 된다(SHALL NOT).

재구성은 **관측 쓰기 데드라인을 초과한 결정에만** 적용된다(SHALL — 진행 중인 쓰기를 결손으로 오인해 재구성하면 실제 쓰기가 뒤이어 착지해 한 결정에 두 행이 남고, 라이브 임계값이 도출될 분포가 이중 계수된다). 발급된 판정의 결정 참조에는 **유일 인덱스**를 둔다(SHALL — 외래키가 아닌 유일성).

탐지·재구성 주기는 결정 행 정리 지평보다 짧아야 하며(SHALL — 결정이 정리되면 결손은 탐지도 복원도 불가능해진다), 재구성 전에 정리되어 영구 손실이 된 건수는 계수되어야 한다(SHALL).

거부 측 기록은 참조할 preimage가 없으므로 결손이 복원 불가하다(SHALL — 이 비대칭을 명시한다). 돈이 걸리지 않으므로 수용하되, 결손 자체는 계수되어야 한다(SHALL).

**관측 기록의 실패는 진입 판정을 바꾸지 않으며 운영 모드·진입 latch에 도달해서도 안 된다**(SHALL NOT): 실패는 critical 알림 등급·durable outbox 경로·미전달 알림에 의한 자동 강화·진입 게이트 차단 중 어느 것에도 진입하지 않는다. 측정 결함이 사람 승인을 요구하는 진입 차단으로 번지는 경로가 있어서는 안 된다(SHALL NOT — 메인 스펙의 "분석·성과 작업 실패는 트리거가 아니다"와 정합).

동시에 조용히 넘어가서도 안 된다(SHALL NOT). 실패는 계수되고 알림은 **`SeverityNormal` 등급**으로 발송된다(SHALL) — 알림의 전달 여부가 거래 가능성을 좌우하지 않으면서도 기록되지 않은 판정의 수를 조회할 수 있어야 한다. 관측 실패 이벤트 타입은 **critical 등급표의 구성원이 되어서는 안 되고**(SHALL NOT — 그 표의 구성원 여부가 진입 게이트 차단과 자동 강화로 이어지는 유일한 구조적 스위치다), 비구성원성을 고정하는 테스트를 둔다(SHALL — 기존 등급표 테스트는 포함만 단언한다).

**열화 계수는 관측 행과 독립된 실패영역에 기록되어야 한다**(SHALL). 관측 테이블 자체에 열화 행을 쓰는 설계는 성립하지 않는다(SHALL NOT — 디스크 소진·I/O 오류·스키마 오류로 관측 INSERT가 실패하면 같은 저장소의 열화 쓰기도 함께 실패하므로 자기모순이다). 독립 저장이 불가능한 상황에서는 구조화 로그와 프로세스 내 단조 카운터로 강등하되 거래 경로는 계속되어야 하고(SHALL), 그 강등된 계수를 재시작 후에도 durable하다고 주장해서는 안 된다(SHALL NOT).

#### Scenario: 관측 실패가 모드를 강화하지 않는다
- **WHEN** 관측 쓰기가 실패하고 그에 따른 알림 전달까지 소진되면
- **THEN** 운영 모드와 진입 latch는 불변이고, 열화 기록과 계수만 남는다

#### Scenario: 결손이 재구성된다
- **WHEN** 발급된 결정에 대응하는 관측이 없는 상태가 탐지되면
- **THEN** preimage에서 결정론적으로 재구성되고, 재구성 표지와 재구성 시점 지문이 함께 기록된다

#### Scenario: 거부 결손은 계수만 된다
- **WHEN** 거부 측 관측 쓰기가 실패하면
- **THEN** 복원되지 않고 결손 계수만 증가하며, 진입 판정은 영향받지 않는다

### Requirement: 순 RR은 관측값이며 게이트 입력이 아니다
순 RR은 `(target − 실질본전) / (실질본전 − stop)`으로 산출한다(SHALL). 실질본전은 손절 계약 rung이 이미 소비하는 값과 동일한 `BreakEvenSellPrice(진입가, 수량 1, 시장)`이어야 한다(SHALL — 하나의 본전 정의만 존재해야 두 소비자가 같은 사실을 말한다).

**최소 RR 게이트의 판정 기준은 총 RR 그대로 유지된다**(SHALL — 이 change는 어떤 의도의 ALLOW/REFUSE도 바꾸지 않는다). 순 RR을 게이트로 승격하는 것은 후속 change의 소관이며, 그 change는 관측된 분포를 근거로 임계값을 정해야 하고 추정으로 고정해서는 안 된다(SHALL NOT).

지표의 이름은 **수수료·세금 차감 후 RR**이다(SHALL — 슬리피지는 포함되지 않으므로 "비용 차감 후"라고 부르지 않는다). 비용 범위는 기록에 표기되어야 하며(SHALL), 범위가 넓어지면 표기로 구분 가능해야 한다.

관측 분포의 두 구조적 한계를 명시한다(SHALL). ① **좌측 절단**: 손절 계약 rung이 본전 미달 목표가를 이미 거부하므로 어떤 ALLOW 관측에도 순 RR ≤ 0이 나타나지 않는다 — 절단을 분포의 성질로 오독해서는 안 된다(SHALL NOT). ② **정밀도**: 실질본전이 비용 모델의 부동소수점 산술에서 오므로 순 RR 관측값은 그 상대오차를 물려받는다. 순 RR을 게이트로 승격하는 change는 판정 경계가 가격의 이진 전개에 좌우되지 않도록 정밀도를 먼저 해결해야 한다(SHALL — 이 패키지가 유리수 산술을 쓰는 이유가 그것이다).

#### Scenario: 순 RR이 낮아도 판정은 변하지 않는다
- **WHEN** 총 RR은 임계값을 넘지만 순 RR이 임계값 미만인 의도가 평가되면
- **THEN** 판정은 종전과 동일하게 ALLOW이고, 두 값의 차이가 관측 기록에 남는다

#### Scenario: 본전 정의의 단일성
- **WHEN** 순 RR 관측이 산출되면
- **THEN** 사용된 실질본전은 손절 계약 rung이 목표가를 비교한 값과 동일하다

#### Scenario: 비용 범위 표기
- **WHEN** 관측 기록이 조회되면
- **THEN** 그 순 RR이 수수료·세금만 차감한 값임이 기록 자체로 식별된다

### Requirement: 티어 상한의 사정거리

등록된 티어의 필드별 최대(이하 **상한**)는 콘솔 쓰기 경로 전용의 백스톱이며 기동 인터록의 판정에 참여하지 않는다(SHALL — 인터록은 다섯 값의 양수·유한·비율 (0,1]·통화 비지 않음만 보고 상한을 보지 않는다). 두 판정은 서로를 대체하지 않는다(SHALL NOT — 상한 통과를 인터록 통과로, 인터록 통과를 상한 통과로 읽지 않는다). 상한을 낮게 유지하는 것으로 시스템 전체의 한도를 낮췄다고 결론내서는 안 된다(SHALL NOT — 상한 아래로 표현할 수 없는 정당한 운영 요구가 있으면 운영자는 설정 파일을 직접 편집하게 되고, 그 경로에는 상한 검사가 없다; 백스톱을 우회하게 만드는 상한은 백스톱이 아니다).

#### Scenario: 상한과 인터록은 다른 질문이다

- **WHEN** 등록된 상한을 넘지만 다섯 값이 전부 양수·유한하고 비율이 (0,1]이며 통화가 있는 블록을 검사하면
- **THEN** 콘솔 저장은 거부되고 기동 인터록 규칙은 통과한다

#### Scenario: 티어 추가가 인터록을 바꾸지 않는다

- **WHEN** 레지스트리에 티어를 추가하면
- **THEN** 그 통화의 상한만 이동하고 기동 인터록이 수용·거부하는 블록의 집합은 변하지 않는다

### Requirement: KR과 US는 하나의 account-base Guardian 권위를 공유한다

Production strategy runtime은 계좌마다 정확히 하나의 account-base-currency Guardian을 사용해야 한다 (SHALL).
그 Guardian은 하나의 account-wide exposure/daily-loss 권위를 소유한다. KR과 US에 quote-currency별
Guardian 또는 독립 account-wide cap을 만들어서는 안 된다 (MUST NOT). `limit_currency`는
account base currency를 뜻하고, 주문 notional·총 노출·일일 손실·equity·risk budget은 그 base
minor unit으로 비교해야 한다 (SHALL). Market cash와 broker order cost는 market quote minor
unit으로 비교해야 하며 base와 quote 숫자를 직접 비교해서는 안 된다 (MUST NOT).

각 exposure-raising 요청은 official source가 mint한 request-scoped frozen FX authority를 사용해야
한다 (SHALL). Authority는 base/quote pair, exact decimal rate, conservative haircut,
source/version/digest, observed-at/fresh-until을 봉인해야 하며 (SHALL), caller가 제공한 raw rate나
digest를 authority로 승격해서는 안 된다 (MUST NOT). KR same-currency path도 봉인된 identity FX를
사용하고 US path는 official quote-to-base FX를 사용해야 한다 (SHALL). 같은 authority가 Guardian
sizing, aggregate와 다섯 bucket reservation, decision limits envelope와 Gateway last-moment
validation에 끝까지 사용되어야 하며 중간 재조회나 다른 환율 사용은 금지한다 (MUST NOT).

Base reservation은 exposure를 작게 계산하지 않도록 ceil하고 admissible quantity는 floor해야 한다
(SHALL). FX 없음, stale, 역방향, market/quote/base/digest 불일치 또는 완화된 haircut은 해당
시장의 신규 exposure만 거부하고 (SHALL), protection, fill, reconciliation과 reduce-only exit를
차단해서는 안 된다 (MUST NOT).

#### Scenario: KR identity와 US official conversion의 동시 admission

- **WHEN** 같은 account-base Guardian에 KR identity-FX 요청과 US official quote-to-base 요청이 동시에 들어온다
- **THEN** 두 요청은 commit 순서와 무관하게 하나의 base exposure 잔여를 공유하며 합산 cap을 초과하지 않고 어느 시장도 peer 안정화를 기다리지 않는다

#### Scenario: US FX 만료

- **WHEN** US q_final precheck에 사용한 official FX가 issue 또는 Gateway 직전 만료된다
- **THEN** US exposure-raising authority는 broker request와 부분 reservation 없이 거부되고 KR eligible work와 양 시장 safety lifecycle은 계속된다

#### Scenario: quote cash와 base limit 분리

- **WHEN** US 주문 비용은 USD이고 Guardian exposure limit은 account base currency다
- **THEN** cash는 USD order cost와, exposure는 동일 frozen FX로 환산한 base amount와 비교되며 USD 숫자를 base limit 숫자와 직접 비교하지 않는다

#### Scenario: cross-currency concurrent cap race

- **WHEN** KR과 US 요청이 동시에 마지막 account-wide capacity를 예약하려 한다
- **THEN** aggregate와 다섯 bucket의 base reservation transaction은 합산 cap 안의 요청만 commit하고 loser는 bounded fresh recollection 또는 atomic refusal이며 중복 권한은 0건이다

#### Scenario: FX outage 중 exit

- **WHEN** entry용 official FX authority를 만들 수 없지만 기존 US 포지션의 reduce-only exit가 필요하다
- **THEN** 신규 US entry만 닫히고 exit, protection, reconciliation과 fill 처리는 계속된다
