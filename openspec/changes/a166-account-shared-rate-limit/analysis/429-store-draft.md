# 429 공유 기록과 poll hint — 검토용 저장 계약

상태: 아래 마지막 절의 미배선 store leaf만 Manager 구현 승인. transport/호출자 통합과 운영은 별도 검토 전이다. 기준 `c412c0e55f2948e16e1f33f1e1e8fadaf2092c56`. 관련 원문은 `accountbudget.Transport.RoundTrip`, `Budget.write/Acquire/Lease.Release`, `Retrier.Query`, official doRequest/token/a112, replay/probe다. `/tmp/a166-429-storage-research.md`와 `/tmp/a166-429-checkpoint-options.md`의 대안을 검토했다. 아래 사항은 기존 제품에 존재한다고 주장하지 않는다.

## 유지할 계약

strict 350ms admission과 보호 대기 정책은 그대로다. 모든 실제 enforcing 응답의 429를 body 읽기 전에 관측하며 header 유무·후속 성공과 무관하게 기록한다. POST의 원래 response/status/body/error, 전송 여부와 원장 판정은 저장 실패가 덮어쓰지 않는다. 새로운 계좌 전체 429 cooldown은 없다. 엔진 poll의 해당 계좌/역할/RequiredQuery에만 유효 hint를 적용한다. token 교환의 429도 공유 사건이지만 실제 poll GET 응답의 hint로 오인하지 않는다.

## 기록과 관측 공백

공유 저장은 계수와 최근 128개 이벤트 창을 함께 제공한다. 모든 429는 단조 증가 event sequence를 받고 창 안에 개별 record로 원자 저장된다. 전체 영구 감사 로그를 새로 만들지 않는다. 창 밖 cursor는 정상 빈 결과가 아니라 명시적 overrun이다. 읽는 쪽은 overrun을 관측 연속성 실패로 처리한다. attempt/event/gap sequence는 uint64이며 wrap하지 않는다. 다음 증가가 overflow할 수 있으면 새 wire 전에 실패한다. 전체 상태는 256KiB, poll slot은 256개 이하로 제한하며 미만료·poisoned slot을 용량 확보를 위해 버리지 않는다. records에는 고정 enum, sequence와 boot 표본만 두고 credentials·URL·raw header/body·계좌/주문 식별 원문은 넣지 않는다.

단순 header 후 기록만으로는 응답을 받은 직후 프로세스가 사라졌을 때 누락을 식별할 수 없다. 이 초안은 enforcing 전송마다 wire lease를 얻은 후 첫 HTTP Write 전에 observation pending을 내구 저장하는 방안을 택한다. pending 저장 실패면 아직 보내지 않은 요청으로 거절한다. 이 비용은 모든 priority에 적용되며 새 고정 cooldown은 아니다. 완료된 RoundTrip의 header 결과 또는 정상적인 무응답 오류를 원자 기록해 pending을 끝낸다. 무응답 오류는 `관측된 429 없음`이지 브로커가 429를 내지 않았다는 증명이 아니다. 정상 timeout까지 자동으로 crash gap으로 만들지 않는다.

프로세스 사망 또는 종료 기록 실패로 남은 pending은 다음 wire 소유자가 orphan으로 판정할 수 있다. 이전 요청이 아직 wire를 보유한 동안에는 새 소유자가 될 수 없기 때문이다. pending에는 실제 wire 종류(enum: engine-poll read / token / non-poll)와 engine-poll인 경우 비원문 poll slot key를 함께 저장한다. token과 non-poll에 poll hint를 임의 귀속하지 않는다. orphan은 원래 attempt sequence를 포함한 gap으로 기록하며 성공/no-429로 덮지 않는다. gap 이후의 전송 자체를 새 429 cooldown으로 막지 않지만 engine-poll orphan은 해당 slot을 poisoned로 보존한다. 종류/key를 입증할 수 없는 파손은 scope를 추정하지 않고 상태 오류로 처리한다. dirty·저장 자체의 오류는 required 상태 무결성 실패로 새 wire를 거절한다. 이를 보호 지연이 없다고 표현하지 않는다. poisoned hint의 해제는 자동 만료가 아니라 별도 오프라인 복구 절차가 필요하며 이 로트에서 운영 복구를 실행하지 않는다.

