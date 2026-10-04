신원: GPT-6 기반 Codex(세부 버전 미공개) · 「~/.codex 미접근」

**FAIL — N2 CLOSED(설계 수준), 신규 P0 0건 · P1 1건.**

절차 이탈: 최초 읽기 전용 조회 셸에 `set -euo pipefail`을 붙이지 않았다. 명령은 exit 0이었고 이후 셸에는 적용했다. brief.md의 안전 규칙상 리뷰 무효 조건에 해당하므로 정식 승인 증거로 사용하지 말 것. 아래는 별도로 확인된 기술 발견이다. 금지 경로 접근은 없었다.

입력: design-brief-v3.1-under-review.md SHA-256 `c2ea7e821d87224b7bc28923a72017bad95886f5149178826e04b1823688671c`. 기준 코드 `4d22d72663dcf3a42be928b3ecf3db817ffc6ac5`. v3→v3.1 diff, codex-recheck3-output.md 및 이전 스케치 확인.

| 항목 | 판정 | v3.1 좌표 | 근거 |
|---|---|---|---|
| N2: 필수 stdlib 의존성과 금지 충돌 | CLOSED, 설계 수준 | §2:42–44 | 금지 대상을 모듈 패키지의 직접 import로 좁히고 stdlib 내부 사용을 제외. encoding/json 내부 reflect/unsafe가 정상 구현을 탈락시키던 반례 해소. 아래 실측 범위 구분 필수. |
| 핀 ①(A), ④, ⑤와 dispatch 앞 작업 | 신규 P0/P1 입증 없음 | §5:113–119 | 명시적인 인자 운반과 같은 잠금의 단일 보관은 I/O·판정·순회를 허용하지 않는다. 다만 accessor와 record 인자 전달의 문법적 예외를 정확히 써야 한다(아래 비차단 노트). |
| `{wave,batch}` 보관과 evaluate 오류/panic | **P1 N3 OPEN** | §5:103–108, §5.1:130–137 | record 전 실패는 새 wave를 만들지 않는다. 시작 생략만으로 이전 관측은 무효가 되지 않는다. |
| 계보 충돌 부재 값 vs 빈 관측 묶음 | 신규 P0/P1 없음 | §4:78–87 | observed=false와 observed=true/len=0을 구별하는 필드·반환 리터럴·행동 시험 명시. 부분 묶음 NO_INPUT 오보고를 설계상 차단. |
| 투영 시점 만료 | 신규 P0/P1 없음 | §5.1:132–137 | now < expiresAt와 경계 ±1ns 핀은 파도 정지 후 만료 문제를 닫는다. 만료 전 오류에 의한 무효화까지 대신하지는 못한다. |

P1 N3 — record 이전 실패 뒤 이전 SHADOW 관측이 남는다

- 주장: v3.1:106–108은 cycle 오류 → shadow 시작 안 함 → UNOBSERVED를 약속한다. :130–134의 투영 조건은 wave 등식과 미만료뿐이다.
- 현재 AST: `strategy_lane_runtime.go:208–209`는 recoverMarketLanes 오류를 즉시 반환한다. :244–246은 레인 goroutine panic을 join 뒤 다시 던진다. `record`는 그 뒤 :249에서만 호출한다. `record`:326–335에서 잠금·wave 증가·관측 기록 수행.
- 실제 오류 원천: `strategy_lane_latch.go:172–178`의 원장 복구 오류. 복구 증거 부족 sentinel 외 오류가 반환된다. 단순 nil runtime 같은 비현실적 입력에 기대지 않는다.
- 반례: 시장의 완료된 wave=7, shadow observation.wave=7, expiresAt 미래 → 다음 evaluate에서 복구 오류(또는 record 전 panic) → record 미도달, wave=7 유지 → shadow kickoff 없음 → 투영의 `7 == 7 && now < expiresAt` 참. 이전 SHADOW/WOULD_EMIT가 남는다. §6:145–146은 LATCHED도 SHADOW 허용하므로 health 변화가 자동으로 이 반례를 닫지 않는다.
- record **뒤** persistMarketLatches/staleLatchError 오류는 구별한다. 정상적으로 wave가 증가했다면 기존 wave=7 관측은 새 wave=8과 달라 폐기된다. 문제를 모든 evaluate 오류로 일반화하지 않는다.
- 수정 요구: record 전 오류·panic 및 cycle 오류의 shadow 무효화 계약을 명시한다. 생산 wave/dispatch 의미를 바꾸지 않는 방법으로 실패한 시장의 이전 관측을 무효화하고, 이미 진행 중인 이전 shadow 작업이 다시 게시해 되살릴 수 없도록 한다. 단순 관측 삭제만으로는 늦은 작업 재게시를 막지 못한다.
- 필요한 핀: 관측 성공 W → 만료 전 복구 오류 / pre-record panic → 즉시 UNOBSERVED·null. 별도로 진행 중 W 작업을 실패 후 완료시켜도 UNOBSERVED·null 유지. 오류 반환·기존 dispatch 궤적 보존.
- 재현: `TestV31PreRecordFailureRetainsShadow` 2갈래에서 이전 관측 유지 확인. **설계 알고리즘 최소 모델 + 현재 함수 AST 근거**이며 아직 존재하지 않는 SHADOW 생산 구현의 통합 재현이 아니다. 신규 실제 주문/활성화 경로는 입증하지 않았으므로 P0가 아닌 P1.

