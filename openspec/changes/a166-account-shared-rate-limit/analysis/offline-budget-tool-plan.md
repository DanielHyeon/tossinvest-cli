# 네트워크 없는 예산 provision와 preflight 도구

독립 검토의 P1 두 건을 아래 경로/취소·clock 계약에서 정정했고 Manager가 제한 구현을 승인한다. 선행 runtime v2 게시 뒤 착수한다. tasks 4.1의 운영 준비를 반복 가능한 로컬 절차로 만들기 위한 최소 도구다. 실제 운영 실행은 별도 사람 단계다. `operator-cutover-draft.md`의 drain·30초 정지·구 root fencing을 자동 실행하거나 입증하지 않는다.

## 연구와 범위

저장소 source 검색에서 core Provision/OpenContext/ProvisionObservation, runtime strict manifest parser와 임시 fixture를 찾았고 별도 운영 installer는 찾지 못했다. 기존 stdlib와 이미 설치된 golang.org/x/sys v0.46.0만 사용한다. 새 패키지 설치·네트워크 검색으로 불필요한 외부 구현을 도입하지 않는다. 대상 Linux ext4의 기존 지원 계약을 유지하며 tmpfs는 fixture다. 지원 밖 플랫폼의 provision은 명시 오류다.

`tools/a166-budget-state`에 명시적 `provision`과 `check` 두 동작만 둔다. 기본은 usage 오류이며 자동 설정·복구·정지·서비스 실행·env 수정·credential 읽기·client 생성·token/브로커 호출은 없다. 시스템 디렉터리를 스캔하거나 계좌를 조회하지 않는다. provision은 root/manifest/scope/instance 인자를 모두 요구하고 check는 manifest만 받는다. path는 absolute clean, symlink 금지다. context 인자를 받는 기존 API에는 취소를 전달한다. core Provision은 context를 받지 않으므로 전후 취소 확인만 하며 fsync/커널 IO 도중 즉시 중단이나 유한 timeout을 보장하지 않는다.

## provision 계약

운영자가 미리 만든 소유자 0700의 빈 전용 root만 사용한다. 기존 directory entry가 하나라도 있으면 수정 전에 거절한다. core Provision가 모르는 기존 파일을 함께 둔 디렉터리에 새 상태를 추가하지 않는다. manifest는 root 밖의 소유자 전용 parent에 두고, 기존 destination·symlink·hardlink·불완전 이전 실행 결과를 덮거나 자동 지우지 않는다. runtime의 같은 canonical path/ancestor 경계를 재사용한다. manifest 파일 생성·rename은 검증한 parent fd에 결속한다. core Provision/OpenContext는 root 경로를 다시 여는 기존 API이므로 root 전체가 최초 fd로만 처리된다고 주장하지 않는다. 최초 root fd identity를 보존하고 Provision 후 OpenContext 및 최종 manifest 게시 직전에 경로/identity를 재확인해 불일치면 게시하지 않는다. 동일 UID의 악의적 관리자에 의한 경로 교체가 없다는 기존 신뢰 전제 아래의 로컬 도구이며, 이 전제를 배포 환경에서 만족하지 못하면 실행하지 않는다. 적대적 동일 UID 방어를 위한 신규 dirfd provisioning API는 이 범위에서 만들지 않는다.

순서는 core Provision → OpenContext → ProvisionObservation → root/queue/wire/event ID의 candidate manifest 구성 → VerifyIdentity 및 close → manifest 게시다. scope/instance의 기존 문자·길이 검증을 사용한다. 생성 실패는 원 오류와 cleanup 오류를 합치고 부분 root를 남겨 재실행을 거절한다. API 자체의 기존 파일 실패를 성공으로 해석하지 않는다. 임시 Budget은 정확히 한 번 소유·종료하며 cleanup은 취소된 context로 무한 재시도하지 않는다.

candidate manifest 작성은 기존 runtime manifest 타입/직렬화를 재사용한다. private 타입을 중복하지 않도록 같은 runtime 패키지의 좁은 offline provision API를 두고 CLI는 인자·출력만 담당한다. reader와 writer의 roundtrip fixture를 둔다. field는 v2 event_lock을 포함하고 observation state의 version1은 유지한다. parent fd에 무작위 이름의 O_EXCL 0600 임시 일반 파일을 만들고 완전 write·file sync·close 후 Linux renameat2 RENAME_NOREPLACE로 최종 이름에 게시한다. 부모 fsync까지 성공해야 성공을 보고한다. 이미 존재하는 manifest를 교체하는 rename fallback은 없다. 오류 시 이 호출이 만든 임시 파일만 정리하며, 최종 이름으로 이미 게시된 파일이나 root를 삭제하지 않는다. 게시 뒤 fsync 실패는 실패로 보고하고 경로가 남을 수 있음을 명시한다.

공통 path 검증을 위한 작은 helper 추출이 필요하면 readManifest의 기존 허용/거절 동작을 사전 FLM/BTM·회귀로 보존한다. 범용 filesystem 프레임워크는 만들지 않는다. 이 도구가 새 root에서 측정한 ID는 후보 신뢰 입력이며 실제 모든 참여 namespace가 같은 객체를 본다는 운영자 검증을 대체하지 않는다. 임의 기존 root를 열어 자기 승인하는 runtime 동작은 추가하지 않는다.

## check 계약

ModeRequired로 현재 v2 Factory.Open을 호출하고 별도로 clock.SystemBootSource().Snapshot()을 한 번 확인한 뒤 Close한다. Factory.Open만으로 현재 boot/time namespace snapshot을 확인했다는 주장을 하지 않는다. NewClient/Accounts/토큰 호출은 없다. 성공은 로컬 manifest·pin·state·지원 clock 조건이 지금 유효하다는 뜻이다. 모든 참여자 포함·전송 중 없음·서비스 준비·금전 권한을 뜻하지 않는다. active owner/clock/fs/dirty 오류는 그대로 실패로 보존한다. 상태 파일을 복구·초기화하거나 poison을 해제하지 않는다. event store Open의 기존 read-only 동작과 잠금만 허용한다.

## 수용 fixture와 문서

임시 root/manifest에서 provision→check 성공, 재실행 거절/기존 내용 불변, nonempty root·inside-root manifest·symlink/권한/owner/지원 filesystem 거절, write/file-sync/rename/dir-sync/close fault 후 결과·잔여 파일·재실행 경계를 확인한다. v1/missing/복사 lock/dirty check 실패, 취소된 context, 출력에 credential/account/body 없음, client/HTTP constructor 도달0을 구조·fixture로 확인한다. fault injection은 test-only이며 실제 disk fault나 powerloss 검증이라고 하지 않는다. 모든 fixture는 개발 임시 경로와 no-route namespace 안에서만 실행한다.

operator 초안에 실제 지원 인자와 fixture SHA를 붙여 runbook으로 바꾸되 tasks4.2 사람 검증은 남긴다. 전체 tests/gate·운영 경로 provision·컨테이너 접근·서비스 중지·설정 적용·브로커·자동 rollback은 범위 밖이다. impeccable: N/A — offline CLI/backend-only. grill-me는 기승인 strict 정책과 운영 금지를 보존하고 실패·권한 경계를 명세한 것으로 적용한다. 독립 설계 검토 후 사전 AST/FLM/BTM→RED/GREEN→변이·scoped gstack→독립 구현 검증을 수행한다.
