# a192 — missing-order 종목 차단은 그 주문들의 종결 증명으로 풀린다

> **상태: 1단계(증거·제안·설계·분기 지도·시험 계획), High-risk(reconciliation).** production Go·시험 편집 0. 구현(2단계)은 Manager 지시 뒤.
> Story: `STORY-TOS-a192`(FEAT-TOS-005). base: `6277ec34e`(`base-commit.txt`).
>
> **개정 2(Manager 결정 2026-11-17 반영)**: 결속은 프로세스 기억(스키마 변경 없음), 체결 id 워터마크(M11), 무장-후-해제(M12), 영구 차단은 운영자 전용(P3).
> 리뷰 1·2회차의 구멍 A·B 는 design D3 C12·C13 이 닫는다. 결정 반영분 리뷰(3회차) 처분은 `review.md`.

## Why

2026-11-16 운영 사고: 엔진이 낸 주문을 사용자가 토스 앱에서 취소했다. 단건 조회가 그 주문을 계속 `PENDING` 으로 돌려줘 체결 감시가 종결을
못 적었고, 대사가 「장부에는 열린 주문인데 계좌 open 목록에 없다」(missing-order)로 종목 차단을 건 뒤 같은 분쟁 3회 관측으로 계좌 전체 영구
차단으로 승격했다. a191(2026-11-17 배포)이 체결 감시가 종결(CLOSED) 목록으로 그 주문의 종결을 확인해 장부에 적게 고쳤다. 그러나 **차단은
스스로 풀리지 않는다** — 운영에서도 종결 기록 뒤 엔진 정지 → `reconcile-resolve --confirm` → 재시작이 필요했다.

코드 근거(전부 `tools/logic-map` AST 열거 — `analysis/cited-ast/README.md`; 값·SQL 은 「손으로 읽음」):

| 단계 | 근거 |
|---|---|
| missing-order 는 첫 blocking 비교에서 즉시 종목 블록이 된다 | `Tracker.Observe` B3(else@684) → `blocksFor`(call@685) B2(range@1418), 새 키면 B17(if@720) → pending, `persist` B2(range@869) → `EnterReconcile`(call@880). 블록 값 `Cause: QUANTITY_MISMATCH`·`Release: adjusted_reconcile`(mismatch.go:1419~1430, 손으로 읽음) |
| 종결이 적힌 다음 비교에서 그 주문은 차이에서 빠진다 | `LocalStateFromJournal` 은 `TrackedFillOrders`(call@201) 결과만 B4(range@212)로 `OpenOrders` 에 넣고, 추적 SQL 은 `terminal = 0` 만 돌려준다(fills.go:1604·1627, 손으로 읽음) |
| 그래도 블록은 남는다 | 차이 없는 비교에서 `Observe` B2(if@637) → B6(range@653) → B7(if@654) 통과 → B8(if@662) `!creditUsableBy(t.adjusted[symbol], …)` → `AwaitingAdjustment`(append@667). 크레딧은 `ConvergeQuantities` 의 defer(call@189)만 주는데 그 함수는 `diff.Quantities` 가 비면 B1(if@144) → return@145 |
| 남은 블록은 보호를 줄인다 | 편입: `judgeHoldings` B10(if@152) `d.blocked` → continue(로그 0). 청산: 블록이 그 종목을 덮거나 영구 블록이 있으면 `reconcileFloor.ConfirmedFloor` B2(if@197)를 지나 매 청산 전 브로커 읽기 2회(call@207·@231)와 수량 캡이 붙고, 읽기 실패면 `applyFloor` B2(if@1761) → `"0"`(return@1768) — **손절이 한 주도 안 나간다** |
| 운영자 해제는 엔진 정지가 전제 | `runEngineReconcileResolve` 가 엔진 잠금(call@133)을 잡고 3회 안정 스냅숏 뒤 `Tracker.Resolve`(call@204)로 **전부** 푼다. 엔진이 멈춘 동안 손절 없음(기억 「엔진 정지 = 손절 없음」) |

사용자 판단(2026-11-17, 원문): 「차단 하는게 더 위험 한거 같은데 1번으로 진행하고 차단 해제도 고려 해줘」. 1번 = 종목 차단에 원인 해소 해제를 추가.

## What Changes

