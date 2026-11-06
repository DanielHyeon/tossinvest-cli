# a169 daily claim 연구 — 2026-11-06

## 결론 및 기존 계약 충돌
- 단일 신뢰 root·정확한 KST 날짜·협조 caller 전제에서 `O_CREAT|O_EXCL` 일별 빈 claim 하나는 유효한 **1회 시작 소모권 후보**다. claim 파일과 부모 디렉터리 sync 성공 후에만 성공을 반환하고 어떤 종료에도 제거하지 않으면, 성공적으로 동기화된 claim이 보존되는 범위에서 동일 root/day의 다음 시작을 거부한다. 정상/abort 이유·실제 종료 결과·보고 적격성은 알려주지 않는다. 이 단계는 dormant이며 actual runtime, 전송 차단, 감시, engine identity, 보고를 증명하지 않는다.
- 명백한 계약 변경: Q2는 재시작 시 universe 원장 복원 (`openspec/changes/a169-breakout-universes-are-compared-in-shadow/proposal.md:96`, `design.md:91-92`)이고 U2 선정 마감 전 재시작은 다시 시도 (`design.md:214-215`), 테스트 요구도 복원을 전제 (`design.md:567`). proposed claim은 정상/장애 구분 없이 같은 날 모두 거부하므로 정면 충돌한다. 기존 abort 뒤 재시작 거절만 정한 `design.md:450-452`보다 범위가 넓다. Manager가 충돌을 명시적으로 정리하기 전에는 이 구현안을 현행 동등 계약이라 주장하지 않는다. 사용자 진행 승인이 구체적 Q2 변경을 뜻하는지는 이 연구에서 판정하지 않는다.

## 현재 AST 근거와 재사용 평가
- 현재 HEAD `ca7a3f4f77f597ed529752d9b89dd81bcd8a4b59`. `go run ./tools/logic-map` 원문 출력으로 생성: `probe.startSession` source sha256 `62c2fe84d04026f1f398a01cb97b270880c82f1f355ad0a6b2841ccb59e62de1`, B3@182, B6@189, B9@200. 소스 `tools/a140-breakout-signal-shadow/shadow.go:176-201`과 대조: 같은 세션 마지막 universe를 복원하고 nil일 때만 새 선정. AST는 구조 근거, 분기 의미는 소스 손읽기다.
- `checkedDir`/`checkedFile` AST source sha256 `d2c6c8bf5a8df35e406fe7c2720770b0a00eb4cfe9c05a3dc4830aec2624b8ff`. `checkedDir` B1@30/B3@37/B7@55/B9@63, `checkedFile` B1@71/B2@75/B3@78. 손읽기: `internal/accountbudget/files_linux.go:29-82`는 각 경로 component `openat`+NOFOLLOW, owner/mode/fs, regular file/owner/link를 확인한다. 둘 다 **비공개**라 별도 패키지 직접 import 불가. 계좌 budget root를 claim root로 재사용하거나 함수 export를 위해 해당 패키지 변경할 이유는 현재 없다.
- `internal/accountbudget/budget_linux.go:68-101`의 `Provision`은 공개지만 특정 budget schema 전용, 파일 sync 후 dir fsync 선례다. `internal/enginelock/enginelock.go:148` Acquire는 engine journal 생존 lock, 일별 영속 일회성 소모권이 아니다. `internal/runlock/runlock.go:17-33,84`도 배타 durable claim이 아니다. 공개된 완전 일치 helper는 대상 범위 검색에서 발견 못 함; 저장소 전체에 없다는 주장은 아님.
- Go 1.25 및 이미 설치된 `golang.org/x/sys v0.46.0` (`go.mod:3,14`)의 unix openat/fstat/fsync를 사용하면 새 의존성 불필요. `os.OpenFile`은 기본 파일 열기 가능하지만 경로 component를 dirfd로 검증하는 현재 보안 패턴과 동등하지 않다. 최종 선택은 좁은 신규 Linux 패키지 + `!linux` fail-closed stub; 필요한 부분만 기존 패턴을 참조해 구현. 전용 dependency나 범용 helper 추출은 이번 범위 밖. 라이선스 변경/신규 설치 없음.

