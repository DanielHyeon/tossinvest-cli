# Tasks: verify-observes-the-trigger

위험도 **High-risk**. 이 도구가 처음으로 체결을 의도한 라이브 주문을 만든다.
1.x는 전부 RED 선행이고, 라이브 실행(3.1)은 앞의 전부가 끝난 뒤 사람 승인으로만 한다.

## 1. 도구 — 체결을 의도한 주문을 담을 수 있게 만든다

### 1.1 넘길 의도의 임계

- [x] 1.1 RED — `NearStopTrigger` 경계: `bid < trigger <= last` 위반, 두 값 사이에 유효 틱
  없음, `last<=0`, `bid<=0`, KR 틱 격자, US 1달러 경계(0.0001↔0.01). 각각 에러 또는 격자 위 값
- [x] 1.2 GREEN — `NearStopTrigger` 구현. **`Far*` 함수군은 서명·본문 무변경**
- [x] 1.3 정적 가드 — `FarBuyLimit`/`FarSellLimit`/`FarStopTrigger`가 방향·근접 인자를 받지
  않고, `NearStopTrigger`를 호출하는 곳이 발동 단계 하나임을 AST로 고정 (D1)

### 1.2 체결을 종결 상태로

- [x] 1.4 RED — 체결된 artifact가 `Outstanding`·`liveCount`·정리 대상·`verify status`
  잔여물 어디에도 남지 않는다. 취소가 아니라 체결로 기록된다
- [x] 1.5 RED — 단조성: 체결 뒤에 쓰인 줄이 객체를 되살리지 못한다 (M22가 취소에서 겪은 형태)
- [x] 1.6 RED — **기존 기록 회귀**: `Filled` 필드가 없는 줄의 판정이 오늘과 동일
- [x] 1.7 GREEN — `Artifact.Filled`/`FilledAt` 가산, `Outstanding` 종결 판정 확장,
  `outstandingLines`의 역행 가드 확장. `FormatVersion` 무변경

### 1.3 옵트인 게이트

- [x] 1.8 RED — `--include-trigger` 없이 실행하면 발동 단계가 오늘과 **바이트 단위로 같은**
  미검증 관측을 남기고, `NearStopTrigger`가 호출되지 않는다 (호가조차 읽지 않는 것으로 강화)
- [x] 1.9 GREEN — `FlagIncludeTrigger` + `Options.IncludeTrigger` + `optedIn` 배선.
  카탈로그의 `Deferred`는 옵트인일 때만 걷힌다

### 1.4 관측

- [x] 1.10 RED — 2단 폴링. **초안 수정**: 대기 국면도 시세와 조건주문을 함께 읽는다 —
  호가 기준이면 임계 도달이 시세에 나타나기 전에 발동하므로 시세만 보면 그 경우를 놓친다.
  간격은 무슨 일이 생기는 순간 조여진다
- [x] 1.11 RED — 네 시각(`condition_crossed_at`, `trigger_first_observed_at`,
  `triggered_order_id_first_seen_at`, `child_order_filled_at`)이 **각각 그때의 폴링 간격과
  함께** 기록된다. 창 안에서 백오프가 걸리면 그 사실도 기록된다
- [x] 1.12 RED — 발동 최초 관측 시점의 bid/ask/last가 기록된다. **초안 수정**: 기준 판정은
  지연 비교가 아니라 **두 관측의 순서**다 — 순서는 거친 폴링 간격에서도 살아남는다
- [x] 1.13 RED — child 주문이 `HeldUntil=conditional-trigger` + 부모의 `ChainID`로 붙잡히고,
  체결 확인 시 `Filled`로 종결된다
- [x] 1.14 GREEN — `stepConditionalTrigger` 본문. SINGLE+MARKET+SELL+1주 고정

### 1.5 결말

- [x] 1.15 RED — 임계 미도달 창 만료: 단계가 **자기가 등록한** 조건주문을 자기 창 안에서
  취소하고 `skipped(INCONCLUSIVE)`로 끝난다. 계좌 잔여물 0
- [x] 1.16 RED — **경합**: 취소 직후 재확인. **초안 수정**: 취소된 조건주문은 다시 읽히지
  않으므로(이미 측정된 `conditional.cancel.gone_after`) 재확인만으로는 가릴 수 없다 —
  최종 근거는 **보유 수량**이고, 그것마저 못 읽으면 `skip`이 아니라 `fail`이다
- [x] 1.17 RED — 발동은 봤으나 창 안에 체결 미확인: `fail`, child가 붙잡힌 채 보고된다
- [x] 1.18 GREEN — 결말 처리. **여섯 번째 결말 추가**(리뷰 A3): 임계에 도달했는데 발동하지
  않으면 `skipped`가 아니라 `fail` + `conditional.fires_when_its_condition_is_met=false`

### 1.6 운영자 종료 경로 (I1)

- [x] 1.19 RED — `verify abort`가 취소 대상을 먼저 나열하고, 대상이 **기록상 outstanding인
  것으로 한정**되며, 종결이 이유와 함께 기록된다
