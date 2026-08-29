# a056 · Tasks

## 1. 근거 고정 (편집 전)

- [x] 1.1 `cmd/tossctl/engineproc.go:startEngine`의 Function Logic Map과 Branch Test Map을
      **편집 전에** 작성한다. 엔진 기동 경로이므로 면제하지 않는다 (design.md 말미).
- [x] 1.2 `enginelock.Read`/`StaleAfter`와 `engineFindProcesses`의 현재 계약을 AST로
      확인하고, 두 검사의 실행 순서를 Branch Test Map에 기록한다.
- [x] 1.3 flock 획득이 기동의 첫 동작이라는 spec 문장이 현재 런타임에서도 참인지
      `engine run` 경로에서 확인한다 (D2가 여기에 기댄다).

## 2. RED

- [x] 2.1 유령 마커 테스트: 마커는 stale 창 안, PID는 존재하지 않음, 프로세스 열거는 빈
      결과 → 기동이 진행되어야 한다. 현재 코드에서 **실패**하는 것을 관측한다.
- [x] 2.2 실제 중복 테스트: 프로세스 열거가 PID를 반환 → 거부되고 안내에 그 PID가 있다.
      현재 코드에서 통과해야 한다 (회귀 방지용 고정).
- [x] 2.3 열거 실패 테스트: `engineFindProcesses`가 error, 마커 신선 → 거부 유지 (D3).
- [x] 2.4 거부 근거 테스트: 마커만으로 거부하는 경로가 소스에 존재하지 않음을 고정한다.

## 3. GREEN

- [x] 3.1 `startEngine`의 사전 확인을 결합한다: 마커 신선 **AND** 프로세스 관측일 때만
      거부. 마커는 안내 문구의 재료로 유지한다 (D1).
- [x] 3.2 열거 오류 시 기존 거부 동작을 유지한다 (D3).
- [x] 3.3 2.1~2.4 전부 GREEN.

## 4. REFACTOR

- [x] 4.1 결합된 판정을 이름 있는 함수로 분리하고, 왜 마커가 거부 근거가 아닌지를 주석에
      남긴다 — 다음 사람이 순서를 되돌리지 않도록.
- [x] 4.2 `engine-run.json`의 "advisory only" 문구와 코드 주석이 같은 말을 하는지 맞춘다.

## 5. VERIFY

- [x] 5.1 변이 검증: 결합 조건을 마커 단독으로 되돌리면 2.1이 RED가 되는지 확인하고 되돌린다.
- [x] 5.2 변이 검증: 열거 결과를 무시하면 2.2가 RED가 되는지 확인하고 되돌린다.
- [x] 5.3 컨테이너 실측 (2026-08-03 02:5x KST, KR·US 휴장, 사람 승인). 엔진이 돌고 마커가
      33초로 신선한 상태에서 `docker compose up -d --force-recreate` → 유령 마커(죽은 pid 16)
      조건. 엔진이 **1초 안에** 떴고 로그는 `엔진 자동 시작: 엔진을 시작했다`. a056 이전의
      같은 조건은 같은 밤에 두 번 관측됐다 — 01:17 건은 8분, 02:44 건은 회복되지 않아
      배포 시점까지 11분간 정지 상태였다. 두 서비스 healthy, journal 무손상(OPEN 5, 활성
      exit 정책 4).
- [x] 5.4 실제 중복 실측 — **가설이 틀렸음을 확인했다.** 컨테이너 안에서 도는 엔진은
      `engineProcessPattern`(`"tossctl engine run"`)에 매칭되지 않는다. cmdline이
      `/usr/local/bin/tossctl --config-dir … --session-file … engine run`이라 그 연속
      문자열이 없기 때문이다(`pgrep -f 'tossctl engine run'` → exit 1, `pgrep -f 'engine run'`
      → 16). 따라서 컨테이너 모드에서 a056의 거부 분기는 도달 불가이고 배타는 flock 단독이다.
      `observed=true` 분기 자체는 변이 검증된 단위 테스트(5.1, 5.2)가 덮는다. 파급과
      `stopEngine`에 미치는 더 큰 영향은 issues.md I2에 기록했다 — 별도 change 대상이다.
- [x] 5.5 `make test`, `make vet`, `make validate`, `make sdd-sync`, `make sdd-check`.
- [x] 5.6 `make gate CHANGE=a056-autostart-survives-container-recreate`.

## 6. 리뷰와 기록

- [x] 6.1 독립 리뷰(gstack)를 받고 `review.md`에 기록한다.
- [x] 6.2 발견 사항을 `issues.md`에 남긴다.
- [x] 6.3 PM story/tracker 동기화.
