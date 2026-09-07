# a065 — 독립 적대 리뷰 결과와 해소 기록

- 일자: 2026-09-07
- 대상 브랜치: `feat/a112-four-family-runtime`, 리뷰 시작 HEAD `f90c25c8`
- a065 코드 커밋: `75cb371a` (2026-08-04). 리뷰 시작 시점까지 `internal/positioncampaign`
  전체와 `internal/journal/position_campaign.go` 는 변경 0.
- 리뷰 축 넷: D4 전이표/D8 재구성, 원장 원자성, EXIT FIRST/손절 단조성, 휴면·주장 검증
- 최초 판정: **넷 모두 BLOCK**

## 0. 전제 정정 — a065 코어는 휴면이 아니다

`status.md` 와 `review.md` 는 "`CampaignApplierBound()` 는 production 에서 false",
"모든 진입 caller 와 live dispatch 는 연결되지 않았다" 고 적었다. 리뷰 시점 HEAD 에서
둘 다 거짓이다.

- `internal/app/engine/gateway.go:188` — 저장소의 **유일한** 비테스트 `SetApplyHooks`
  호출이 `Campaign: journal.ApplyPositionCampaignFill` 을 건다.
  `internal/app/engine/engine.go` 가 기동 때 무조건 부른다.
- `internal/journal/strategy_dispatch_runtime.go:1184` (a072) 가 첫 leg 진입 경로에서
  `campaignExposureBlockedInTx` 를 부르고 `campaign_order_watermarks` 를 쓴다.
- a065 가 그 휴면 주장을 지키라고 넣은 가드는 a072(`8022f578`)가 이름을 바꾸고
  **뒤집었다** — 이제 "배선되어 있어야 한다" 고 단언한다.

측정으로 확인한 live/비live 경계:

| 진입점 | 비테스트 호출자 |
|---|---|
| `ApplyPositionCampaignFill` | **1** (engine gateway, 무조건) |
| `campaignExposureBlockedInTx` | **3** (그중 a072 first-leg 경로가 live) |
| `(*Journal).LinkCampaignOrder` | 0 |
| `(*Journal).PlanCampaignLeg` | 0 |
| `(*Journal).UpdateCampaignStop` | 0 |

따라서 아래 P0 중 I-4 는 **오늘은 생산에서 도달 불가**(래치를 거는 유일한 writer 인
`UpdateCampaignStop` 에 비테스트 호출자가 0)이고, I-5·I-8 은 **live 경로**다.

## 1. 뿌리는 하나 — 같은 규칙이 두 자리에 있었다

거의 모든 발견이 같은 기전이다. 규칙 하나가 두 자리에 서로 다른 방법으로 구현돼 있고,
시험은 자기가 보고 쓴 쪽을 통과시킨다. 그래서 둘 다 변이에 살아남았다.

| 규칙 | 구현 A (원장) | 구현 B (도메인/재구성) | 해소 |
|---|---|---|---|
| leg 상태 | `ApplyPositionCampaignFill` 안 `if` | `TransitionLeg` 표 | `positioncampaign.LegStateAfterFill` 하나 |
| `entry_blocked` | 체결이 상태에서 재계산 | `UpdateCampaignStop` 의 단조 래치 | `positioncampaign.LatchEntryBlocked` 하나 |
| successor 잔량 | cap 기준 (`campaignRemaining`) | leg 잔여 기준 (`LegLedger.Observe`) | `positioncampaign.StoredOrderRemaining` 하나 |
| cap 잔여 산술 | `campaignRemaining` | `remainingFromCap` | `positioncampaign.RemainingFromCap` 하나 |

새 파일 `internal/positioncampaign/fill.go` 가 그 판정들의 정본이다.

## 2. 해소한 항목

각 항목은 (a) 행동 시험과 (b) 판정이 한 곳인지 보는 AST 시험을 갖고, (c) 수정을
되돌리는 뮤테이션으로 **RED 를 확인**했다. 원복은 전부 바이트 동일로 검증했다.

### P0

**I-1 — 교체가 있는 campaign 이 offline 재구성되지 않았다**
`replay.go` 가 같은 leg 에서 `AttemptID` 변경을 `LEG_IDENTITY_CHANGED` 로 봤다. D5 는
교체가 **같은 leg 위에 새 order/attempt** 를 만든다고 규정하고, 스키마도 `attempt_id` 를
`campaign_order_watermarks` 에만 둔다. leg 수준 attempt 불변은 replay 가 지어낸 것이다.
→ leg 에서 그 절을 없애고, attempt 불변을 **order 수준**으로 옮겼다(`replayOrder.attempt`,
위반 시 `ORDER_IDENTITY_CHANGED`). 시험 `TestCampaignWithReplacementReconstructsFromEvidence`.