- [x] 1.20 RED — 시각 경과만으로는 아무것도 취소되지 않는다 (D3 거부의 재확인)
- [x] 1.21 GREEN — `tossctl verify abort`. **콘솔에 새 타이핑 확인·추가 승인 마찰을 넣지 않는다**

### 1.7 노출 상한 (I2 → ③)

- [x] 1.22 RED — 발동 단계는 조건주문 1건이 이미 살아 있어도 자기 것을 등록할 수 있고,
  같은 상태에서 `conditional-register`는 `ErrExposureCap`으로 막힌다
- [x] 1.23 GREEN — `MaxLiveConditionalsTrigger`. **`MaxLiveOrders`·`checkOrderCap`은
  무변경** (design D6 — child는 접수가 아니라 발견이다)

### 1.8 유예형과 옵트인의 상호작용 (구현 중 발견)

- [x] 1.24 RED — 옵트인 없는 발동 단계는 승인 목록(`Plan`)에 **줄이 오르지 않고**,
  기록 줄의 `mutating`이 오늘처럼 false다
- [x] 1.25 GREEN — `deferredForm`/`mutatesNow` 도입, `preflightStatic`·`preflight`·
  `Plan`·`entryFor` 배선. `Deferred` 조기 반환이 옵트인 게이트를 건너뛰지 않게 한다
- [x] 1.26 RED — `sweepStep`이 `Deliberate`로 표시된 객체를 취소하지 않는다 (결말 ②가
  "붙잡힌 채 보고된다"를 지키려면 필요; 기존 단계에는 무영향)
- [x] 1.27 GREEN — `sweepStep` 스킵 가산

### 1.9 적대적 리뷰가 연 것 (review.md A1·A2·A4·A5)

- [x] 1.28 GREEN — 발동 단계의 조건주문은 `markHeld`가 아니라 `joinChain`. 중단되면
  실행이 시끄럽게 실패하고 다음 실행의 prologue가 취소를 승인 목록에 올린다 (A1)
- [x] 1.29 GREEN — `liveConditional()`이 출처로 가른다. 조건주문이 둘일 때
  `conditional-cancel`이 발동 측정의 대상을 취소하던 결함 (A2)
- [x] 1.30 GREEN — `fail` 경로에서 조건주문을 종결로 **단정하지 않는다** (A4)
- [x] 1.31 GREEN — 발동 관측 후 시세·호가 폴링 중단 — 창 안 요청량 tick당 4 → 1~2 (A5)

## 2. 검증

- [x] 2.1 Function Logic Map — `analysis/function-logic-map.md`. 기존 함수 내부를 고치는
  전 대상 + Branch Test Map. 계약과 어긋난 것 2건을 여기서 발견해 계약을 고쳤다
- [x] 2.2 Pre-Edit 선언 기록 (High-risk, review.md)
- [x] 2.3 **실기록 재생 회귀** — `capability-verify.jsonl` 61 entry / artifact 줄 16개,
  `capability-verify-us.jsonl` 34 entry / 18개. 전 줄이 변경 전 규칙과 동일하게 분류,
  outstanding 0, 정리 대상 0. 스냅샷이 아니라 옛 규칙을 테스트 안에 복제한 A/B다
- [x] 2.4 `make test` / `make vet` / `make validate` — 3806 passed, 상속 테스트 회귀 0
- [x] 2.5 review.md — 적대적 Eng 리뷰. P0 2건·P1 2건·P2 1건을 찾아 전부 구현으로 닫았다
- [x] 2.6 `make sdd-sync` → `make sdd-check` → `make gate`. **CodeGraph hard evidence는 신선**
  (`sdd-check`: "CodeGraph hard-evidence index matches the worktree", exit 0).
  `sdd-sync`는 GBrain에서 non-zero로 끝난다 — PGLite lock 대기 timeout이고 GBrain은 advisory라
  현재 HEAD·OpenSpec·테스트를 대체하지 않는다(`.claude/CLAUDE.md`). gate는 3.x가 남아 있으므로
  **아직 통과하지 않는다** — 그것이 정확한 상태다: 측정이 아직 실행되지 않았다
- [x] 2.7 PM 등록 — `docs/pm/portfolio/_registry.yaml` + `tools/pm/test_generate_master_tracker.py`

## 3. 실측 — 사람 승인 없이 실행하지 않는다

- [ ] 3.1 **사용자 확인**: `engine.adoption.exclude_symbols`에 측정 종목 고정 (설정 변경이므로
  사용자가 한다, §0.7)
- [ ] 3.2 측정 전 soak·candidate watch 정지 (폴링 예산 확보)
- [ ] 3.3 `--list`로 전 요청 목록 확인 → 사람 입회 하 US 1회 실행 (TSLA 1주, SINGLE+MARKET)
- [ ] 3.4 결과를 `verify-execution-capability/measurements.md`에 기록
- [ ] 3.5 계좌 잔여물 0 확인, 남은 미측정 갱신

**3.x 이후에만** `ProtectiveCapability` 산출(별도 change)과 2c-A GREEN 배선이 열린다.
