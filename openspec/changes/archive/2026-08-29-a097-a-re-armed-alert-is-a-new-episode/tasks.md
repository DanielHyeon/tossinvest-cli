# a097 tasks — 재무장된 알림은 새 에피소드다

a096 2라운드 독립 리뷰가 낸 **P2 6건을 닫는다.** 다만 마지막 두 건은 리뷰가 쓴 방식으로
닫히지 않았다 — 처방 하나(start barrier)는 측정이 기각했고, 우려 하나(50ms sleep의 거짓
통과)는 재현되지 않았다. 근거는 2.12·2.14의 실측이며 `review.md` §6에 정리했다.

> **개정 2.** 초판은 6건을 닫는다고 썼고, proposal-freeze 리뷰 지적으로 "5건만"으로
> 낮췄으며, 그 뒤 뮤테이션 실측이 6건 전부 검증된다고 말해 되돌렸다. 되돌린 근거는
> 리뷰에 대한 반론이 아니라 숫자다.

> **개정 1.** proposal-freeze 적대적 리뷰가 이 목록에서 **P1 blocker 하나**를 냈다:
> 초판 2.10은 2.1~2.9가 전부 현재 코드에서 실패할 것을 요구했는데, 2.4·2.5는 의도적으로
> 바뀌지 않는 동작을 고정하는 테스트이고 2.7~2.9는 이미 옳은 동작을 경화하는 것이라
> **초록으로 도착한다.** 그 요구를 그대로 두면 구현자는 RED 증거를 지어내거나 게이트를
> 미완으로 두는 수밖에 없다. 아래 §2는 그것을 두 종류로 나눈다.
> Pre-Edit 선언도 GREEN 뒤에서 앞으로 옮겼다.

체크박스는 **이 change의 완료 조건**이다. 완료 조건이 아닌 것은 체크박스로 두지 않는다
(§7, §8).

## 1. 증거 (구현 전에 끝난다)

- [x] 1.1 memory recall — `scripts/memory-recall.sh`와 MEMORY.md 색인 확인
- [x] 1.2 CodeGraph hard evidence — `ClaimAlertForDelivery`·`claimOwed`·`claimAndDeliver`·
      `notifyCritical`·`Flush`·`Saga.event`의 정의·호출자·호출 대상 확인
- [x] 1.3 **AST 산출물을 proposal보다 먼저 생성** — 위 6개 함수. proposal·design의 분기
      주장은 이 열거를 인용한다
- [x] 1.4 `attempts`·`last_attempt_at`·`delivered_at`을 판단에 쓰는 비테스트 코드 확인 (D1의 전제)
- [x] 1.5 `EntryGate.Block`이 신규 진입만 막고 청산에 영향이 없음을 확인 (D2의 전제)
- [x] 1.6 **`EntryGate`의 래치가 메모리에만 있음**과 `Notifier.Acknowledge`에 비테스트
      호출자가 없음을 확인 (D2가 뒤집힌 근거)
- [x] 1.7 Function Logic Map·Branch Test Map 6쌍 작성 — **RED 커버리지를 측정해서** 채운다
- [x] 1.8 구현 후 AST와 두 Map을 최신화하고 GREEN 측정값을 기재

## 2. RED

새 코드는 **새 파일**에 쓴다. 기존 테스트 파일에 더하면 logic-map 대상이 번진다.

### 2a. 지금 실패해야 하는 것 (동작 변경)

- [x] 2.1 `internal/journal/a097_rearm_is_a_new_episode_test.go` —
      다른 원인으로 재무장된 행의 `title`·`body`·`payload`가 이번 관측값인가
- [x] 2.2 같은 파일 — 재무장 뒤 `attempts`가 0인가
- [x] 2.3 같은 파일 — 재무장 뒤 `delivered_at`과 `last_attempt_at`이 **비어 있는가**
- [x] 2.4 `internal/obs/a097_claim_failure_blocks_entry_test.go` —
      claim 실패 시 gate가 잠기고 구조화 로그가 남는가
