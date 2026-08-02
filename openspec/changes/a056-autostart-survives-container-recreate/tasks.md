# a056 · Tasks

## 1. 근거 고정 (편집 전)

- [ ] 1.1 `cmd/tossctl/engineproc.go:startEngine`의 Function Logic Map과 Branch Test Map을
      **편집 전에** 작성한다. 엔진 기동 경로이므로 면제하지 않는다 (design.md 말미).
- [ ] 1.2 `enginelock.Read`/`StaleAfter`와 `engineFindProcesses`의 현재 계약을 AST로
      확인하고, 두 검사의 실행 순서를 Branch Test Map에 기록한다.
- [ ] 1.3 flock 획득이 기동의 첫 동작이라는 spec 문장이 현재 런타임에서도 참인지
      `engine run` 경로에서 확인한다 (D2가 여기에 기댄다).

## 2. RED

- [ ] 2.1 유령 마커 테스트: 마커는 stale 창 안, PID는 존재하지 않음, 프로세스 열거는 빈
      결과 → 기동이 진행되어야 한다. 현재 코드에서 **실패**하는 것을 관측한다.
- [ ] 2.2 실제 중복 테스트: 프로세스 열거가 PID를 반환 → 거부되고 안내에 그 PID가 있다.
      현재 코드에서 통과해야 한다 (회귀 방지용 고정).
- [ ] 2.3 열거 실패 테스트: `engineFindProcesses`가 error, 마커 신선 → 거부 유지 (D3).
- [ ] 2.4 거부 근거 테스트: 마커만으로 거부하는 경로가 소스에 존재하지 않음을 고정한다.

## 3. GREEN

- [ ] 3.1 `startEngine`의 사전 확인을 결합한다: 마커 신선 **AND** 프로세스 관측일 때만
      거부. 마커는 안내 문구의 재료로 유지한다 (D1).
- [ ] 3.2 열거 오류 시 기존 거부 동작을 유지한다 (D3).
- [ ] 3.3 2.1~2.4 전부 GREEN.

## 4. REFACTOR

- [ ] 4.1 결합된 판정을 이름 있는 함수로 분리하고, 왜 마커가 거부 근거가 아닌지를 주석에
      남긴다 — 다음 사람이 순서를 되돌리지 않도록.
- [ ] 4.2 `engine-run.json`의 "advisory only" 문구와 코드 주석이 같은 말을 하는지 맞춘다.

## 5. VERIFY

- [ ] 5.1 변이 검증: 결합 조건을 마커 단독으로 되돌리면 2.1이 RED가 되는지 확인하고 되돌린다.
- [ ] 5.2 변이 검증: 열거 결과를 무시하면 2.2가 RED가 되는지 확인하고 되돌린다.
- [ ] 5.3 컨테이너 실측: `docker compose up -d`로 재생성한 직후 엔진이 8분 기다리지 않고
      뜨는지 확인한다. **휴장 시간에만 수행하고 사람이 승인한다.**
- [ ] 5.4 실제 중복 실측: 엔진이 실행 중인 상태에서 기동을 시도하면 거부되는지 확인한다.
- [ ] 5.5 `make test`, `make vet`, `make validate`, `make sdd-sync`, `make sdd-check`.
- [ ] 5.6 `make gate CHANGE=a056-autostart-survives-container-recreate`.

## 6. 리뷰와 기록

- [ ] 6.1 독립 리뷰(gstack)를 받고 `review.md`에 기록한다.
- [ ] 6.2 발견 사항을 `issues.md`에 남긴다.
- [ ] 6.3 PM story/tracker 동기화.
