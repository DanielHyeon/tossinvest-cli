## Why

`//go:build tossos_testseams`를 단 테스트 파일이 **20개**, 그 안의 최상위 테스트 함수가 **71개**다
(2026-08-29, `36d6145f` 실측). `make test`는 `go test -timeout 30m ./...`(`Makefile:35-36`)이고
CI는 정확히 그 타깃을 돌린다(`.github/workflows/ci.yml:28-30`). `tools/gate.sh:256`의 완료 게이트도
`sdd-check test vet validate`만 돌린다. 어느 쪽도 `-tags tossos_testseams`를 주지 않는다.
`grep -rn tossos_testseams Makefile .github/workflows tools/ scripts/` = **0건**.
그 71개는 저장소의 어떤 게이트에서도 **한 번도 실행된 적이 없다.**

그 71개가 어느 패키지에 있는지가 이 제안의 전부다. 목록은 `.claude/CLAUDE.md`의 High-risk 정의
("주문·손절·익절·사이징·Guardian·원장·대사·인증·체결")와 사실상 같다 —
`internal/app/engine` 19(라우트·제안·리스크 권한, dispatch, 첫 다리 인정),
`internal/execgw` 18(Guardian 계좌 베이스, 보호 권한), `internal/journal` 7(원장 투영),
세 레인 생산 제안 7, `riskbucket`·`risk`·`strategyaccount` 7,
`strategyproposal`·`cmd/tossctl`·`strategyflow` 13.

**이것은 추정이 아니라 이미 값을 치렀다.** `TestActualEngineRecoveryStillFailsClosedOnASnapshot429`
(`cmd/tossctl/engine_account_seq_recovery_test.go:169`)는 a102가 재시도 계약을 의도적으로 교체한
2026-08-13(`1c76a580`) 이후 실패해 왔다. 단언은 `93165f96`(2026-07-30) 한 커밋뿐이고 갱신되지 않았다.
a102 자신의 검증 명령(`a102/tasks.md:124`)에 태그가 없어 이 파일이 빌드에서 통째로 빠졌다.
같은 함정이 `a092/review.md:411-417`에 **이미 기록되어 있었고**, 기록만 되고 배선되지 않았다.
a112 L3에서도 무태그 첫 실행이 초록인 동안 태그 테스트 8개가 실패하고 있었다(`a112/review.md` 결정 50).

**부채의 실제 크기는 측정했고, 하나다.** 전체 태그 스위트(`./...`)를 돌린 결과 95개 패키지가 `ok`,
실패는 위 테스트 **정확히 하나**다. 나머지 70개 함수는 지금 초록이다. 따라서 이 change는
"미지의 부채를 파낸다"가 아니라 **"단언 하나를 고치고 게이트에 배선한다"**이다.

**21은 재시도 정책이 아니라 두 상수의 나눗셈이다.** `Recovery.stableSnapshot`의 AST 열거
(`analysis/function-logic/internal-reconcile--recovery.stablesnapshot/ast.json`, 분기 7개)를 근거로:
루프 B1(`:376`)은 `attempt <= MaxAttempts`로 돌지만, rate limit 팔은 B4(`:382`)에서
`waitOutRateLimit`을 부른 뒤 `attempt++` **없이** `continue` 한다(`:385` 주석이 그렇게 선언한다).
따라서 루프 상한은 이 경로를 멈추지 않는다. 멈추는 것은 `Recovery.waitOutRateLimit`의
B1(`ratelimit.go:88`) — `progress.RateLimitWaited + backoff > MaxRateLimitWait`이면
`ErrRecoveryIncomplete`를 돌려준다. 그래서 `Collect` 호출 수는
`MaxRateLimitWait / RateLimitBackoff + 1` = `5m / 15s + 1` = **21**이고,
앞의 20회가 B3(`ratelimit.go:107`)에서 각 15초씩 실제로 자기 때문에
`clock.System()`을 넘겨준 이 테스트는 **300초**를 태운다.

**안전 단언은 전부 통과한다.** 복구는 제대로 fail-closed 된다 — `ErrRecoveryIncomplete`,
`!report.Complete`, `report.Snapshot.Empty()`, accounts 1회. 낡은 것은 호출 횟수 하나뿐이다.

## What Changes

**순서가 계약의 일부다.** 배선을 먼저 하면 게이트와 CI가 매 실행 300초를 태우고 빨개진다.

1. **낡은 단언을 landed 계약에 맞춘다.** `internal/reconcile/a102_recovery_rate_limit_test.go`가 이미
   쓰는 패턴을 적용한다 — 잠들지 않는 클럭을 주입해 벽시계 300초를 없애고, 테스트가
   `Stabilise.MaxRateLimitWait`를 명시하며, 기대 호출 수를 `MaxRateLimitWait / RateLimitBackoff + 1`로
   **상수에서 유도**한다. 기존 안전 단언 4개는 그대로 두고 `RateLimitWaits`와 오류 문구 단언을 더한다.
2. **`make test-seams`를 만들고 완료 게이트와 CI에 배선한다.** `tools/gate.sh:256`의 타깃 목록에
   `test-seams`를 넣고, CI에는 주 job과 **병렬로 도는 별도 job**을 만든다.
3. **timeout을 올린다.** 태그 스위트 총 소요는 측정값 1844초(30.7분)이고 단언 수정 후 약 25분이다.
   `make test`의 `-timeout 30m`을 그대로 쓰면 여유가 없다.

`want 21` 하드코딩은 거부한다. 21은 두 상수에서 유도되는 값이라 상수가 움직이면 썩고,
그 상수를 바꾼 change에게 아무 신호도 주지 않는다 — 이 change가 존재하는 이유가 정확히 그 실패다.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `sdd-workflow`: "테스트 동반 구현"이 요구하는 "통과"는 저장소 자신의 게이트가 실행하는 통과여야 한다.
  build tag 뒤에 있어 게이트가 실행하지 않는 테스트는 그 요구를 만족시키지 못한다.

## Impact

- 수정: `cmd/tossctl/engine_account_seq_recovery_test.go`, `Makefile`, `tools/gate.sh`,
  `.github/workflows/ci.yml`
- 독립 리뷰가 추가로 요구한 것: `make lint` 이 태그를 준 두 번째 `go vet` 을 함께 돌린다.
  태그 뒤 20개 파일은 무태그 vet 이 빌드조차 하지 않아 저장소의 어떤 정적 분석도 받지 못했다.
- **생산 코드는 한 줄도 바뀌지 않는다.** 주문·손절·익절·사이징·Guardian·원장·대사·인증·체결 경로 무변경.
- 토글 이동 없음, 컨테이너 교체 없음, 실계좌 접촉 없음, 서명 매니페스트 무관.
- 태그 뒤 **seam 본체** 파일 18개는 이 change의 대상이 아니다. 테스트 실행만 다룬다.
