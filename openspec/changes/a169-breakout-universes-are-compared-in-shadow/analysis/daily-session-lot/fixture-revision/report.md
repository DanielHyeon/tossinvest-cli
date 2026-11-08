# 독립 리뷰 후 fixture 보강

- 독립 P1 원문 `independent-original.md`(SHA-256 `7f248056cb562e3d7b6961437734597f50ee6b2588aa77f578792ba228f27927`), 보조 검토 원문 `secondary-original.md`(`cde267cd0b603ba4379f5145bb19da260571bc54545134be2fe1d62466de7655`)을 변경 없이 보존했다. 원 판정 FAIL은 새 독립 재검토로만 해소된다.
- 테스트 커밋 `8e8b6db24`, production source `ce79892b33822c1f670eab66bf9d2bffc239c9e972b43a64cc9e1491eed8d29a`는 `ec254d8ca`와 동일. 테스트 SHA: `session_linux_test.go d6836a3a114499cc2f920cbfb7bf97b3533af59392bf4094010fe48c7dac47a8`, `session_contract_linux_test.go 7068e43619e8ceff3c5d206ad9df663a4ec9aa1268b8083df5bca965572da429`.
- fixture는 명시된 `A169_SHADOWSESSION_TEST_ROOT`의 ext4 parent를 사용한다. 설정 오류는 실패, 미설정 기본 temp가 ext4가 아니면 skip; 시스템 TMPDIR은 변경하지 않는다. 지정 `/dev/shm`은 예상대로 실패, 미지정 기본 경로는 이 환경에서 PASS. `/dev/shm` fixture 자체를 사용한 non-ext4 거절 시험도 PASS.
- 두 child가 lock-open 뒤 pipe barrier에서 출발: 한 child만 성공하고 다른 child와 세 번째 Begin은 Busy, 소유자 종료 후 직접 flock은 성공하지만 active 기록은 ErrInvalid로 거절. 해당 시험 10회 PASS. 기존 kill 시험도 lock 재획득과 active 검증을 구분한다.
- 초기·paused 시작에 대해 named operation before/after 오류 후 nil handle과 현재 파일 상태 및 재시작 판정을 확인했다. Pause/Abort/Complete 각각 write·state sync·dir sync·unlock·close before/after 오류도 동일 방식으로 확인한다. 실제 동작 후 오류를 주입한 경우는 실제 syscall failure나 crash-safe durability 증명이 아니다. root/lock/state 교체와 유효 paused prefix 뒤 부분 EOF도 거절 확인했다.
- `A169_SHADOWSESSION_TEST_ROOT=/home/daniel/tossos-work go test -json -count=1 ./internal/shadowsession`과 동일 환경의 `go test -race -json -count=1 -tags tossos_testseams ./internal/shadowsession` 각각 116 run, skip 0, PASS. 원문 `normal-final.jsonl`, `race-final.jsonl`. 앞선 정상 실행 `normal.jsonl`도 보존한다. 전체 테스트/gate·실측·운영 접근은 미실행이다.
- 분기 목록의 `related_tests`는 함수 단위 구조 연관이며 각 분기의 실행 coverage를 뜻하지 않는다. `test:null`은 해당 분기별 실행 여부 미확인으로 유지한다.