독자 snapshot은 최근 확정 records와 pending/gap/overrun을 별도로 제공한다. 정상 pending을 clean 또는 crash로 단정하지 않는다. pending의 생존 확인이 필요한 독자는 event lock을 놓은 뒤 검증된 wire lock을 비차단으로 검사하고, wire를 잡았다면 event state를 다시 읽어 동일 generation인지 확인한다. wire→event 순서를 지키고 검사 뒤 즉시 해제한다. 이 조회는 API 전송 권한을 얻지 않으며 잠금을 기다리는 retry loop를 만들지 않는다. a087/a169가 어느 상태에서 멈출지는 각 change에서 이 명시적 상태를 소비하도록 정한다.

## 잠금·실패·신뢰 경계

queue state/dirty 파일은 재사용하지 않는다. 별도 event lock과 제한된 상태 파일을 같은 trusted budget root에 둔다. 고정 inode 결속은 root와 영속 queue/wire/event lock에만 적용한다. rename으로 교체되는 state와 생성/삭제되는 dirty는 고정 inode를 manifest에 넣지 않으며 dirfd 기반 open과 regular/owner/mode/nlink·bounded schema를 검증한다. 전체 required factory는 manifest version 2와 event lock identity를 요구한다. 현재 version 1을 조용히 승격하거나 새 파일을 자동 생성하지 않는다. 기존 version 1은 첫 wire 전 명시적 거절이며, 별도 오프라인 provisioning으로 새 root와 version 2 manifest를 만드는 절차·fixture를 제공한다. 현재 운영 설정은 바꾸지 않는다.

허용 순서는 wire→event뿐이다. event lock 보유자는 queue/wire를 취득하지 않는다. poll reader는 event만 사용하며 pending 생존 조회는 위 재검증 절차를 따른다. queue.lock 안에서 event IO를 하지 않는다. 기존 atomic temp/write/fsync/rename/dir-fsync와 전용 dirty marker 선례를 확장한다. 원자성 실패를 임의 초기화·빈 상태·OFF로 복구하지 않는다. 요청 context와 종료 오류를 보존하고, 이미 받은 응답의 기록 실패는 기존 callback 및 request-local 관측 결과로 드러낸다. callback 성공을 내구 저장 성공으로 간주하지 않는다.

이 모델은 신뢰된 파일 시스템·fsync·동일 namespace·관리자 비변조 전제다. 파일 시스템이 성공을 거짓 보고하거나 모든 내구 기록 자체를 유실시키는 장애까지 검출한다고 주장하지 않는다. 추가 서비스·DB·비동기 메모리 큐·계좌 동등성 DB는 도입하지 않는다.

## poll hint의 저장과 전달

RequiredQuery 호출은 명시적 poll scope를 전달한다. slot key는 domain separator와 길이 구분한 group instance·기존 canonical account reference·고정 engine-poll role·RequiredQuery enum의 SHA-256이다. 이 값은 파일 경로·event payload·로그에 쓰지 않고 0600 상태 내부의 key로만 쓴다. credential/token을 hash하지 않으며, hash는 계좌 동일성이나 권한 증명이 아니다. 기존 caller의 계좌 결속을 먼저 확인하고 byte identity 및 SHA-256 충돌 저항을 전제로 사용한다. raw account는 저장하지 않지만 pseudonymous key도 비공개 정보로 취급한다.

실제 engine poll의 GET 사슬(그 poll이 필요로 하는 Accounts 선행 읽기 포함)은 같은 논리 slot을 가진다. token 교환은 priority/deadline과 request-local collector를 계승하되 wire 종류를 token으로 지정해 poll hint 쓰기를 하지 않는다. token의 429도 collector에 남아 해당 논리 Query의 같은 호출 재시도를 중단한다. POST·flatten·별도 walk에는 engine-poll hint scope를 전달하지 않는다. 실제 GET과 token 종류를 enum으로 구별하는 생산 경계를 구조·통합 시험으로 고정한다.

