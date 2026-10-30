# 공통 예산 v2의 운영 준비와 장애 처분

개발용 절차 초안. 아직 실행 가능한 운영 승인서나 검증 완료 runbook이 아니다. actual host/container namespace·UID·파일시스템·동일 inode 및 모든 참여자 포함 여부는 tasks 4.2의 사람 단계다. 이 문서를 쓰거나 개발 시험을 통과해도 서비스 중지·설정 변경·실주문·배포를 실행하지 않는다.

## 사전 조건

모든 공식 API 송신 참여자를 표로 고정한다. 엔진, root CLI/console/verify/soak/candidate, flatten, token 교환과 401 재전송, 별도 도구·수동 probe가 대상이다. WTS/judge는 이 공식 API 예산의 적용 대상과 구분한다. 별도 키라고 별도 예산 root를 만들지 않는다. 계좌 동등성을 입증할 수 없으므로 모든 참여자가 하나의 보수적인 공통 group 총량을 나눈다. 게시된 정확 SHA와 지원되는 required 경로 census가 준비되지 않으면 전환하지 않는다.

운영자는 보호 주문 필요성, 유지보수 창, 이미 전송된 요청·열린 연결·미확정 원장을 확인하고 기존 송신 참여자 전부의 quiescence를 확보해야 한다. 파일 lock을 잡았거나 특정 PID가 없다는 사실만으로 전체 참여자 종료를 증명하지 않는다. 마지막 참여자의 로컬 wire 종료/drain과 서비스·타이머·콘솔·독립 도구·컨테이너 자동 재시작 차단을 확인한 뒤, 신뢰 monotonic/boot clock으로 30초 이상 계속 정지했음을 확인한다. 확인이 깨지거나 clock/namespace가 바뀌면 처음부터 다시 확인한다. 이는 정상 기동 cooldown이 아니라 `429-store-draft.md`의 v1→v2/새 instance 오프라인 cutover 계약이다. 브로커 처리 완료·서버 도착률의 증명은 아니다. 개발 에이전트가 이 단계를 대신 실행하지 않는다.

## provision과 신뢰 입력

개발 후보에는 네트워크 없는 `tools/a166-budget-state`가 있다. 운영 실행은 별도 사람 승인과 설치·바이너리 SHA 확인 이후에만 한다. 아래는 인자 형식이며 지금 실행하는 지시가 아니다.

```text
a166-budget-state provision --root <absolute-root> --manifest <absolute-manifest> --scope <opaque-scope> --instance <opaque-instance>
a166-budget-state check --manifest <absolute-manifest>
```

provision은 빈 root와 외부 manifest를 명시적으로 준비하고 check는 local pin/state 및 현재 boot source만 검사한다. 부분 실패 파일은 자동 수리하지 않는다. 임시 fixture 결과를 실제 운영 provision·동일 namespace·drain 증거로 취급하지 않는다. 독립 검증된 후보와 fixture SHA는 해당 lot의 review 및 영수증에서 확인한다.

installer 수용 계약은 다음과 같다. 다른 내용이 없는 새 절대·정규 경로, effective UID 소유 0700, symlink 없는 ext4 영속 디렉터리를 사용한다. tmpfs는 시험에서만 쓰며 reboot 영속성을 주장하지 않는다. scope/instance는 1–128자의 영문숫자·하이픈·밑줄이며 계좌 원문이나 credential을 넣지 않는다. 모든 instance 교체는 새 운영 epoch이므로 기존 hint/gap을 숨기는 자동 복구로 사용하지 않는다.

core Provision 이후 임시 Budget을 열어 ProvisionObservation을 한 번 호출한다. 두 API의 O_EXCL·fsync 오류를 그대로 실패 처리하고 부분 결과를 자동 삭제하거나 기존 파일 위에 덮지 않는다. root/queue.lock/wire.lock의 실제 device/inode와 반환된 event.lock ID를 배포자가 별도로 확인한다. 이 값은 runtime가 임의 root를 열고 스스로 채우는 신뢰 입력이 아니다.

v2 manifest는 budget root 외부의 소유자 전용 일반 파일 0600, nlink=1로 둔다. `version`, `root`, `scope`, `instance`, `identity.root/queue/wire`, `event_lock`을 정확하게 기록한다. 비밀키·토큰·계좌 원문은 없다. 원자적 게시와 파일/부모 fsync가 성공한 뒤에만 사용하며, 이전 manifest를 조용히 업그레이드하지 않는다. v1은 required에서 거절된다.

