# a112 0.5 전체 diff 리뷰 — voice4 (performance) SubagentHandback 원문 (agent aee74bdbdd2b4ba17, 저장 2026-10-05)

~/.codex 와 리뷰 디렉터리의 `_codex*/` 는 열지 않았다. 안전 규칙 위반 없음.

**전체 판정: SHIP.** 이 범위(성능 · 동시성 · 누수)에는 P0/P1 이 없다. P2 둘은 시험 시간 문제다.

| # | 등급 | file:line | 원문 인용 | 문제 | 재현/변이와 결과 | 고치는 법 |
|---|---|---|---|---|---|---|
| 1 | P2 | internal/app/engine/a112_shadow_structure_test.go:68-71 | `func a112CheckedEngine(t *testing.T) testenv.CheckedPackage { … return testenv.TypeCheckProduction(t, ".", a112EnginePath) }` | census 시험 7개가 각자 engine 패키지 전체(의존 포함)를 source importer 로 다시 타입 검사한다. 시험 사이 캐시가 없다. | 4cbcfb36 사본에서 `go test -tags tossos_testseams -run <구조 시험 8개> -json` 실행. 시험마다 5.63~5.96s, 전체 wall 40.9s, RSS 346MB. overlay 로 `sync.Once` 캐시만 넣으면 첫 시험 6.23s, 나머지 6개 0s, 8/8 PASS(판정 불변), wall 10.0s. 시험당 약 34s 절감. | 패키지 수준 `sync.OnceValues` 로 한 번만 검사한다. 단, 지금 `TypeCheckProduction` 은 안에서 `t.Fatalf` 를 부르므로 Once 안에서 실패하면 첫 시험에 귀속되고 나머지는 영값을 받는다. 그래서 error 를 돌려주는 변형을 만들어 각 시험이 저장된 오류로 실패하게 해야 한다. |
| 2 | P2 | Makefile:110 (+ :106 `RACE_ENGINE_TESTS`) | `go test -race -timeout 5m -count=1 -tags tossos_testseams -run '$(RACE_ENGINE_TESTS)' $(RACE_ENGINE_PACKAGE)` | 목록이 base 18개에서 head 74개로 늘었다. shadow 시험 하나가 -race 아래 15~17s 걸린다. 5분 한도를 이미 70% 쓰고 있어, 코어가 적은 CI 에서는 시간 초과로 흔들릴 위험이 있다. | 사본(i9-12900H, 20 스레드)에서 그 명령 그대로 실행. 74/74 PASS, DATA RACE 0, 패키지 210.4s(wall 211s). 가장 긴 시험: TestTheOrderLeaseCannotOutliveTheFamilyActivation 20.7s, TestTheShadowPinLeavesTheDispatchTraceUnchanged 17.1s, TestAFailingShadowStepChangesNothingButItsOwnObservation 16.7s, TestARestartNeverRestoresShadow 16.4s. | 한도를 10m 로 올리거나, shadow 시험이 매번 새로 만드는 SQLite dispatch fixture 를 공유한다. CI 실측으로 여유를 정할 것. |

**확인했으나 문제 없음** (전부 4cbcfb36 격리 사본에서 실측. 측정 파일은 사본에만 있고 저장소에는 넣지 않았다)

- **(a) dispatch 앞에 새로 들어온 일.** 벤치마크 결과:
  - `shadowConfig`: 2.0~2.2µs/op, 할당 17회. 대부분은 `strategyRuntimeBuildDigest` 의 `debug.ReadBuildInfo` + sha256(1.8µs).
  - 8.5 의 관문 재계산 1회 추가(미선언 시장): 2.2µs. base 에는 `familyGateFor` 호출이 :371 한 곳뿐이었고 head 는 :337, :399 두 곳이다.
  - 핀이 선언된 시장의 활성화 적재 1회: 24~26µs, 12KB, 할당 84회. 소형 로컬 파일 기준.
  - `record` 의 칸 대입: 87ns, 할당 0.
  - cycle 클로저 꼬리(`strategyLanesMu` + `startShadowStep`): 144ns.
  - schedule 재검증은 판정만 순수 함수로 옮긴 것이다. diff 상 수집 I/O 는 base 와 같다.
  - 합계는 시장 주기당 수 µs~수십 µs 다. 5s 폴 간격이나 원격 I/O 에 비하면 무시할 수준이다.
  - 7.5 D1 의 레인 병렬화는 dispatch 앞 지연의 최악값을 레인 합에서 최댓값으로 줄인다. 레인당 고루틴 1개를 쓰고 join 한다.
- **(c) 메모리 · 복사.**
  - 크기: `Proposal` 1832B(= `ShadowInput`), `strategyShadowBatch` 184B, `strategyShadowCell` 272B.
  - collect 비용: 제안 8개 6.1~6.7µs/30KB, 64개 47µs/260KB, 512개 0.5ms/2MB.
  - 보관은 시장당 칸 하나로 덮어쓴다. 더 붙잡는 것은 직전 물결 하나뿐이고, map 은 시장 2개 기준이라 무한히 자라지 않는다.
  - 선택 사항: 칸 용량을 `batch.Len()` 으로 미리 잡으면 할당이 약 절반으로 준다. 다만 그 자리는 AST 로 모양이 고정돼 있다.