429의 hint는 해당 응답의 Retry-After만 파싱한다. 유효 0..30s는 기존 미만료 값과 최대 잔여시간을 유지하고, missing/invalid/>30/503/후속 성공은 기존 값을 지우지 않는다. HTTP-date의 응답 수신 시각 변환은 신뢰되는 UTC를 제공하는 주입 clock에서 하고 이후 경과는 boot 표본만 사용한다. UTC가 정확하지 않은 환경의 date hint 정확성은 보장하지 않으며 elapsed 판정으로 wall clock을 재사용하지 않는다. overflow·음수는 invalid, duration은 더 짧아지도록 정수 초로 버리지 않는다. event 추가·해당 hint max·pending 완료는 같은 원자 상태 갱신이다. 부분 성공을 저장 commit 성공으로 보고하지 않으며 기존 HTTP 응답은 그대로 반환한다. 다른 boot의 레코드는 원래 duration과 최초 복원 anchor를 같은 상태 잠금 아래 원자 보존한다. 레코드 없는 조회에 일괄 기동 지연을 넣지 않는다.

Query 진입의 사전 확인만으로는 예산 대기 중 다른 429가 만든 hint를 놓친다. 실제 wire 소유권을 얻은 후에도 poll scope의 hint를 다시 확인하고, 제한 중이면 새 Write 0으로 생략한다. 전송 전 생략·응답 429·기록 오류는 같은 Query 안에서 재전송하거나 freshness 성공으로 기록하지 않는다. request-local collector는 header 429를 body 오류·후속 결과와 별도로 보존하며 concurrent Query끼리 전역 마지막 값을 교환하지 않는다.

## 미해결과 수용

복구는 in-place poison 삭제나 자동 초기화로 제공하지 않는다. 운영자가 모든 참여 프로세스를 종료하고 그 사실을 확인한 뒤, 신뢰 monotonic/boot clock으로 30초 이상 계속 정지한 상태를 확인한 경우에만 새 empty root의 오프라인 Provision와 새 instance/version 2 manifest 배포로 전환하는 절차를 문서화한다. 남았을 수 있는 유효 poll hint의 상한이 30초이므로 모든 관측 주체 종료 이후 이 정지 구간을 요구한다. 정지 입증이 깨지거나 clock/namespace가 바뀌면 확인을 처음부터 다시 하며 자동으로 운영을 재개하지 않는다. 이 절차는 정상 기동의 30초 지연이나 전송 시 429 cooldown이 아니다. 사람이 승인·실행하는 손상 복구이며 이 개발 작업에서 실행하지 않는다. 기존 원장·미확정 주문은 보존하고, 구 root의 gap 이력은 증거로 보존한다. event cursor는 instance에 결속되며 instance 변경을 연속된 관측이라고 표시하지 않는다. 실제 모든 참여자 확인이 불가능하면 복구·측정 재개를 보류한다.

version 1→2 전환도 모든 참여자 정지와 위 보수적 정지 확인을 포함한 같은 오프라인 cutover로 한다. 혼합 version이 같은 group을 조용히 사용하지 않도록 v2 factory는 event identity 없는 v1을 거절한다. 실제 운영 manifest를 자동 수정하지 않으며 새 Provision·manifest 검증·누락/복사 lock 거부는 임시 경로 fixture로 검증한다.

독립 복구 검토의 보완: 정지 구간은 마지막 참여자의 로컬 wire 종료/drain과 자동 재시작 차단을 확인한 뒤 시작한다. 서비스·타이머·콘솔·독립 도구 등 모든 구 launch를 차단하고, 중지된 구 root를 증거 보존 위치로 옮겨 구 manifest의 경로/identity로 다시 열 수 없도록 한다. 새 root·단일 v2 manifest를 모든 참여자에게 배포한 뒤, 구 바이너리/구 manifest/구 root로의 재기동이 거절됨을 검증하고서만 새 참여자를 시작한다. 운영자는 구 root를 원래 경로나 alias로 되살리지 않는다. v2 parser 하나가 비협조 구 바이너리·관리자를 차단한다고 주장하지 않는다. OFF launch도 이 실행 전제에서 허용하지 않는다. 이 절차는 별도 사람 승인과 실제 배포 증거가 필요한 운영 작업이다. 로컬 wire 종료·30초 정지는 관측된 유효 hint와 로컬 시작 상한에 대한 보수 처리이며, 브로커 처리 완료나 서버 도착률의 증명이 아니다.

