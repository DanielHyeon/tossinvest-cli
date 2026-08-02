# a059 · Tasks

## 1. 근거 고정 (편집 전)

- [x] 1.1 `pgrepEngine`·`startEngine`·`stopEngine`의 Function Logic Map과 Branch Test Map을
      **편집 전에** 작성한다. 엔진 기동·종료 경로이므로 면제하지 않는다.
- [x] 1.2 프로덕션 컨테이너에서 현재 패턴과 후보 패턴의 실제 매칭을 측정하고 기록한다
      (읽기 전용 `pgrep`).
- [x] 1.3 soak 경로가 같은 결함을 갖는지 확인한다 — `spawnDetachedSoak`의 argv를 읽고
      결론을 proposal에 남긴다.

## 2. RED

- [x] 2.1 패턴↔argv 계약: `engineArgs`가 만드는 명령줄에 `engineProcessPattern`이 맞는지
      검사한다. 현재 코드에서 **실패**하는 것을 관측한다.
- [x] 2.2 다른 하위 명령 배제: `console`·`httpapi`·`soak` 명령줄에 맞지 않는다. 현재
      코드에서 통과해야 한다 (회귀 방지용 고정).
- [x] 2.3 소유 판정: 두 프로필의 엔진이 열거될 때 이 콘솔의 journal 디렉터리에 해당하는
      pid만 남는다. 현재 코드에서 **실패**한다(그런 판정이 없다).
- [x] 2.4 기본 경로 동치: `--config-dir`를 명시한 쪽과 생략한 쪽이 같은 journal
      디렉터리를 가리키면 같은 인스턴스로 판정된다.
- [x] 2.5 소유 미증명 배제: 명령줄이 다른 `--config-dir`를 갖거나 판정 근거가 없으면
      시그널 대상에서 빠진다.
- [x] 2.6 열거 실패 보존: pgrep 오류는 여전히 부재가 아니다(a056 D3 회귀 고정).

## 3. GREEN

- [x] 3.1 `engineProcessPattern`을 `tossctl( .*)? engine run`으로 바꾼다 (D1).
- [x] 3.2 `pgrepEngine`을 `pgrep -a -f`로 바꾸고 pid와 명령줄을 함께 읽는다.
- [x] 3.3 소유 판정을 순수 함수로 분리하고 journal 디렉터리로 거른다 (D2, D3).
- [x] 3.4 `startEngine`·`stopEngine`이 소유한 엔진만 보게 호출부를 맞춘다.
- [x] 3.5 `tools/engine-autostart.sh`의 `ENGINE_PATTERN`과 drift 테스트를 함께 옮긴다 (D4).
- [x] 3.6 2.1~2.6 전부 GREEN.

## 4. REFACTOR

- [x] 4.1 `stopEngine`이 journal 디렉터리를 두 번 구하지 않게 정리한다.
- [x] 4.2 왜 후보 선별과 소유 판정이 나뉘어 있는지 — 그리고 왜 스크립트는 앞의 절반만
      하는지 — 를 코드와 스크립트 주석이 같은 말로 적게 한다.

## 5. VERIFY

- [x] 5.1 변이 검증: 패턴을 원래 값으로 되돌리면 2.1이 RED가 되는지 확인하고 되돌린다.
- [x] 5.2 변이 검증: 소유 판정을 건너뛰면 2.3/2.5가 RED가 되는지 확인하고 되돌린다.
- [x] 5.3 변이 검증: 패턴에서 토큰 경계를 빼면 2.2가 RED가 되는지 확인하고 되돌린다.
- [x] 5.4 컨테이너 실측 (2026-08-03 07:45~07:46 KST, KR·US 휴장, 사람 승인). 콘솔
      `/verify-console`의 [엔진 정지] → 도는 엔진 pid 16을 실제로 찾아 SIGTERM을 보내고
      루프 완주까지 기다렸다. 응답은 `"16를 종료시켰지만 활성 마커가 아직 신선하다
      (22:43:35Z) — 최대 5m0s 뒤 사라진다"`. a059 이전 같은 조건의 응답은
      `"실행 중인 엔진을 찾지 못했다."`였고 엔진은 계속 돌았다. 엔진 로그에
      `"the runtime was cancelled; every loop was drained and the journal can be closed"`.
      이어 [엔진 시작] → 3.0초(= `engineStartProbe`) 뒤 pid 110으로 재기동.
      journal.db md5 `d1102d49…` 시험 전후 동일, `/positions` 5행 전후 동일, 두 서비스
      healthy, 엔진 정지 구간 32초(대부분 명령 간 지연).
- [x] 5.5 컨테이너 실측 (2026-08-03 07:46 KST). 엔진이 pid 110으로 도는 중 [엔진 시작]을
      다시 눌렀다 → 4.5ms 만에 거부, 응답은 `"엔진이 기동을 거부했다: 엔진이 이미 실행
      중이다 (pid 110, 마지막 갱신 2026-08-02T22:45:54Z)"`. 엔진은 여전히 하나뿐이다.
      **a056의 거부 분기가 컨테이너에서 처음으로 도달했다** — a056 `issues.md` I2가
      "도달 불가"로 기록한 그 분기다. a059 이전에는 두 번째 프로세스가 spawn되어 flock에서
      죽었고, 운영자는 pid도 갱신 시각도 보지 못했다.
- [x] 5.6 `make test`, `make vet`, `make validate`, `make sdd-sync`, `make sdd-check`.
- [x] 5.7 `make gate CHANGE=a059-console-finds-the-engine-it-owns`.

## 6. 리뷰와 기록

- [x] 6.1 독립 리뷰를 받고 `review.md`에 기록한다.
- [x] 6.2 발견 사항을 `issues.md`에 남긴다.
- [x] 6.3 PM story/tracker 동기화.
