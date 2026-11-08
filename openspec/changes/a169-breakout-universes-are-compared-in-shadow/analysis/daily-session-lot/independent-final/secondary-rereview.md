# a169 fixture 보조 재검토 — 8e8b6db24

대상 HEAD `8e8b6db24abc6c003a6ff913d818f2a52d92c0d0`. `session_linux.go` SHA-256 `ce79892b33822c1f670eab66bf9d2bffc239c9e972b43a64cc9e1491eed8d29a`; `session_contract_linux_test.go` `7068e43619e8ceff3c5d206ad9df663a4ec9aa1268b8083df5bca965572da429`; `session_linux_test.go` `d6836a3a114499cc2f920cbfb7bf97b3533af59392bf4094010fe48c7dac47a8`. source/fixture 손읽기, 시험 미실행.

- 이전 flock false green 해결: `session_linux_test.go:246-273`은 child kill/wait 뒤 실제 lock 파일을 열어 nonblocking flock 획득·해제하고 별도로 state=active 확인 및 `ErrInvalid`/`ErrBusy` 배제를 요구한다. lock이 남는 반례는 실패한다.
- 동시 최초 시작: `session_contract_linux_test.go:25-33,88-114`는 두 child가 lock-open 뒤 barrier에 모인 후 함께 진행하고 winner 한 명/busy 한 명 및 소유 중 제3자 busy를 확인한다. winner가 release 신호 전 lock 유지하므로 단순 순차 재시작으로 혼동하지 않는다.
- Begin/terminal 오류: `:178-217,224-263`은 before/after 훅에서 nil session·오류 및 관찰된 state별 재시작 결과를 확인한다. Pause의 write/sync 후 결과가 paused면 재개, terminal 전 write 오류의 active는 거절한다. 다만 `session_linux.go:61-70,321-360`의 before/after는 syscall 직전 생략 또는 성공 직후 가상 오류다. 실제 fsync/unlock syscall 실패, 기기 전원 손실 또는 crash durability 시험은 아니다.
- root/lock/state 교체: `session_contract_linux_test.go:269-301`은 active 세션 중 각각 실제 rename/대체하고 Pause와 재시작 거절을 검사한다. root dev/inode, 파일명과 열린 fd 불일치 경로에 대해 이전의 기대 inode 숫자만 틀린 fixture보다 직접적이다. 운영자/비협조 process가 일시 교체 후 복구하거나 root를 rollback하는 범위는 입증하지 않는다.
- partial EOF: `:306-318`은 canonical active→paused 뒤 active 프레임의 첫 글자, 중간, 완전 JSON에서 newline만 제거한 세 절단을 각각 invalid/재시작 거절로 확인한다. 모든 절단 위치를 열거하지는 않지만 요청된 이전 반례들은 포착한다.
- ext4: `session_linux_test.go:23-39`은 명시 환경변수 경로 검증을 추가하고, `session_contract_linux_test.go:322-338`은 non-ext4를 별도 거절한다. 기본 temp가 non-ext4면 fixture 전체가 `Skip`이므로 지정 ext4 root 없이 전체 테스트 명령이 성공해도 해당 fixture 실행을 증명하지 못한다. 실행자가 skip 수와 지정 ext4 root 사용을 확인해야 한다.
- 남은 좁은 범위: `session_linux_test.go:210-275`는 active 뒤 kill만 다룬다. `session_contract_linux_test.go:224-263`의 paused postcommit 재진입은 같은 프로세스 내 확인이다. **Pause 커밋 후 실제 child kill/restart**가 요구라면 해당 별도 fixture는 보이지 않는다. 단 같은 process에서의 재진입 검증을 power-loss 보장으로 확대해서는 안 된다.

파일 편집·테스트·커밋 없음. 해당 범위에 대한 정적 확인이며 전체 리뷰 또는 PASS가 아니다.
