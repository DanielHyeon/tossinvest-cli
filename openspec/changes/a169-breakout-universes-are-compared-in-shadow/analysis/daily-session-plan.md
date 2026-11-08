# 거래일별 측정 세션 영속 상태

2026-11-06 Manager. 기준 `232d6b86c024dfd71cc8ffe6fd0fbba342b3e9ec`. **독립 설계 리뷰 통과, 한정 계약 READY — scoped strict/PM 및 Pre-Edit 선커밋 뒤 구현**. 사용자 진행 승인과 기존 Q2의 정상 재개를 유지한다. 처음 검토한 일별 immutable claim은 정상 재개까지 막으므로 채택하지 않는다. 정상 재개 질문은 기존 승인으로 해소되어 철회했다.

## 1. 범위와 정책

새 `internal/shadowsession`에 runtime 미연결 Linux 저장 컴포넌트와 unsupported stub만 추가한다. 실제 요청·측정·엔진·기존 관측 저장소·주문 원장은 바꾸지 않는다. UI 영향 없음(impeccable: N/A — backend/documentation-only). 기존 함수 본문 편집은 예상하지 않으며 필요하면 멈추고 사전 FLM/BTM부터 고정한다.

상태는 `active`, `paused`, `aborted`, `completed` 네 가지다. 부재→active, paused→active, active→paused/aborted/completed만 허용한다. 기존 active는 이전 실행 중 비정상 종료 또는 결과 불확실을 뜻하므로 새 시작을 거절한다. aborted/completed도 같은 날 수집을 다시 시작하지 못한다. 정상 일시중단만 paused로 남겨 기존 universe 복원 기반 재개를 지원한다. 정상 수집 완료 후 보고서 재생성은 별도 읽기 경로이며 재수집을 허용하지 않는다.

오류·강제 종료의 의미는 마지막 검증 가능한 영속 상태로 판정한다. active 실행 중 kill이면 거절하지만, 안전한 pause가 커밋된 뒤 kill이면 paused 재개가 가능하다. 함수가 오류를 반환했다고 쓰기가 반드시 미반영이거나 미래 시작이 무조건 금지된다고 주장하지 않는다. 이 규칙은 기존 과거 gap 제외 정책을 바꾸지 않는다.

## 2. API와 호출자 경계

`Root{Path string, Device uint64, Inode uint64}`와 `Begin(ctx, root, day) (*Session, error)`, `Session.Pause/Abort/Complete/Close() error`를 제공한다. Root identity는 신뢰하는 실행 경계가 사전에 전달해야 한다. expected identity를 대상에서 자동 읽어 채우는 API, 자동 root 생성·복구·삭제·전환은 없다. day는 엄격한 `YYYY-MM-DD` 유효 날짜(연도 1~9999)다. 현재 KST 날짜·공식 거래일·자정 경계 일치 여부는 후속 runtime의 책임이며 이 모듈은 시계를 읽지 않는다.

Begin은 active 기록과 file/directory sync 성공 및 마지막 context 확인까지 끝나야 Session을 반환한다. 오류이면 Session=nil이며 해당 호출자는 wire를 시작할 수 없다. 성공 이후 요청을 실제로 이 상태와 결속하는 기능은 이번 로트에 없다.

Pause/Complete의 사전조건은 caller가 신규 wire 진입을 막고 기존 모든 요청·응답·작업을 drain했으며 해당 Session으로 앞으로 전송하지 않는 것이다. Abort도 신규 작업을 멈춘 뒤 사용한다. 이 모듈만으로 HTTP drain/first Write를 증명하지 않는다. 각 종료 메서드는 시도 후 성공 여부와 무관하게 해당 handle을 폐기하고 자원을 정리한다. 다시 종료 전이를 요청하면 ErrClosed, Close 반복은 원래 정리 결과를 반환한다. 전이 없는 Close는 active 기록을 그대로 둬 재시작을 막는다. Session mutex로 메서드 간 경쟁을 직렬화한다. 외부 callback·timer·재시도 framework는 만들지 않는다.

## 3. 신뢰 root와 배타성

절대·정규·root 자체가 아닌 경로만 받는다. 각 component를 dirfd/openat/O_NOFOLLOW로 열고 최종 디렉터리 owner=euid, mode0700, expected device/inode 및 ext4를 확인한다. tmpfs·다른 filesystem은 재부팅 내구성을 주장하지 않고 거절한다. 비Linux는 fail closed다.

