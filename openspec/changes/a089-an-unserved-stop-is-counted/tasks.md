# a089 tasks

> **High-risk.** 손절 경로의 함수를 편집한다. Pre-Edit 선언과 proposal-freeze 리뷰
> (적대적 Eng 필수)가 구현 착수 전에 필요하다.
>
> **이 change는 판정·발의·제출의 동작을 바꾸지 않는다.** 바꾸는 것은 계수·시계의 해제
> 조건·장부·기록이다. 어떤 task도 주문의 가격·유형·시점을 건드리지 않는다.

## 0. 게이트 선행

- [ ] 0.1 `capture_change_base.py --change a089-an-unserved-stop-is-counted` (재고정)
- [ ] 0.2 `openspec validate a089-an-unserved-stop-is-counted --strict --no-interactive`
- [ ] 0.3 **proposal-freeze 재리뷰** (적대적 Eng 필수) → `review.md`에 2라운드 추가
- [ ] 0.4 `make sdd-sync` 후 `noteDelay`·`clearDelay`·`submit`·`judge`·`EnqueueAlert`·
      `MarkAlertDelivered`·`classifyStatus`의 definition/callers/impact 확인
- [ ] 0.5 **반증 조회 선행**(리뷰가 남긴 규율): 각 요구사항에 대해 "이 주장이 거짓이라면
      어디에 흔적이 남는가"를 먼저 조회하고 결과를 `issues.md`에 적는다

## 1. 지연 시계의 의미를 하나로 만든다 (D1)

- [ ] 1.0 **Pre-Edit 선언** — `ExitObserver.judge`(1117 분기)와 `ExitObserver.submit`
- [ ] 1.1 **Function Logic Map** — `internal/app/engine/exitloop.go` /
      `ExitObserver.judge`·`ExitObserver.submit`·`noteDelay`·`clearDelay`
      (`ast.json`은 저장소 상대 경로)
- [ ] 1.2 **Branch Test Map** — 위 표 P1~P9 각 행에 테스트 1건
- [ ] 1.3 **RED (C1 회귀)** — `clearTheSymbol`이 성공해도 보호 제출이 거부되면
      시계가 해제되지 않는다. **이 테스트가 첫 판의 D1을 죽인 결함을 잡는다**
- [ ] 1.4 **RED** — P1~P9 각 경로에서 시계가 시작·지속된다
- [ ] 1.5 **RED** — `StateConfirmed`가 해제한다 / 포지션 종료가 해제한다
- [ ] 1.6 **RED** — `StateInDoubt`는 시작도 해제도 하지 않는다
- [ ] 1.7 **RED** — 익절 제안은 시계에 들어가지 않는다(a074 주석의 위험 회귀)
- [ ] 1.8 **RED** — 한계 초과 시 critical 1회, `delayAlerted` latch로 반복 없음
- [ ] 1.9 **GREEN** — 주기 끝에서 불변식을 한 지점으로 판정. 분기마다 손으로 넣지 않는다
- [ ] 1.10 포지션이 working set에서 빠질 때 맵 항목이 지워지는지 — 누수 회귀 테스트

## 2. 연속 미제출을 센다 (D2)

- [ ] 2.1 **RED** — P1~P9에서 +1, `StateConfirmed`에서 0, 종료에서 삭제
- [ ] 2.2 **RED** — 계수와 최초 미제출 시각이 시계와 **같은 조건**에서 갱신·초기화된다
      (두 값이 어긋나면 실패하는 테스트)
- [ ] 2.3 **RED** — `exit.proposal_refused`·`exit.liquidation_delayed` 필드에 계수가 실린다
- [ ] 2.4 **GREEN** — `ExitObserver`의 기존 맵과 같은 자리, 프로세스 메모리
- [ ] 2.5 재시작 시 0으로 돌아가는 성질을 **주석과 `issues.md`에 명시**. 스키마 미변경

## 3. outbox 재발 (D3)

