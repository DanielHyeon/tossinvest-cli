# Review — a118-a-sealed-test-still-runs

## 착수 전 상태 (2026-08-29, base `36d6145f`)

`//go:build tossos_testseams` 뒤에 테스트 파일 20개, 최상위 테스트 함수 71개가 있었고
저장소의 어떤 게이트도 그것을 실행하지 않았다. `make test`(`Makefile:35-36`),
CI(`.github/workflows/ci.yml`), 완료 게이트(`tools/gate.sh`) 셋 다 태그를 주지 않았다.

패키지 분포가 `.claude/CLAUDE.md`의 High-risk 정의와 거의 겹친다:
`internal/app/engine` 19, `internal/execgw` 18, `internal/journal` 7,
세 레인 생산 제안 7, `riskbucket`·`risk`·`strategyaccount` 7,
`strategyproposal`·`cmd/tossctl`·`strategyflow` 13.

## 측정 — 부채의 실제 크기는 하나였다

착수 전 전체 태그 스위트를 1회 돌렸다(태스크 1.2).

| 항목 | 값 |
|---|---|
| `ok` 패키지 | 95 |
| 실패 | **1** — `TestActualEngineRecoveryStillFailsClosedOnASnapshot429` |
| 합계 소요(cold) | 1844초 = 30.7분 |
| 최상위 소요 | `internal/journal` 551s, `cmd/tossctl` 355s(그중 300s가 이 실패), `tools/a112-mb-us-source` 313s, `internal/app/engine` 209s, `internal/execgw` 140s |

나머지 70개 함수는 초록이었다. 그래서 이 change 는 "미지의 부채를 파낸다"가 아니라
"단언 하나를 고치고 게이트에 배선한다"로 확정됐고, 그 확정을 **측정 뒤에** 했다.

## 분기 주장보다 AST를 먼저 만들었다

proposal 과 design 이 "루프 상한이 아니라 예산이 멈춘다"를 근거로 쓰기 때문에,
그 문장을 쓰기 전에 세 함수의 묶음을 만들었다.

| 대상 | 분기 | 편집 여부 |
|---|---|---|
| `Recovery.stableSnapshot` (`recovery.go:371-405`) | 7 | 편집 0줄 |
| `Recovery.waitOutRateLimit` (`ratelimit.go:82-113`) | 3 | 편집 0줄 |
| `TestActualEngineRecoveryStillFailsClosedOnASnapshot429` | 15 | 이 change 가 수정 |

열거가 말한 것: B4(`recovery.go:382`) 뒤의 `continue` 는 `attempt++` 를 하지 않는다
(`:385` 가 그렇게 선언한다). 그러므로 429 반복 경로에서 B1 의 루프 상한은 아무것도 멈추지
않고, 멈추는 것은 `waitOutRateLimit` B1(`ratelimit.go:88`)의 예산 검사뿐이다.
읽기 횟수 = `MaxRateLimitWait / RateLimitBackoff + 1` = `5m / 15s + 1` = **21**,
그리고 앞의 20회가 각 15초를 실제로 자서 **300초**가 됐다.

## 수정과 그 결과

| | 전 | 후 |
|---|---|---|
| 결과 | FAIL | PASS |
| 소요 | 300.50s | **0.24s** |

예산과 백오프를 테스트가 명시하고(`budget = 3 * backoff`), 기대 호출 수를
`budget/backoff + 1` 로 유도한다. 안전 단언 4개(`ErrRecoveryIncomplete`, `!report.Complete`,
`report.Snapshot.Empty()`, accounts 1회)는 문구 하나 바뀌지 않았고, `RateLimitWaits`·
실제 대기 횟수·사유 문구 단언 3개가 더해졌다.

## 계획을 실행 중에 정정했다

tasks 3.1 은 처음에 "`RateLimitBackoff`/`MaxRateLimitWait` 기본값을 바꾸는 뮤테이션"이었다.
**그 뮤테이션은 이 테스트를 죽이지 못한다** — 기대값이 같은 상수에서 유도되므로 기대와 실제가
함께 움직인다. 반증 불가능한 뮤테이션을 영수증으로 쓰면 죽지 않는 커버리지 주장이 하나 더
생길 뿐이다. 죽여야 하는 것은 상수가 아니라 계약이라서, 태스크를 계약 뮤테이션 3종으로 바꿨다.

## 뮤테이션 영수증

각 뮤테이션은 `go vet -tags tossos_testseams` 를 **먼저** 통과시켜 가짜 KILL(빌드 실패)이
아님을 확인했다.

| # | 뮤테이션 | 결과 |
|---|---|---|
| M1 | `waitOutRateLimit` B1 의 예산 검사를 참이 될 수 없게 | KILLED |
| M2 | rate limit 팔에 `attempt++` 를 넣어 루프 상한이 멈추게 | KILLED |
| M3 | `progress.RateLimitWaited` 누적 제거 | KILLED |

