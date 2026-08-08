# a096 작업

Base commit: `ec29dc72` (`base-commit.txt`).
전제: 증거(§1)가 문서(§2)보다 먼저 만들어졌다 — AST 산출물의 mtime이 proposal·design보다 이르다.

**2판이다.** 독립 리뷰 1라운드(Codex)가 blocker 4건으로 1판을 깼다. §3bis에 두 번의 설계
정정을 기록한다.

## 1. 증거 — 완료

- [x] 1.1 `Notifier.Notify` AST — 1분기
- [x] 1.2 `Notifier.notifyCritical` AST — 4분기 (B4의 조건이 바뀌었다)
- [x] 1.3 `Notifier.deliver` AST — 10분기. **B5(`markErr != nil`)가 결함의 실행 지점**
- [x] 1.4 `Notifier.claimAndDeliver` AST — 2분기 (2판 신규)
- [x] 1.5 `Notifier.Flush` AST — 6분기 (2판이 잠금을 더했다)
- [x] 1.6 `Journal.EnqueueAlert` AST — RED 9분기 → GREEN 0분기(위임)
- [x] 1.7 `Journal.ClaimAlertForDelivery` AST — 11분기 (신규)
- [x] 1.8 `claimOwed` AST — 7분기 (2판 신규, 순수 판정)
- [x] 1.9 `Journal.MarkAlertDelivered` AST — 1분기. `no such alert`의 발생지
- [x] 1.10 커버리지 실측 RED — obs 84.8%, journal 74.9% (325.2s)
- [x] 1.11 격리 측정 — `deliver` B5에 진입하는 테스트를 `TestNotifierIsConcurrencySafe`로 특정
- [x] 1.12 운영 실측 — `alert_outbox` 14행 / `engine.log` 재전송 53건 / 거절 60건 / ntfy push
- [x] 1.13 FLM 헤더를 `ast.json`에서 기계 생성 (a085 §12.3 교훈; 1판이 stale 헤더 5건을 남겼다)
- [x] 1.14 BTM 표의 `:줄` 참조를 `ast.json`과 **기계 대조**한다. 헤더를 생성으로 바꾼 뒤에도
      표 안의 숫자는 사람이 타이핑하므로 같은 방식으로 썩는다. 첫 실행에서 5건을 잡았다.
- [x] 1.15 측정하지 않은 RED 열을 `미측정`으로 표시한다 — `Acknowledge`·`Flush`는 base 시점의
      분기 대응을 뜨지 않았다. 추정한 칸은 측정한 칸과 구별되지 않는다(리뷰 1라운드 B4의 교훈).

## 2. RED — 완료

- [x] 2.1 `internal/obs/a096_one_send_per_condition_test.go` 신규.
      1판 관측: `sends = 3, want 1`, `sends = 5, want 1`, `log lines = 9, want 5`.
- [x] 2.2 과잉 억제 반증 `TestAnUndeliveredConditionIsStillRetried` — **RED에서도 통과**.
      통과하지 않으면 R2가 과잉 억제라는 뜻이므로, 이 테스트는 양쪽에서 초록이어야 한다.
- [x] 2.3 `internal/journal/a096_claim_for_delivery_test.go` 신규 — 상태 6종을 각각 밟는다.
- [x] 2.4 **2판 추가**: `TestADeadTransportIsStillFoundAfterASuccessfulDelivery` (blocker 2),
      `TestALaterOccurrenceOfTheSameKeyIsNotSwallowed` (blocker 3),
      `TestConcurrentObservationsOfOneConditionSendOnce` (blocker 1, `-race`),
      `TestTheSameConditionIsRemindedOncePerWindow`, `TestAnUnrecognisedAlertStateOwesDelivery`,
      `TestAZeroReminderWindowNeverReArms`. 2차 gate 전 셀프리뷰에서 unknown 상태도 PENDING으로
      재무장되어 `MarkAlertDelivered`가 성공하는지 RED를 보강했다.

## 3. GREEN — 완료

