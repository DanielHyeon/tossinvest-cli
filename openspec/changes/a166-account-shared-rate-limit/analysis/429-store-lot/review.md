# 429 observation store leaf 결과

기준 c412c0e55f2948e16e1f33f1e1e8fadaf2092c56. Manager 승인 계약은 analysis/429-store-draft.md, 선행 범위·검색·잠금 지도는 pre-edit.md다. 신규 미배선 leaf이며 기존 Provision/Open/Acquire/Release/Transport, runtime parser·official·Retrier·caller 본문과 schema는 변경하지 않았다. Manager가 승인한 Lease의 private started/sequence 필드 두 개만 기존 struct에 추가했다. 기존 함수 본문 변경이 없어 기존 함수 FLM/BTM은 해당 없음. 신규 Begin/Finish/Snapshot/PollRemaining/write/validate의 Go AST는 `ast-*.json`이다. CodeGraph는 미초기화; 그래프 증명은 주장하지 않는다. impeccable N/A — backend-only, 신규 의존성 없음.

## 상태·경계

명시적 ProvisionObservation은 기존 trusted root에 별도 event.lock/state를 O_EXCL 생성한다. OpenObservation은 외부 expected event-lock inode와 Budget의 root/queue/wire 결속을 확인한다. state/dirty는 dirfd의 regular/owner/mode/nlink/bounded 검증을 받고 state는 원자 교체한다. state에는 opaque scope/instance·typed pending·마지막 완료 결과·각 128개 429/gap ring·최대 256개 poll slot만 둔다. raw account·header·body·URL은 기록하지 않으며 slot key는 snapshot/event payload에 내보내지 않는다. cursor는 scope/instance와 event/gap sequence에 결속된다.

Begin/Finish는 실제 동일 Budget의 활성 Lease와 Lease.mu에 결속하며 Release와 직렬화한다. 한 Lease는 Begin을 한 번만 시도할 수 있다. 실패한 Begin의 추정 sequence로 Finish하지 못한다. wire→event만 허용하며 event 보유 중 wire/queue를 취득하지 않는다. reader는 event를 놓고 wire를 비차단 확인한 뒤 동일 generation을 재확인한다. Busy는 당시 wire가 점유됐다는 뜻이며 원래 pending의 생존을 단정하지 않는다. 새 wire owner가 남은 pending을 gap으로 기록하고 poll slot을 poison한다. 정상 NoResponse는 마지막 완료로 저장하며 429가 없었다는 원격 사실을 주장하지 않는다. dirty 저장 오류는 unavailable, 자동 복구 없음.

동일 boot hint의 남은 시간과 새 유효 hint 중 큰 값을 보존한다. 음수·30초 초과·비429 hint는 무시하되 429 사건은 기록한다. 재부팅 후 현재 uptime을 차감하지 않고 최초 복원 anchor를 같은 상태 갱신에 저장한다. 만료된 정상 slot만 회수하며 poison은 자동 삭제하지 않는다. 모든 sequence 및 generation overflow를 거절하고 pending에는 Finish용 generation을 남긴다. Begin 쓰기 중 취소되면 성공을 반환하지 않고 이미 내구 저장된 pending은 보존한다. Finish는 명시 ctx를 받으며 취소된 요청 ctx와 별도 bounded settle ctx 선택은 후속 transport 설계 책임이다.

## 실행 근거

- `initial-red.log`/`initial-green.log`, `invalid-hint-red.log`/`invalid-hint-green.log`, `generation-reserve-red.log`/`generation-reserve-green.log`, `cancel-sync-red.log`/`cancel-sync-green.log`, `noresponse-state-red.log`/`noresponse-state-green.log`에 주요 RED/GREEN을 보존했다.
- Go 1.26.5 `GOPROXY=off go test ./internal/accountbudget -run '^TestA166Observation' -count=1 -timeout=120s` normal `final-normal.log` 17.292초, 같은 선택 race `final-race.log` 20.698초. Go 1.25.0 선택 normal `final-go125.log` 17.398초. 정상 무응답·취소 settle·dirty 단계·scope/instance·overflow·ring/cursor·hint/poison·reader generation·Lease/Close·2process 생존/종료를 포함한다.
- 기존 Lease 기능의 선택 race `existing-lease-race.log`: TestCloseRejectsLiveLease, TestWaitRespectsCallerDeadline, TestReleaseRepeatsCleanupError, TestPriorityOvertakesQueuedNormal, TestDirtySyncFaultBlocksOtherProcess. Darwin amd64 Go 1.25 컴파일만 성공 (`darwin-final-compile.log` 무출력); 실행하지 않았다.
- `final-mutation-*.log` 13건은 Lease 중복, dirty 무시, inode 불일치, generation 재검증, poison, 취소, ring, reboot anchor, 외부 state 결속, invalid hint, Begin/reader의 Finish generation 예약, NoResponse 완료 누락을 각각 assertion 실패로 검출했다. 해당 patch gzip 원문과 압축 해제 SHA-256은 `final-mutation-*.patch.gz`, `mutation-sha256.txt`에 보존한다. 주입된 fsync/write/rename 오류는 실제 전원 상실 실측이 아니다.

수동 gstack review는 두 pass에서 파일 신뢰 경계·락 순서·Lease/Close 경쟁·원자 파일 교체·정수 overflow·enum·bounded ring·clock 계보·cursor/poison을 검토했다. 검토 중 취소 전파, invalid hint와 event 분리, Finish generation 예약, state scope/instance 및 NoResponse 완료 누락을 보완했다. 추가 P0/P1 발견 없음. 설치·업그레이드·telemetry·별도 specialist는 수행하지 않았다. 이 결과는 독립 검증을 대체하지 않는다.

기존 v1 runtime은 바꾸지 않았고 이 leaf는 어떤 요청에도 아직 연결되지 않았다. 사전 checkpoint 이후 실제 poll hint 재확인, 원래 HTTP response/status/body/error 보존, callback·request-local 신호, parser·v2 manifest·모든 caller 결속 및 실제 TLS 통합은 후속이다. 신뢰 filesystem·정상 fsync·관리자 비변조와 신뢰 BootSource 전제이며 성공을 거짓 보고하거나 전체 저장소가 사라지는 장애를 증명하지 않는다. 모든 참여자 중지·새 root/instance cutover는 사람이 승인할 별도 운영 절차이고 실행하지 않았다. 전체 테스트/gate·네트워크·실계좌·운영·활성화·push·archive 미실행.
