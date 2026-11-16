# a191 — 체결 감지가 미체결 목록을 떠난 추적 주문을 종결 목록으로 확인한다

> **상태: 1단계(제안·설계·분기 지도·시험 계획), High-risk.** production Go 편집 0. 구현(2단계)은 Manager 가 1단계를 검증한 뒤 지시한다.
> Story: `STORY-TOS-a191`(FEAT-TOS-002). base: `a392308e8`(`base-commit.txt`).

## Why

2026-11-16 운영(Manager 보고)에서 처음 드러난 결함이다. 엔진이 낸 위험 축소 매도가 접수된 뒤 **체결 없이** 미체결(OPEN) 목록을 떠났다.
체결 감지는 그 주문을 번호로 단건 조회했고, 단건 조회는 매 주기 `PENDING`(취소 시각 없음)을 돌려줬다. 그래서 감지는 그 주문을 매 주기
「열림·미체결」로 다시 읽었고, 장부와 같으니 아무것도 쓰지 않았다(오류·로그 없음). 대사는 「장부에는 열린 주문인데 계좌 open 목록에 없다」
(missing-order)를 반복 관측해 계좌 전체 영구 차단으로 승격했다. 운영자 해제 명령 `tossctl engine reconcile-resolve` 는 차이가 남아 있어
거절한다. **지금 있는 수단으로는 못 푼다.**

코드 근거(전부 `tools/logic-map` AST 열거 — `analysis/cited-ast/README.md`, `analysis/function-logic/internal-filldetect--detector.collect/`):

| 단계 | 근거 |
|---|---|
| 목록을 떠난 추적 주문은 단건 조회 하나로 상태가 정해진다 | `Detector.collect` B13(range@423)·B14(if@424 `seen` → continue)가 목록에서 본 것만 거르고, 나머지는 `d.Order.OrderRaw`(call@428) → `parseSnapshot`(@433) → `brokerstate.Derive`(@454)뿐이다. 「목록에 없다」는 사실을 읽는 분기는 B14 하나이고 그것은 단건 조회를 **부르는** 조건일 뿐이다 |
| 단건 `PENDING`·취소 증거 없음·후속 없음·체결 0 은 `OPEN_UNFILLED` | `Derive` B9(case@387) → B10(switch@388) → B14(default@398) → `settle(OPEN_UNFILLED)` return@399 |
| 그 스냅숏은 비종결이라 추적 집합에 남는다 | `TrackedFillOrders` SQL 의 `f.terminal = 0`(fills.go:1604·1627) — **SQL 문자열이라 AST 가 열거하지 못함, 손으로 읽음** |
| 대사는 그 주문을 열린 주문으로 센다 → missing-order | `LocalStateFromJournal` `TrackedFillOrders`(call@201) → B4(range@212)로 `OpenOrders` 를 채움, `Comparer.Compare` B23(range@477)·B24(if@478)·B25(if@483)로 `MissingOrders`, `Diff.BlocksEntry`(return@313, 분기 0)가 그 길이를 셈 |
| 해제 명령은 차이가 남으면 거절 | `runEngineReconcileResolve` B21(if@198 `diff.BlocksEntry()`) → return@199 |
| 감지는 아무 말도 하지 않는다 | `Detector.Run` 분기 B1(for@518)·B2(if@519)·B3(if@523)뿐이고 `PollOnce`(call@522)의 결과를 읽는 분기·로그 호출이 0 |

브로커 근거(영수증): 토스 단건 조회는 KR 주문을 앱에서 취소해도 `PENDING`·`canceledAt` 없음을 돌려주고, 정정해도 옛 id 를 `PENDING` 으로 읽는다
(`openspec/changes/verify-execution-capability/measurements.md` M32, 「status 하나로 취소를 판정하면 안 된다」). 목록 조회는 `CANCELED`/`FILLED`/
`REJECTED` 를 보여 준다(M31). openapi 는 `status=CLOSED` 를 「FILLED, CANCELED, REJECTED, REPLACED 등 종료 상태 주문」의 그룹으로 정의하고
`from`/`to`(주문 생성일 KST)·`symbol`·`limit`(최대 100)·`cursor` 를 받는다(`docs/migration/openapi.latest.json` paths./api/v1/orders.get).

이것은 손절 즉시성(안전 불변식 4)에도 닿는다: 보호 청산 매도가 이 상태에 빠지면 `ApplyExitFill` B11(if@1009 `fill.Terminal && state.Pending() …`)이
영영 참이 안 돼 청산 발의가 무장된 채로 남는다. 그 뒤의 재발의 억제(ratchet/ladder `SuppressedPending`)는 a167 proposal 이 a087 AST 에서 옮긴
**빌린 증거**이며 이 change 는 HEAD 에서 그 함수들의 AST 를 만들지 않았다(안 봄).

## What Changes