- [x] 2.5 같은 파일 — claim 실패 시 **운영 모드 승격이 시도되는가**
- [x] 2.6 `internal/obs/a097_exclusion_is_an_event_test.go` —
      Flush 뮤텍스: publisher 재진입 카운터가 동시 진입을 잡는가
- [x] 2.7 2.1~2.6이 현재 코드에서 **실패하는 것**을 실행 로그로 확인하고 인용한다

### 2b. 초록으로 도착하는 것 (동작 고정·테스트 경화)

이 항목들은 **RED를 요구하지 않는다.** 가치의 증명은 뮤테이션과 반복 실행이다.

- [x] 2.8 `internal/journal/a097_rearm_is_a_new_episode_test.go` —
      `remindAfter=0`으로 claim한 **미지 상태** 행이 PENDING으로 복구되는가 (R3 고정)
- [x] 2.9 같은 파일 — `remindAfter=0`으로 claim한 **DELIVERED** 행은 재무장되지 않는가
- [x] 2.10 2.8·2.9의 가치를 **뮤테이션으로** 증명: `claimOwed`의 `default`가
      `return true, false`를 돌려주게 바꾸면 2.8만 실패하는가
- [x] 2.11 `internal/obs/a096_one_send_per_condition_test.go` 수정 —
      `TestConcurrentObservationsOfOneConditionSendOnce`가 오통과하지 않게 만든다
- [x] 2.12 2.11의 가치를 **측정으로** 정한다: claim-and-send 잠금을 제거한 뮤턴트에
      대해 각 변형을 `GOMAXPROCS=1`에서 100회 돌려 탐지율을 비교한다.
      **리뷰가 제안한 해법(start barrier)이 듣지 않으면 그렇게 기록하고 듣는 것을 쓴다.**
      실측 결과 — 원본 96/100, barrier만 91/100(개선 없음, 1.4σ),
      **차단 publisher + barrier 100/100** (정상 코드 오검출 0/30).
      따라서 a096 P2 ⑤의 진단(오통과 존재)은 맞고 처방(barrier)은 틀렸다
- [x] 2.13 같은 파일 수정 — `TestAcknowledgeCannotClearTheGateMidSend`의 50ms 시계를
      원장 효과 관측 + 순서 검사로 교체
- [x] 2.14 2.13의 효과도 **측정으로** 판정한다: `Acknowledge`의 `n.mu.Lock`을 제거한
      뮤턴트에 구판·신판을 각각 걸어 본다.
      실측 — 유휴 각 50/50, 20코어 부하 각 30/30. **리뷰가 걱정한 거짓 통과는
      재현되지 않았다.** 이 잠금은 검증된 것으로 세고, 신판의 이점은 탐지율이 아니라
      판정의 종류라고 기록한다

## 3. Pre-Edit 선언 (High-risk — GREEN보다 먼저)

- [x] 3.1 `internal/journal/outbox.go` 편집 직전 Pre-Edit Gate 6항목 기록
- [x] 3.2 `internal/obs/notifier.go` 편집 직전 Pre-Edit Gate 6항목 기록
- [x] 3.3 두 선언에서 안전 불변식 §0 위반 여부를 명시적으로 판정한다

## 4. GREEN — 최소 구현

- [x] 4.1 `ClaimAlertForDelivery` 재무장 UPDATE에 `title`·`body`·`payload`·`attempts=0`·
      `last_attempt_at=NULL`·`delivered_at=NULL` 추가
- [x] 4.2 `claimAndDeliver`의 claim 오류 분기에 구조화 로그 + `Gate.Block` 추가.
      오류는 그대로 반환한다 (삼키지 않는다)
- [x] 4.3 `notifyCritical`의 오류 분기에 `n.escalate` 추가 — **`n.mu` 밖이어야 한다**
- [x] 4.4 `ClaimAlertForDelivery` 문서의 `remindAfter` 문단을 두 규칙으로 나눈다.
      **동작은 바꾸지 않는다**
- [x] 4.5 §2b의 테스트들은 프로덕션 변경 없이 통과해야 한다. 통과하지 않으면 그것은
      경화가 아니라 결함이 하나 더 있다는 뜻이므로 멈추고 기록한다

## 5. VERIFY

