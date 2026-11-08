# a169 shadowsession 독립 구현 리뷰

- 대상: commit `ec254d8ca3b8b40220910f73e73d556e1043142b`, 계약 `1a53b8917`. 검토 전후 SHA-256: Linux source `ce79892b33822c1f670eab66bf9d2bffc239c9e972b43a64cc9e1491eed8d29a`, test `85ddd35359cac1c24b6d69dc45dd375cd39f48bf27fe3107f155f5775b84081e`, unsupported `7f759c52141012b7d48356bdefdb5363a490f222be625012d328be1cbdcc4745`, types `3affc279a5116e5f51b748e295f1d417b301242c90bc6c56d19a2c6aa90ff51f`.
- 판정 **FAIL — P1 1건**. runtime 안전성·운영·전체 a169 완료 판정 아님.

## P1: 테스트 fixture가 작성자 개인 절대 경로에 의존

`internal/shadowsession/session_linux_test.go:23`은 모든 fixture를 `/home/daniel/tossos-work`에서 생성한다. 해당 디렉터리가 없거나 다른 계정이 쓰지 못하는 Linux 개발기·CI에서 `go test ./internal/shadowsession`의 모든 fixture 기반 테스트가 실패한다. 패키지와 프로젝트 검증을 특정 개인 환경에 종속시키는 회귀다. 최소 수정: 테스트 전용 ext4 root를 환경에서 명시 주입하고, 기본 경로는 실행 환경에 적합하게 선택·검증하며 부재를 명확히 처리한다. production root나 시스템 전체 TMPDIR을 변경하지 않는다.

## Coverage gap

`TestChildLockAndCrash`는 자식이 Begin에 성공해 READY를 보낸 뒤 부모가 경쟁하므로, 두 프로세스의 lock 파일 최초 생성 경합은 직접 검증하지 않는다. 설계 §5가 요구한 동시 최초 시작을 동기화된 두 프로세스로 별도 검증할 필요가 있다. 이는 미검증 범위이며 별도 제품 결함으로 계산하지 않는다.

## 검증과 한계

- `/home/daniel/tossos-work`가 ext4임을 확인하고 fixture만 사용. `go test -count=1 ./internal/shadowsession` PASS; `GOPROXY=off go test -race -count=1 -tags tossos_testseams ./internal/shadowsession` PASS. 추가 `integration` tag race도 PASS이나 계획된 tag 증거로 대체하지 않는다.
- 코드·fixture 수동 검토: canonical byte 대조, 상태 전이, context 반환 경계, lock·identity 검증, postcommit 오류, cleanup을 확인했다. production Go source에서 `shadowsession` 참조 검색 결과 없음; 두 지정 commit 사이 기존 Go source 변경 없음. subprocess 결과 및 fault fixture 통과는 전원 손실 내구성이나 HTTP drain 증명이 아니다.
- gstack review의 범위형 수동 적대 점검만 적용했다. 사용자 read-only 제한 때문에 preamble의 state write/update check, 자동 수정·다른 모델 호출·리뷰 로그 기록은 수행하지 않았다. full gstack gate가 아니다. 전체 tests/gate, 운영·실브로커·외부 전송, source/test 수정 미실행.

Recommendation: 수정 후 범위 재검증 because 개인 절대경로가 다른 Linux 환경의 패키지 시험을 실패시킨다.
