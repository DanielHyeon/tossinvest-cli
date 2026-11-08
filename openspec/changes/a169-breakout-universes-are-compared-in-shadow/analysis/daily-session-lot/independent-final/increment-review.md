# a169 최종 증분 독립 검증

- 후보 `80ca5dbfb46b480a6148d281f6e2878846d95e8c`: source SHA-256 `ce79892b33822c1f670eab66bf9d2bffc239c9e972b43a64cc9e1491eed8d29a`, contract test `1ce4271e4a8fa23471050a7554580532119edd028b68e73db05c6608a3d6ec47`; 시험 전후 일치.
- **증분 PASS, skip 0**. `TestSimultaneousInitialLock`의 absent/paused 각각 두 프로세스 open barrier 후 한 승자·한 busy, 종료 후 active 거절을 확인한다. `TestPauseCommitThenChildKill`은 자식의 Pause 성공·lock 해제 이후 kill, paused 판독과 재개를 확인한다. 이는 Pause 도중 kill이나 전원 손실 시험이 아니다.
- ext4 전용 임시 parent를 명시했고 신규 raw 경로 `/tmp/a169-final-increment-xAZOKcBG/`에 보존했다. `GOPROXY=off go test -c` normal과 `-race -tags tossos_testseams -c`를 빌드한 뒤 두 실행 파일 각각 `-test.v -test.count=1 -test.run='^(TestSimultaneousInitialLock|TestPauseCommitThenChildKill)$'`로 실행했다.

| 파일 | SHA-256 |
| --- | --- |
| normal.test | `5751411bd3b35917ddeb81c2cf9c47db2f432d73d77e0fe1bb399ed1b277e7a3` |
| race.test | `7c220cda42210c032be58ce6019e8bc299fa4260c3b522c97c82cf148ac0d120` |
| normal.log | `57aa3c139e786dc0a057c884baa9eced1c57056d303b291a1a0c00ceec5005dd` |
| race.log | `8ce0ab96fb8fa3709eee5b97b006e3e3081aad69d15df8bbb7058235e9668d46` |

바이너리 실행 전후 해시 동일. 기존 패키지 전체 선택 PASS는 이전 candidate `8e8b6db24`의 `/tmp/a169-session-rereview.md`에 기록돼 있으며, 이번 candidate는 변경 없는 production source에 대한 두 신규 시나리오만 추가 검증했다. 이번 실행을 현재 candidate의 전체 패키지 재시험으로 보고하지 않는다. 실제 요청 drain, 운영 저장장치 내구성, 런타임 결속·실측, 전체 테스트/gate는 미검증이다. 기존 로그·source/test·운영은 변경하지 않았다.
