# a060 · Tasks

## 1. 근거 고정 (편집 전)

- [x] 1.1 `restartSoak`·`pgrepSoak`·`spawnDetachedSoak`과 추출 대상
      `enginePIDsForJournal`의 Function Logic Map과 Branch Test Map을 **편집 전에**
      작성한다.
- [x] 1.2 프로덕션에서 결함을 측정하고 기록한다 — `config/soak.log`의 실패 사유,
      `config/capability-soak.jsonl` 부재, 자격증명 위치 (읽기 전용).
- [x] 1.3 설치된 `~/.local/share/tossos/bin/soak-autostart.sh`의 패턴과 spawn 형태를
      읽고 D4의 전제를 확인한다.

## 2. RED

- [x] 2.1 프로필 상속: 콘솔이 `--config-dir`/`--session-file`을 갖고 있으면 spawn된
      argv가 그것을 담는다. 현재 코드에서 **실패**한다.
- [x] 2.2 패턴↔argv 계약: `soakArgs`가 만드는 명령줄에 `soakProcessPattern`이 맞는지
      검사한다. 프로필 상속을 넣는 순간 **실패**한다(넣기 전에는 통과 — 그 사실 자체가
      기존 drift 테스트의 무력함을 보인다).
- [x] 2.3 다른 하위 명령 배제: `engine run`·`console`·`httpapi` 명령줄에 맞지 않는다.
- [x] 2.4 소유 판정: 두 프로필의 서베이가 열거될 때 이 콘솔의 기록 경로에 해당하는
      pid만 남는다. 현재 코드에서 **실패**한다.
- [x] 2.5 기본 경로 동치: `--config-dir`를 명시한 쪽과 생략한 쪽이 같은 기록을 가리키면
      같은 서베이로 판정된다.
- [x] 2.6 열거 실패 보존: pgrep 오류는 여전히 부재가 아니다.

## 3. GREEN

- [x] 3.1 `soakArgs(root)`를 `engineArgs`와 같은 모양으로 만든다 (D1).
- [x] 3.2 `spawnDetachedSoak`이 그 argv로 spawn한다.
- [x] 3.3 `soakProcessPattern`을 `tossctl( .*)? soak run`으로 바꾼다.
- [x] 3.4 `enginePIDsForJournal`을 `pidsOwnedBy`로 추출해 두 프로세스가 공유한다 (D3).
      엔진 쪽 동작은 바뀌지 않는다 — a059 테스트 전부 green 유지.
- [x] 3.5 `pgrepSoak`이 기록 경로로 소유를 판정한다 (D2).
- [x] 3.6 `restartSoak`과 `console.go` 호출부가 root를 넘긴다.
- [x] 3.7 2.1~2.6 전부 GREEN.

## 4. REFACTOR

- [x] 4.1 기존 drift 테스트를 실재하는 계약으로 교체한다 (D5). 왜 리터럴 비교를
      버렸는지 테스트 주석에 남긴다.
- [x] 4.2 `soakproc.go`와 `engineproc.go`가 공유 헬퍼를 같은 말로 설명하게 한다.

## 5. VERIFY

- [x] 5.1 변이 검증: 프로필 상속을 되돌리면 2.1이 RED가 되는지 확인하고 되돌린다.
- [x] 5.2 변이 검증: 패턴을 `tossctl soak run`으로 되돌리면 2.2가 RED가 되는지 확인하고
      되돌린다.
- [x] 5.3 변이 검증: 소유 판정을 약화하면 2.4가 RED가 되는지 확인하고 되돌린다.
- [x] 5.4 a059 회귀: 엔진 쪽 테스트 전부 green (공유 헬퍼 추출이 동작을 바꾸지 않았음).
- [x] 5.5 `make test`, `make vet`, `make validate`, `make sdd-sync`, `make sdd-check`.
- [x] 5.6 `make gate CHANGE=a060-soak-restart-uses-this-profile`.
- [x] 5.7 컨테이너 실측 (2026-08-03 09:24~09:27 KST, **KRX 장중**, 사람 승인
      "배포 하세요"). 배포는 12초(00:24:38→00:24:50 UTC)만에 끝났고 엔진은 그 안에
      돌아왔다. journal.db md5 `d1102d49…` 배포 전후 동일.
      ① `config/capability-soak.jsonl`이 **생겼다** — 배포 전에는 없던 파일이다.
      ② `soak.log`의 "no Open API credentials"가 **멎었다**. 대신 서베이가 실제로 돌며
      API를 읽는다.
      ③ 두 번째 누름이 도는 서베이(pid 107)를 찾아 SIGINT하고 pid 146으로 재기동했다 —
      `ps | grep -c soak run` = 1, 중복되지 않았다.
      측정 후 서베이는 SIGINT로 정지시켜 측정 전 상태로 되돌렸다(`issues.md` I6).
      2 사이클이 기록에 남았고 콘솔이 그것을 읽는다 — "아직 기록이 없다"가 사라졌다.

## 6. 리뷰와 기록

- [x] 6.1 독립 리뷰를 받고 `review.md`에 기록한다.
- [x] 6.2 발견 사항을 `issues.md`에 남긴다.
- [x] 6.3 PM story/tracker 동기화.
- [x] 6.4 운영자 승인 후속 처분: 실행 중인 프로세스와 systemd·cron·shell rc·
      `console-launch.sh`의 실행/등록 참조가 없음을 확인하고 만료된 저장소 밖
      `~/.local/share/tossos/bin/soak-autostart.sh`를 제거한다. 런타임 코드는 바꾸지 않는다.
- [x] 6.5 외부 스크립트 은퇴 뒤 계약을 문서·주석·PM acceptance에 일치시키고, 발견된
      attestation 갱신 프로필 불일치는 `STORY-TOS-a063` /
      `a063-align-attestation-renewal-profile`로 분리해 추적한다.
- [x] 6.6 a060 런타임 구현이 이미 포함된 최신 `main`(a061/a062 통합 후)을 Function
      Logic Map 비교 기준으로 재고정한다. 최종 후속은 문서와 주석만 바꾸므로 무관한 후속
      Go 함수 변경을 a060 누락으로 계산하지 않는다.
