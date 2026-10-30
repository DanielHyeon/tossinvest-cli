# a166 lot 1 Terra 검증 기록

기준 `da67f3e1e9282e181160467fb2a476345a08627a`. 대상은 신규 `internal/accountbudget`이며 공식 client/engine/CLI·환경값을 연결하지 않았다. 이 기록은 Manager 독립 delta 검증 전 제출물이지 전체 a166 완료·운영 활성화 승인이 아니다. UI 없음: impeccable N/A — backend-only. `tasks.md`의 1.4는 독립 최종 검증 전 체크하지 않는다.

## 구현과 범위

- 운영자가 준비한 Linux root의 scope·instance와 잠금 inode를 검증하며 `OpenContext`는 대기 context를 사용한다. `Open`은 호환용 Background 래퍼이므로 제한 시간 필요한 호출자는 `OpenContext`를 사용한다. `Provision`은 별도 명시적 준비 함수이고 기존 앱에 자동 연결하지 않았다.
- 경로 walk에서 NOFOLLOW/CLOEXEC, UID·권한·regular·hardlink·filesystem, 읽기 크기, state 형식·ticket ID/순번·상한을 검사한다. queue와 wire는 분리되며 dirty 선기록·state sync/rename·dirty 제거 순서를 사용한다. 큐가 꽉 찬 경우 dead liveness를 확인한 뒤 등록한다. 모호한 liveness는 실패한다. Waiter는 단일 소유자다.
- `WithPriority`/`PriorityFromContext`는 기본 Normal과 명시 Critical/Auxiliary를 전달한다. `NewTransport`는 거래 권한을 부여하지 않는다. 공식 constructor 소유의 direct DialContext만 caller가 전달해야 한다. proxy 함수가 실제 proxy 또는 오류를 반환하면 거절한다. custom TLS 검증 callback, insecure 검증, legacy dialer 등은 지원하지 않는다. 클라이언트 timeout은 외부 `http.Client`에 남으며 TLS handshake timeout은 base transport 값을 사용한다.
- transport는 본문 EOF/Close/취소 시 revoke/drain 후 wire를 반환한다. 오류 경로 `ErrNotSent`는 drain 후 현재 HTTP 시도의 첫 Write 여부를 기준으로 한다. 응답 handoff 이후 취소로 발생한 비동기 cleanup 오류는 `NewTransport`의 concurrency-safe·빠르게 반환하는 callback으로 보고하며 동기 오류는 호출자에게 반환한다. 이벤트 저장/429 hint는 후속 로트다.
- 기존 `clock.Clock`은 신뢰된 주입 전제다. fake/조기 반환 clock은 rate를 보장하지 않는다. 제품 wiring에서 기존 `clock.System` 계열을 봉인하고 AST/테스트로 검증해야 한다. 실제 배포는 모든 계좌 참여자가 같은 namespace·device/inode·instance를 쓰는지 별도 확인해야 한다. 서로 다른 root에 같은 marker를 복사한 경우 module만으로 동등성을 증명하지 못한다. tmpfs는 재부팅 영속성을 보장하지 않으며 누락은 fail closed. 실제 지원 파일시스템·namespace는 운영 전 확인 대상이다.
- Linux 구현 외 `!linux` fail-closed API stub을 제공한다. 다른 OS 컴파일이나 OFF 통합 테스트는 실행하지 않았다. GOFLAGS=-buildvcs=false는 테스트 VCS 메타데이터만 비활성화한다.

## 수행 증거

- `go125-normal-final.log`, `go125-race-final.log`, `go126-normal-final.log`, `go126-race-final.log`: Go 1.25.0/1.26.5 각각 `go test [-race] -count=1 ./internal/accountbudget`. Go 1.25.0은 `GOTOOLCHAIN=go1.25.0` 호출 시 Go가 toolchain을 다운로드해 사용했다. 프로젝트 go.mod는 바꾸지 않았다.
- 제품 테스트는 2개 별도 프로세스·각 2개 실제 TLS 요청에서 총 4회 첫 TLS plaintext Write 직전 test hook 시각을 기록하고 인접 350ms 및 1초 rolling 4회 부재를 검사한다. 이 시각은 테스트 hook에서 파일에 기록한 wall-clock 표본이며 브로커 도착 시각이나 실제 운영 증거가 아니다. owner kill, dirty 주입 후 별도 프로세스 거절, full dead queue, missing liveness, priority, context, 파일 오류, TLS body/인증, unread response 취소 및 실제 TLS blocked Write 취소/drain 시험 포함.
- `mutation-final-{classify,dirty,priority,cleanup,full}.log`: 각각 drain 전 sent snapshot, dirty 무시, priority 무시, wire cleanup 오류 폐기, full dead queue cleanup 제거의 임시 사본 변이가 기대대로 실패했다. classify는 TLS 인증 실패 후 gated Write를 주입한 동기화 fixture이며 실제 네트워크 경합 실측과 구분한다. priority 변이 RED의 active owners 오류는 테스트 실패 후 잔존 waiter의 부수 출력이다.
- `full-queue-red.log`는 기존 256 dead ticket에서 critical이 queue full로 거절된 실패, `full-queue-green.log`는 보정 후 focused 시험이다. 최종 시험은 normal/critical 두 등급과 missing liveness를 함께 검사한다.
- gstack review: 설치·업그레이드·telemetry 없이 로컬 checklist를 수동 적용했다. race/데이터 무결성, 실패 시 lock release, TLS 수명 및 input 경계를 검토했고 proxy 선택, sent snapshot, full dead queue 및 release 오류를 보정했다. 이는 별도 독립 리뷰나 전체 gate를 대체하지 않는다.

## 남은 검증과 제한

CodeGraph는 worktree에서 미초기화; 현재 증거는 원문 검색과 저장한 특정 함수 pre-edit AST이며 전체 graph freshness/전체 FLM/BTM PASS를 주장하지 않는다. 첫 lot 신규 함수에는 기존 제품 함수 변경이 없지만 개발 중 자체 함수 수정용 pre AST는 보존했다. independent draft `abaa22aaa` 리뷰는 후속 delta를 검증하지 않는다.

실제 kernel fsync EIO·전원차단/재부팅·다른 OS 실행·모든 파일/close 오류 조합·시계 점프·비협조 파일 교체·실계좌·브로커 도착률·운영 namespace는 미검증이다. 임의 무한 지연되는 OS 호출이나 request body의 종료 보장을 주장하지 않는다. OFF golden, official token/401/redirect·authority·429/event·전체 호출자 배선은 후속 로트. 전체 저장소 테스트·전체 gate·실운영·배포·활성화·push/merge/archive는 수행하지 않았다.