**I-4 — 평범한 체결 하나가 손절 fail-closed 래치를 지웠다**
`entry_blocked` 를 campaign 상태만으로 되계산해 저장 열을 덮어썼다. 상태에 encode 되지
않는 유일한 래치 출처가 D7 의 그것이다. spec `position-campaigns` 위반: "stop evidence 가
missing 또는 invalid … 새 exposure-raising leg 는 fail closed 된다".
→ `LatchEntryBlocked(저장값, 전이)` 로 단조화하고, 원장·재구성·두 링크 경로가 모두 같은
함수를 부른다. 시험 `TestCampaignFillDoesNotClearTheStopEntryLatch` (열 값이 아니라
**다음 leg 가 실제로 거부되는지**까지 본다) + 구조 시험
`TestEveryEntryBlockedWriterGoesThroughTheLatch`.

**I-5 — Position CLOSED 가 새 노출을 막지 않았다**
`campaignExposureBlockedInTx` 가 `state='CLOSING'` 만 봤다. 청산 완료는 `instance_seq` 를
바꾸지 않으므로 체결 경로의 generation 검사에도 걸리지 않는다.
→ campaign 이 **묶인** generation(`actual_position_generation`)의 행이 CLOSED 면 차단한다.
판정을 묶인 generation 으로 한정한 것이 핵심이다 — "최신 행이 CLOSED" 로 막으면 spec 의
"청산 후 재진입" 시나리오가 영구히 불가능해진다. 시험 둘:
`TestClosedBoundGenerationBlocksTheNextEntryLeg`(막는다)와
`TestReEntryAfterCloseIsAdmittedOnANewCampaign`(정상 입력은 막지 않는다).

### P1

**I-2/I-3 — 체결+잔량취소가 한 관측으로 올 때 원장과 재구성이 갈렸다**
재구성이 `delta > 0` 을 `OrderTerminal` 보다 먼저 봐서 `RESIDUAL_CANCELLED` 를 영원히
유도하지 못했고, D4 leg 표에는 `SUBMITTED + RESIDUAL_CANCELLED` 행 자체가 없었다
("첫 체결이 곧 마지막 체결" 인 경우를 놓쳤다).
→ 분류 순서를 공용 함수 하나로 고정하고 표에 그 행을 넣었다. 시험
`TestResidualCancelInOneObservationAgreesWithReconstruction`.

**I-6 — `updateCampaignSuccessorRemaining` 이 증명 가능한 no-op 이었다**
세 writer 전부 그 주문의 `max(0, cap − 자기 누적)` 을 썼다. predecessor 의 늦은 체결은
두 피연산자 어디에도 들어가지 않는다(본문을 `return nil` 로 바꿔도 스위트가 통과했다).
D5 가 요구하는 재계산은 실제로 일어나지 않았고, 도메인 모델(`watermark.go`)만 leg 잔여
기준으로 계산하되 아무도 그것을 저장하지 않았다.
→ `StoredOrderRemaining(hasPredecessor, terminal, cap, 누적, leg 잔여)` 하나로 통일했다.
값은 두 상한의 **작은 쪽**이다. 커밋된 시험이 못 박던 cap 기준 `"4"` 를 `"3"` 으로
정정하고, 늦은 체결 **직전/직후**를 둘 다 재서 재계산이 실제로 일어남을 보인다.

**I-7 — EXIT FIRST 거절이 노출이 아니라 장부만 거절했다**
`LinkCampaignOrder` 는 attempt 가 이미 CONFIRMED 이고 broker order id 를 가질 것을
요구한다. 그래서 그 시점의 거절은 이미 나간 주문을 되돌리지 못하고 기록만 막았고,
형제 거절들과 달리 `latchCampaignLinkConflict` 를 부르지 않아 **그 주문의 실제 체결이
원장에서 영구히 보이지 않았다**.
→ 형제 거절과 같은 자리에 세웠다: RECONCILE/entry-block 으로 격리하고
`ORDER_LINK_REFUSED` 를 남긴 뒤 commit 한다. 거절은 그대로 거절이다.
시험 `TestExposureRefusalAtLinkLeavesRefusalEvidence`.

**I-8 — 미해결 risk-reducing 판정에 시간 경계가 없었다**
그 종목의 **모든 시점** SELL intent 를 훑었다. 해소 경로가 없는 모양이 둘 있다(attempt 가
아예 없는 intent, terminal fill snapshot 이 끝내 오지 않는 CONFIRMED 주문 — 거절 관측은
`terminal=0, fail_closed=1` 로 영원히 남는다). 2020 년의 죽은 intent 하나가 2026 년
campaign 의 첫 leg 를 막았다.
→ 현재 position generation 이 열린 뒤(`positions.opened_at`)의 intent 만 본다. 경계를
모르면(행 없음·`opened_at` 빈 값) `coalesce(...,'')` 로 예전처럼 전 구간을 훑는다.
시험 둘: 옛 intent 는 막지 않고(`…DoesNotBlockForever`), 현재 generation 의 미해결
intent 는 여전히 막는다(`…InTheLiveGenerationStillBlocks` = 양성 대조군).