원복은 심볼 계수(예산 검사 1, "Deliberately no attempt++" 1, 누적 1)와
`git diff --quiet internal/reconcile/` 둘 다로 확인했다.

**유도임을 따로 증명했다(3.4).** 테스트의 예산 배수를 3 → 5 → 7 로 바꿔도 전부 통과한다.
리터럴이 남아 있었다면 실패했을 자리다.

## 배선이 무는 것을 증명했다

배선을 넣고 초록인 것은 "무태그가 초록"과 구별되지 않으므로 증거가 아니다(태스크 4.4).

| 실행 | 결과 |
|---|---|
| 온전한 트리에서 `make test-seams` | **exit 0**, 96 패키지 ok, 실패 0, 571초 |
| 태그 테스트를 일부러 깨뜨린 뒤 `make test-seams` | **exit 2**, 깨뜨린 그 테스트만 실패 |

571초는 **캐시가 더운 실행**이다(`make test` 와 마찬가지로 `-count=1` 을 주지 않는다).
timeout 근거로 쓰는 값은 cold 측정 1844초에서 이 change 가 없앤 300초를 뺀 약 1544초(26분)이고,
`test-seams` 의 `-timeout 45m` 은 거기에 73% 여유다. `make test` 의 30m 를 그대로 쓰면
여유가 15% 남짓이라 `internal/journal` 이 마이그레이션 하나로 늘어나는 순간 —
그 타깃의 주석이 기록한 a084 전례가 정확히 그것이다 — 정당한 실행이 timeout 으로 잘못 보고된다.

## 배선 지점

- `Makefile`: `test-seams` 타깃 추가. timeout 근거를 recipe 주석에 남겼다.
- `tools/gate.sh`: 타깃 목록 `sdd-check test vet validate` → `sdd-check test test-seams vet validate`.
  `TOTAL_STEPS` 를 9 → 10 으로 함께 올렸다. 올리지 않으면 단계 표시가 어긋난다.
- `.github/workflows/ci.yml`: 주 job 과 **병렬로 도는 별도 job**. 임계 경로를 늘리지 않고,
  실패 시 "무태그가 깨졌는가 태그가 깨졌는가"가 job 이름으로 갈린다. YAML 구조 검사 통과.

## 안전 판정

- **생산 코드 편집 0줄.** 바뀐 것은 테스트 1파일과 빌드/게이트 배선 3파일뿐이다.
- 주문·손절·익절·사이징·Guardian·원장·대사·인증·체결 경로 무변경. 토글 이동 없음,
  컨테이너 교체 없음, 서명 매니페스트 무관, 실 API·실계좌 접촉 0(fake `httptest` 서버만 사용).
- 이 테스트가 지키는 대상(부팅 복구의 fail-closed)은 High-risk 이지만, 이 change 는
  그 계약을 바꾸지 않고 **낡은 기대치를 계약에 맞췄을 뿐**이다.

## 인접 발견 — 고치지 않고 기록만 한다

`internal/app/engine/runtime.go:288-297` 은 `Recover` 를 모든 루프보다 **먼저** 완주시키고,
그 루프에 `exit`(손절 관측)이 들어 있다(`cmd/tossctl/engine.go:679-710`). 부팅 시 429 가
지속되면 손절 루프 시작이 최대 5분 늦는다. **안전 불변식 4번의 약화는 아니다** —
`1c76a580`(2026-08-13) 이전에는 같은 429 가 복구를 즉시 끝내 엔진이 아예 뜨지 않았고,
그때는 손절 루프가 아예 돌지 않았다. 상한은 유한하고 초과는 fail-closed 다.
a112 review.md 에도 같은 판정을 남겼다.

## 소유권

낡은 단언만 보면 계약을 바꾼 a102(34/34 완료, archive 전)의 것이다. 게이트 배선은 저장소
전역이라 별개다. 그럼에도 한 change 에 둔 이유는 **둘의 순서가 계약으로 묶여 있기 때문이다** —
배선을 먼저 하면 게이트와 CI 가 매 실행 300초를 태우고 빨개진다. 사용자 판정으로 통합했다.

## 독립 리뷰 (gstack, 2026-08-29)

무태그·태그 스위트가 모두 초록인 상태에서 돌렸다. **교차 모델(codex) 패스가 내 자체 검증이
놓친 결함 두 개를 잡았고, 둘 다 진짜였다.**

### P1 — `test-seams` 가 `.PHONY` 에 없었다

`Makefile:11` 의 `.PHONY` 목록에 새 타깃을 넣지 않았다. `test-seams` 라는 이름의 파일이나
디렉터리가 생기는 순간 `make test-seams` 는 **아무것도 하지 않고 성공한다.** 그러면 새 CI job 도
완료 게이트도 태그 테스트를 한 줄도 돌리지 않은 채 통과한다. 이 change 의 전체 목적을 조용히
무효로 만드는 한 줄이었다. 수정: 목록에 추가.

### P2 — 소진액 단언이 예산과 실제 소진액을 혼동했다

