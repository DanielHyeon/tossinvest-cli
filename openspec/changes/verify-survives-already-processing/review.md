# Review: verify-survives-already-processing

Function Logic Map: applied — `analysis/function-logic/`, `check_analysis.py` 통과.

## 1. Proposal freeze (Eng)

두 수정 모두 **같은 실행의 실측**이 근거다(run-IXCQU5UBZE, measurements.md M24·M26). 추측으로
브로커 동작을 가정한 부분이 없다 — 대기값은 브로커가 응답에 담아 보낸 `retryAfterSeconds`이고,
정정이 같은 코드로 거절된다는 사실은 requestId `si1tUiUvi8DzWXr5`가 있는 실제 호출이다.

**같은 실행이 직전 change를 실증했다**: `apply-us-measurement-fixes`의 취소 재시도가 실계좌에서
409를 넘겨 order-cancel을 pass로 만들었다(M25, `order.cancel.retries=1`). 이 change는 그 규칙을
같은 근거 위에서 정정까지 넓힌다.

## 2. Code review

- **재시도 확장의 경계**: 판정은 `transientRefusal` 하나이고 조건은 종전과 동일하게 좁다 —
  HTTP 409 **그리고** 본문 `code == "already-processing"`. 적용 지점은 `cancelOrder`와
  `amendOrder` 두 곳뿐이며, `placeOrder`·`createConditional`은 손대지 않았다.
- **왜 정정은 되고 접수는 안 되는가**: 선을 "그 요청이 **생성**할 수 있는가"에 그었다. 취소는
  주문을 없애고, 정정은 같은 주문을 시장에서 한 호가 **더 먼** 가격으로 대체한다
  (`OneTickFurther`) — 몇 번을 보내도 두 번째 라이브 주문이 생기지 않는다. 접수와 조건주문
  생성은 반복하면 생길 수 있다. 이 문장이 규칙의 전부이고, 상수 주석에 그대로 적었다.
- **정적 가드가 다시 설계를 결정했다**: 재시도 루프를 두 함수가 공유하도록 빼면
  `TestTheApprovedPlanIsTheOnlyThingMutateGoActsOn`이 "gate에서 한 홉 떨어진 mutating 호출"로
  실패한다. `cancelOrder`에서 그랬듯 `amendOrder`에도 **인라인**했고, 왜 중복인지를 주석에 남겼다.
- **sweep의 범위**: `sweepStep`은 **그 단계가 만들고 취소하지 않은** 산출물만 본다
  (`Outstanding`을 그 단계의 artifacts만 담은 합성 Entry에 적용). 기록의 이전 실행 잔여물은
  건드리지 않는다 — 그것은 prologue의 몫이고, prologue는 사람이 읽는 목록에 올린다. 범위를
  `cancelLiveOrders`(기록 전체를 보는 기존 헬퍼)로 잡았다면 이 단계의 계획 줄로 다른 실행의
  주문을 취소하게 됐을 것이다.
- **sweep이 게이트를 지나는가**: `r.cancelOrder`를 부르므로 `r.gate` → `Plan.Authorises`를
  그대로 지난다. 계획에 그 단계의 취소 줄이 없으면 전송되지 않는다
  (`TestTheSweepGoesThroughTheGate`가 계획에 그 줄이 있음을 고정).
- **sweep이 하지 않는 것**: `sr.abort`가 설정됐거나 컨텍스트가 죽었으면 아무것도 보내지 않는다.
  `ErrOutsidePlan`으로 멈춘 실행이 "정리니까 한 건만 더"를 보내는 것은 그 오류의 의미
  (승인 밖 요청은 보내지 않는다)를 정면으로 부정한다.
- **판정을 미화하지 않는다**: sweep은 단계의 verdict를 바꾸지 않는다. 실패한 측정은 실패로
  남고, sweep 실패는 `step.sweep.failed`로 기록되며 산출물은 계속 잔여물로 보고된다.

## 3. Security review (CSO 관점)

**mutation 자동 재시도 금지의 두 번째 예외**(2b task 1.7③)를 만든다. 좁히는 근거:

- 정정은 **생성 연산이 아니다**. 이 도구의 정정은 대상 주문 ID와 목표가가 고정된, 이미 승인된
  같은 요청이며, 가격은 항상 시장에서 **멀어지는** 방향이다. 재시도가 노출을 늘릴 수 없다.
- 브로커가 재시도를 **명시적으로 요청한다**(`retryAfterSeconds`). 409 `already-processing`은
  "처리하지 않았다"는 확정 응답이다.
- 실제 피해가 이미 관측됐다(M26): 정정 실패 1건이 주문을 남겨 sell-boundary 측정을 통째로
  막았다.
- 상한이 있다: 추가 2회, 대기 상한 5초, 주입된 sleep seam이라 테스트가 실제로 자지 않는다.

**sweep의 새 위험**: 실행이 취소를 한 건 더 보낼 수 있다. 방향은 **노출 감소뿐**이고, 대상은
그 단계가 방금 만든 주문뿐이며, 같은 게이트를 지난다. 이것은 도구가 이미 성공 경로에서 하던
일이고, 이 change는 실패 경로에서도 하게 만들 뿐이다.

**잔여 위험(수용)**: ① 브로커가 정정을 실제로 처리하고도 409로 답하는 경우(응답 유실 포함)
재시도는 옛 ID로 404를 받고 새 주문이 기록에 없는 채 살아 있을 수 있다 — 재시도가 **없어도**
동일한 위험이며(응답 유실), 다음 실행의 화면·리포트가 계좌 잔여물로 드러낸다. ② 상한까지
실패하면 주문이 남는다 — 종전과 같고 prologue가 다음 실행에서 정리를 제안한다.

**게이트**: `ProtectionReady`·automation gate·엔진 기동에 손대지 않는다.

## 4. QA

- `go build ./...`, `go vet ./...` clean. `go test ./...` **3131 passed**.
- RED 관측: 409 1회 후 성공하는 fake에서 order-amend가 fail(실계좌 실패를 그대로 재현) →
  GREEN(pass, `order.amend.retries=1`); 정정이 상한까지 실패할 때 그 단계가 낸 주문이 남고
  **sell-boundary가 `ErrExposureCap`으로 실패**(실계좌 연쇄 그대로) → GREEN(주문 없음,
  연쇄 없음); 접수 거절은 재시도되지 않고 그 단계가 실패.
- 실계좌 확인은 다음 US 재측정에서: order-amend가 pass가 되고 sell-boundary가 상한이 아닌
  자기 이유로 판정되어야 한다.

## 5. 완료 조건

- 미완료 태스크 0, FLM 통과, PM check 통과, gate 8/8.
- 사용자 조치: 새 빌드 설치 후 콘솔 재시작 → `/verify?market=US` → `[재측정]` → 승인 클릭.
