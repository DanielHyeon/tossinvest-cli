# a112 0.5 전체 diff 리뷰 — voice1 (testing) SubagentHandback 원문 (agent a2f93a2390322cb6b, 저장 2026-10-05)

안전 규칙 위반 없음(`~/.codex` 와 `_codex*/` 는 열지 않음, 저장소 편집 0, 모든 실험은 4cbcfb36 `git archive` 격리 사본에서만 함).

**전체 판정: SHIP-WITH-FIXES** — P0 0건. P1 2건은 모두 시험이 주장을 받치지 못한다는 증거 문제이고, 생산 결함은 재현하지 못함.

| # | 등급 | file:line | 원문 인용 | 문제 | 재현/변이와 결과 | 고치는 법 |
|---|---|---|---|---|---|---|
| 1 | P1 | internal/app/engine/a112_shadow_cycle_test.go:213, :227 | `partial := strategyrouter.FamilyActivationForTest(... KRContinuationLaneID: true})` / `if got, want := with.trace(t), without.trace(t); got != want {` | 차등 dispatch 의 「부분 ON」 배치가 빈 표본끼리 비교함. 두 실행 모두 dispatch 가 0 이라 0 == 0 으로 통과함. 브리프 §8 은 「0 은 틀린 척도 — ON 레인이 있으면 정상 dispatch 가 있다」 고 적음. 전제 단언(:222)은 WOULD_EMIT ≥1 만 보고 dispatch ≥1 은 보지 않음. | 사본에서 t.Logf 를 넣어 2회 실행. 두 번 다 `partial ON trace with={observed:0 placed:0 leases:0 laneLatches:0 latchRecoveries:0} without={…전부 0}`. undeclared 배치는 `placed:1 leases:1` | 부분 ON 활성화가 dispatch 픽스처 제안을 소유한 가족을 켜게 함. 전제로 `without.trace(t).placed >= 1` 을 단언함 |
| 2 | P1 | internal/app/engine/a112_shadow_structure_test.go:608 (변이 자리 strategy_lane_shadow.go:195 뒤) | `forbidden := []string{"refreshPaired…", …, "Submit", "dispatch", "Record", "Recover", "Place", …}` | 단계 폐포 핀이 금지 단어 목록(blacklist) 방식이라 레인 상태를 바꾸는 호출(`Offer`·`Fail`·`Run`)을 보지 못함. 성공 경로의 행동 시험도 레인 상태 불변을 재지 않음. 「SHADOW 는 레인 상태를 바꾸지 않는다」(worker/shadow.go 의 ShadowOutcomeOver 주석)가 강제되지 않음. | runShadowStep 루프에 `_ = lane.Offer()` 를 넣고 SET_731S 시험 6명령을 그대로 돌림. **6/6 ok — 변이 생존**. 이 변이는 OFF 레인의 pending 을 올리고, 레인이 나중에 ON 이 되면 queueDepth 에서 TriggerFull 로 진짜 투입이 버려짐. `lane.Fail(...)` 변이는 deadline 하위시험 하나만 우연히 잡음 | 단계 폐포 핀을 허용 목록(allowlist)으로 바꿈: strategyworker.Lane 메서드 중 Key·ShadowEligible·ShadowOutcomeOver 만 허용. 성공한 단계의 앞뒤로 8 레인 `Status()` 가 같다는 행동 단언을 추가함 |
| 3 | P2 | a112_shadow_cycle_test.go:129 · strategy_lane_shadow.go:154, :164 | `busy := runtime.shadowInFlight[market]` / `runtime.shadowInFlight[market] = false` / `runtime.publishShadow(market, …)` | 대기 helper 가 기다리는 표식(in-flight 해제)은 안쪽 goroutine 이 결과 전송 직후 내리고, 게시는 그 뒤 감독 goroutine 이 함. helper 가 게시 전에 돌아올 수 있음 → 양성 단언은 거짓 실패, 음성 단언은 변이 거짓 통과 가능 | publishShadow 앞에 10ms 지연(의미 변화 없음)만 넣어도 양성 시험 3개(ProjectsWouldEmit · PinLeavesTheDispatchTrace · TheGapBetween) FAIL. S18 변이 단독은 30/30 잡힘, S18 + 지연이면 deadline 하위시험 10/10 통과(변이 생존). 원본은 150회·-race 8회 모두 0 flake → 잠복 위험 | 감독 goroutine 종료를 알리는 test-seam 신호를 두고 helper 가 그 신호를 기다리게 함 |
| 4 | P2 | a112_shadow_cycle_test.go:276 | `time.Sleep(20 * time.Millisecond)` | deadline 하위시험이 실시간 20ms 안에 감시견 select 가 마감 갈래를 소비한다고 가정함. 부하 상태에서 결과·마감이 동시에 준비되면 select 가 무작위로 고름 → 게시 → 실패 | 감시견이 깨어난 뒤 마감 전달 전에 50ms 지연을 넣음 → 20/20 FAIL `a failed shadow step left 2 SHADOW lanes` | close(release) 전에 감독 goroutine 종료를 seam 으로 기다림 |
| 5 | P2 | internal/strategyshadow/shadow.go:238 · loader_order_test.go:65 | `); len(fields) != 0 {` / `calls = []string{normalize.Replace(exprText(guard.Cond))}` | shadow 적재기의 설정 결속 거절(owner·market·config_dir·observed_at·digest·calibration…) 시험이 0건. 순서 AST 핀은 호출이 있는 조건에서 연산자를 안 보므로 `&& false` 를 못 봄. 활성화 쪽 원본에는 같은 거절 시험이 있음 | `len(fields) != 0 && false` → 로트 시험 전부 ok(생존). 같은 변이를 활성화 적재기에 넣으면 TestEveryActivationRefusalNamesItsFieldAndKeepsItsSentinel 의 config binding 하위시험 9개 FAIL. lifetime·actor 하위검사(`issued.After(now) && false` · `issued.Before(approved) && false` · actor `&& false`)도 생존 | 활성화의 config-binding · lifetime 거절 사례를 shadow_test 로 옮겨 적음(옮겨 적은 코드는 양쪽을 못 박음) |
| 6 | P2 | internal/strategyshadow/shadow_test.go:187 | `{"unknown shadow state", run(marshal(unknownState), nil), ErrProductionFamilyShadowUnavailable, "shadow"},` | 필드명 단언이 공허함. sentinel 문장 자체에 「shadow」 가 들어 있어 어떤 오류든 통과함(verified 단언이 대신 잡고 있음) | 원문 대조: `"strategyshadow: family shadow unavailable"` | 단언 문자열을 `descriptors[0]: shadow` 로 바꿈 |
| 7 | P2 | a112_shadow_cycle_test.go:302–331 | `baselineUSObserved, baselinePlaced := baseline.spy.observed["protection-us"], …` | 교차 시장 시험에서 US 주기가 기준·실험 둘 다 같은 실패로 끝남. US 배치가 0 이라 「US dispatch 불변」 은 「같은 오류」 만 잼 | 로그: `errUS=…ATOMIC_ADMISSION_FAILED … BUCKET_USAGE_STALE … ledger holds 808`, `placed=1/1`(KR 것). 단, 잠금을 쥔 채 멈추는 변이는 US 주기 교착(60s timeout, :323)으로 잡힘 → 「막지 않음」 축은 유효함 | US 위험 버킷 스냅숏을 따로 둬 US 기준 배치 ≥1 을 전제로 단언함 |
| 8 | P2 | a112_shadow_absent_restart_test.go:99 | `if id, ok := value.Type.(*ast.Ident); ok && id.Name == typeName {` | 어휘 census 가 타입 있는 const 만 셈 | `var LaneRuntimeLive LaneRuntime = "LIVE"` 추가 → census · Shadow 시험 모두 ok. Validate 의 `LaneRuntimes()` 멤버십이 실제 관문이라 영향은 낮음 | var 선언과 타입 없는 그룹 const 도 셈 |
| 9 | P2 | tools/a112-family-shadow/main.go | `func parseFamilies(value string) …` | 저작 도구에 시험 0건(parseMarket · parseFamilies · O_EXCL · 0400). golden_test 는 손으로 만든 Document 로 Encode 만 부르고 도구 경로는 부르지 않음 | `ls tools/a112-family-shadow` → main.go 하나. 활성화 도구도 같은 관행 | `run()`/`options.document()` 를 부르는 시험으로 golden 바이트를 재생산함 |
| 10 | P2 | analysis/measurements/lot-7.3.1-shadow/mutation-7.3.1-S.tsv (S19) · harness/a112_lot_mutate.py `run_tests` | `S19 … CAUGHT 0 failing:  \| FAIL` | 하네스가 실패 이름 없는 비영 종료(프로세스 crash)도 RED=CAUGHT 로 적음. 어느 시험이 잡았는지 귀속 불가 | 원장 원문. 다른 CAUGHT 표본(S16 · S17 · S18 · S21 · S26)은 재대조해 정상 | crash 는 별도 판정(CRASH)으로 적고 panic 스택의 시험 이름을 기록함 |
| 11 | P2 | lot-7.3.1-shadow/branch-coverage-engine.json | 묶음 13개가 전부 기존 함수의 FLM 묶음 | 새 shadow 함수의 분기 진입은 영수증에 없음 | 직접 잰 커버리지에서 진입 0 블록: strategy_shadow_batch.go:38(boundTo 부재 갈래) · strategyshadow/shadow.go:137, 142, 171(인코더 거절) · 216, 219(ctx nil · 취소) · 238(#5) · 262 · 271 · 281 | 새 함수도 측정에 넣음 |

**확인했으나 문제 없음**
- 경계의 등호 쪽: 늦은 nil 주기는 한도 −1ns 면 단계 시작, 정확히 한도면 실패(S13 · S12b 가 잡음). 투영 만료는 −1ns 면 SHADOW, 정각이면 UNOBSERVED. 나이 74s 도 같은 규칙. 적재기 만료 정각(L02). breakout RVOL 1_199_999 / 1_200_000 / 1_499_999. wick 은 `<=` → `<` 변이를 넣어 시험 2개가 FAIL 로 잡음.
- 가짜 시계 sleeper 수 가정: 감시견 등록을 5ms 늦춰도 deadline 하위시험 통과. 남는 sleeper 가 없어 `WaitForSleepers(1)` 가정이 성립함.
- S21 의 「동등」 주장: 투영의 파도 등식(:47)이 이전 물결 관측을 버리므로 성립함.
- 변이 하네스는 소스 동결 시험을 `-skip` 으로 빼고 돌림. 그래서 L 계열 CAUGHT 는 행동 시험이 잡은 것임(SET_731S_TESTS 확인).
- safety cadence 시험은 synctest 가상 시계라 flaky 위험 없음.
- 경쟁 검사: shadow 엔진 시험 10개를 `-race -count=8` 로 돌림 → ok(692s). 양성 시험 `-count=150` → 0 flake.
- 활성화 적재기 설정 결속에 같은 `&& false` 변이 → 하위시험 9개가 잡음.
- 투영 검증기 거절 사례는 각각 다른 규칙 하나만 깨는 최소 입력이라 판별력 있음(V01 · V02).

**저장소 무변경**
- 시작 HEAD `4cbcfb36365c5910a681883841edb9e7fad9cc04`, 끝 HEAD 동일.
- `status --short` 는 시작과 끝이 같음: `M …/repin-1e25b3a3/README.md`, `?? .reticle-setup-crash.log`, `?? …/function-logic/internal-strategyrouter--allfourfamiliesfortest/`, `?? …/check-analysis-at-4cbcfb36.txt`, `?? …/window-receipt-1e25b3a3-at-4cbcfb36.tsv`, `?? …/review-0.5-full-2026-10/`, `?? w4.log`.
- 격리 사본(scratchpad/c4cb)은 삭제함.
- 증거 파일은 /tmp/claude-1000/-mnt-D-Axipient-workspace-TossOS/4b178b0a-b918-4419-8142-f266a03d8e82/scratchpad/ 에 있음: set731s.sh, race8.txt, flake1.txt, cov-engine.out, cov-shadow.out, cov-worker.out.
