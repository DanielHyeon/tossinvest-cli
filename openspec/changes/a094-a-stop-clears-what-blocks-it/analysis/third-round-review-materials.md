# a094 proposal-freeze 리뷰 3라운드(task 0.5d) — 검토 자료 목록

작성 2026-09-27. 3라운드를 돌릴 사람·모델에게 넘길 입력을 한 곳에 모은다. **리뷰 자체는 Manager 지시 뒤다.**
판정 기준은 `review.md` §2.11(「3라운드가 받는 것」) 여덟 항목이다.

## A. 교차 모델 요구 (task 0.5d 원문)

> 0.5d **proposal-freeze 리뷰 3라운드**. **교차 모델을 여기서 지킨다** — a092 여섯 + a094 두 라운드가 미충족이다

1라운드는 Claude 보이스 셋(사용자 지시 「클로드로 돌리세요」), 2라운드는 Codex 사용 한도로 Claude 대체였다
(`review.md` §2.0). 3라운드는 **Claude 가 아닌 모델 보이스 최소 하나**가 필요하다. a124 review §0.13 이 기록한 codex
CLI 인증 오류(`401 Unauthorized`)는 **2026-09-27 재가동이 확인됐다** — 기본 `~/.codex` 인증으로 TossOS 에서 정상 완료
(`task_complete`, 401 흔적 0)한 세션 셋: `01a0e235-4a0f-72b1-8cee-3fb977773772`(18:32 KST) ·
`01a0e291-5dae-7450-84f3-be59e4b526f3`(20:12, a112 로트) · `01a0e2ad-e633-74e1-88f2-e4f0a66e4d7a`(20:44, a124 로트 PONG 프로브).

## B. 리뷰 대상 (3판 본문 — 고치지 않은 채로)

| 자료 | 경로 | 비고 |
|---|---|---|
| proposal 3판 | `proposal.md` | 줄 인용은 base `ec29dc72` 기준 — C 의 정오표로 옮겨 읽는다 |
| design 3판 | `design.md` | D−1·D1 소급·D2 축소·D3 교체(tasks 0.5c) |
| tasks 3판 | `tasks.md` | §3 R2 · §4 R3 · §4bis R1 소급 · §8 배포(사람 승인) |
| spec delta 2 | `specs/order-execution/spec.md` · `specs/exit-policy/spec.md` | order-execution 은 「IN_DOUBT 해소」 전문 재현(MODIFIED) |
| 이전 판정 | `review.md` §1(1라운드) · §2(2라운드) · §2.11 | 3라운드가 닫았는지 볼 차단 8건 |
| issues | `issues.md` | a087·a089·a091·a092 상호작용 기록 자리(task 6.3) |

## C. 이번에 새로 만든 입력

| 자료 | 경로 | 내용 |
|---|---|---|
| 정오표 | `analysis/third-round-errata.md` | 인용 91건 대조(맞음 57 · 이동 32 · 변경 2), `record` 분기 재번호, 이웃이 바꾼 것 3가지, a094 R1 ↔ a089 R2 문장 대조 |
| FLM 번들 15 | `analysis/function-logic/` | stale 7 을 HEAD 로 refresh(3937e341) — record 는 14→16 분기, 나머지 항등. 신선 8 은 바이트 불변 |
| base 재고정 | `base-commit.txt` = 3937e341(d2f5d3f1) | 자기 Go 커밋 0 영수증은 커밋 메시지 d2f5d3f1 과 `review.md` 「증거 재생성」 |
| 깨끗한 판정 | 격리 detached 워크트리 @d2f5d3f1 에서 `check_analysis --change a094` rc 0 · required 0 | 공유 워크트리는 병행 로트의 미커밋 Go 편집으로 오염돼 있어 거기서 잰 값은 쓰지 않는다 |

## D. 1라운드가 확인하지 못했다고 적은 것 (`review.md` §1.9) — 3라운드가 채울 후보

- 원장 실측(문서 주장을 인용만 했다) · 브로커가 이 code 를 낼 때 부분 접수 후 되돌리는지
- `OrdersPageRaw` 의 실 지연·rate limit 예산 · KRX 예약·조건부주문이 OPEN 목록에 나오는지
- `ast.json` 좌표 1:1 재대조(이번 refresh 로 번들은 HEAD 좌표다 — 3판 본문 좌표는 C 의 정오표로)
- **a087·a089·a091·a092 의 delta 본문 대조**(세 보이스 모두 미실행) — C 의 §5 가 a089 한 쌍을 했다
- `-race` 회귀 · `make sdd-check` · `make gate`(mutating — 사람 승인)
- `official.OrdersFilter` 의 status 그룹이 두 시장 모두 서버측 심볼 필터를 지원하는지

## E. 2라운드 차단 8건 (`review.md` §2.11) — 3판이 답했다고 주장하는 자리

