# a059 · 콘솔은 자기가 소유한 엔진을 찾는다

- **Feature**: `FEAT-TOS-003` — Engine interlocks and Guardian controls
- **Story**: `STORY-TOS-a059`
- **Spec**: `engine-safety`

## Why

콘솔의 **정지 버튼이 도는 엔진을 찾지 못한다.** 2026-08-03 a056 tasks 5.4 실측에서
발견했고(a056 `issues.md` I2), 프로덕션 컨테이너에서 지금 이 순간 재현된다.

```
docker exec tossos-tossos-1 pgrep -f 'tossctl engine run'   → exit 1 (매칭 없음)
docker exec tossos-tossos-1 pgrep -af 'engine run'          → 16 /usr/local/bin/tossctl … engine run
```

엔진은 pid 16으로 살아 있는데 `stopEngine`은 `"실행 중인 엔진을 찾지 못했다."`를
돌려준다. 운영자가 엔진을 세우려고 눌러도 아무 일도 일어나지 않고, 화면은 성공처럼
보이는 문장을 보여 준다.

## 원인 — 컨테이너 문제가 아니다

`engineProcessPattern = "tossctl engine run"`이 `pgrep -f`에 그대로 넘어간다.
`pgrep -f`는 argv를 공백으로 이은 **전체 명령줄**에 이 패턴을 ERE로 맞춘다. 그런데
콘솔이 스스로 띄우는 엔진의 명령줄은 이렇다.

```
/usr/local/bin/tossctl --config-dir /var/lib/tossos/config \
  --session-file /run/tossos/session.json engine run
```

`engineArgs`가 콘솔 자신의 `--config-dir`·`--session-file`을 **앞에** 붙이기 때문에
`tossctl engine run`이라는 연속 문자열이 존재하지 않는다. 즉 이 결함은 컨테이너 고유가
아니라 **root 플래그를 갖고 뜬 모든 콘솔**에 해당한다. 플래그 없이 뜬 콘솔에서만
우연히 맞았을 뿐이다.

같은 자리의 soak은 멀쩡하다. `spawnDetachedSoak`이 `exec.Command(binary, "soak", "run")`
으로 플래그 없이 띄우므로 `"tossctl soak run"`이 그대로 맞는다. 고장난 것은 엔진
쪽 비대칭이다.

## 파급

1. **정지 버튼이 동작하지 않는다** — `stopEngine`의 `pids`가 비어 "찾지 못했다"로 끝난다.
2. **a056의 거부 분기가 컨테이너에서 도달 불가다** — `observed`가 언제나 false다.
   배타는 flock 단독이고 그것이 spec이 정한 정본 배타이므로 정합성은 유지되지만,
   a056이 약속한 "관측되면 안내하며 거부한다"는 컨테이너 밖에서만 참이다.

## What Changes

패턴을 콘솔이 실제로 spawn하는 명령줄에 맞춘다. 그리고 그렇게 넓힌 순간 생기는
**반대 방향 위험**을 같은 change에서 닫는다.

패턴만 넓히면 이 콘솔이 **다른 프로필의 엔진**까지 보게 된다. 호스트 PID namespace는
컨테이너 프로세스를 보므로, 기본 프로필로 뜬 호스트 콘솔의 정지 버튼이 컨테이너에서
포지션을 지키고 있는 엔진에 SIGTERM을 보낼 수 있다. 이 저장소가 도는 이 머신이 정확히
그 형상이다. 손절 즉시성을 약화하지 않는다는 불변식이 그것을 금지한다.

그래서 발견을 두 단계로 나눈다.

- **후보 선별(pgrep)**: 넓은 패턴으로 엔진 후보를 고른다. 셸 스크립트도 할 수 있는 일.
- **소유 판정(Go)**: 후보의 명령줄에서 journal 디렉터리를 되뽑아 이 콘솔의 것과 같을
  때만 우리 엔진으로 인정한다. spec이 엔진 인스턴스의 정체를 journal flock으로
  정의하므로, 소유의 기준도 journal 디렉터리여야 한다.

## Non-Goals

- flock 배타·stale 창(5분)·마커 갱신 주기(1분)를 바꾸지 않는다.
- a056의 결합 판정(`markerRefusesStart`)을 바꾸지 않는다. 이 change는 그 판정에
  **참인 입력이 들어가게** 만들 뿐이다.
- soak의 패턴을 바꾸지 않는다. 위에서 확인한 대로 결함이 없다.
- autostart 스크립트를 소유 판정까지 시키지 않는다. 셸에서 프로필을 되뽑는 것은
  Go에서 하는 일의 조잡한 재구현이 되고, 스크립트의 오탐은 "기동하지 않음"이라는
  보수적 방향으로만 틀린다.
- 다른 프로필의 엔진을 발견했을 때 별도 안내를 만들지 않는다(`issues.md` I3).

## Impact

- `cmd/tossctl/engineproc.go` — 패턴, `pgrepEngine`, `engineFindProcesses` seam,
  `startEngine`·`stopEngine`의 호출부.
- `tools/engine-autostart.sh` — `ENGINE_PATTERN` 한 줄과 그 의미를 적은 주석.
- `engine-safety` spec — ADDED 1건. `엔진 런타임 수명주기`는 MODIFY하지 않는다
  (a055 `issues.md` I1의 미아카이브 MODIFY 부채를 늘리지 않는다 — a056과 같은 이유).