1. **결속은 프로세스 기억(스키마 36 그대로)**: 차단 행을 쓰는 `EnterReconcile` 은 이미 `(행, entered bool, 오류)` 를 돌려준다. 추적기가 missing-order 블록을 쓸 때
   `entered == true`(이 관측이 행을 **실제로 삽입**함)일 때만 그 행 id 에 결속(관측된 missing-order 신원 집합, 수량·외부 포지션 흔적, 삽입 전 체결 id 워터마크,
   무장 상태)을 기억에 만든다. 기존 행에 겹친 관측, 재기동 전 행, 다른 생산자 행은 결속이 없다 → 이 경로로 안 풀린다(design D2).
2. **판정 함수 하나(새 파일)**: 결속이 있고, 행 id 가 그대로이고, 흔적이 없고, 결속 주문 **전부**가 완전한 신원·같은 계좌·lineage 무이동이며 장부에
   `terminal = 1 ∧ fail_closed = 0` 이고, 비교 as-of 가 그 종결 기록들보다 엄격히 나중이며, 그 비교가 그 종목에 차이·같은 번호 외부 주문을 보지 않고, 삽입 뒤 그
   종목(계좌 무관)이나 종목 빈 체결이 없을 때 — **처음엔 무장만, 같은 증명(digest)으로 더 나중 비교가 다시 통과하면 해제**(design D3 C1~C13).
3. **해제 자리는 오늘의 자동 해제 자리 그대로**: 비교 전체가 진입을 막지 않을 때, 조정 크레딧이 없어 `AwaitingAdjustment` 로 가던 블록에 한해. 수량 규칙이
   먼저이고 그대로다(I2).
4. **해제는 같은 트랜잭션 안 재판정**: 새 journal 메서드가 행 id 고정 + `released_at IS NULL` 로, 증명을 tx 안에서 다시 읽어 같은 판정 함수를 통과할 때만 해제하고
   새 release cause `ORDER_TERMINAL_CONFIRMED` 와 근거를 남긴다. 계좌·주문번호·수량·가격 없는 로그 하나(design D5·D7).
5. **바꾸지 않는 것**: 장부 스키마, 차단을 거는 조건·시점, 수량 불일치 해제 규칙, 영구 승격과 영구 차단의 운영자 전용 해제, `EnterReconcile`, 청산·손절 경로,
   `reconcile-resolve`, a191 체결 감시.

## 이 change 가 하지 않는 것(범위 밖 — 후속 후보, design §7)

- **청산 확정 하한의 적용 범위**(블록 중 손절이 읽기 실패 시 0 주 — design Q5 ③)와 **편입 건너뛰기 가시성**(`judgeHoldings` B10) — 보호를 직접 줄이는 자리라 후속 1·2순위.
- 계좌 전체 영구 차단의 자동 해제(P1'). Manager 결정 M2 = P3 — **별도 Story**.
- 엔진을 세우지 않는 운영자 해제(P2 — 콘솔·엔진 내 명령).
- `reconcile.clean` 이벤트 의미(a191 후속), `Tracker.ExitAllowed` 가 production 호출자 0 인 장식 메서드인 점.
- 정정으로 옛 번호가 남은 주문(a167) — 블록 유지.
- 운영 장부 수리·배포·실계좌 조회 — 별도 사람 승인.

## Capabilities

### Modified Capabilities

- `reconciliation`: missing-order 종목 차단의 주문 종결 증명 해제(ADDED Requirement 3 — 해제 조건, 삽입 증명 결속, 같은 tx 재판정), 「자동 해제는 조정 뒤
  재조회 일치에만」 문장 두 곳을 그 예외를 포함하도록 수정(MODIFIED Requirement 2 — 영구 차단의 운영자 전용 해제는 문장 그대로).

## Impact

- 새 파일: `internal/reconcile/order_terminal_release.go`(결속·무장 기억, 판정, 로그 함수), `internal/journal/reconcile_order_terminal_release.go`
  (`MaxFillEventID`, tx 해제 메서드·tx 읽기), 시험 파일 2. 기존 함수 본문 편집(gate ⑤ FLM 대상, design D11): `Tracker.Observe`, `Tracker.persist`, `blocksFor`,
  `ValidReconcileReleaseCause`, `buildGateway`, `ReconcileDriver.RunOnce`, `Tracker.Restore`(결속 지우기 1줄). `EnterReconcile` 무편집.
- 원장 스키마 변경 없음(36). 롤백은 이전 이미지 교체(백업 복원 불필요).
- 브로커 호출 0. retry matrix 변경 없음. 설정·토글 없음(`Tracker.Terminal == nil` = 오늘과 같음).
- 재기동하면 남은 missing-order 행은 이 경로로 안 풀린다 — 운영자 해제(오늘과 같음).
- 위험 버킷 broker-zero 는 새 release cause 를 근거로 받지 않는다(fail-closed — M7).
