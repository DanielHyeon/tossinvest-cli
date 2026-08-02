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
- [ ] 5.4 컨테이너 실측 (휴장 시간, 사람 승인): 정지 버튼이 도는 엔진을 실제로 세우고,
      이어서 기동이 다시 뜨는지 확인한다. journal 정합 close와 포지션 무손상을 확인한다.
- [ ] 5.5 컨테이너 실측: a056의 거부 분기가 이제 도달하는지 — 엔진이 도는 중 기동을
      요청하면 프로세스를 안내하며 거부하는지 확인한다.
- [ ] 5.6 `make test`, `make vet`, `make validate`, `make sdd-sync`, `make sdd-check`.
- [ ] 5.7 `make gate CHANGE=a059-console-finds-the-engine-it-owns`.

## 6. 리뷰와 기록

- [ ] 6.1 독립 리뷰를 받고 `review.md`에 기록한다.
- [ ] 6.2 발견 사항을 `issues.md`에 남긴다.
- [ ] 6.3 PM story/tracker 동기화.