- **(b) 고루틴 수명.**
  - M1: Fake 시계로 빠른 단계 500 주기를 돌렸다. 500/500 nil, 고루틴 3 → 3. 감시견 sleeper 는 `cancel()` 과 `Fake.drop` 으로 정리된다.
  - M3: 실제 시계로 200 단계. 고루틴 3 → 3, 타이머 고루틴 잔류 없음(`systemClock.Sleep` 이 ctx 를 따르고 `timer.Stop` 한다).
  - M2: ctx 를 무시하는 적재가 멈춘 채 300 물결을 돌렸다. 적재 1회, 건너뜀 299, 고루틴은 마감 전 +3, 마감 후 +1, 해제 후 기준으로 복귀. 단일 비행이 상한을 지킨다.
  - 멈춘 단계가 돌아오지 않으면 그 시장의 shadow 는 프로세스 수명 동안 꺼진다. 설계대로이고, 고루틴은 시장당 1개로 묶인다.
  - `result` 와 `deadline` 채널이 버퍼 1이라 송신 쪽 고루틴이 막히지 않는다.
  - 감독(supervisor) ctx 는 주기마다가 아니라 `Run` 의 childCtx 다. 그래서 주기가 끝난 직후 shadow 단계가 즉시 취소되는 일은 없다. 종료 시에는 ctx 취소로 곧 끝난다. 단, `Run` 의 WaitGroup 이 shadow 고루틴을 join 하지는 않는다. 이 부분은 추론이고 재현하지 않았다.
- **(b) 잠금 순서와 교착.**
  - `strategyLanesMu` → `runtime.mu` 를 동시에 쥐는 곳은 `productionStrategyLanes` → `restoreLatches` 한 곳이다. 원장 I/O 는 첫 호출에만 일어난다.
  - `strategyLaneRuntimeIfAny` 와 `Read` 는 `strategyLanesMu` 를 놓은 뒤에 `runtime.mu` 를 잡는다.
  - 투영은 `runtime.mu.RLock` → `lane.mu`(`Status`) 순서다. 반대 순서는 없다: 레인 step 은 runtime 에 닿지 않는다.
  - 투영 안에서 RLock 을 다시 잡는 재귀 RLock 은 없다. `lanesFor` 를 부르지 않고 필드를 직접 읽는다.
  - shadow 의 수동 Unlock 구간(strategy_lane_shadow.go:119-133, :153-155)에는 panic 날 거리가 없다. 생산 생성자가 map 다섯을 모두 만든다.
  - M7 스트레스(-race, 3s): 주기 5855 · 투영 92631 · 폐기 12780 · 단계 시작 17898 을 동시에 돌렸다. 교착 0, DATA RACE 0.
- **-race 전수.**
  - shadow 와 레인 시험 22개 × count=3: PASS, 360s.
  - strategyworker · strategyshadow · strategyprojection -race: PASS.
  - Makefile 의 엔진 race 목록 74개: PASS, DATA RACE 0.
- **(d) 그 밖의 census 비용.** `ListDeps`(go list, 걸음 4회)는 시험당 0.05~0.66s 다. 패키지별 `TypeCheckProduction` 1회는 strategyproposal 2.85s, officialbars 2.38s 다. 문제는 #1 의 engine 7회 반복뿐이다.

**저장소 무변경.** 시작과 끝 모두 HEAD `4cbcfb36365c5910a681883841edb9e7fad9cc04` 이고 `status --short` 가 같다(` M …/repin-1e25b3a3/README.md`, `?? .reticle-setup-crash.log`, `?? …/allfourfamiliesfortest/`, `?? …/check-analysis-at-4cbcfb36.txt`, `?? …/window-receipt-1e25b3a3-at-4cbcfb36.tsv`, `?? …/review-0.5-full-2026-10/`, `?? w4.log`). 실험 파일은 스크래치 사본에만 있다:
- `/tmp/claude-1000/-mnt-D-Axipient-workspace-TossOS/4b178b0a-b918-4419-8142-f266a03d8e82/scratchpad/copy/internal/app/engine/zz_perf_review_testseam_test.go`
- `/tmp/claude-1000/-mnt-D-Axipient-workspace-TossOS/4b178b0a-b918-4419-8142-f266a03d8e82/scratchpad/copy/internal/strategyrouter/zz_perf_review_test.go`
- overlay: `/tmp/claude-1000/-mnt-D-Axipient-workspace-TossOS/4b178b0a-b918-4419-8142-f266a03d8e82/scratchpad/overlay.json`, `/tmp/claude-1000/-mnt-D-Axipient-workspace-TossOS/4b178b0a-b918-4419-8142-f266a03d8e82/scratchpad/structure_cached.go`
- race 결과: `/tmp/claude-1000/-mnt-D-Axipient-workspace-TossOS/4b178b0a-b918-4419-8142-f266a03d8e82/scratchpad/race_engine.json`
