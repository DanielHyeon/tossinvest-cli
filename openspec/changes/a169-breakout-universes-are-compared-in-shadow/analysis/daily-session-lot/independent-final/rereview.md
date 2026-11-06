# a169 shadowsession 독립 재리뷰

- 대상 candidate `8e8b6db24abc6c003a6ff913d818f2a52d92c0d0`. 시험 전후 Linux source `ce79892b33822c1f670eab66bf9d2bffc239c9e972b43a64cc9e1491eed8d29a`, Linux test `d6836a3a114499cc2f920cbfb7bf97b3533af59392bf4094010fe48c7dac47a8`, contract test `7068e43619e8ceff3c5d206ad9df663a4ec9aa1268b8083df5bca965572da429`, unsupported `7f759c52141012b7d48356bdefdb5363a490f222be625012d328be1cbdcc4745`, types `3affc279a5116e5f51b748e295f1d417b301242c90bc6c56d19a2c6aa90ff51f` 일치. 시험 종료 시 HEAD는 병행 문서 커밋 `756c0042947589b79d708e92f2d4d1f165dd47bb`로 전진했으나 candidate 대비 변경은 evidence 문서뿐이며 패키지 변경 없음.
- **범위 PASS — P0/P1 발견 없음.** 기존 P1 개인 절대경로 제거. `A169_SHADOWSESSION_TEST_ROOT`가 지정되지 않고 기본 임시 경로가 ext4가 아닌 환경에서는 fixture가 명시적으로 skip될 수 있으므로 그 환경의 일반적인 녹색 결과만으로 coverage를 주장할 수 없다.
- 새로운 두 프로세스 barrier는 둘 다 lock open 이후 동시 flock 경쟁, 하나 승리·하나 busy 및 마지막 active 거절을 검증한다. kill 이후 lock 획득과 active 거절 분리, Begin/종결 fault 전후 재시작, identity 교체, paused 뒤 부분 EOF 및 비 ext4 거절 시험을 확인했다.

## 독립 실행 증거

전용 ext4 fixture parent를 별도 생성해 `A169_SHADOWSESSION_TEST_ROOT`로 지정하고 삭제했다. `GOPROXY=off go test -c`와 `GOPROXY=off go test -race -tags tossos_testseams -c`로 해당 패키지의 두 실행 파일을 빌드한 뒤 각 `-test.v -test.count=1` 실행. normal·tagged race 모두 PASS, SKIP 0. raw log 및 실행 파일은 새 경로 `/tmp/a169-rereview-5eP3xeos/`에 보존했다.

| 파일 | SHA-256 |
| --- | --- |
| normal.test | `d8323417e8a083c0ab12f5437f91b883dbfde3f2effd96b58e27f591d9deb101` |
| race.test | `1d88ddd7ade833c1677287d9de4129312ae558d1e477e9a524fac4276312a6a6` |
| normal.log | `6fb7652013f90d9fb2587ce7bebdfdba295605f039c7658b7cc7eadaa7918a5c` |
| race.log | `ff6c1171f0f120b9b01c6a8e05e0cd83bdef92e9f7fc5272a7ce21f5d5462a10` |

실행 전후 바이너리 해시 일치. gstack review는 수동·읽기 범위만 적용했다. preamble의 외부 상태 쓰기·자동 수정·다른 모델 호출·review log는 수행하지 않았으므로 전체 gstack gate가 아니다. 대상 패키지 외 전체 시험·소스 수정·운영·브로커 실행 없음. fixture fault 및 process kill은 실제 전원 손실 증거가 아니며 HTTP drain·실제 날짜와 엔진 신원 결속·실측은 후속 검증이다.

Recommendation: Manager의 범위 검증으로 진행 because 개인 경로 의존과 최초 lock 경합 coverage가 해소됐고 지정 시험이 skip 없이 통과했다.