1. **새 파일 `internal/filldetect/closed_lookup.go`** — 목록을 떠났는데 단건 파생이 비종결인 추적 주문 하나를, 같은 주기 안에서 종결(CLOSED)
   목록으로 확인한다. 질의는 기존 `d.Orders`(`execgw.OrderPager`)와 기존 `execgw.ScanOrders`(페이지 상한·커서 반복 거절)를 그대로 쓴다.
   필터는 `status=CLOSED`·`symbol`·`from`/`to`(추적 주문 거래일 ±1일)·`limit=100`, 페이지 상한 3.
2. **받아들이는 경우는 하나뿐이다**: 같은 주문번호(바이트 동일)·같은 정규 범위(계좌·시장·거래일·종목·방향) 행이 **정확히 하나**이고, 그 행을
   추적 lineage 로 `brokerstate.Derive` 한 결과가 **깨끗한 종결**(`Terminal && !FailClosed`)이며, 그 행의 주문 수량·누적 체결 수량이 단건 조회와
   **같다**. 그때만 그 주기 단건 스냅숏의 **상태 필드**(status·취소 시각·파생)를 종결 목록 행의 것으로 바꾼다. 수량·평균가·금액·체결 시각은
   단건 것을 유지한다 — 체결은 종결 목록 혼자서 절대 정하지 않는다.
3. **그 밖의 모든 결과는 오늘과 같다**(단건 스냅숏을 그대로 씀, 종결 기록 없음, 새 refusal·래치 없음, 주기 실패 없음). 대신 사유 코드를 남기고,
   같은 주문이 정해진 시간 이상 그 상태면 경보한다(경보 통로는 Manager 결정 M1).
4. 같은 주문의 종결 목록 조회는 최소 간격으로 제한한다(첫 조회는 즉시). retry matrix 정본 표에 이 호출을 계상한다(§0.4).
5. `Detector.collect` 본문에는 호출 두 줄만 더한다(분기 추가 0). `pollLocked`·`Run`·journal·reconcile·스키마·설정·토글은 바꾸지 않는다.

## 이 change 가 하지 않는 것

- 종결 목록의 체결 수량으로 체결을 기록하지 않는다(수량 불일치는 경보로 끝낸다 — design D3).
- 대사(reconcile)의 차단을 스스로 풀지 않는다. missing-order 심볼 차단은 `ReleaseOnAdjustedReconcile` 라 조정 크레딧 없이는 안 풀리고
  (`Tracker.Observe` B6~B9, `blocksFor` B2, 수렴은 수량 차이가 없으면 크레딧을 안 줌), 계좌 영구 차단은 운영자 전용이다. 이 change 뒤에는
  그 주문의 차이가 사라져 `reconcile-resolve` 가 **조건부로** 실행 가능해진다 — 그 id 에 빈 범위 legacy 스냅숏 행이 없고 다른 차이가 없을 때
  (design Q5). 실행은 사람 몫이다.
- `reconcile.clean` 로그가 「차이 없음」이 아니라 「주기 오류 없음」을 뜻해 운영자를 속인 문제(`ReconcileDriver.logCycle`) — 후속(아래).
- 운영 장부 수리·배포·실계좌 조회 — 별도 사람 승인.

## 후속(범위 밖)

- `reconcile.clean` 이벤트의 의미(`internal/app/engine/reconcileloop.go` `logCycle`, `obs.EventReconcileClean`) — 오류 없는 주기와 차이 없는
  주기를 가르는 별도 change.
- 체결 감지 주기 결과가 어디에도 기록되지 않는다(`Run` 이 `PollOnce` 결과를 버림) — M1 결정에 따라 일부만 이 change 가 다룬다.
- a167(보호 잔량 재가격)은 「비종결 = terminal 0」을 전제한다. 이 change 가 그 전제를 참에 가깝게 만든다. a167 은 종결 확인 판정을 새로
  만들지 말고 이 change 의 종결 스냅숏을 써야 한다(design D9).

## Capabilities

### Modified Capabilities

- `fill-detection`: 목록을 떠난 추적 주문의 종결 확인(ADDED Requirement 1개).

## Impact

- 새 파일 1(`internal/filldetect/closed_lookup.go`)·새 시험 파일 1. 기존 함수 본문 편집은 `Detector.collect`(호출 2줄) — FLM 번들 필수.
  M1 결정이 「로그+경보 배선」이면 `cmd/tossctl.engineFillDetector`(필드 2줄)도 번들 대상.
- 브로커 호출: 정상 경로 0. 목록 이탈 + 단건 비종결일 때만 주문당 최대 1회/간격(페이지 ≤3). a166 계좌 공유 3 req/s 예산의 Normal 우선순위.
- 원장 스키마·설정·토글 변경 없음. TossOS 전용 경로(엔진의 `filldetect` 루프)라 토글 OFF = upstream 동일(불변식 3)에 닿지 않음 — design D8.
- High-risk(체결 감지). Pre-Edit·FLM/BTM·RED/GREEN·변이·적대 리뷰 의무.
