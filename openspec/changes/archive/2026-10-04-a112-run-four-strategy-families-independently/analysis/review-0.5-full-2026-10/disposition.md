# a112 0.5 전체 diff 프리랜딩 리뷰 — 4판 합본 처분 (2026-10-05)

좌표: base `1e25b3a3` → 착지 `4cbcfb36`. 리뷰 출력 원문은 이 디렉터리 `voice1-testing-output.md` · `voice2-maintainability-output.md` ·
`voice3-security-output.md` · `voice4-performance-output.md`. 판정: voice1 SHIP-WITH-FIXES · voice2 SHIP-WITH-FIXES · voice3 SHIP-WITH-FIXES · voice4 SHIP.
**P0 0건(네 판 모두).** P1 3건(시험#1 · 시험#2 · 보안#1)은 모두 「시험이 주장을 못 받침」 — 생산 결함 재현 0.

Manager 합본 판정: HOLD → 응답 로트 → SHIP. 응답 로트의 생산 편집은 2줄(유지#4=보안#2 `%w`→`%v` · 보안#4 `os.Remove`)이고, 나머지는 시험 · 시험 seam · 문서 ·
하네스다. 증거는 `analysis/measurements/lot-0.5-response/`. 변이 원장 `mutation-0.5-R-run1.tsv`(28 정의) · `mutation-0.5-R-run2-fixed.tsv`(고친 5 재실행).

## 처분 표

| 리뷰 # | 등급 | 처분 | 고친 자리 | 재확인(변이 · 측정) |
|---|---|---|---|---|
| 시험#1 | P1 | 고침(시험) | `a112_shadow_cycle_test.go` `newA112ShadowWorld` — 활성화된 KR 묶음에 `ProposalSetDigest` 를 조립과 같은 식으로 적음 · `TestTheShadowPinLeavesTheDispatchTraceUnchanged` 에 전제 `without.trace(t).placed >= 1` | 부분 ON 이 실제로 dispatch(편집 전 placed 0 → 1). 두 하위시험 PASS |
| 시험#2 | P1 | 고침(시험) | `a112_shadow_structure_test.go` 단계 폐포 핀에 strategyworker **허용 목록**(Lane.Key · Lane.ShadowEligible · Lane.ShadowOutcomeOver, 셋 다 닿아야 함 + 주기 함수 양성 대조) · 새 `TestASuccessfulShadowStepLeavesEveryLaneStatusUnchanged`(성공 단계 전후 8 레인 `Status()` 같음, 전제 WOULD_EMIT ≥ 1) | A01 `lane.Offer` CAUGHT(행동 + 구조) · A02 `lane.Fail` CAUGHT |
| 보안#1 | P1 | 고침(시험) | `TestTheCoordinatorCollectsWithOneStatement…` 에 지역 객체 자리 전수 핀: coordinate `shadow` = 정의 · collect · return 셋, collectMarket `shadow` = 매개변수 · 두 대입 셋, `collected` = 정의 · 사용 둘(types 객체로 셈) | C01(M4 묶음 크기가 관문을 내림) · C02(M2) · C03(collectMarket 추가 읽기) CAUGHT |
| 유지#4 = 보안#2 | P2 | 고침(**생산 1줄**) | `production_family_activation.go` `decodeProductionFamilyActivation` 둘째 `%w` → `%v` · 새 `TestActivationDecodeFaultCarriesOnlyTheUnavailableSentinel` + shadow 짝 `TestShadowDecodeFaultCarriesOnlyTheUnavailableSentinel` | RED `red-decode-identity.log`(4/4 FAIL) → GREEN · D01 · D02 CAUGHT · BTM B2 갱신 |
| 보안#4 | P2 | 고침(**생산 도구 1자리**) | `tools/a112-family-shadow/main.go` 쓰기 · 닫기 실패 시 `os.Remove(opts.out)` | 시험 없음 — 정규 파일 쓰기 실패 주입 seam 없음(남은 위험, `branch-coverage-new-functions.txt` 처분 절) |
| 유지#1 | P2 | 고침(시험 · 주석) | 새 `a112_shadow_binding_test.go` — shadowConfig 결속으로 활성화 매니페스트를 써서 **생산 활성화 적재기**가 받는지 KR · US 로 봄(활성화 무편집, 달력 원천 둘을 일부러 벌림) · `shadowConfig` 주석의 「같은 env 이름」 정정 | B01 · B02 · B03b · B04 · B05b(M1 세 필드) CAUGHT |
| 유지#2 | P2 | 고침(시험) | 새 무태그 `TestTheProductionLoadShadowIsTheProductionLoaderOutsideTestSeams`(laneStepFor 핀 방식 — 정의 수 2 · 태그 · 몸통) | L01(M4 몸통 비움) CAUGHT |
| 유지#3 | P2 | 고침(시험) | 새 `TestTheProjectionShadowsOnlyAnOffOffObservation`(네 조합 직접 호출 — desired ON/effective OFF 포함) | P01(M2 Desired 항 삭제) · P02(Effective 항 삭제) CAUGHT |
| 시험#3 · #4 | P2 | 고침(시험 seam) | `strategy_lane_shadow_load_testseam.go`(tagged) 가 단계 stepCtx 를 기록 — 감독 goroutine 은 `defer cancel()` 로 끝나므로 stepCtx 닫힘 = 감독 종료. `a112WaitShadowIdle` 이 그것까지 기다리고, deadline 하위시험의 실시간 20ms sleep 을 `a112WaitShadowSupervisionSettled` 로 바꿈. 생산 파일 무편집 | N01(게시 10ms 지연) · N02(마감 전달 50ms 지연) GREEN-AS-EXPECTED(리뷰 때는 각각 3 · 20/20 FAIL) · S18D(S18 + 게시 지연 — 리뷰 때 생존) CAUGHT |
| 시험#5 | P2 | 고침(시험) | 새 `strategyshadow/refusal_fields_test.go` — 활성화 시험의 설정 결속 · 몸통 결속(actor · generation) · 수명 사례를 옮김, 필드 목록 정확 대조 | R10(`len(fields) != 0 && false`) · R11 · R12 · R13 CAUGHT |
| 시험#6 | P2 | 고침(시험) | `shadow_test.go` unknown shadow state 단언을 `descriptors[0]: shadow` 로 | (단언 문자열 — refusal_fields 쪽 같은 사례도 그룹 정확 대조) |
| 시험#7 | P2 | 고침(시험) | `TestAStuckShadowStepNeitherLatchesNorRefreshesTheOtherMarket` — KR 주기를 dispatch 없이 돌려 공유 위험 버킷 원장을 움직이지 않게 하고, US 기준 무오류 · placed ≥ 1 · observed ≥ 1 을 전제로 단언, KR 단계가 적재기에 닿았는지(멈춤 성립) 단언 | US 기준이 실제 주문 1(편집 전 BUCKET_USAGE_STALE 로 0) |
| 시험#8 | P2 | 고침(시험) | `TestTheRouterRuntimeStays…` census 를 타입 검사기 선언 객체로: 패키지 범위 const · var 전부 + 그 타입을 가진 상수 값 식 전부(빈 문자열 영값 비교만 제외) | V01(var 어휘) · V02(타입 없는 상수 암묵 변환) CAUGHT |
| 유지#5 | P2 | 고침(시험) | 새 `a112_shadow_vocabulary_test.go` — `strategyworker.ShadowOutcomes()` 와 `strategyprojection.LaneShadowOutcomes()` 직접 대조 + 각 목록의 상수 완전성 | V03(M3 WOULD_EMIT → WOULD_FIRE) CAUGHT(새 시험 포함 19) |
| 시험#9 | P2 | 고침(시험) | 새 `tools/a112-family-shadow/main_test.go` — run() 이 커밋된 골든 바이트 · 0400 · 핀 줄 전체를 재생산, 기존 파일(0600 — 권한이 대신 막지 않게) 비덮어쓰기, 옮길 수 없는 인자 여섯을 **그 플래그 이름으로** 거절하고 파일 0, parseMarket 두 시장 | T01 · T02 는 1회차 SURVIVED(인코더 · 0400 권한이 대신 막음) → 시험 보강 뒤 2회차 CAUGHT · T03 CAUGHT |
| 시험#10 | P2 | 고침(하네스) | `a112_lot_mutate.py` `run_tests` — 실패 이름 없는 panic · 시간 초과는 `CRASH` 로 따로, panic 스택의 시험 이름 기록 | 이번 원장 CRASH 0 |
| 시험#11 | P2 | 고침(영수증 + 시험) | `lot-0.5-response/branch-coverage-new-functions.txt`(새 함수 커버 · 진입 0 블록) · 새 `strategyshadow/guard_branches_test.go` · `engine/a112_shadow_guard_branches_test.go` | 첫 측정 진입 0 블록 25(그중 6 은 공유 래퍼 `production_shared_export.go` — 같은 패키지 시험만 센 측정 범위 탓, `-coverpkg` 로 재면 0) → 시험 추가 뒤 7(처분은 영수증 끝 절) |
| 성능#1 | P2 | 고침(시험 부품) | `testenv.TypeCheckProductionErr`(오류 반환판) + 엔진 census `sync.OnceValues` 캐시 | 판정 불변(같은 시험 전부 PASS) |
| 성능#2 | P2 | 고침(Makefile) | 엔진 race 한도 5m → 10m · 새 시험 둘을 `RACE_ENGINE_TESTS` 에 | `race-0.5-response.log` 76/76 PASS, DATA RACE 0, wall 3:03 |
| 유지#10 = 보안#3 | P2 | 고침(문서) | `docs/operations.md` SHADOW 절 — 「급히」 삭제, 폐기 매니페스트는 핀 교체 + 재시작이 필요함 · 재시작 없이 끄기 = 파일 치우기, 세대 단조 증가는 「권장」 | — |
| 보안#5 | P2 | 고침(문서) | OpenAPI `shadowOutcome` 설명에 「latch 미반영」 | JSON 유효 · httpapi 시험 PASS |
| 유지#7 | P2 | 고침(주석) | `shadowSkipped` 를 시험 관측 전용으로 | — |
| 유지#8 | P2 | 고침(주석) | 「반환 갈래 열다섯」 → 「열넷(닫힘 13 + 성공 1 — fail 클로저 안 return 제외)」. AST 셈: 본문 return 14 · 함수 리터럴 안 1 | — |
| 유지#11 | P2 | 고침(Makefile) | `RACE_ENGINE_FILES` 한 줄에 셋이던 경로를 한 줄 하나로 | race 가드 `tools/sdd/test_race_detector_actually_runs.py` 8/8 OK |
| 유지#6 | P2 | **기록만** | — | 상수 전달 리팩터(나이 상한 · 성공 판정이 supervisor 한도를 상수로 가정). 두 생산 자리가 모두 Maximum 을 넘기고, 어긋나도 R1(abandon 시 SHADOW 숨김)이 덮어 안전 효과 0 — 리팩터는 supervisor 편집을 부르므로 이 로트 밖 |
| 유지#9 | P2 | **기록만** | — | breakout RVOL 1_200_000 리터럴 이름 붙이기 · 반사실 필드 분리 — 봉인 밖 진단 필드이고, breakout 로트를 다시 여는 편집이라 피함 |
| 유지#12 | P2 | 기록만 | — | 의도된 패키지 간 사본(DRY). 이번 로트가 shadow 쪽 결속 · 수명 · 설정 거절 시험(시험#5)과 해석 단일 신원 짝 시험(유지#4)을 더해 두 사본을 시험으로 묶었음 |
| 유지#13 | P2 | 처분 불요 | — | 주석 종결 어미는 저장소 관례(리뷰어 자신이 「No action」) |

**각주(안전 규칙).** voice3 은 4cbcfb36 사본 전체 grep(`LoadProductionFamilyActivation|familyGateFor`)에서 다른 리뷰 디렉터리
`analysis/review-8.5-2026-10/_codex-r2/*.go` 의 줄 몇 개가 함께 걸렸다고 출력 맨 위에 스스로 적었다. 이번 디렉터리의 `_codex*` 는 열지 않았고, 걸린 줄은 어느 판정에도 쓰지
않았으며, 그 뒤 검색 범위에서 openspec 을 뺐다. `~/.codex` 접근은 네 판 모두 0.

**남은 위험.** (1) 도구의 쓰기 · 닫기 실패 갈래(`os.Remove`)는 시험으로 실행되지 않음. (2) shadow 적재기의 반환 직전 `ctx.Err()` 갈래는 결정적 주입 자리가 없어 미실행.
(3) 시험#3 · #4 의 대기 신호는 tossos_testseams seam 이 기록한 stepCtx 에 기댐 — 감독이 `defer cancel()` 을 잃으면 신호도 사라짐(그 경우 대기 helper 가 5s 뒤 실패하므로 조용히 통과하지는 않음).