- [ ] 3.1 **Function Logic Map** — `internal/journal/outbox.go` / `EnqueueAlert`
- [ ] 3.2 **RED** — `DELIVERED` 행 재발 → `PENDING` 복귀 + 제목·본문·payload 갱신
- [ ] 3.3 **RED** — `ACKNOWLEDGED` 행 재발 → `PENDING` 복귀
- [ ] 3.4 **RED** — 재발 후 `MarkAlertDelivered`가 성공하고 `attempts`가 누적된다
- [ ] 3.5 **RED** — 재발의 전달 실패가 `UndeliveredCount`에 잡히고 `PendingAlerts`에 뜬다
- [ ] 3.6 **RED** — `Acknowledge`가 재발 중인 행을 남겨 둔 채 게이트를 풀지 않는다
- [ ] 3.7 **RED (회귀)** — 첫 발생의 동작·반환 id·행 개수가 무변화
- [ ] 3.8 **GREEN** — `EnqueueAlert` 한 함수. `notifier.go`와 다른 outbox 함수는 무편집
- [ ] 3.9 `no such alert: N (or it is no longer pending)`이 재발 경로에서 더는 나오지 않고,
      **행이 실제로 없을 때만** 나오는지 확인 — 오류 문구의 이중 의미 해소
- [ ] 3.10 재발이 잦은 다른 critical 이벤트에 대한 영향 점검. 발송량 무변화를 테스트로 고정

## 4. 브로커 사유 기록 (D4)

- [ ] 4.1 **RED** — 8/5 실제 응답 본문 fixture에서 `code`·`data.field` 추출
- [ ] 4.2 **RED** — 사유가 없는 본문·JSON이 아닌 본문에서 빈 값이고 오류가 아니다
- [ ] 4.3 **RED** — 원문 `detail`이 보존된다
- [ ] 4.4 **RED** — 기록된 사유가 제안·재발의·제출의 동작을 바꾸지 않는다
      (같은 사유 두 종류로 동일 경로를 타는지)
- [ ] 4.5 **GREEN** — `official.APIError`에 가산 메서드. `classifyStatus`·`ShouldFallback`
      및 기존 타입은 무편집
- [ ] 4.6 §0.8 — 옮기는 값에 계좌·세션·개인정보가 없음을 확인

## 5. 실측 재생

- [ ] 5.1 2026-08-05의 여섯 판정·다섯 거부를 fixture로 재생. 확인할 것:
      알림이 00:55:02에 경과 62초로 **한 번**, 마지막 거부의 계수가 **5**,
      outbox 행의 `attempts`가 **5**
- [ ] 5.2 D1의 받아들인 결과 재생 — 손절 1회 거부 후 가격이 회복해 며칠 뒤 익절로
      끝나는 포지션이 그 사이 알림을 **한 번만** 내는지
- [ ] 5.3 재생 결과를 `issues.md`에 기록

## 6. 게이트

- [ ] 6.1 `go test ./... -count=1 -race` 회귀 0, upstream 650 green
- [ ] 6.2 §0.3 확인 — 판정·발의·제출의 **시점과 내용이 무변화**임을 diff로 보인다
- [ ] 6.3 §0.4 확인 — 브로커 요청 수 무변화(재시도 미추가)
- [ ] 6.4 `make sdd-sync` 재실행 → `make sdd-check`
- [ ] 6.5 **격리 worktree에서** `make gate CHANGE=a089-an-unserved-stop-is-counted`
- [ ] 6.6 독립 검증 (구현과 분리된 컨텍스트)
- [ ] 6.7 PM 동기화 → `openspec archive`

## 선후 관계

| change | 관계 |
| --- | --- |
| a087 (보호 청산의 가격) | **독립.** a089는 가격을 건드리지 않고, a087은 계측을 필요로 하지 않는다. a087이 8/5의 다섯 거부를 원천 제거하고, a089는 여섯 번째 원인이 왔을 때 그것이 보이게 한다 |
| a088 (익절 호가 그리드) | 독립 |
| 2c (브로커측 보호주문) | 지연 알림 본문이 이미 2c를 참조한다 — "브로커에 상주하는 손절이 생기기 전까지 이 지연은 보호되지 않은 노출이다" |
