# Issues — a056-autostart-survives-container-recreate

## I1. `stopEngine`의 마커 읽기는 결함이 아니다

**분류**: 기록용 (수정 없음)

`stopEngine`(`cmd/tossctl/engineproc.go:207`)도 같은 마커를 읽는다.

```go
if status := enginelock.Read(enginelock.MarkerPath(dir), time.Now()); status.Running {
        return fmt.Sprintf("%s를 종료시켰지만 활성 마커가 아직 신선하다 (%s) — 최대 %s 뒤 사라진다", ...)
}
```

표현식 모양은 이 change가 고친 것과 같지만 **하는 일이 다르다.** 이것은 종료 성공 뒤의
보고이며 아무것도 거부하지 않는다. 자문 신호를 의도대로 쓰는 자리다.

이 change의 소스 규칙(`TestNoPathRefusesOnMarkerFreshnessAlone`)을 `startEngine` 본문으로
좁힌 이유가 이것이다. 파일 전체를 대상으로 하면 정당한 사용을 결함으로 신고한다.

## I2. `engineProcessPattern`이 컨테이너의 엔진과 매칭되지 않는다

**분류**: 기존 결함, a056보다 넓음 — 별도 change 필요
**발견**: 2026-08-03, tasks 5.4 실측 중

`engineProcessPattern = "tossctl engine run"`이고 `pgrep -f`에 그대로 넘어간다. 그런데
컨테이너에서 도는 엔진의 cmdline은 이렇다.

```
/usr/local/bin/tossctl --config-dir /var/lib/tossos/config \
  --session-file /run/tossos/session.json engine run
```

`tossctl`과 `engine run` 사이에 플래그가 있으므로 `tossctl engine run`이라는 연속 문자열은
**존재하지 않는다.** 실측:

```
docker exec tossos-tossos-1 pgrep -f 'tossctl engine run'   → exit 1 (매칭 없음)
docker exec tossos-tossos-1 pgrep -f 'engine run'           → 16    (엔진)
ps                                                          → 16 ... engine run (살아 있음)
```

즉 컨테이너 모드에서 `engineFindProcesses`는 **엔진이 도는 중에도 항상 빈 결과**다.

### 파급 (a056보다 넓다)

1. **`stopEngine`이 컨테이너에서 동작하지 않는다.** `pids`가 비므로
   `"실행 중인 엔진을 찾지 못했다."`를 반환한다. 콘솔의 정지 버튼이 도는 엔진을 못 찾는다.
   a056과 무관한 기존 결함이며 더 심각하다.
2. **a056의 거부 분기가 컨테이너에서는 도달 불가다.** `observed`가 언제나 false라
   `markerRefusesStart`는 언제나 false를 반환하고, 콘솔은 기동을 거부하지 않는다. 배타는
   전적으로 flock이 담당한다 — spec이 정한 정본 배타이므로 **정합성은 유지되지만**, 이
   change가 약속한 "관측되면 안내하며 거부한다"는 컨테이너 밖에서만 참이다.

### 왜 여기서 고치지 않는가

패턴 하나가 아니다. `tools/engine-autostart.sh`의 `ENGINE_PATTERN`과 짝이고
`TestTheEngineProcessPatternMatchesTheAutostartScript`가 둘의 일치를 고정한다. 패턴을
바꾸면 stop·autostart 스크립트·drift 테스트가 함께 움직이고, 느슨한 패턴(`engine run`)은
다른 프로세스를 잡을 위험을 새로 만든다. a056의 범위(기동 거부 근거)를 넘고 자체 RED/GREEN이
필요하다.

**a056은 이 결함 위에서도 안전하다**: 컨테이너에서 `observed`가 false여서 유령 마커가
기동을 막지 못하는 것이 바로 이 change가 원한 결과이고, 중복 방지는 flock이 그대로 한다.

## I3. 마커 정리는 종료 경로만 고칠 수 있다

**분류**: 대안 검토 결과, 채택하지 않음

"컨테이너 종료 시 마커를 지우면 되지 않나"는 정상 종료(SIGTERM)에만 듣는다. SIGKILL,
호스트 재부팅, OOM kill은 프로세스만 지우고 파일을 남긴다. 유령 마커는 없앨 수 있는
부류가 아니라 **견뎌야 하는** 부류라서, 생성을 막는 대신 소비를 고쳤다.

## I4. autostart 재시도 주기는 건드리지 않았다

**분류**: 범위 밖 (의도)
**정정**: 2026-08-03 — 아래 취소선 문단은 **틀렸다**

~~사고 당시 콘솔은 8분 뒤 스스로 복구했다 — autostart는 부팅 때 한 번만 도는 것이 아니라
재평가한다.~~

`runConfiguredEngineAutostart`는 `cmd/tossctl/console.go:334`에서 **콘솔 기동 시 한 번만**
호출된다. 재평가는 없다. 같은 밤 02:44에 난 두 번째 사고가 그것을 보였다 — 마커가 stale이
되고도 엔진은 11분간 올라오지 않았고, 배포가 사람 손으로 이루어질 때까지 복구되지 않았다.
01:25의 8분 복구는 재평가가 아니라 다른 원인이었다.

고칠 것이 **최초 거부**라는 결론 자체는 그대로다. 오히려 재평가가 없으므로 최초 거부의
비용이 이 문단이 가정했던 것보다 크다 — 놓치면 사람이 알아챌 때까지 정지 상태다.

재시도 주기 도입은 여전히 이 change의 범위 밖이고, 별도 change가 필요하면 그 근거는
"재평가가 있으니 불필요"가 아니라 "재평가가 없으니 최초 거부가 유일한 기회"다.