## 제안하는 경계·실패 의미
- root는 외부에서 사전 생성·신뢰·식별된 단일 root: 절대 정규 경로, symlink 거부, owner/mode 및 지원 filesystem 검증과 기대 dev/inode 등 identity 대조를 명세. 재기동 후에도 모든 caller가 동일 root/identity를 사용해야 함; 경로·마운트 대체, 파일 삭제, snapshot rollback, 별도 namespace/복제, 비협조 프로세스는 root 검증만으로 방지하지 못하므로 명시적 신뢰 전제. `--out`에 종속시키지 않는다. tmpfs는 휘발성이라 재부팅 후 durable claim 보장용으로 허용하지 않는다. ext4도 장치·mount·sync 신뢰 전제와 별도 실제 검증이 필요하다.
- 날짜는 정확한 `YYYY-MM-DD` 유효 civil date로 제한해 단일 안전 파일명에 매핑. 호출자가 신뢰할 수 있는 KST 시각/거래일을 확정하는 별도 경계가 필요하다. 유효 형식만 검증하는 함수는 받은 날짜가 현재 KST 거래일임을 입증하지 못한다. 시장 달력 GET 이전에 claim하려면 KST clock 기반 날짜와 나중 공식 거래일의 일치·자정 교차 실패를 runtime 계약에서 다뤄야 한다.
- openat에 `O_CREAT|O_EXCL|O_NOFOLLOW|O_CLOEXEC` 등 사용, 단일 root/day 생성 성공자만 다음 단계 진행; EEXIST는 파일 내용/유형과 무관하게 거절하고 오류나 identity 불확실성도 거절. 생성 후 file fsync 및 close 오류와 directory fsync 오류 중 하나라도 실패면 성공 반환 금지, 자동 unlink 금지. 명시적으로 반환 성공 이후에만 후속 wire 가능. 각 과정 예외와 process 경쟁은 fixture 및 독립 리뷰로 검증한다.
- Linux man-pages [open(2)](https://man7.org/linux/man-pages/man2/open.2.html), [fsync(2)](https://man7.org/linux/man-pages/man2/fsync.2.html): exclusive create와 파일/디렉터리 sync의 구분 근거. `fsync(file)`만으로 디렉터리 엔트리 내구성을 보장하지 않는다. 지원 파일시스템의 전원 손실/디스크 장애 보장은 API 결과·환경 전제에 종속된다.

## fsync 실패 전후와 retry 반례
- claim 생성 뒤 file sync 또는 dir sync 실패로 no-wire 반환했는데 파일은 현재 보일 수 있다. 즉시 재시도는 EEXIST로 거절해야 하며 오류가 실패 없는 미생성을 뜻하지 않는다. file sync 성공·dir sync 실패 뒤 crash 시 파일이 남거나 사라질 수 있다. 사라진 경우, **모든 caller가 sync 성공 전 wire 금지를 지켰고 이전에 이 day에 성공한 claim이 없었다는 추가 사실이 보장된다면** 재시도 자체는 중복 wire를 만들지 않는다. 그러나 재부팅 뒤 파일 부재만으로 이 추가 사실을 증명할 수 없고, 실패 전 다른 성공 실행의 claim 삭제/rollback 또는 디스크가 내구성 약속을 어긴 경우에는 부재가 이전 wire 없음을 뜻하지 않는다. 따라서 무조건 retry 가능·당일 abort가 재부팅 후도 반드시 보존된다고 주장하지 않는다. 명세는 failed attempt의 no-wire 보장과 storage/root 보존 전제를 분리하고 불확실 저장장치 상태를 fail closed 하도록 한다.
- claim 선점 후 프로세스만 죽으면 파일은 보통 남아 후속 거절되지만 sync 완료 전 시스템 crash는 엔트리 소실 가능. 최초 정상 wire 전에 성공적 file+directory sync를 요구하면 crash 시 0 wire 또는 보존 claim이라는 조건부 안전 논리가 성립한다. 이후 crash와 종료 이유를 구별하려면 claim 외 후속 ledger/protocol 필요.

## 범위·검증 상태
- 허용 범위에서 Go AST 생성 및 문서/소스 손읽기. 테스트, 실제 시장/운영 파일, 계좌·시크릿, 소스/테스트 수정, Git mutation 미실행. ai-memory는 이전 분석에서 scoped a169 빈 결과, Hindsight 유효 결과 없음. MCP 도구 목록·기존 skill 목록 및 로컬 저장소/Go API 확인; 외부 패키지 레지스트리·GitHub 구현 검색은 하지 않음(새 의존성 금지, 기존 OS primitives 채택). Linux man-pages 확인. 검증 제안: 날짜 유효/경계, root identity 치환, symlink/hardlink/권한/FS 거절, 두 프로세스 동시 create, file/dir sync·close 각 fault, process kill 및 시스템 crash 모델 구분, `!linux` fail closed; 실제 테스트 결과 아님.

## 대안 연구: 정상 재개를 보존하는 durable 상태 (질문 답변 대기)

아래는 후보이며 확정 계약이 아니다. user가 동일일 정상 재개를 택하면 Q2 원장 복원을 허용하면서 미완 run의 재개를 거절하기 위해, 일별 상태를 최소 `active` / `paused` / `aborted`로 구분하고 초기 미존재→active, 정상 중단 active→paused, 재개 paused→active, 오류 active→aborted로 전이한다. active가 남은 상태에서 owner 없이 재시작하면 crash/불확실로 간주하고 당일 거절한다. aborted·dirty·손상·읽기 오류도 거절. 새 날짜는 독립 키. 성공한 전체 수집 완료는 정상 일시중단과 같지 않으므로 `completed` 종결 상태를 추가하거나, `paused`가 재개 가능한 수집 중지이며 완료 여부는 별도 권위 있는 계약에 의해 판정된다는 결정을 필요로 한다. 보고서 재생성은 수집을 재개하지 않는 별도 읽기 경로로 구분해야 한다. 진단 ledger를 state와 원자 결합하지 않은 상태에서는 ledger만으로 완료 사실을 권위 있게 증명하지 않는다.

- 동일 root/day에 process 간 kernel flock 등의 owner lock을 최초 검사 **전** 획득하고 전체 active 실행 동안 유지한다. lock 충돌은 거절. lock 자체는 프로세스 사망 시 풀리므로 durable 상태를 대신하지 않는다. 상태 및 dirty marker 변경은 같은 lock 아래 직렬화, 서로 다른 output도 같은 root/day/lock에 결속한다. lock inode 교체·삭제·root rollback이 없는 신뢰 전제를 별도 명기한다.
- 최초 실행 및 paused 재개는 active 상태와 디렉터리 변경의 durable 성공이 반환되기 전까지 wire 0. dirty marker 생성 및 fsync → 임시 상태 write/file sync → rename/dir sync → dirty 제거/dir sync 순서를 이용하는 `internal/accountbudget/observation_linux.go:298-359`는 참조 패턴이다. dirty 존재는 무조건 거절하며 복구·자동 삭제하지 않는다. 단일 기록만 존재하는 모형에서 최초 상태 미존재와 storage rollback·파일 삭제는 구별 불가하다.
- Pause의 필수 사전조건: caller가 새 wire 진입을 막고 이미 시작된 모든 요청/응답·관련 작업이 종료되었음을 확인한 뒤 pause를 호출하고, pause 이후 다시 wire하지 않는다. 이 컴포넌트는 HTTP에 연결되지 않으므로 해당 drain 및 first Write와의 선형화를 **증명할 수 없다**. paused durable 성공 후 unlock하며, unlock/close 실패나 sync 오류에는 성공 대신 거절한다. pause 성공 직후 crash와 응답 유실의 재시도는 paused가 이미 durable이고 caller의 무전송/완전 drain 전제가 지켜졌을 때만 안전하다. 일시중단 전 crash는 active 또는 dirty로 남아 fail closed해야 한다.
- abort 기록 실패는 active 또는 dirty를 보존해 이후 거절하도록 설계한다. fsync 실패 이후 파일이 보였다가 crash로 사라지는 가능성과 저장장치 장애/rollback은 이전 절의 한계를 그대로 따른다. 어떤 실패에도 자동 unlink/재기준화하지 않는다. 서로 다른 state 결과만으로 관측 실패와 actual wire 0, 실측 적격성을 도출하지 않는다.
- fixture 검증 제안: 동시 process 잠금, paused→active 및 중복 재개, active→paused 후 정상 재개, 실행 중 kill→active 거절, 상태 전이 각 marker/write/sync/rename/unlink/fsync fault 및 restart, lock inode/root 교체, abort persistence, completed 후 수집 금지 및 report 읽기 경계, caller drain 위반을 드러내는 fake-inflight 계약 시험. 프로세스 kill은 전원손실 시험이 아니다. 재개 모드 선택·완료 의미·권한/root 신뢰 및 운영 바인딩은 아직 미확정이다.

## 추가 기술 비교: 정상 재개 승인 기준 (2026-11-06)

Manager 정정: Q2 정상 재개는 승인됨. 앞 절의 사용자 승인 여부 유보 문구는 철회한다. immutable daily claim은 Q2와 충돌하므로 이 작업의 채택안이 아니다. 아래는 Manager 명세에 넣을 후보이며 실제 binding/안전성 입증이 아니다.

### 파일 프로토콜 vs 기존 SQLite
- SQLite는 이미 `modernc.org/sqlite v1.54.0` (`go.mod:16`, `go.sum:471-472`), 추가 패키지 설치는 필요 없다. `internal/journal/journal.go:212-224`의 WAL + synchronous FULL은 프로젝트의 참고 사례다. `internal/candidate/store.go:424-433`은 NORMAL이므로 그대로 차용할 수 없다. 두 저장소 어느 것도 4상태 owner 수명, 신뢰 root, KST 날짜와 전송 drain 계약에 대한 구현은 아니다. SQLite는 transaction·atomic commit 활용 가능하나 `database/sql` 연결 수명 외에 실행 전체의 별도 flock, DB/WAL/SHM identity·경로, 생성/초기화·손상·pragma 검증 및 commit 에러 결과 불명 처리가 여전히 필요하다. DB 내부 transaction만으로 실행 수명 전체 owner 독점은 되지 않는다. SQLite driver 파일 경로와 열린 root dirfd를 같은 객체에 결속하는 안전한 방법도 미검증이다.
- 최소 후보는 **사전 provision된 파일 하나의 append-only state frame + 수명 전체 flock**이다. pre-provision된 전용 root에 고정 lock 파일과 state 파일을 만들고 최초 생성 시 파일/디렉터리 sync까지 명세해야 한다. 파일 fd를 검증해 `O_APPEND`로 쓰고, sequence·버전·정확한 day·상태·고정 크기 또는 엄격한 길이·무결성 검사 필드를 담은 frame을 단일 writer가 기록하고 full write/파일 fsync 확인. 읽을 때 전체 프레임 순서·전이·길이·무결성을 검사하고 부분 마지막 frame도 거절한다; 자동 truncate/복구 금지. checksum은 우발 손상 검출이며 악의적 변조·전체 프레임 삭제/rollback 증명은 못한다. 성장 제한 초과도 거절. `tools/a140-breakout-signal-shadow/ledger.go:107-130`의 append+sync만으로는 framing·다중 writer·durable authority를 충족하지 않는다.
- append 후보도 기존 production-ready helper는 없다. 검증 로직·fault injection·정상/비정상 전이의 충분한 시험 비용이 있으므로 코드량만으로 SQLite보다 안전하다고 단정하지 않는다. 둘 다 새 의존성 0이며, 특정 환경의 storage/kernel/fsync 전제·first-creation 내구성·경로 신원 및 I/O 오류 모델이 공통으로 필요하다. no-network 제약으로 upstream SQLite 문서는 이번 연구에서 조회하지 않았다.

### 최소 4상태 계약
- `Begin`: 잠금 전 검증된 고정 root dirfd와 lock/state 파일 identity, 정확한 `YYYY-MM-DD` 및 caller가 제공한 KST 거래일 검증. 그날 초기/paused만 active 전이. stale active, aborted, completed, 잘못된 순서·손상·잠금 불가·날짜 불일치/불확실성은 거절. active frame 쓰기+fsync 확인 전 성공/전송 허가 0, 오류 후에도 해당 호출은 허가 0. 활성 상태에 있는 전체 수명 lock 보유. paused에서 당일 재개할 때 기존 Q2 universe 복원은 후속 runtime의 책임.
- `Pause`: caller가 신규 전송을 차단하고 같은 세션의 in-flight 전부 drain, 이후 그 실행의 전송이 없음을 보장한 뒤에만 호출. paused frame의 durable commit 후 lock release. commit 전 crash면 검증 가능한 마지막 active/손상에 따라 거절; commit 후 crash면 paused 재개 가능. lock unlock/close 오류는 paused commit 자체를 되돌리지 않는다: 호출 결과는 실패/불명일 수 있어도 다음 실행이 검증된 paused를 읽으면 재개 가능. 모듈 단독으로 drain을 증명할 수 없다.
- `Abort`: active→aborted; 전이 write/sync 실패 뒤 기존 확정 active 또는 invalid/torn log는 다음 시작을 거절. 그러나 write/commit의 오류가 실제 미반영을 뜻하는 것은 아니다. aborted가 남아도 거절. 이미 전송된 요청은 취소된 것으로 소급 주장하지 않는다.
- `Complete`: caller가 Pause와 동일한 차단/drain을 선행한 뒤 active→completed. completed는 같은 날 수집 재개를 거절하며 읽기 전용 보고는 별도 경로; 측정 유효성·원장 완결성은 별도 검증이다. 완료 시점과 증거 결합은 런타임 후속 계약.
- `paused`에서 다시 `active`가 durable 된 뒤만 새로운 run의 wire 허용. 전이의 지원 범위, 최대 frame 개수, 승인된 KST date source, clock rollback/자정 경계, state 파일에 담을 session identity와 실제 runtime 재확인은 명세에서 구체화한다. 환경변수는 신뢰되지 않은 입력이므로 환경변수 이름 자체는 권위가 아니다. 프로덕션 고정 이름/경로를 선결하지 말고 Manager가 root의 명시적 주입과 동일 identity 보증을 정한다.

### postcommit 반례 및 오류 계약
- dirty-marker 방식은 marker unlink 성공 후 최종 directory fsync 실패하면 `internal/accountbudget/observation_linux.go:356-359`: 같은 부팅에서 dirty가 보이지 않고 새 paused state가 이미 보일 수 있다. crash 후에는 제거 전 dirty 또는 제거 후 상태가 보일 수 있다. 따라서 **모든 I/O 오류가 이후 실행을 영구 차단한다는 보장은 거짓**. append 방식도 full frame 쓰기 후 fsync 실패라도 유효 frame이 즉시 보이거나 crash 후 남을 수 있다.
- `Begin` 오류 후 active가 보이면 stale active로 거절; 오류 후 이전 paused만 남았다면 새 Begin을 허용할 수 있으나, 실패한 Begin 호출은 wire 0이었고 이전 paused의 caller drain 조건·저장소 보존 전제가 성립해야 한다. storage error/rollback이 이미 승인된 active frame을 사라지게 만든 경우 파일만으로 과거 wire 없음을 증명할 수 없다. 이때 신뢰 가능한 storage health 복구 및 root 보존은 외부 전제이며 무조건 재개라고 단정하지 않는다.
- `Pause` 오류라도 유효 paused가 commit되었다면 그것은 정상 재개 가능 상태일 수 있다. 전송 금지/drain 조건이 깨졌다면 위험하므로 이 조건은 실패 전후 무조건 유지한다. `Abort` 완료/오류 후 active 또는 aborted만 허용하는 단조 상태 전이와 malformed 거절을 시험한다. 오류 시 caller의 해당 호출 전송 금지와 미래 판정은 분리한다.

### 필수 표적 검증 (미실행)
- 경쟁 두 process: Begin 하나만 성공, lock 보유 중 pause/begin 경쟁 거절, lock inode 교체·다른 root·hardlink/symlink/owner/mode·root identity·지원 FS·틀린 날짜 거절.
- frame partial write, short write, 무결성 불일치, invalid sequence, file sync 전/후, close/unlock 오류, creation dir sync, crash 전/후의 살아남은 파일 상태 각각 재시작 판정; 반환 오류지만 유효 paused frame이 관찰되는 사례, Begin fsync 오류인데 active가 관찰되거나 이전 paused만 관찰되는 사례, Abort 실패 후 active/aborted/부분 frame을 명시 검증. dirty 대안을 쓴다면 unlink 직후 마지막 dirsync 오류 필수.
- kill before pause commit → active/불완전 기록 거절; kill after verified pause commit → 정상 재개 허용; `completed`는 수집 거절·읽기 전용 보고 분리; fake caller로 drain 전 pause 금지 계약, 이후 write 금지 위반 탐지. process kill은 전원 차단이 아니며 고의 I/O fault fixture만으로 디스크 내구성 보장 입증 불가. 실제 시험·운영·네트워크·소스 수정 없음.