| # | 차단 | 3판의 답이 있는 곳 |
|---|---|---|
| 1 | 잠금 재지목(`pending_action`, submit B8 의 release) | design D−1 「잠금을 잘못 지목하고 있었다」(`design.md:6`) · D3 「왜 B8의 안전 논거를 깨지 않는가」(`:517`) · D6 표(`:705`) |
| 2 | R2 자기 방향 부재 확인 철회 | design D2 「자기 방향 미체결의 부재 확인 — 3판에서 철회했다」(`:386`) · `tasks.md:106` |
| 3 | spec 새 모순 쌍 | spec delta 2 재작성(tasks 0.5c) |
| 4 | a087 선후(`floatOf` 저널분) | `tasks.md:273` 「3판에서 제약 해소」 |
| 5 | R3 `Context.Resolver`·`resolveCancel` 의 `r.Order` · "주문 side effect 없음" 정정 | design D3 「세션 중 해소는 별도 change로 분리한다」(`:573-576`) — 3판은 그 부분을 떼어 냈다 |
| 6 | R2↔R3 상호작용(오염으로 인한 park) | design D5 「셋의 상호작용 (3판)」(`:660`) |
| 7 | §0.4 호출 빈도 · detector OPEN 스냅샷 재사용 | design D2 「§0.4 — 동기 조회를 넣지 않는다. 이미 있는 스냅샷을 읽는다 (3판)」(`:307`) · tasks 3.E2 |
| 8 | 좌표 3건 정정 + 2판 FLM·AST 재생성 | 3판 AST 6 추가(0.5c) · 이번 refresh(C) |

「분할 권고」(R3 을 별도 change 로) 에 대해 3판은 세션 중 해소만 떼어 냈다(design D3 `:573`) — 그 선택이 권고를 충족하는지도 리뷰 대상이다.

## F. 여전히 사람 몫 (리뷰가 대신할 수 없다)

- `tasks.md` §8 배포와 운영 — 8.2 「엔진 재시작은 사람이 직접 승인한다」, 8.3 첫 409 사건의 실물 확인,
  8.4 현재 열린 세 포지션(475150·080220·272210)은 사람이 처리
- `make sdd-sync`·`make gate` 실행 승인(`review.md` §2.6)

## G. a124 12회차가 찾은 AC1 — 3판의 논거에 닿는 자리 (Manager 지시로 추가, 2026-09-27)

a124 codex 12회차(`01a0e2b7-a927-7522-8c27-2847fc08c6ae`, a124 review §0.14)의 AC1: durable 운영 모드 투영이
생산에 배선돼 있지 않다 — `Journal.SetModeProjector`(`internal/journal/operating_mode.go:294`)와
`Journal.RestoreOperatingModeProjection`(`:574`)의 **비시험 호출자가 0** 이다(이 로트에서 재측정: `internal`·`cmd` 비시험
Go 에서 정의 외 출현 0). 전달 실패한 critical 알림은 `Notifier.escalate` 가 `EscalateOperatingMode(…,
ModeTriggerCriticalAlertUndelivered)` 로 원장에 모드를 쓰지만(`internal/obs/notifier.go:378-383`), 재시작 뒤 그 모드를
진입 게이트로 되돌리는 경로가 없다. 같은 주석(`notifier.go:370-377`)이 적듯 프로세스 안 차단(게이트 latch)은 선다.

3판이 이 집행을 전제한 문장 — 리뷰어가 AC1 을 알고 읽어야 할 자리:

- `design.md:441` · `tasks.md:169` — "critical 전달 실패는 `ENTRY_BLOCKED`까지 간다(`notifier.go:216-218`)"
  (현재 줄은 `:381-383`, 정오표 §2). 이 문장은 PENDING_CANCEL 제외의 **근거**(거짓 critical 의 비용)로 쓰인다.
- `specs/exit-policy/spec.md:79` — "등급이 오른 알림의 전달 실패는 신규 진입을 막는다" (같은 SHALL 의 이유절).

프로세스 안 latch 로는 참이고, 재시작을 건넌 durable 집행으로는 AC1 때문에 거짓이다. 3판의 결론(정상 취소를 계수에서 뺀다)은
어느 쪽이든 유지될 수 있으나, 근거 문장이 둘 중 무엇을 뜻하는지는 리뷰가 판정한다.

**정정 (3라운드 F9, 2026-09-27).** 위 "재시작 뒤 그 모드를 진입 게이트로 되돌리는 경로가 없다" 는 **운영 모드에 한해서만** 참이다.
재시작 시 진입 차단은 따로 복원된다 — `internal/app/engine/gateway.go:269` 가 `restoreAlertEntryLatch`(`:153-168`)를 부르고,
그 함수는 미전달 critical 알림이 있으면(`UndeliveredCount > 0`) `ReasonAlertUndelivered` 로 게이트를 막는다. 배선되지 않은 것은
durable 운영 모드의 투영(`SetModeProjector`·`RestoreOperatingModeProjection`)뿐이며, 그래서 전달이 **성공한 뒤**에는 모드의 사람
해제 요구가 재시작을 건너지 못한다(a124 AC1 과 같은 사실, a124 D10 대조 중). 4판의 처리: `design.md` D−2.9.

`design.md:543` "park는 그 자체로 계정 전역 진입을 막는다(`indoubt.go:379-382`)" 는 **다른 기전**이다 — `EntryGate.Block`
(`internal/execgw/indoubt.go:379-381`, 프로세스 안)이지 운영 모드 투영이 아니다. AC1 과 무관한지 리뷰어가 확인한다.