날짜별 고정 이름 `<day>.lock`, `<day>.state`만 dirfd 아래 연다. 파일은 NOFOLLOW/CLOEXEC/NONBLOCK, regular·owner=euid·0600·nlink1을 검증한다. lock 파일 생성/열기는 flock보다 먼저다. O_EXCL 생성이 EEXIST이면 동일 이름을 다시 열어 검증하고, 열린 inode와 이름의 identity를 대조한 뒤 nonblocking exclusive flock을 획득한다. 이 잠금은 **state 읽기/생성 전** 획득하여 Session 수명 전체 유지한다. 최초 동시 생성에서도 하나의 검증된 lock inode를 공유해야 한다. 충돌은 즉시 ErrBusy, 임의 대기 없음. lock 파일은 제거하지 않는다. 생성 여부와 무관하게 잠금을 획득한 각 Begin에서 lock file/directory sync를 마치며 오류이면 시작 허가를 주지 않는다. state는 O_EXCL 생성과 기존 파일 구분을 유지한다. 기존 빈 파일은 손상/미완으로 거절한다.

root·lock·state identity를 최초 open/lock 후, 기록 전, 반환 전 다시 대조한다. 재검사 자체가 비협조 프로세스의 무제한 치환을 원자적으로 막지는 않는다. 모든 실행이 같은 신뢰 root와 identity를 쓰고 파일 삭제·교체·snapshot rollback·독립 복제·다른 namespace 우회가 없다는 전제가 필요하다. 다른 output을 쓴다는 이유로 root/day를 바꾸지 못한다는 실제 CLI 계약은 후속이다. 지원 storage가 성공한 sync의 내구성을 지킨다는 전제이며 실제 운영 환경 확인은 별도다.

## 4. 저장 프로토콜

단일 state 파일의 append-only canonical JSONL을 사용한다. Go 고정 struct의 필드 순서와 JSON 이름을 `version`, `day`, `sequence`, `state`, `checksum`으로 고정하고 encoding/json.Marshal 결과 뒤 LF 1개를 붙인다. version은 정수1, sequence는 uint64의 1부터 연속 증가하는 정수(overflow 거절), day는 API와 같은 유효 날짜, state는 네 값 중 하나다. checksum은 마지막 필드를 제외한 같은 순서의 struct를 json.Marshal한 바이트의 SHA-256을 소문자64자리 hex로 표현한다. decoder는 unknown/duplicate/trailing 데이터와 비정규 인코딩을 canonical round-trip 바이트 대조로 거절하며 마지막 LF 없는 부분 행도 거절한다. 모든 행의 version/day/sequence/checksum/전이와 마지막 상태를 검사한다. 최대 파일 크기는 128 KiB이며 초과 또는 다음 행 추가로 초과하면 거절한다. truncate·자동 복구는 없다. 원장에는 계좌·종목·주문·자격 정보를 넣지 않는다.

생성한 state는 각 행 전체 Write→file Sync→directory Sync를 마친 뒤에만 커밋 성공으로 반환한다. 기존 paused 재개도 active 행을 같은 순서로 저장한다. 파일은 lock 아래에서 읽고 append하며 inode/size 변화도 검사한다. 짧은 Write·각 IO 오류는 전파하고 해당 handle을 폐기한다. 단순 정리 오류는 앞선 저장 결과를 되돌리지 않는다. 같은 root/day의 다음 Begin이 관측한 유효한 기록이 최종 재개 판단의 근거다.

예를 들어 Pause 행이 저장된 뒤 Sync 또는 unlock/Close가 실패하면 호출자는 오류를 받더라도 다음 실행이 유효 paused를 읽을 수 있다. Pause 전에 drain하고 이후 절대로 전송하지 않는 조건 때문에 이는 안전한 재개다. 반대로 Begin의 active 저장 결과가 모호하면 해당 Begin은 허가하지 않고, 이후 active가 보이면 거절한다. 이전 paused만 남았으면 이전 호출이 wire를 시작하지 않았다는 Begin 계약 아래 다시 시작할 수 있다. Abort 쓰기 실패에서는 기존 active 또는 손상 상태가 재시작을 막는다. 내구성 약속 위반·파일 삭제/rollback 뒤 과거 실행 부재를 증명하지 않는다.