단일 v2 manifest에 root 절대 경로가 하나이므로 모든 참여 namespace에서 그 동일 문자열 경로로 같은 root/lock 객체를 열도록 mount를 구성한다. namespace별 별도 manifest나 자동 경로 변환은 이 절차에서 허용하지 않는다. observed device/inode, effective UID, BOOTTIME/time namespace가 지원 계약에 맞는지 사람이 확인한다. 복사 디렉터리·별도 volume·다른 event.lock을 동일 group으로 승인하지 않는다. 경로 문자열만 같다는 것으로 충분하지 않다. 아래 복구 절의 구 root/manifest/바이너리 재기동 fencing은 v1→v2 첫 전환에도 동일하게 적용한다.

## 실행 전 확인과 전환

네트워크 없는 검증 helper는 manifest v2 parser와 required Factory.Open/Close만 실행한다. client 생성·Accounts·token exchange·브로커 probe를 포함하지 않는다. event state/dirty/root/lock 검사가 실패하면 wire0 상태로 중단한다. 별도의 두 local process·local TLS fixture로 같은 group의 wire 간격과 event 공유를 검증하지만 실제 namespace 증거를 대체하지 않는다.

launch 입력은 `TOSSOS_ACCOUNT_BUDGET_MODE=required`와 명시 manifest 경로 두 값이다. 둘 모두 없는 상태는 기존 OFF이며 required 운영 보장이라고 부르지 않는다. required 누락·빈 값·unknown mode를 OFF로 내리지 않는다. 설정 적용은 별도 승인 후 운영자가 수행하고 각 참여자의 실제 launch 입력을 다시 확인한다. a087 admission은 계속 nil, a169 측정은 계속 보류다.

## dirty·pending·gap·hint 오류

dirty/무결성 오류는 기록 상태를 모르는 장애다. marker 삭제, event.state 손편집, lock 교체, root 복사, instance 변경, chmod 완화로 정상처럼 보이게 하지 않는다. 증거는 접근 제한을 유지해 보존하고 실행 중인 다른 참여자와 보낸 요청 여부를 먼저 확인한다. cleanup callback이 오류를 냈는데 HTTP 응답은 수용인 경우 원장 결과를 NOT_SENT로 바꾸지 않는다.

pending Busy는 wire lock이 바쁘다는 뜻이며 원래 process 생존 증명은 아니다. orphan은 다음 Begin이 gap으로 남기며 poll slot을 poison할 수 있다. poison의 자동 만료·자동 삭제 기능은 없다. event/gap cursor overrun이나 instance 변경을 조용히 따라가지 말고 소비자 측 fail-closed와 장애 보고를 유지한다. 단순 재시작으로 hint를 초기화하지 않는다. 같은 boot anchor는 재사용하고 반복 reboot의 wall-time 상한은 주장하지 않는다.

복구 시에는 모든 writer 종료와 원장·브로커 사실의 별도 확인을 근거로 사람이 처분한다. 현재 API는 손상 state 자동 수리나 poison 해제를 제공하지 않는다. 위 drain·자동 재시작 차단·30초 정지 확인 뒤에만 새 empty root/instance로 전환할 수 있다. 구 root를 증거 보존 위치로 옮겨 구 manifest의 경로/identity로 열 수 없도록 fencing하고 원래 경로나 alias로 되살리지 않는다. 단일 v2 manifest를 배포한 뒤 구 바이너리/구 manifest/구 root 재기동의 거절을 확인하고 새 참여자를 시작한다. parser 하나가 비협조 구 바이너리나 관리자를 차단한다고 주장하지 않는다. 모든 참여자 확인이 불가능하면 복구·측정 재개를 보류한다. 사고 범위·손실한 관측·새 instance의 연속성 단절을 기록하며 검증되지 않은 복구 명령을 여기서 만들어 실행하지 않는다.

## rollback

required 오류를 이유로 프로세스 하나만 OFF나 raw client로 내리는 것은 rollback이 아니라 총량 계약 이탈이다. 배포 rollback은 전체 참여자와 v2 manifest 호환성을 함께 검토하고 사람이 승인한다. v2를 읽지 못하는 이전 바이너리를 실행하기 위해 manifest version을 수동으로 바꾸지 않는다. 기존 원장과 관측 파일을 보존하고, 운영 복구 동안 보호 동작의 제약을 분리 보고한다.

## 남은 검증

독립 runbook 검토, 운영용 installer/preflight의 필요 최소 구현 또는 검증된 수동 대안, 임시 디렉터리 fixture 및 fault 시험, 실제 사람 namespace/참여자 확인이 남아 있다. impeccable: N/A — 운영 문서/backend-only. 검색한 저장소에는 runtime factory·provision leaf API와 fixture가 있으며 별도 운영 installer는 찾지 못했다. 새 의존성을 설치하거나 운영 값을 읽지 않았다.