`report.RateLimitWaited != budget` 은 예산이 백오프의 **배수일 때만** 참이다. 마지막 대기는
예산을 넘기므로 실행되지 않는다 — 예산 40초 / 백오프 15초면 실제 소진은 30초지 40초가 아니다.
올바른 기대값은 `wantWaits × backoff` 다.

**내 자체 검증이 이것을 놓친 이유를 기록해 둔다.** 태스크 3.4에서 "유도임을 증명"하려고 예산
배수를 3 → 5 → 7 로 바꿔 전부 통과하는 것을 확인했는데, **셋 다 배수여서** 예산과 소진액이
우연히 같았다. 반증하려고 만든 실험이 반증할 수 없는 값만 골랐다. 정정과 함께 예산을 일부러
배수가 아니게(`3*backoff + 5s`) 바꿔, 이 구분이 이론이 아니라 실행되게 했다.

### P2 — 대기 횟수만 세고 시간은 보지 않았다

`len(clk.waits)` 만 확인하면 "요청보다 짧게 자 놓고 장부에는 백오프만큼 적는" 버그가 통과한다.
기록된 각 대기가 백오프와 같은지 확인하는 루프를 더했다. 선례는
`TestRateLimitDefaultsMatchTheSurveyDiscipline`(`a102_recovery_rate_limit_test.go:254`)이
이미 `got[0] != DefaultRateLimitBackoff` 로 같은 것을 본다.

### 태그 뒤 파일은 `go vet` 을 한 번도 받은 적이 없다

`make lint` 은 `go vet ./...` 하나였다. 무태그 vet 은 태그 뒤 20개 파일을 **빌드조차 하지
않는다.** `make test-seams` 가 컴파일은 시키지만 vet 이 잡는 것(포맷 문자열 불일치 등)은 잡지
못한다. `make lint` 에 `go vet -tags tossos_testseams ./...` 를 더했다 — 새 타깃도 새 게이트
단계도 없이 게이트·CI·로컬 세 경로가 한 번에 덮인다. 실행 결과 exit 0.

### 문서 불일치

`tools/gate.sh:6` 이 "아래 9개 조건"이라고 남아 있었다. 10으로 정정. 단계 번호 자체
(`1/$TOTAL_STEPS` ~ 루프의 6~10)는 일관됐음을 확인했다.

### 정정에 대한 뮤테이션 영수증

| # | 뮤테이션 | 결과 |
|---|---|---|
| M4 | 소진액 단언을 옛 형태(`!= budget`)로 되돌림 | KILLED — 배수가 아닌 예산이 구분을 살렸다 |
| M5 | 시계가 요청의 절반만 자도록(짧은 수면 버그) | 시간 단언을 끄면 **SURVIVED**, 살리면 **KILLED** |

M5는 차등 증명이다. 같은 버그가 단언 유무로 갈리는 것을 보였으므로, 새 단언은 장식이 아니다.

### 검토했고 지적이 아니라고 판정한 것

- **기본값을 더 이상 시험하지 않는다** — 테스트가 예산·백오프를 명시하므로 `DefaultMaxRateLimitWait`
  (5m)와 `DefaultRateLimitBackoff`(15s)를 지나지 않는다. 공백이 아니다:
  `TestRateLimitDefaultsMatchTheSurveyDiscipline`(`a102_recovery_rate_limit_test.go:223-257`)이
  두 상수와 **zero 값 배선**까지 고정한다.
- **`-race`** — 주입한 시계(`a102Clock`)는 뮤텍스가 없지만 이 경로는 단일 goroutine 이다.
  `go test -race` 실행 결과 ok.
- **읽기 횟수 공식** — `int(budget/backoff)` 는 배수가 아닌 예산에서도 옳다. 예산 검사가
  `>` 이지 `>=` 가 아니기 때문이다. 기본값에서 20회 대기 / 21회 읽기로 관측값과 일치한다.
- **CI 별도 job 이 주 job 의 Python 설정을 안 받는다** — `make test-seams` 는 순수 Go 라 불필요하다.
- **게이트가 스위트를 두 번 돈다** — 태그 실행은 무태그의 상위집합이지만(부정 태그 `!tossos_testseams`
  사용처 0건, 96 vs 95 패키지) 무태그 실행은 **production 빌드 구성**을 검증하므로 값이 다르다.
  seam 이 있어야만 통과하는 테스트를 무태그 실행이 잡는다. 둘 다 유지한다. 게이트 소요는
  대략 두 배가 되며, change 하나당 한 번 무는 비용이다.
- **Agent 기반 specialist 는 돌리지 않았다** — 이 세션의 지시가 사용자가 요청하지 않은 서브에이전트
  기동을 금지한다. 구조적 패스와 교차 모델(codex) 패스는 수행했고, 생산 코드 편집이 0줄이라
  security/performance/data-migration/api-contract specialist 는 어차피 scope 로 걸리지 않는다.
  이것은 축소된 커버리지이며, 숨기지 않고 적는다.
