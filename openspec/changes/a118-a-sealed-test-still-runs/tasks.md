## 1. Evidence

- [x] 1.1 `Recovery.stableSnapshot`, `Recovery.waitOutRateLimit`, 그리고 수정 대상 테스트 함수의 AST/Function Logic Map/Branch Test Map/risk-pattern 묶음을 **분기를 근거로 쓰기 전에** 만든다.
- [x] 1.2 전체 태그 스위트(`go test -tags tossos_testseams ./...`)를 1회 돌려 실패 총수와 패키지별 소요를 측정하고 기록한다. 측정 전에는 이 change의 크기를 주장하지 않는다.
- [x] 1.3 태그 뒤 테스트 파일 수와 테스트 함수 수를 세고, 그 패키지 목록을 High-risk 정의와 대조해 기록한다.

## 2. Correct the stale expectation

- [x] 2.1 `TestActualEngineRecoveryStillFailsClosedOnASnapshot429`에 잠들지 않는 클럭을 주입해 벽시계 300초를 없애고, `Stabilise.MaxRateLimitWait`를 테스트가 명시하게 한다.
- [x] 2.2 기대 `/api/v1/orders` 호출 수를 `MaxRateLimitWait / RateLimitBackoff + 1`로 상수에서 유도한다. 리터럴 금지.
- [x] 2.3 기존 안전 단언 4개(`ErrRecoveryIncomplete`, `!report.Complete`, `report.Snapshot.Empty()`, accounts 1회)를 그대로 유지하고, `RateLimitWaits`와 예산 소진 오류 문구 단언을 더한다.
- [x] 2.4 수정한 테스트 함수의 AST 묶음을 편집 후 재생성한다(파일 SHA-256이 디스크와 일치해야 한다).

## 3. Prove the corrected assertions are not decorative

기본 상수를 바꾸는 뮤테이션은 쓰지 않는다 — 기대값이 그 상수에서 유도되므로 기대와 실제가
함께 움직여 테스트가 죽지 않는다. 죽여야 하는 것은 **계약**이다.

- [x] 3.1 `waitOutRateLimit`의 예산 검사(B1, `ratelimit.go:88`)를 참이 될 수 없게 만드는 뮤테이션이 이 테스트를 죽이는지 확인한다.
- [x] 3.2 rate limit 팔에 `attempt++`를 넣어 예산이 아니라 루프 상한이 멈추게 하는 뮤테이션이 이 테스트를 죽이는지 확인한다(`recovery.go:385`의 선언을 뒤집는다).
- [x] 3.3 `progress.RateLimitWaited` 누적을 지워 예산이 영원히 소진되지 않게 하는 뮤테이션이 이 테스트를 죽이는지 확인한다.
- [x] 3.4 기대값이 리터럴이 아니라 유도임을 증명한다: 테스트의 예산 배수를 바꿔도 통과해야 한다.
- [x] 3.5 각 뮤테이션이 **빌드를 깨서** 통과한 것이 아님을 확인하고(가짜 KILL), 원복을 심볼 계수로 검증한다.

## 4. Wire the seam suite into the gate

- [x] 4.1 `Makefile`에 `test-seams` 타깃을 추가한다. timeout은 측정된 소요에 근거해 정하고 그 근거를 recipe 주석에 남긴다.
- [x] 4.2 `tools/gate.sh`의 타깃 목록에 `test-seams`를 넣는다.
- [x] 4.3 CI에 주 job과 병렬로 도는 별도 job을 추가한다.
- [x] 4.4 배선이 실제로 실행됨을 증명한다: 태그 뒤 테스트를 일부러 깨뜨린 상태에서 게이트가 실패하는지 확인하고 되돌린다.

## 5. Verify

- [x] 5.1 `make lint`, `make test`, `make test-seams`, `openspec validate --strict`, `check_analysis.py` 전부 통과.
- [x] 5.2 `make sdd-sync` 후 `make sdd-check` 통과.
- [x] 5.3 독립 리뷰(gstack) 통과.
- [x] 5.4 `make gate CHANGE=a118-a-sealed-test-still-runs` 통과.

## 6. 독립 리뷰가 요구한 정정

- [x] 6.1 `Makefile` 의 `.PHONY` 에 `test-seams` 를 넣는다. 없으면 같은 이름의 경로 하나로 게이트와 CI가 태그 테스트를 한 줄도 안 돌린 채 통과한다.
- [x] 6.2 소진액 단언을 `budget` 이 아니라 `wantWaits × backoff` 로 고치고, 예산을 일부러 백오프의 배수가 아니게 바꿔 그 구분이 실행되게 한다.
- [x] 6.3 대기 횟수만이 아니라 기록된 각 대기 시간이 백오프와 같은지 확인한다.
- [x] 6.4 `make lint` 에 `go vet -tags tossos_testseams ./...` 를 더한다. 태그 뒤 20개 파일은 무태그 vet 이 빌드조차 하지 않아 정적 분석을 한 번도 받지 못했다.
- [x] 6.5 `tools/gate.sh` 헤더의 "9개 조건" 을 10으로 고친다.
- [x] 6.6 6.2와 6.3의 정정이 장식이 아님을 뮤테이션으로 확인한다(M4, M5).