모듈 import 실측

`go/build.ImportDir`로 현재 플랫폼의 파일/태그 선택과 직접 import를 읽었다. 루트의 test 모드만 TestImports/XTestImports를 더하고, 의존 패키지는 생산 Imports로 재귀한다. `go list`는 실행하지 않았다.

| 측정 루트 | 모드 | 도달 모듈 패키지 | 직접 reflect/unsafe 패키지 수 |
|---|---|---:|---:|
| strategyrouter | deps | 9 | 0 |
| strategyrouter | deps-test | 9 | 1 |
| strategyrouter | deps-tagged | 9 | 0 |
| strategyrouter | deps-test-tagged | 22 | 3 |
| strategyworker | deps | 15 | 0 |
| strategyworker | deps-test | 17 | 1 |
| strategyworker | deps-tagged | 15 | 0 |
| strategyworker | deps-test-tagged | 17 | 1 |

시험 루트의 양성은 `strategyrouter/scheduler_test.go:4`, `strategyworker/a112_projection_table_test.go:11`의 reflect. router tagged 시험은 `a127_real_journal_test.go:14`에서 journal로 들어가 `journal/exit_snapshot.go:11`, `journal/risk_bucket_relaxation.go:20`, `exitpolicy/recovery.go:6`의 생산 reflect도 도달한다. 전체 패키지별 직접 import 목록은 tests.log에 있다.

중요한 범위 구분: router 자체의 `-test` 그래프는 새 strategyshadow를 루트로 하는 `-test` 그래프와 다르다. 의존 패키지의 자체 시험은 자동으로 포함되지 않는다. 따라서 위 시험 양성을 N2 재개방 근거로 삼지 않는다. §2의 실제 핀은 strategyshadow 루트 네 모드를 써야 하며, router/worker 자체 시험 그래프까지 일괄 0이라고 주장해서는 안 된다.

한계: 외부 모듈 경계 `modernc.org/sqlite`, `golang.org/x/sys/unix`에서 소스 추적을 멈췄다. 외부 의존성을 경유해 본 모듈로 돌아오는 경로까지 검증한 완전한 ListDeps 대체가 아니다. strategyshadow는 미구현이라 그 실제 네 모드 그래프 통과도 주장하지 않는다. N2 CLOSED는 원래의 필수 stdlib 반례가 제거됐다는 설계 판정이다.

비차단 핀 정밀화 노트

- :113–116은 `fresh.<shadow>.forMarket(market)`을 허용하면서 shadow 함수 호출 0이라고 쓴다. forMarket을 순수 시장 선택 접근자로 예외 명시하고 몸체를 핀해야 검사 해석이 일치한다.
- evaluate가 받은 batch를 record에 넘기는 인자 사용도 필요하다. ④의 “보관 대입 한 자리에만”은 운반 인자 전달과 유일한 저장 위치를 구별해야 한다. 동작상 I/O·판정 허용으로 읽지 않았다.
- §4의 “다른 shadow 식별자 0”도 선언/반환 운반값까지 금지하는 뜻이 아닌 조정 경로 사용 금지로 구체화할 것. 수집·운반 자체의 명시적 의도가 있으므로 새 P1로 중복 집계하지 않는다.

검증·절차

- 신규 Go 파일은 전부 `_codex/recheck4/` 안. `extract_go_ast.go`는 기존 tools/logic-map 파일을 그대로 복사하여 go test에서 analyze를 호출했다. AST 6개 생성, 이후 분기 판정 작성.
- 함수 논리 요약: evaluate = 복구(error exit) → 병렬 레인 실행/join → panic 재전파 → record → durable latch 저장(error exit) → stale latch 오류. record = nil/빈 관측 guard → 잠금 → wave 증가 → 관측 저장. 분기 시험 지도: pre-record 오류/panic 모델, post-record 오류는 AST 순서 분석, 실제 생산 fault/race 시험 미수행.
- 최종 go test: 최상위 4개 PASS, import 모드 8개·실패 갈래 모델 2개 포함. PASS는 설계가 안전하다는 뜻이 아니라 측정값과 반례가 재현됐다는 뜻.
- 최초 탐색 시험은 모든 시험 루트에도 import 0 및 외부 경계 없음이라는 강한 가정을 두어 FAIL. tests-initial.log 보존. 최종 시험은 실측 양성과 외부 경계를 명시하고 예상 수치를 검증한다. 생산 코드를 수정해 통과시킨 것이 아니다.
- 명령: `GOCACHE="$PWD/.gocache" GOTMPDIR="$PWD/.gotmp" GOFLAGS=-mod=mod GOPROXY=off GOSUMDB=off GOTOOLCHAIN=local go test -count=1 -v ./_codex/recheck4`.
- Function Logic Map/Pre-Edit: 생산 함수 편집 not-applicable; 분기 근거용 AST는 생성. CodeGraph/GBrain/memory/외부 skill/make 게이트는 사용자 범위·시험 제한으로 미실행. 구현 완료·SDD 통과 주장 없음.
- 실저장소 시작·끝 status와 HEAD 동일(cmp exit 0). 원래 dirty 상태 유지. 시작 값은 최초 도구 출력에서 전사했으며 이를 숨기지 않는다. repository-check.txt와 source-comparison.txt 참고.
