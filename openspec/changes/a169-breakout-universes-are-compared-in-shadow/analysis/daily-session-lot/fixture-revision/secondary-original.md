# a169 shadowsession fixture 보조 검토 — 잠정

대상 checkout `/home/daniel/tossos-work/a169-daily-session`, 검토 시 HEAD `ec254d8ca3b8b40220910f73e73d556e1043142b`. `session_linux.go` SHA-256 `ce79892b33822c1f670eab66bf9d2bffc239c9e972b43a64cc9e1491eed8d29a`; `session_linux_test.go` SHA-256 `85ddd35359cac1c24b6d69dc45dd375cd39f48bf27fe3107f155f5775b84081e`. 손읽기만 수행. 시험 미실행, commit SHA 통지 전 최종 PASS 없음.

1. **P2 — kill 뒤 flock 해제와 active 재시작 거절이 분리 검증되지 않음.** `internal/shadowsession/session_linux_test.go:227-235`는 child kill 전 `ErrBusy`, kill 후에는 `err != nil`만 검사한다. 반례: kill 뒤에도 락이 계속 busy인 잘못된 구현이면 마지막 Begin이 `ErrBusy`라 시험이 통과한다. child 종료 뒤 별도 파일 디스크립터로 nonblocking flock을 성공적으로 잡았다 놓은 다음 `Begin`이 stale active로 거절하는지 확인해야 두 조건을 구분한다. child 죽이기 전 최초 active sync도 재확인할 수 있다.

2. **P2 — Begin fault fixture는 재시작 상태·postcommit 불명 결과를 확인하지 않음.** `session_linux_test.go:161-176`은 before injection 시 nil session/error만 확인한다. `session_linux.go:223-240,309-316`에서 full write 후 sync 전 실패면 active가 이미 읽힐 수 있는데 재시도 결과를 검사하지 않는다. `after`는 취소 한 건(:178-190)만 있으며 Begin의 after `write`, `state-sync`, `state-dir-sync` 오류에서 반환 허가 0 및 재시작 판정이 없다. 반례: Begin이 오류에 nil 반환하지만 후속 호출이 이미 써진 active를 잘못 paused로 처리해 성공해도 이 표는 통과한다. before/after 각각 활성 프레임의 실제 존재·오류 반환·재시도 결과를 분리해야 한다. 반대로 최초 쓰기 전 state 빈 파일이 남으면 이후 거절되는 현재 동작도 확인해야 한다.

3. **P2 — pause postcommit 시뮬레이션은 오류 반환이지만 실 동기화/프로세스 생존 전제만 검증.** `session_linux_test.go:123-142`의 hook은 `session_linux.go:61-70`에 따라 실제 syscall이 성공한 뒤 오류를 반환한다. 특히 `state-sync` after 오류에서는 `state-dir-sync`가 건너뛰어지지만 파일이 같은 부팅에서 보이므로 재개할 수 있다. 반례: 이 시험을 파일·디렉터리 sync 오류 후 crash-safe durable pause가 검증됐다고 해석하면, dir sync 미수행 후 crash에서 엔트리 보존은 시험된 바 없다. `before` 실패나 실제 kernel sync 오류 시뮬레이션과 프로세스 kill 전후를 별도로 구분해야 한다. kill은 power loss 증명도 아니다. cleanup after `unlock`은 실제 unlock 완료 후 가짜 오류이며 실제 unlock syscall 실패를 주입하지 않는다 (`session_linux.go:337-350`).

4. **P2 — 파일·root 치환 검사는 일부 fixture만 있음.** `session_linux_test.go:107-110`은 기대 inode 숫자만 틀리게 만들고 실제 root 교체가 아니다. `:287-299`는 열린 state 파일만 교체해 Pause가 오류인지 확인한다. 반례: 동일 경로 root 디렉터리를 rename하고 새 root로 대체하거나, 보유 중 lock 파일을 다른 inode로 교체했을 때의 검증이 회귀해도 해당 fixture는 통과한다. `session_linux.go:145-156`의 root 재개방과 lock/state 이름-열린 fd의 identity 검사를 각각 분리 검증하고 Begin 도중·terminal 전 root/lock/state 교체 및 교체 후 재진입을 체크해야 한다. root rollback·사전 root identity 신뢰는 fixture 하나로 증명할 수 없는 전제다.

5. **P2 — 유효한 paused prefix 뒤 partial EOF에서 재개 거절 검증 없음.** `session_linux_test.go:94-103,151-157`은 유효 prefix 없는 partial을 만들고 `:239-264`는 깨진 checksum/중복 필드/잘못된 전이를 확인한다. 반례: parser가 `active\npaused\n` 뒤 잘린 다음 active 프레임을 조용히 버리고 paused로 처리하면 실제 새 Begin을 허용해도 현재 partial fixture는 통과한다. 실제 `marshal`로 유효 active→paused prefix 뒤 다음 active frame의 각 위치를 절단해 newline 누락/부분 JSON/완전 JSON의 newline 누락 모두 거절하는 검증이 필요하다 (`session_linux.go:171-198`). valid prefix 전체 프레임 삭제·rollback은 parser가 알 수 없다는 한계도 별개다.

6. **P3 — fixture root가 개발자 고정 경로이며 ext4 조건을 명시적으로 주장하지 않음.** `session_linux_test.go:21-32`는 `/home/daniel/tossos-work`에 생성한다. 다른 Linux checkout/CI에서 해당 경로가 없거나 non-ext4면 시험 자체가 실패한다 (`session_linux.go:82-105`). 반례: ext4 test root 지원이 없는 CI에서는 정상/오류 fixture를 선택적으로 실행할 수 없다. 외부에서 제공하는 전용 ext4 테스트 root를 명시적으로 검증하고 없을 때 skip 또는 명확한 실패 정책을 계약해야 하며, tempfs/non-ext4 거절도 독립 시험해야 한다. 현 fixture 성공은 실제 power-loss 내구성을 증명하지 않는다.

이 보고는 위 범위만 다룬다. 전체 정적 리뷰·시험·소스 수정은 수행하지 않았다.