- [x] 5.1 `go test ./internal/journal/` 통과
- [x] 5.2 `go test -race ./internal/obs/` 통과
- [x] 5.3 `go test ./internal/execgw/` 통과 (gate 계약)
- [x] 5.4 `go test ./internal/flatten/` 통과 (읽기만 했음을 확인)
- [x] 5.5 `go test ./...` 전체 통과, upstream 회귀 없음
- [x] 5.6 `go vet ./...` 통과
- [x] 5.7 커버리지 GREEN 실측 — journal·obs. 새 분기의 진입/미진입을 프로파일로 직접
      대조한다. **주장하지 말고 측정한다**
- [x] 5.8 뮤테이션 검증 — `Flush`의 `n.mu.Lock`을 제거하면 2.6만 실패하는가.
      변이 전에 파일을 scratchpad로 복사하고 `cp`로 복원한다 (`git checkout` 금지)
- [x] 5.9 `claimOwed`의 BTM 값이 GREEN에서 바뀌지 않았는지 확인 (편집 대상이 아니다)

## 6. 게이트

- [x] 6.1 `python3 tools/logic-map/check_analysis.py --change a097-...` evidence complete
- [x] 6.2 `openspec validate a097-... --strict --no-interactive` valid
- [x] 6.3 proposal-freeze 리뷰 기록 — 적대적 교차모델 판정 BLOCK과 그 반영을 `review.md`에
- [x] 6.4 구현 후 독립 리뷰 1회 — 구현 컨텍스트와 분리된 패스
- [x] 6.5 `make sdd-sync` 결과 기록. advisory 실패는 **명시적으로** 적는다 (침묵 금지)
- [x] 6.6 `make gate CHANGE=a097-a-re-armed-alert-is-a-new-episode` 8/8 통과
- [x] 6.7 PM 동기화 — `python3 tools/pm/generate_master_tracker.py --check`

## 7. 운영 실측 (배포 후, 사람 승인)

이 목록은 **이 change의 완료 조건이 아니다.**

- 재무장된 행을 원장에서 직접 읽어 본문이 최신 원인이고 시각 칸이 비어 있는지 확인.
- claim 실패를 주입한 적 없이 `ENTRY_BLOCKED`나 운영 모드 승격이 일어나지 않는지 확인
  (거짓 잠금·거짓 승격 없음).

## 8. 범위 밖 — 별도로 남긴다

이 목록도 **이 change의 과제가 아니다.**

- **세 동시성 테스트의 결정론적 구성.** 지금의 검증은 뮤테이션 실측이고, 통과 경로는
  세 곳 모두 스케줄러 기회에 의존한다. 구성으로 결정론을 얻으려면 프로덕션에 seam이
  필요하며 그것은 별도 결정이다 (design D4).
- **`Notifier.Acknowledge`에 비테스트 호출자가 없다.** 운영자 해제 경로 자체가 미배선이며,
  그래서 오늘의 실질적 탈출구는 재시작이다. a097은 승격으로 재시작 내구성을 확보할 뿐
  해제 경로를 배선하지 않는다.
- **`Flush`가 `payload`를 보내지 않는다.** `deliver`는 `notificationFor`로 필드를 본문에
  덧붙이지만 `Flush`는 `Title`·`Body`만 쓴다. R1이 payload를 갱신해도 backlog 경로로는 그
  문맥이 전달되지 않는다. spec 시나리오를 title/body로 좁혔다.
- **`Notifier.Flush`에 비테스트 호출자가 없다.** a096 review C3 그대로. a097은 그 경로의
  잠금을 검증하지만 배선하지는 않는다.
- **`Flush`가 `MarkAlertAttemptFailed` 오류를 버린다.** 같은 함수의 기존 문제이며 a096도
  a097도 만들지 않았다.
- **장 운영 시간 게이트 없음.** a096 tasks §7 그대로. 비거래일 익절 발의와 브로커 거절은
  계속된다. 별도 change 필요.
- **event key에 원인이 없다.** a097 R1은 *행*이 원인을 담게 하지만 *key*는 그대로다.
  같은 key의 두 원인은 여전히 한 창을 공유한다.
