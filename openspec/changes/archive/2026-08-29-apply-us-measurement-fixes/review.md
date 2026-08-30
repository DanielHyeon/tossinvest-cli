# Review: apply-us-measurement-fixes

Function Logic Map: applied — `analysis/function-logic/` 19 target, `check_analysis.py` 통과.

## 1. Proposal freeze (Eng)

두 수정 모두 **실측이 근거**다(measurements.md M16·M17, run-M6WQZ5WKGGE4KS4C). 추측으로
브로커 동작을 가정한 부분이 없다 — 재시도 대기값은 브로커가 응답에 담아 보낸
`retryAfterSeconds`이고, 상태 필터 허용값은 브로커가 400 본문의 `allowedValues`로 알려준
것이다.

## 2. Code review

- **재시도 범위가 좁은가**: `transientCancelRefusal`은 HTTP 409 **그리고** 본문
  `code == "already-processing"`일 때만 참이다. 다른 409(예: 상태 충돌)는 재시도하지 않는다.
  적용 지점은 `cancelOrder` 한 곳이며, 접수·정정·조건주문 생성 경로는 손대지 않았다.
- **승인 범위를 넓히지 않는가**: 재시도는 `r.gate(...)`를 통과한 **뒤** 같은 주문 ID에 대해
  같은 취소를 반복한다. 새 요청 종류도, 새 심볼도, 새 수량도 만들지 않는다.
- **정적 가드가 잡은 설계 오류 하나**: 처음에 재시도 루프를 별도 함수(`cancelWithRetry`)로
  빼자 `TestTheApprovedPlanIsTheOnlyThingMutateGoActsOn`이 "gate/authorise를 거치지 않고
  mutating 메서드를 호출한다"로 실패했다. 가드를 고치는 대신 **루프를 `cancelOrder` 안으로
  되돌렸다** — 게이트와 브로커 호출이 한 함수에 있어야 한다는 규칙이 옳다.
- **상한**: 추가 2회, 대기 상한 5초. 대기는 주입된 `r.sleep` seam이라 테스트가 실제로 자지
  않는다. 무한 루프 불가(attempts > CancelRetryAttempts에서 종료).
- **정직성**: 재시도 횟수를 `order.cancel.retries`로 기록한다. 세 번 만에 된 취소를 한 번에
  된 것처럼 남기지 않는다.
- **필터 교정**: `WATCHING`(조건주문 자신의 상태) → `OPEN`(목록 필터의 어휘). 문서상
  `OPEN`은 "진행 중(감시 중·일시중지·주문 진행 중 포함)"이라 이 단계가 확인하려던 집합과
  같다. 상수로 뽑아 두 곳(요청 파라미터 기록·실제 호출)이 갈라지지 않게 했다.

## 3. Security review (CSO 관점)

**mutation 자동 재시도 금지(2b task 1.7③)의 예외를 만든다.** 좁히는 근거:

- 취소는 **노출을 줄이는 유일한 방향**의 mutation이다. §0.3이 취소 앞에 확인 프롬프트를
  두지 못하게 하는 것과 같은 이유로, 취소를 못 하고 남기는 쪽이 더 위험하다.
- 실제 피해가 이미 관측됐다: 취소하지 못한 주문 1건이 노출 상한을 채워 두 단계를 추가로
  막았고, 계좌에 미체결 주문이 남았다.
- 재시도는 이미 승인된 동일 행위의 반복이며 새 주문을 만들 수 없다(취소는 생성 연산이
  아니다). 접수·정정은 반복하면 두 번째 주문이 생길 수 있어 계속 금지한다.
- 브로커가 재시도를 명시적으로 요청한다(`retryAfterSeconds`).

**잔여 위험**: 상한까지 실패하면 주문이 계좌에 남는다 — 종전과 같고, 기록·화면이 잔여물로
보고한다(사람이 처리). 이 change는 그 경우를 조용히 성공으로 만들지 않는다.

## 4. QA

- `go build`·`go vet` clean, `go test ./...` 3115 passed.
- RED 관측: 409 1회 후 성공하는 fake에서 기존 코드가 잔여 주문을 남기고 단계를 fail로
  끝냄 → GREEN(재시도 후 취소, 잔여 없음). 상한 소진 시 fail 유지. 접수 거절은 재시도되지
  않고 그 단계가 실패. 잘못된 status 필터를 400으로 거절하는 fake에서 목록 관측이 실패 →
  GREEN.
- 실계좌 확인은 다음 US 재개 실행에서 이뤄진다(잔여 주문 1건이 이 경로로 취소되는지).

## 5. 완료 조건

- 미완료 태스크 0, FLM 통과, PM check 통과, gate 8/8.
- 사용자 조치: 새 빌드 설치 후 콘솔 재시작 → `/verify?market=US` → [이어하기].