**I-10 — 전략 중립 가드가 토큰 존재 검사라 아무것도 막지 못했다**
금지 부분문자열 넷 중 하나(`internal/broker`)는 존재하지 않는 디렉터리를 이름했고,
무엇보다 **철자를 고르면 통과했다** — 리뷰가 그 리터럴을 하나도 쓰지 않은 진짜 7-leg
비율표와 시장별 cap 을 붙였는데 스위트가 초록이었다.
→ 판정을 역할로 바꿨다. (a) import **폐포** 전체를 걸어 `internal/riskcalc` 외 어떤
패키지도, 어떤 제3자 모듈도 들어오지 못하게 한다(전이 의존까지 잡는다). (b) 이 패키지가
담은 모든 숫자 리터럴(10진 문자열 포함)을 **측정해서** 얼렸다: `0`, `1`, `128`, `"0"`.
비율표는 숫자 없이 쓸 수 없으므로 철자와 무관하게 걸린다. 두 시험 모두 리뷰가 쓴
그 반증(비율표, 전이 의존)으로 RED 를 확인했다.

**I-9 — review.md 가 존재하지 않는 시험의 PASS 를 보고했다**
→ `review.md` 에서 그 줄을 정정했다(아래 3절).

## 3. 문서 정정

- `status.md` — 휴면·미배선 주장을 삭제하고 측정한 live 경계로 교체했다.
- `review.md` — 존재하지 않는 시험의 PASS 줄을 정정하고, 6.4 독립 리뷰 결과와 해소를 기록했다.
- `design.md` — D4 leg 표에 `SUBMITTED + RESIDUAL_CANCELLED` 행을 넣고, D5 의 successor
  remaining 정의를 실제 구현과 일치시켰다.
- `tasks.md` — 6.3/6.4 완료.

## 4. 남은 부채 (해소하지 않음 — 이유를 적는다)

- **D-1. D4 의 "ACTIVE | Position CLOSED | CLOSED" 상태 전이 자체**
  이번에 넣은 것은 admission 차단(spec 의 MUST NOT 절)이다. campaign 을 실제로 CLOSED 로
  옮기려면 position 수명주기를 관측하는 **writer 가 새로 필요**하고, 그 경로는 a065 의
  writer 범위 밖이다. 그 결과 청산 뒤에도 옛 campaign 이 ACTIVE 로 남아
  `idx_position_campaign_active_scope` 가 새 campaign 생성을 막는 liveness 문제가 남는다.
- **D-2. `EXIT_FIRST_BLOCKED` 거절 코드**
  spec 시나리오가 이름하지만 `campaign_commands.result_error` 의 CHECK 가
  `'INVALID_IDENTITY'` 하나만 허용한다. 별도 값으로 저장하려면 스키마 migration 이
  필요하고 그것은 a065 의 v20 소유 범위 밖이다. 지금은 기존 refusal 코드를 쓴다.
- **D-3. a072 first-leg 경로의 같은 거절 구멍**
  `linkConfirmedStrategyCampaignTx` 도 `ErrExposureBlocked` 를 흔적 없이 반환한다.
  거기서 래치하려면 caller 의 트랜잭션이 롤백하므로 a072 의 트랜잭션 설계를 바꿔야 한다.
- **D-4. 이전 generation 에서 걸어 둔 SELL 주문이 broker 에 살아 있는 경우**
  I-8 의 경계는 그것을 더 보지 않는다. 그 위험은 주문 수명 대사의 일이며, 경계 없는
  판본도 실질적으로는 막지 못했다(모든 진입을 영구히 막아 우회를 부른다).
- **D-5. P2 목록** — 버전 CAS 다섯 자리가 `RowsAffected` 를 보지 않음(현재
  `_txlock=immediate` 때문에 경합으로 도달 불가, 잠복), 종결 자동 close 가
  `TransitionCampaign` 을 우회, `lineage_ambiguous`/`carry_baseline` 이 증명 가능하게
  비활성, campaign reconcile scope 는 계좌 전체인데 projection 은 종목 범위.

## 5. 반증된 가설 (깨끗한 축)

- 최종 UPDATE 의 zero-row 도달 — `_txlock=immediate` 가 모든 트랜잭션을 직렬화하고
  트랜잭션 안에서 `position_campaigns.version` 을 움직이는 문장이 없다. 잠복이지 경합 아님.
- 원시 `tx.tx` 가 apply 가드를 우회 — 디렉터리의 모든 비테스트 파일을 훑는 시험이 있고
  `live()` 는 hook 중간에 뒤집히지 않는다.
- append-only 규칙 이중화 — 트리거가 `campaign_commands`/`campaign_events` 의 유일한 집행자.
- 손절 비교의 십진 정확성 — `riskcalc` 문자열 산술이고, 빈 값·NULL·`0`·비정상 값 어느
  것도 저장된 stop 을 낮추지 못한다.
- campaign 코어가 SELL 체결·청산·대사를 지연 — `side=='BUY'` 강제라 SELL 체결은 아무것과도
  맞지 않고 Exit hook 앞에서 반환한다.
- 스키마 소유권 — `schemaV20` 리터럴이 `75cb371a` 와 바이트 동일(9008 바이트). v21~v32 와
  `.sql` 전체에서 v20 의 18개 객체에 대한 `ALTER`/`DROP`/`CREATE TRIGGER … ON` 일치 0.
- Position 수량 권위 — `positions` writer 는 저장소 전체에서 셋뿐이고 campaign 코어는
  그중에 없다.