- [x] 3.1 `Journal.ClaimAlertForDelivery(ctx, Alert, remindAfter)` 신규.
      base `EnqueueAlert` 본체를 옮기고 SELECT를 `id, state, delivered_at, acknowledged_at`으로
      넓힌다. 창을 넘긴 종결 행은 **같은 트랜잭션 안에서 PENDING으로 재무장**한다.
- [x] 3.2 `claimOwed` 신규 — 순수 판정. PENDING·창·모르는 상태를 한 표로 볼 수 있게 분리한다.
      모르는 상태는 `owed=true, rearm=true`다. rearm 없이 보내면 완료 표시가 실패해 폭주가 남는다.
- [x] 3.3 `Journal.EnqueueAlert`는 **서명·계약 그대로** 두고 `remindAfter = 0`으로 위임한다.
      기존 호출자(`execgw.Gateway.parkAlert`, `outbox_test.go` 7개 테스트) **무수정**.
- [x] 3.4 `Notifier.claimAndDeliver` 신규 — `n.mu`를 claim부터 send까지 잡는다.
- [x] 3.5 `Notifier.deliver`는 잠금을 **잡지 않고** 전제를 `PRECONDITION` 주석으로 명시한다.
- [x] 3.6 `Notifier.Flush`가 같은 잠금을 잡는다.
- [x] 3.7 `notifyCritical` B4를 `owed && !sent`로 — 보낼 필요가 없어서 안 보낸 것은
      전달 실패가 아니므로 gate를 잠그지 않는다.
- [x] 3.8 `Notifier.RemindAfter` + `DefaultRemindAfter = 1시간`.
- [x] 3.9 `Notifier.Acknowledge`도 같은 잠금을 잡는다. 해제는 세고-판정하는 read-then-decide인데,
      재무장이 그 사이에 PENDING 행을 새로 만들면 **미전달 critical 알림이 있는 채로 gate가
      열린다.** 이 위험은 a096이 키웠다 — 이전에는 그 사이에 생기는 PENDING의 유일한 원천이
      "새 조건의 첫 알림"이었다. `TestAcknowledgeCannotClearTheGateMidSend`가 상호 배제를 건다.

### 3bis. 설계를 두 번 고쳤다 (기록)

**게이트가 한 번.** 처음에는 `EnqueueAlert`의 서명을 `(int64, string, error)`로 넓혔다.
`check_analysis.py`가 증거 누락 8건을 보고했다 — `Gateway.parkAlert` + `outbox_test.go`의
기존 테스트 7개. arity 하나 때문에 원장 패키지의 기존 테스트 전부가 "수정된 기존 함수"가
됐다. 새 함수 + 위임으로 좁혔다.

**독립 리뷰가 한 번.** 1판의 억제는 **영구**였다. Codex가 blocker 4건을 냈고 전부 사실이었다:
동시 claim의 이중 전송(B1), 영구 억제가 진입 차단을 없앰(B2), 같은 key의 다른 원인을
영구히 가림(B3), 증거 불일치와 `covermode=set` 오독(B4). 2판은 창 + 재무장 + 배타 구간이다.

## 4. VERIFY

- [x] 4.1 `go build ./...` 통과
- [x] 4.2 `go vet ./internal/journal/ ./internal/obs/ ./internal/execgw/` 통과
- [x] 4.3 `go test -race ./internal/obs/` 통과 (70.6s) — 동시성·상호배제 테스트 포함
- [x] 4.4 `go test ./internal/journal/... ./internal/execgw/...` 통과 (493s, 81s)
- [x] 4.5 AST 재생성 후 `check_analysis.py --change a096` → `evidence complete`
- [x] 4.6 `openspec validate --strict` 통과
- [x] 4.7 GREEN 커버리지 obs 84.8% — `deliver` B5 블록 **미진입**(RED은 진입).
      이 블록은 이미 전달된 행에 다시 전달 표시를 시도했을 때만 들어간다. `set` 모드이므로
      읽을 수 있는 것은 발생 여부까지이고 횟수가 아니다.