## 5. 검증

Terra가 새 package 영향·API·오류·락 수명·미실행 범위를 Pre-Edit로 먼저 커밋한다. 기존 함수 변경이 없다면 FLM/BTM 해당 없음 사유를 명시하고 새 함수의 구현 후 AST/분기 지도를 보존한다. 기존 함수 편집이 생기면 완전한 사전 지도 없이 진행하지 않는다.

정확 선택 RED/GREEN: 최초 시작, 정상 pause/resume, abort/completed 차단, Close/kill 후 active 차단, pause 커밋 후 종료 재개, 두 process 동시 시작/재개, 서로 다른 output 논리 입력에서 동일 root/day 배제, 날짜·다른 날, nil/context 취소, 닫힌 handle·동시 종료, root/inode·symlink/hardlink/permissions/filesystem, 빈/부분/손상/unknown/duplicate/checksum/sequence/전이/크기 오류. 새 시험만 추가하고 기존 test 함수도 임의 편집하지 않는다.

결정적 private nil-default IO seam으로 create/write/short write/file sync/directory sync/unlock/close 오류를 주입한다. fail-before 및 실제 commit/실제 unlock 뒤 오류를 구분하고, 실제 fd 정리를 막는 가짜 오류로 누수를 숨기지 않는다. 같은 프로세스와 subprocess 재시작에서 결과를 확인한다. kill은 자신이 만든 fixture child만 대상으로 하며 전원 손실 시험으로 주장하지 않는다. critical 경계 변이(active sync 생략, active/aborted 재개, flock 제거, identity/checksum 무시)는 행동 시험으로 검출하고 동등 변이는 검출 수에 넣지 않는다.

정확 package 범위 normal·tagged race, 필요 subprocess 반복, unsupported compile, scoped gstack·독립 적대 리뷰 후 Manager 검증한다. 전체 tests/gate·실브로커·운영·측정·배포는 금지다. production caller 0과 기존 코드 diff0을 확인한다. runtime/caller/engine identity/실측이 남으므로 a169 전체를 archive하지 않는다.

## 6. 연구와 준비 근거

`daily-session-research.md`에 사전 조사·4상태 및 파일/SQLite 비교와 도구 한계를 보존한다. 기존 공개 helper는 해당 범위의 완전 일치 구현이 아니며 accountbudget 비공개 함수 export나 DB 의존 확대 대신 현재 설치된 Go1.25/x/sys v0.46.0와 stdlib만 사용한다. 설치된 SQLite v1.54.0는 검토했으나 이 작은 단일 writer 상태에 별도 DB/sidecar 수명과 경로 경계를 추가하지 않기로 했다. 새 dependency·라이선스 변경 없음. 외부 registry/GitHub 검색은 하지 않았고 syscall 문서·로컬 코드/AST를 사용했다.

원본 main의 무관한 a180 커밋과 .claude/.codex/a128 변경은 보존한다. ext4 격리 worktree에서 CodeGraph1.6.0 index up to date를 확인했으나 별도 `.sdd/index-state` freshness gate PASS는 아니다. 기존 source 분기 근거는 연구 AST와 원문을 source SHA로 다시 대조한다.

grill-me/grilling: 정상 재개는 이미 승인된 Q2로 해소했다. root 보존·storage 내구성·caller drain·실제 운영 결속을 가정/미검증으로 명시했다. 연구·선택 시험·독립 리뷰는 실제 활성화나 전체 완료 증거를 대신하지 않는다.

## 7. Manager 제한 READY

독립 reviewer의 P2(최초 lock 생성/동기화 순서, canonical 형식 범위)를 §3/4에 반영했다. 수정 재검토는 bounded READY, 새 P0/P1 없음이다(`daily-session-design-review.md`). Manager는 이 미배선 저장 컴포넌트만 승인한다. Terra는 계약·연구·검토 원문과 Pre-Edit를 strict/PM 검사 후 먼저 커밋하고 그 뒤 RED/GREEN을 시작한다. 기존 함수 변경0 경계를 지키며 필요 변경이 생기면 사전 FLM/BTM과 범위를 다시 검토한다. 전체 측정 모드 freeze, 실제 caller drain, 운영, archive 승인으로 확대하지 않는다.
