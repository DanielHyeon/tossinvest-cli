# Design — a118

## 1. 태그를 없애지 않는다

`tossos_testseams`의 목적은 production 빌드에서 seam을 빼는 것이다. 태그 뒤에는 테스트 파일 20개
말고도 **seam 본체 파일 18개**가 있고, 그것들은 production 바이너리에 들어가면 안 된다.
이 change는 태그를 지우지 않고 **테스트 실행만** 게이트에 배선한다. 테스트 바이너리는 production이
아니므로 이 배선은 태그가 존재하는 이유와 충돌하지 않는다.

## 2. 기대 호출 수는 유도하고 리터럴로 굳히지 않는다

현재 단언은 `want 1`이고, 실제는 21이다. 21을 그대로 써넣는 수정은 두 가지 이유로 거부한다.

**첫째, 21은 계약이 아니라 산술이다.** `Recovery.stableSnapshot`의 AST 열거(분기 7개)를 근거로:
rate limit 팔 B4(`recovery.go:382`)는 `waitOutRateLimit` 호출 뒤 `attempt++` 없이 `continue` 한다
(`:385`의 주석이 그렇게 선언한다). 그래서 루프 B1(`:376`)의 `attempt <= MaxAttempts`는 이 경로를
멈추지 않는다. 멈추는 것은 `Recovery.waitOutRateLimit`의 B1(`ratelimit.go:88`)
`progress.RateLimitWaited + backoff > MaxRateLimitWait`뿐이다. 따라서 호출 수는
`MaxRateLimitWait / RateLimitBackoff + 1`이고, 두 상수 중 하나만 움직여도 값이 변한다.

**둘째, 리터럴은 신호를 죽인다.** 상수를 바꾸는 change가 이 테스트에서 아무 경고도 받지 못한다.
그것이 정확히 이 change가 존재하는 이유의 재발이다.

## 3. "예산을 backoff보다 작게" 대안을 쓰지 않는 이유

`MaxRateLimitWait`를 15초 미만으로 주면 B1이 자기 **전에** 거부하므로 `want 1`이 다시 참이 되고
테스트는 초록이 된다. 쓰지 않는다 — **통과 이유가 달라지기 때문이다.** 그 경로는
`waitOutRateLimit` B2(`ratelimit.go:91`, `RateLimitWaits == 0`)의 다른 메시지로 끝나며,
"예산이 backoff 하나도 못 덮는다"는 별개의 계약이다. 원래 테스트가 증명하려던 것 — 429가 반복될 때
복구가 유한한 예산을 쓰고 fail-closed 한다 — 을 더 이상 증명하지 않는다.
초록으로 만드는 가장 짧은 길이 증명을 지우는 길인 경우다.

## 4. 벽시계 300초를 없앤다

현재 테스트는 `clock.System()`을 넘겨서 20회 × 15초를 **실제로** 잔다. `internal/reconcile`의
a102 계약 테스트가 이미 가상 클럭으로 같은 산술을 검증한다
(`a102_recovery_rate_limit_test.go`, 예산 40s → 대기 2회 → 읽기 3회). 같은 패턴을 쓴다.
300초는 정보가 아니라 비용이다 — 그리고 이 change가 배선을 넣는 순간 그 비용은 매 CI 실행마다 든다.

## 5. 게이트와 CI 둘 다, 그리고 CI는 별도 job

- **완료 게이트**(`tools/gate.sh:256`)에 `test-seams`를 넣는다. change 완료 선언이
  "테스트가 존재하고 통과한다"를 기계적으로 강제하는 유일한 지점이고, 지금은 그 강제가
  무태그 스위트에만 걸린다. change 하나당 한 번 무는 비용이다.
- **CI**에는 주 job과 **병렬로 도는 별도 job**을 만든다. 같은 job에 단계를 더하면 임계 경로가
  그만큼 길어지지만, 별도 job은 러너가 있는 한 벽시계를 늘리지 않는다. 또 실패했을 때
  "무태그가 깨졌는가 태그가 깨졌는가"가 job 이름으로 즉시 갈린다.

## 6. timeout 근거

측정(2026-08-29, `36d6145f`): 태그 스위트 `./...` 합계 **1844초 = 30.7분**. 상위 패키지는
`internal/journal` 551s, `cmd/tossctl` 355s(그중 300s가 이 change가 없애는 실패),
`tools/a112-mb-us-source` 313s, `internal/app/engine` 209s, `internal/execgw` 140s.
단언 수정 후 약 25분이 예상되지만 **예상이지 측정이 아니다.** `make test`의 `-timeout 30m`을
그대로 쓰면 여유가 15% 남짓이고, 새 마이그레이션 하나가 `internal/journal`을 늘리면
(그 타깃 주석이 기록한 a084 전례가 정확히 그것이다) 정당한 실행이 timeout으로 잘못 보고된다.
따라서 `test-seams`의 timeout은 `test`보다 크게 잡고, 값의 근거를 recipe 주석에 남긴다.
확정값은 4.1에서 수정 후 재측정한 뒤 정한다.

## 7. 배선이 실제로 무는지 증명한다

배선 자체가 검증 없이 초록일 수 있다. 태스크 4.4는 태그 뒤 테스트를 일부러 깨뜨린 상태에서
게이트가 **실패하는지** 확인하고 되돌린다. 배선을 넣고 초록인 것은 "무태그가 초록"과 구별되지 않으므로
증거가 아니다.

## 8. 소유권

낡은 단언은 계약을 바꾼 a102(34/34 완료, archive 전)의 것이라고 볼 수도 있으나, 게이트 배선이
저장소 전역이고 **둘의 순서가 계약으로 묶여 있어**(배선을 먼저 하면 CI가 300초를 태우고 빨개진다)
같은 change에 둔다. a102를 다시 여는 비용도 피한다.
