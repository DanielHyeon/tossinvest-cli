# 영속 poll hint의 시계 — Manager 세부 계약 후보

2026-10-31. 4판의 응답별 최대 30초 hint를 재시작 뒤 복원하기 위한 작은 선행 로트다. 독립 delta 검토 후 Manager가 시계 표본·순수 산술 leaf의 구현 착수를 승인했다. 일반 rate admission의 매 소유자 350ms 대기는 이 시계나 영속 timestamp에 의존하지 않는다.

## 기존 근거와 선택

현재 `internal/clock/clock.go`의 Clock/LeaseAnchor는 프로세스 안의 elapsed-time용이며 직렬화 가능한 boot/namespace 표본을 제공하지 않는다. `/tmp/a166-429-source-map.md`와 AST 자료가 이 공백을 기록한다. 기존 clock 인터페이스와 fake들을 넓히지 않고 새 별도 BootSource/BootStamp leaf를 추가하는 안을 채택한다. 생산 provider는 clock 패키지 안에서만 syscall/proc를 읽고 주문 경로에는 주입한다. 기존 x/sys/unix와 stdlib를 사용하며 신규 의존성이나 서비스는 없다.

Linux 문서에 따르면 BOOTTIME은 suspend 시간을 포함하고 time namespace의 offset 영향을 받는다. offset은 initial namespace 기준이며 첫 member 진입 뒤 변경할 수 없다. `/proc/.../ns/time`과 `time_for_children`이 다를 수 있으므로 offsets 파일 하나만 읽고 현재 clock의 기준이라고 추정하지 않는다. 출처: [Linux man-pages time_namespaces(7)](https://man7.org/linux/man-pages/man7/time_namespaces.7.html). boot_id는 한 boot 동안 변하지 않는 UUID다. 출처: [Linux 6.9 sysctl random](https://www.kernel.org/doc/html/v6.9/admin-guide/sysctl/kernel.html#random). 문서 사실과 아래 제품 지원 범위를 구별한다.

## 첫 지원 범위

- Linux만 지원한다. 신뢰된 kernel procfs, 실행 중 setns/unshare/time namespace 교체 없음, 현재 namespace와 children namespace가 동일하며 monotonic/boottime offset 모두 0인 환경으로 제한한다. 실제 host/container는 운영 전 별도 확인한다. 관리자에 의한 procfs bind 대체나 런타임 thread별 namespace 조작을 방어한다고 주장하지 않는다.
- 표본은 canonical boot UUID, 시간 namespace device/inode 식별자, CLOCK_BOOTTIME의 비음수 nanosecond duration을 갖는다. namespace ID는 계좌나 인증정보가 아니다. 원문 proc 파일을 로그에 쓰지 않는다.
- provider는 bounded proc 읽기와 strict parsing을 수행한다. unknown/duplicate/missing offset 행, 음수/overflow, offset 비영, UUID 오류, namespace 읽기 실패/불일치, syscall 오류는 명시 오류다. 지원 불명 환경을 wall-clock으로 대체하지 않는다. namespace 및 boot ID를 표본 앞뒤로 비교해 관찰 중 변경도 거절한다. 모든 입력을 무한 읽거나 자동 수정하지 않는다.
- non-Linux는 이 기능을 unsupported로 반환한다. 일반 Clock와 OFF 동작을 바꾸지 않는다. BOOTTIME 지원이 없거나 proc 접근이 막히면 poll hint 복원은 fail-closed이며 새 boot로 추정하지 않는다.

## 복원 산술 계약

영속 hint는 원래 duration(0~30초), 수신 당시 BootStamp와 poll key를 보존한다. 동일 boot 및 동일 namespace이면 now.uptime >= saved.uptime일 때만 차이를 빼며 음의 남은 시간은 0으로 포화한다. 동일 boot인데 namespace가 다르거나 uptime이 역행하면 오류다. 다른 boot에서는 현재 uptime을 빼지 않는다. boot ID 불일치만으로 이전 응답 이후의 재부팅이라고 입증할 수 없기 때문이다.

다른 boot의 유효 hint는 해당 레코드의 원래 duration만큼 보수적으로 보류한다. 첫 복원 시 현재 BootStamp를 별도 복원 anchor로 저장하고 원래 수신 provenance는 보존한다. 동일 boot의 후속 프로세스 재시작은 이 anchor의 elapsed만 빼며 매번 다시 기간을 시작하지 않는다. anchor의 원자적 공유 저장이 성공하기 전에는 그 poll을 허용하지 않는다. 동시에 복원하는 참여자는 저장 잠금 아래 첫 anchor 하나를 선택한다. 실제 reboot가 반복되면 보류가 연장될 수 있으며 wall-time 상한을 주장하지 않는다. 새 30초 기본값이나 레코드가 없는 조회의 기동 정지는 도입하지 않는다. duration0은 보류 없이 만료다. 복원 값은 해당 원래 duration을 넘지 않는다. 모든 덧셈·뺄셈 overflow를 검사하며 wall-clock 변경으로 만료를 앞당기지 않는다. anchor의 잠금·영속 저장은 후속 store 로트이며 이 clock leaf가 구현했다고 주장하지 않는다.

HTTP-date 파싱은 정확한 응답 수신 시각의 주입 wall clock을 사용하고 이 leaf와 별도다. 이 로트는 response/header/event 저장이나 global cooldown을 구현하지 않는다. poll key별 기존 미만료 hint와 max 선택, invalid/누락/>30초 보존 규칙도 후속 로트다.

## 수용 증거

proc parser는 fake 입력으로 strict·중복·범위·오류를 검증한다. snapshot provider는 주입 reader/stat/syscall로 앞뒤 변경·namespace mismatch·syscall 실패를 반증한다. 실제 로컬 kernel read의 선택 smoke는 식별자 값을 출력하지 않고 성공/지원 불가를 구분한다. hint 산술은 fake 표본으로 같은 boot·재시작·reboot·suspend·역행·namespace 변경·duration 경계를 검증한다. production provider를 사용하는 실제 container namespace 검증과 power loss는 이 시험으로 대체하지 않는다. 주요 fail-open 변이를 잡고 Go 1.25/1.26 선택 normal/race·수동 gstack·독립 리뷰를 수행한다. 전체 시험/gate·운영 접근·설정 변경 없음.

search-first: 기존 clock/lease 선례와 공식 Linux 문서를 재사용한다. 새 의존성/라이선스 도입 없음. grill-me/grilling: 사용자가 결정한 strict cap·보호 대기/응답별 hint 계약을 재질문하지 않으며 미지원 환경을 명시한다. impeccable N/A — backend/documentation-only.

독립 검토 처분: `/tmp/a166-poll-clock-independent.md`는 다른 host/restore에서 boot 계보 없이 uptime을 빼면 조기 만료한다는 P1을 지적했다. Manager는 uptime 차감을 제거하고 유효 레코드별 duration·한 번 저장한 복원 anchor로 바꿨다. 이는 보호 대기를 허용한 사용자 정책 안에서 보수적 복원이며 전체 계좌 cooldown은 아니다. 수정 후 Terra 독립 delta 검토에서 해당 P1 해소와 leaf 착수 가능을 확인했다. Manager는 표본·순수 산술에 한해 구현 착수를 승인한다. 다른 boot의 산술 결과는 원 duration과 복원 anchor 필요 여부를 명시적으로 반환하며, 호출자가 anchor 저장 없이 허용 결정을 내릴 수 있다는 뜻이 아니다. store·HTTP-date·실제 poll 연결은 후속 수용 조건이다.
