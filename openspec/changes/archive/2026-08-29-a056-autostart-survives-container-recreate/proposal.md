# a056 · 자문 마커는 기동을 거부하지 못한다

- **Feature**: `FEAT-TOS-003` — Engine interlocks and Guardian controls
- **Story**: `STORY-TOS-a056`
- **Spec**: `engine-safety`

## Why

2026-08-02 컨테이너 재배포에서 엔진이 약 8분간 뜨지 않았다. OPEN 포지션 5건과 활성
exit 정책 4건이 그동안 감시 없이 남았다. 두 시장 모두 휴장이라 손실은 없었지만,
개장 중 배포였다면 같은 창이 장중에 열린다.

로그가 원인을 그대로 말한다.

```
01:16:35  옛 컨테이너 엔진(pid 16)의 마지막 마커 갱신
01:17:04  새 컨테이너 기동 (29초 뒤)
01:17:04  엔진 자동 시작 거부: 엔진이 이미 실행 중이다 (pid 16, 마지막 갱신 01:16:35)
01:25:xx  엔진 자동 시작: 엔진을 시작했다
```

pid 16은 새 PID namespace에 존재하지 않는다. 컨테이너와 함께 죽었다. 남은 것은
`engine-run.json` 파일 하나이고, 그 파일 자신이 이렇게 적고 있다.

> advisory only. The exclusion is the flock on engine.lock; this file exists so the
> console can draw the engine's status and the autostart script can check before
> spawning. A stale one (older than 5m) is ignored.

## 계약 위반

`engine-safety`의 `엔진 런타임 수명주기`는 이미 이렇게 정한다.

> **엔진 활성 마커**(갱신 1분·stale 5분 — runlock 선례 수치)를 유지해 콘솔의 엔진 상태
> 표시·autostart의 **사전 확인**이 소비한다(SHALL — **자문 신호이며 배타는 flock이
> 담당함을 명시**).

`cmd/tossctl/engineproc.go:128`의 `startEngine`은 그 자문 신호 하나로 기동을 **거부한다**.

```go
// Advisory, and only to give a better answer than "the engine refused because
// it could not take the lock". The exclusion is the engine's own flock.
if status := enginelock.Read(enginelock.MarkerPath(dir), time.Now()); status.Running {
        return "", fmt.Errorf("엔진이 이미 실행 중이다 (pid %d, 마지막 갱신 %s)", ...)
}
if pids, perr := engineFindProcesses(); perr == nil && len(pids) > 0 {
        return "", fmt.Errorf("엔진 프로세스가 이미 있다 (%s)", joinPIDs(pids))
}
```

주석이 스스로 "advisory"라고 말하는 검사가 `return`으로 끝난다. 바로 아래에 **실제
프로세스를 세는 검사**가 이미 있는데, 자문 검사가 먼저라 거기까지 가지 못한다.

`두 번째 인스턴스 기동` scenario는 "엔진이 **실행 중인** 머신"을 전제한다. 마커만 남고
프로세스는 없는 상태는 그 scenario가 말하는 경우가 아니다. 지금 코드는 그 거부를
해당하지 않는 경우까지 넓혀 적용한다.

## What Changes

자문 마커 단독으로는 기동을 거부하지 못하게 한다. 마커가 신선한데 엔진 프로세스가
관측되지 않으면 유령 마커이므로, 기동을 진행하고 **배타는 flock에 맡긴다** — 정말로
엔진이 살아 있다면 flock 획득이 실패하고 그것이 정본 거부가 된다.

거부 문구는 사라지지 않는다. 프로세스가 실제로 관측되면 지금과 같은 안내가 나가고,
flock이 거부하면 flock의 사유가 나간다.

## Non-Goals

- stale 창(5분)이나 마커 갱신 주기(1분)를 바꾸지 않는다.
- flock 배타 자체를 바꾸지 않는다. 이 change는 flock에 **도달하게** 만들 뿐이다.
- autostart 재시도 주기를 도입하지 않는다. 콘솔은 이미 재평가하므로(01:25 복구가 그
  증거다) 필요한 것은 더 빠른 재시도가 아니라 애초에 거부하지 않는 것이다.
- 컨테이너 종료 시 마커를 지우지 않는다. 그것은 정상 종료 경로에만 듣고, SIGKILL·호스트
  재부팅·OOM에는 듣지 않는다. 유령 마커는 없앨 수 없으므로 견뎌야 한다.

## Impact

- `cmd/tossctl/engineproc.go` — `startEngine`의 사전 확인 순서와 조건.
- `engine-safety` spec — ADDED 1건. `엔진 런타임 수명주기`는 MODIFY하지 않는다.
  그 요구사항을 MODIFY하는 미아카이브 delta가 쌓여 있고, 새 MODIFIED를 더하면
  a055 issues.md I1이 기록한 부채가 늘어난다.
