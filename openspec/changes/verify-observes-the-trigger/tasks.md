# Tasks: verify-observes-the-trigger

위험도 **High-risk**. 이 도구가 처음으로 체결을 의도한 라이브 주문을 만든다.
1.x는 전부 RED 선행이고, 라이브 실행(3.1)은 앞의 전부가 끝난 뒤 사람 승인으로만 한다.

## 1. 도구 — 체결을 의도한 주문을 담을 수 있게 만든다

### 1.1 넘길 의도의 임계

- [ ] 1.1 RED — `NearStopTrigger` 경계: `bid < trigger <= last` 위반, 두 값 사이에 유효 틱
  없음, `last<=0`, `bid<=0`, KR 틱 격자, US 1달러 경계(0.0001↔0.01). 각각 에러 또는 격자 위 값
- [ ] 1.2 GREEN — `NearStopTrigger` 구현. **`Far*` 함수군은 서명·본문 무변경**
- [ ] 1.3 정적 가드 — `FarBuyLimit`/`FarSellLimit`/`FarStopTrigger`가 방향·근접 인자를 받지
  않고, `NearStopTrigger`를 호출하는 곳이 발동 단계 하나임을 AST로 고정 (D1)

### 1.2 체결을 종결 상태로

- [ ] 1.4 RED — 체결된 artifact가 `Outstanding`·`liveCount`·정리 대상·`verify status`
  잔여물 어디에도 남지 않는다. 취소가 아니라 체결로 기록된다
- [ ] 1.5 RED — 단조성: 체결 뒤에 쓰인 줄이 객체를 되살리지 못한다 (M22가 취소에서 겪은 형태)
- [ ] 1.6 RED — **기존 기록 회귀**: `Filled` 필드가 없는 줄의 판정이 오늘과 동일
- [ ] 1.7 GREEN — `Artifact.Filled`/`FilledAt` 가산, `Outstanding` 종결 판정 확장,
  `outstandingLines`의 역행 가드 확장. `FormatVersion` 무변경

### 1.3 옵트인 게이트

- [ ] 1.8 RED — `--include-trigger` 없이 실행하면 발동 단계가 오늘과 **바이트 단위로 같은**
  미검증 관측을 남기고, `NearStopTrigger`가 호출되지 않는다
- [ ] 1.9 GREEN — `FlagIncludeTrigger` + `Options.IncludeTrigger` + `optedIn` 배선.
  카탈로그의 `Deferred`는 옵트인일 때만 걷힌다

### 1.4 관측

- [ ] 1.10 RED — 2단 폴링: 대기 국면은 시세만, 관측 국면은 시세+조건주문+주문. 국면 전환이
  임계 도달 관측에서 일어난다
- [ ] 1.11 RED — 네 시각(`condition_crossed_at`, `trigger_first_observed_at`,
  `triggered_order_id_first_seen_at`, `child_order_filled_at`)이 **각각 그때의 폴링 간격과
  함께** 기록된다. 창 안에서 백오프가 걸리면 그 사실도 기록된다
- [ ] 1.12 RED — 발동 최초 관측 시점의 bid/ask/last가 기록된다 (`trigger_price_basis` 근거)
- [ ] 1.13 RED — child 주문이 `HeldUntil=conditional-trigger` + 부모의 `ChainID`로 붙잡히고,
  체결 확인 시 `Filled`로 종결된다
- [ ] 1.14 GREEN — `stepConditionalTrigger` 본문. SINGLE+MARKET+SELL+1주 고정

### 1.5 결말

- [ ] 1.15 RED — 임계 미도달 창 만료: 단계가 **자기가 등록한** 조건주문을 자기 창 안에서
  취소하고 `skipped(INCONCLUSIVE)`로 끝난다. 계좌 잔여물 0
- [ ] 1.16 RED — **경합**: 취소 직후 재확인에서 발동 흔적(발동 상태 또는 `triggeredOrderId`)이
  보이면 미도달로 끝내지 않고 관측 국면으로 전환한다
- [ ] 1.17 RED — 발동은 봤으나 창 안에 체결 미확인: `fail`, child가 붙잡힌 채 보고된다
- [ ] 1.18 GREEN — 다섯 결말 처리 (design D4 표)

### 1.6 운영자 종료 경로 (I1)

- [ ] 1.19 RED — `verify abort`가 취소 대상을 먼저 나열하고, 대상이 **기록상 outstanding인
  것으로 한정**되며, 종결이 이유와 함께 기록된다
- [ ] 1.20 RED — 시각 경과만으로는 아무것도 취소되지 않는다 (D3 거부의 재확인)
- [ ] 1.21 GREEN — `tossctl verify abort`. **콘솔에 새 타이핑 확인·추가 승인 마찰을 넣지 않는다**

### 1.7 노출 상한 (I2 → ③)

- [ ] 1.22 RED — 발동 단계 창 안에서는 전용 상한이 적용되고, 창 밖 단계는 `MaxLiveOrders`
  그대로다
- [ ] 1.23 GREEN — `MaxLiveOrdersTrigger`

## 2. 검증

- [ ] 2.1 Function Logic Map — 기존 함수 내부를 고치는 전 대상
  (`Outstanding`/`outstandingLines`/`cleanupFrom`/`liveCount`/`preflightStatic`/`optedIn`/
  `stepConditionalTrigger`) + Branch Test Map
- [ ] 2.2 Pre-Edit 선언 기록 (High-risk, review.md)
- [ ] 2.3 **실기록 재생 회귀** — `capability-verify.jsonl`(KR)·`capability-verify-us.jsonl`(US)를
  새 코드로 재생해 outstanding·pendingCleanup이 오늘과 같음을 확인
- [ ] 2.4 `make test` / `make vet` / `make validate` — 상속 테스트 회귀 0
- [ ] 2.5 review.md — 적대적 Eng 리뷰 (High-risk 필수)
- [ ] 2.6 `make sdd-sync` → `make sdd-check` → `make gate CHANGE=verify-observes-the-trigger`
- [ ] 2.7 PM 등록 — `docs/pm/portfolio/_registry.yaml` + `tools/pm/test_generate_master_tracker.py`

## 3. 실측 — 사람 승인 없이 실행하지 않는다

- [ ] 3.1 **사용자 확인**: `engine.adoption.exclude_symbols`에 측정 종목 고정 (설정 변경이므로
  사용자가 한다, §0.7)
- [ ] 3.2 측정 전 soak·candidate watch 정지 (폴링 예산 확보)
- [ ] 3.3 `--list`로 전 요청 목록 확인 → 사람 입회 하 US 1회 실행 (TSLA 1주, SINGLE+MARKET)
- [ ] 3.4 결과를 `verify-execution-capability/measurements.md`에 기록
- [ ] 3.5 계좌 잔여물 0 확인, 남은 미측정 갱신

**3.x 이후에만** `ProtectiveCapability` 산출(별도 change)과 2c-A GREEN 배선이 열린다.
