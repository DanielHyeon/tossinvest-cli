# poll clock leaf 결과

기준 `ccc39f6f137236740efa924f96fa2a8f2d953bd9`. 신규 leaf만 추가했다. 기존 Clock/System/Lease와 호출자는 수정하지 않았다. Manager가 승인한 복원 계약 및 design 해당 문단을 복사했다. CodeGraph는 이 worktree에 초기화되지 않았다. 기존 함수 변경이 없으므로 기존 함수 FLM/BTM은 해당 없음; 신규 함수와 시험의 변경 전 AST는 `pre-*.json`, `pre2-*.json`, `pre3-*.json`에 보존했다. search-first 근거와 UI 비해당 사유는 `pre-edit.md`에 있다.

Linux `SystemBootSource`는 신뢰 procfs, zero time namespace offset, 현재/자식 namespace 동등 및 관측 앞뒤 boot/namespace 일치, bounded read와 CLOCK_BOOTTIME을 검사한다. 다른 boot에서 이전 duration을 전부 보류하고 `NeedsAnchor`를 반환하며 uptime을 차감하지 않는다. 동일 boot에서는 namespace와 시간 순서를 검증한다. anchor의 실제 원자 저장 및 읽기, 재시작 통합, 429 처리와 운영 주입은 후속 로트다. 신뢰 procfs와 namespace 조작 금지 전제이며 불변 조건을 위반하면 오류다. non-Linux는 명시적 unsupported 오류다.

## 검증

- 초기 컴파일 RED/GREEN: `initial-red.log`, `initial-green.log`. 계보, zero duration, UUID, anchor 선행 부정 각각 `*-red.log`/`*-green.log`. 개발 중 1ns를 1s로 잘못 기대한 시험의 최초 실패 `full-development-race.log`와 시험 수정 후 `full-development-race2.log`를 함께 보존한다.
- Go 1.25.0과 1.26 각각 `GOFLAGS=-buildvcs=false go test [-race] -count=1 -v -run '^(TestBoot.*|TestSystemClockNowIsUTC|TestSystemLeaseAnchorRetainsMonotonicReading|TestSystemClockSleepRespectsContext|TestFakeClockNowAdvance|TestFakeClockSleepCancel|TestFakeClockSatisfiesInterface)$' ./internal/clock`: 네 로그 `go125-normal.log`, `go125-race.log`, `go126-normal.log`, `go126-race.log` PASS. Linux 실제 proc 읽기 smoke는 boot ID를 출력하지 않았다.
- `go test -race -count=1 -run '^(TestA090ElapsedUsesOnlyTheLeaseHelpers|TestA111ObserverUsesClockLeaseHelpersForTheUseLease)$' ./internal/app/engine`: `existing-ast-pins-race.log` PASS. Go 1.25 Darwin amd64 `go test -c`는 성공했고 실행하지 않았다 (`darwin-compile.log`는 정상 무출력).
- 최종 코드에 대한 overlay 변이 8건은 각각 `mutation-final-*.log`의 의도된 assertion 실패로 검출: 다른 boot uptime 차감, anchor 필요 누락, saved/current 관계 누락, saved/anchor 선행 검사 누락, UUID metadata 검사 누락, nonzero offset 허용, current/children 불일치 허용, boot 변경 허용. 초기 namespace 변이 생존 `mutation-namespace.log`를 보존했고 persistent mismatch fixture 보강 후 `mutation-namespace2.log` 및 최종 변이가 검출했다. overlay 원본은 `/tmp/a166-clock-final-mut-*.go` 및 `*.json`에 있다.

## 수동 gstack review

`~/.agents/skills/gstack/review/SKILL.md`와 checklist를 범위 한정 수동 적용했다. 설치·업그레이드·telemetry 및 별도 specialist 실행은 하지 않았다. 두 pass에서 상태 영속·SQL·셸·LLM·UI·새 enum 소비자는 해당 없음. 읽기 경계·수치 overflow·시간 계보·오류 전파·전후 관측과 경쟁 전제를 손읽기했다. 발견한 anchor 선행·zero 무결성·UUID 검사와 namespace 시험 공백은 RED 후 수정했다. 추가 발견 0. 이 검토는 독립 검증을 대체하지 않는다.

전체 테스트/전체 gate, 다른 플랫폼 실행, namespace 변경·proc 쓰기, 실제 계좌·운영 검증 및 적용은 수행하지 않았다. 제품에 아직 연결되지 않았고 이 결과는 운영 활성화나 실제 배포 환경의 신뢰성을 보증하지 않는다.
