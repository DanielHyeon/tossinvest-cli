# a059 · Design

## Context

`cmd/tossctl/engineproc.go`의 발견 경로는 한 줄이다.

```go
const engineProcessPattern = "tossctl engine run"
exec.Command("pgrep", "-f", engineProcessPattern)
```

그리고 spawn 경로는 다른 명령줄을 만든다.

```go
func engineArgs(root *rootOptions) []string {
        args := []string{"engine", "run"}
        if root.sessionFile != "" { args = append([]string{"--session-file", …}, args...) }
        if root.configDir  != "" { args = append([]string{"--config-dir",  …}, args...) }
        return args
}
```

두 경로가 같은 프로세스를 두고 서로 다른 문자열을 가정한다. 지금까지 이 불일치를
잡는 테스트가 없었다 — drift 테스트 세 개는 **Go 상수와 셸 스크립트**만 대조하고,
상수와 **실제 argv**는 대조하지 않는다. 양쪽이 같은 값으로 틀려 있어도 통과한다.

## Goals

- 콘솔이 spawn한 엔진을 콘솔이 찾는다.
- 이 콘솔이 소유하지 않은 엔진에는 시그널이 가지 않는다.
- 발견 실패는 여전히 부재가 아니다(a056 D3 유지).

## Non-Goals

- flock·stale 창·마커 주기 변경.
- a056 결합 판정의 형태 변경.
- 스크립트의 소유 판정.

## Decisions

### D1. 패턴은 argv 토큰 경계를 지키는 ERE로 넓힌다

```
tossctl( .*)? engine run
```

`tossctl` 뒤에 아무 플래그가 와도, 하나도 안 와도 맞는다. `engine`과 `run` 앞의 공백을
요구하므로 `…myengine run` 같은 부분 토큰에는 걸리지 않는다.

`$` 앵커는 쓰지 않는다. `engine run` 뒤에 플래그가 붙는 호출(`tossctl engine run
--config-dir X`)을 사람이 할 수 있고, 앵커는 그 경우를 **조용히** 놓친다 — 지금 고치는
결함이 정확히 그 종류다.

실측(2026-08-03, 프로덕션 컨테이너, BusyBox pgrep 1.37):

```
pgrep -af 'tossctl( .*)? engine run'
  → 16 /usr/local/bin/tossctl --config-dir … --session-file … engine run
```

콘솔(pid 7)도 httpapi도 매칭되지 않는다.

### D2. 소유는 journal 디렉터리로 판정한다 — 플래그 문자열이 아니라

패턴만 넓히면 이 콘솔이 다른 프로필의 엔진에 SIGTERM을 보낼 수 있다(proposal 참조).
그래서 pgrep 결과를 그대로 쓰지 않고 한 단계 거른다.

기준은 journal 디렉터리다. `엔진 런타임 수명주기`가 인스턴스의 배타를 "journal
디렉터리 flock"으로 정의하므로, 두 프로세스가 같은 엔진 인스턴스인지도 journal
디렉터리가 정한다. 명령줄에서 `--config-dir`를 되뽑아 `engineJournalDir`에 **같은
함수로** 통과시키면, `--config-dir`를 명시한 콘솔과 생략한 autostart가 같은 기본
경로를 가리키는 경우도 자동으로 같다고 판정된다. 플래그 문자열을 비교하면 그 경우가
틀린다.

`pgrep -a -f`가 pid와 명령줄을 함께 준다(procps-ng, BusyBox 둘 다). 파싱과 판정은
프로세스 테이블 없이 테스트되는 순수 함수로 둔다.

### D3. 소유를 증명할 수 없으면 우리 것이 아니다

명령줄을 못 읽거나 기본 journal 경로를 해석하지 못하면 그 pid는 **제외**한다.
방향이 D3(a056)와 반대로 보이지만 같은 원리다: 확신 없이 하면 안 되는 쪽으로 기운다.

| | 확신 없음 → | 왜 |
|---|---|---|
| 기동 거부 (a056) | 거부 유지 | 잘못 허용하면 flock 하나에 기댄다 |
| 종료 시그널 (a059) | 시그널 안 보냄 | 잘못 보내면 포지션을 지키던 엔진이 멈춘다 |

열거 **자체**가 실패했을 때(pgrep 오류)는 a056 D3 그대로 error를 올려 보내고,
`startEngine`은 거부를 유지한다. "빈 목록"과 "오류"는 계속 다른 것이다.

### D4. 스크립트는 후보 선별까지만 한다

`tools/engine-autostart.sh`는 같은 넓은 패턴을 쓰고 소유 판정은 하지 않는다. 셸에서
`--config-dir` 파싱과 기본 경로 해석을 재구현하면 Go와 달라질 수 있고, 그 차이가
바로 이 change가 고치는 종류의 버그다.

스크립트의 오탐이 만드는 결과는 "기동하지 않는다" 하나뿐이다. 스크립트는 플래그
없이 기본 프로필만 띄우므로, 다른 프로필의 엔진을 보고 물러서는 것은 보수적 방향이다.

drift 테스트의 의미도 그에 맞춰 바뀐다: 두 반쪽이 **같은 후보 패턴**을 쓴다는 계약은
그대로고, 소유 판정이 Go 쪽에만 있다는 사실을 테스트가 명시한다.

## Risks

| 위험 | 완화 |
|---|---|
| 넓힌 패턴이 엉뚱한 프로세스를 잡아 SIGTERM | 소유 판정(D2)이 journal 디렉터리로 거른다. `os.Getpid()` 제외는 유지 |
| 소유 판정이 우리 엔진을 놓침 | 정지 버튼이 "찾지 못했다"를 돌려준다 — 지금과 같은 상태이고 더 나빠지지 않는다. 기동 쪽은 flock이 받는다 |
| `pgrep -a` 미지원 환경 | pgrep이 오류를 반환하고 열거 실패로 처리된다(a056 D3) — 부재로 읽지 않는다 |
| a056 거부 분기가 이제 실제로 도달 | 그것이 의도다. 변이 검증으로 두 분기 모두 고정한다 |

## Function Logic Map 대상

기존 함수 내부를 바꾸므로 편집 **전에** 작성한다. 엔진 기동·종료 경로이므로 면제하지
않는다.

- `cmd/tossctl/engineproc.go:pgrepEngine`
- `cmd/tossctl/engineproc.go:startEngine`
- `cmd/tossctl/engineproc.go:stopEngine`