- [x] 4.8 기존 `TestTheSameConditionEnqueuesOnce`·`TestNotifierIsConcurrencySafe` 계속 통과
- [x] 4.9 journal GREEN 커버리지 75.0% (RED 74.9%). BTM의 GREEN 열을 실측으로 채웠다.
      `claimOwed` B2·B3·B4·B6·B7 진입 / B5 미진입(정상 mutator로 도달 불가) / B1 자기 블록 없음.
      `ClaimAlertForDelivery` **B6(재무장) 진입**, B1·B2·B5·B11 진입, B3·B8·B9·B10 미진입.
- [x] 4.10 `go test ./...` 전체 통과 (3.9 반영 후) — FAIL 패키지 없음

## 5. 게이트

- [x] 5.1 `make sdd-sync` → CodeGraph hard-evidence 일치.
      `codegraphcontext`(300초 초과)와 GBrain(병행 프로세스 점유)은 advisory 경고.
- [x] 5.2 `make sdd-check` 통과 (PM Story 3건 신규 생성 후)
- [x] 5.3 `make gate CHANGE=a096-one-condition-is-one-alert` — code/test/vet 통과 후 a095 제안서 형식 오류를 수정하고 전체 validate 84/84 통과
- [x] 5.4 독립 리뷰 1라운드 — **Codex, 교차 모델 충족.** 판정 FAIL, blocker 4건.
      a092부터 이어진 미충족 연쇄가 여기서 끊겼다.
- [x] 5.5 독립 리뷰 2라운드 — Claude Sonnet, 교차 모델. PASS, blocker 0.

## 6. 운영 실측 (배포 후, 사람 승인)

이 목록은 **이 change의 완료 조건이 아니다.** 배포하는 사람이 배포 뒤에 확인하는 것이며,
오늘은 배포하지 않고 엔진도 정지 대상이다. 체크박스로 두지 않는 이유가 그것이다.

- `engine.log`에서 `engine.alert_undelivered`의 `no such alert` 오류 0건 확인.
  이 오류는 이미 전달된 행에 다시 표시를 시도했을 때만 나온다.
- 같은 조건이 지속되는 동안 push가 **창당 1건**만 도착하는지 운영자 확인.
- 전달 뒤 transport를 의도적으로 끊고 창을 넘겼을 때 `ENTRY_BLOCKED`가 걸리는지 확인 —
  design D5의 운영 측 대응물. 사람 승인 필요.
- `alert_outbox`의 PENDING 행(현재 6건)이 전송 복구 시 여전히 전달되는지 확인.

## 7. 범위 밖 — 별도로 남긴다

이 목록도 **이 change의 과제가 아니다.**

- **장 운영 시간 게이트 없음.** 비거래일에 익절 발의가 만들어지고 브로커가
  `order-hours-closed`로 거절하는 경로. a096은 알림만 고치고 **거절된 주문은 계속 나간다.**
  `GetTradingHours`는 CLI만 쓴다. 별도 change 필요 — a094 R1의 확정 거절 분류에 붙일 수 있다.
- **event key에 원인이 없다.** `exit.proposal_refused`의 key는 `type|position_id|action|level`이며
  거절 사유를 담지 않는다. 창이 최대 한 창의 지연으로 이를 완화하지만 근본 해결은 아니다(design D4).
- **`TestNotifierIsConcurrencySafe`에 단언이 없다.** a096은 그것이 만들던 결함을 없앴지만
  테스트 자체는 그대로다. 단언을 넣는 것은 그 테스트의 소관이다.
- **`Notifier.Flush`에 non-test 호출자가 없다.** 따라서 `Gateway.parkAlert`가 넣는 행은
  현재 배선에서 아무도 집어 가지 않는다(`replay.go:101` 주석이 전제하는 것과 다르다).
  `Flush`가 `MarkAlertAttemptFailed` 오류를 버리는 것도 같은 함수의 기존 문제다.
  독립 리뷰 1라운드 concern 3·4이며 a096이 만든 것이 아니다.
- **a092와의 잠금 구간.** a096이 배타 구간을 claim까지 넓혔다. a092가 알림을 비동기로
  옮기면 이 구간의 위치를 다시 정해야 한다.