구현 전 독립 검토: pending checkpoint의 IO 비용과 잠금, manifest version 2 전환, poll key 결속, header/store 오류의 caller 전파 및 위 복구 절차. counter-only·header-only 대안의 crash 공백을 숨기지 않는다. 독립 1차 P1인 pending의 wire 종류/slot과 교체 파일 inode 오류는 위 구분으로 보완했다. 이 초안은 검토를 끝냈다는 뜻이 아니다.

수용 fixture는 header429 뒤 body 오류/후속 성공/token429, pending 전 실패 wire0·header 직후 crash·정상 timeout의 no-observed 구분·finish 실패 원래 POST 보존, 두 프로세스 잠금·reader generation 재검증·overrun, 유효 hint max·invalid 보존·대기 중 hint 생성·boot anchor 경쟁·저장 실패를 포함한다. 전체 tests/gate·실계좌·운영·활성화는 미승인이다. search-first는 기존 atomic filesystem/clock/context를 확장하고 신규 의존성은 없다. impeccable N/A — backend/documentation-only. grill-me/grilling 미해결 항목은 위에 명시했다.

## 첫 구현 로트의 한정 승인

독립 검토에서 typed pending, 고정 lock과 교체 파일 구분, offline cutover fencing의 P1이 해소됐음을 확인했다. Manager는 신규 미배선 store leaf의 구현만 승인한다. 명시적 임시 경로 provisioning/open, externally expected event-lock identity, typed pending/begin/finish, ring cursor/gap/overrun, poll slot max/poison/anchor 및 snapshot의 상태 기계와 내구 IO가 대상이다. public begin/finish가 임의 caller의 wire 소유 주장을 믿지 않도록 실제 동일 Budget의 활성 Lease에 결속하고 Release와 경쟁을 검증한다. reader가 관측만으로 전송 권한을 얻지 않게 한다.

기존 Budget.Provision/Open, runtime manifest parser, Transport.RoundTrip, official/Retrier 및 운영 caller는 첫 로트에서 연결하거나 바꾸지 않는다. 새 API가 소비되지 않는 상태임을 명시하고 기존 version 1 runtime을 version 2라고 보고하지 않는다. 새로운 leaf provisioning은 명시 호출만 허용하고 기존 상태를 덮거나 자동 복구하지 않는다. API 구체화 중 기존 함수 편집이 필요하면 먼저 해당 FLM/BTM과 책임을 명시한다. 전체 caller 통합 때만 새 version 2 강제를 적용하고 앞선 모든 manifest fixture를 갱신한다.

기존 계좌 reference의 byte 결속은 hash의 입력일 뿐 다른 프로필·앱 키 간 물리 계좌 동일성 증명이 아니다. poll producer가 실제 사용하는 canonical reference와 slot의 일치, alias/unknown 계좌의 거절 또는 지원 범위는 transport/engine 통합 전 다시 확인한다. 첫 leaf는 고정 크기 비원문 key의 상태를 처리하며 임의 account alias 판정 기능을 만들지 않는다.

이 첫 로트는 상태 기계·디스크 오류·동시성과 lease 결속·clock 복원에 대한 scoped normal/race·변이·수동 gstack·독립 검증을 받아야 한다. network가 필요하지 않은 leaf 시험을 우선하며, 실제 TLS 통합은 다음 로트에서 외부 egress 없는 namespace로만 실행한다. 이 승인으로 현재 pending 기록이나 429 수집이 운영에서 실행된다고 주장하지 않는다.
