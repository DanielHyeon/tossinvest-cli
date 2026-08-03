# a060 · Design

## Context

두 spawn 경로가 정반대로 틀려 있다.

```go
// engine — 콘솔의 프로필을 물려준다. 그래서 a059 이전에는 패턴이 못 맞았다.
func engineArgs(root *rootOptions) []string { … "--config-dir", …, "engine", "run" }

// soak — 물려주지 않는다. 그래서 패턴은 맞고, 자식이 아무것도 못 찾는다.
exec.Command(binary, "soak", "run")
```

콘솔은 `soakRecord`를 자기 프로필로 계산해서 화면에 그리고 로그 경로도 거기서 뽑는데
(`soakLogPath(recordPath)`), 정작 자식은 기본 프로필로 뜬다. 로그만 프로필 안에 남고
기록과 자격증명 조회는 밖에서 일어난다.

프로덕션 관측(2026-08-03): `config/soak.log`는 "no Open API credentials" 반복,
`config/capability-soak.jsonl` 부재, `config/openapi-credentials.json` 존재.

## Goals

- 콘솔이 띄운 서베이가 콘솔이 보는 기록에 쓴다.
- 콘솔이 자기가 띄운 서베이를 찾는다.
- 이 콘솔이 소유하지 않은 서베이에는 시그널이 가지 않는다.

## Non-Goals

- `soak run`의 판정·주기·조회 전용 성질 변경.
- 기존 기록 이전, 자격증명 복사.
- 저장소 밖 설치 스크립트 수정.

## Decisions

### D1. `soakArgs`는 `engineArgs`와 같은 모양이다

같은 순서로 같은 두 플래그를 앞에 붙인다. 다르게 만들 이유가 없고, 다르면 다음 사람이
둘 중 어느 쪽이 맞는지 고민한다.

기록 경로는 **넘기지 않는다.** `resolveSoakRecord`가 `--config-dir`에서 같은 경로를
유도하므로 중복이고, 두 경로가 어긋날 수 있는 자리를 새로 만드는 셈이다. `--record`
override는 콘솔이 쓰지 않는 CLI 전용 통로로 남긴다.

### D2. soak의 소유 기준은 journal 디렉터리가 아니라 **기록 경로**다

엔진의 정체는 journal flock이 정한다(a059 D2). 서베이의 정체를 정하는 것은 그것이
append하는 기록이다 — soakproc.go 자신이 "두 서베이가 한 기록에 append하는 것"을
피해야 할 상태로 적고 있다.

그래서 판정은 명령줄에서 되뽑은 `--config-dir`를 `resolveSoakRecord`에 **콘솔이 쓰는
그 함수 그대로** 통과시켜 비교한다. a059가 `engineJournalDir`로 한 것과 같은 방식이다.

### D3. 소유 판정은 두 프로세스가 공유한다

a059의 `enginePIDsForJournal`을 `pidsOwnedBy`로 올린다. 인자는 후보 줄, 매처, 원하는
정체, 그리고 "명령줄의 `--config-dir` → 정체" 해석 함수다. 해석 함수가 기본값도
담당하므로(빈 문자열을 받으면 기본 프로필) 별도 fallback 인자가 사라진다.

세 번째 복사본을 만들지 않는 이유는 `joinPIDs`·`pidAlive`가 이미 두 파일이 공유하는
선례이기 때문이다. 판정 로직이 갈라지면 그 차이가 바로 a059·a060이 고치는 종류의 버그다.

### D4. 저장소 밖 autostart 스크립트는 좁은 패턴으로 남는다

`~/.local/share/tossos/bin/soak-autostart.sh`(저장소에 없음)는 `pgrep -f "tossctl soak
run"`을 쓰고 자신은 플래그 없이 spawn한다. 그래서 **자기 자식은 계속 본다.**

넓힌 Go 패턴과의 관계는 한 방향으로만 어긋난다: 스크립트는 콘솔이 띄운(플래그 달린)
서베이를 못 본다. 결과는 "호스트 기본 프로필에서 서베이를 하나 더 시작할 수 있다"이며,
그 둘은 서로 다른 기록에 쓰므로 기록 경합은 없고 계좌 rate budget만 공유한다.

이 change에서 스크립트를 고치지 않는 이유는 저장소 밖 운영 산출물이기 때문이다.
사실과 권고는 `issues.md` I3에 남긴다.

### D5. drift 테스트를 실재하는 계약으로 바꾼다

기존 테스트는 상수를 자기가 적은 리터럴과 비교한다. 저장소에 대조 상대가 없으므로 그
비교는 값이 바뀌었다는 것만 알려 주고 **그 값이 맞는지는 영원히 말하지 못한다.**
a059에서 엔진 패턴이 세 개의 drift 테스트를 통과하며 틀려 있었던 것과 같은 구조다.

대체 계약: `soakArgs`가 만드는 명령줄에 패턴이 맞는가, 그리고 다른 하위 명령
(`engine run`·`console`·`httpapi`)에는 안 맞는가.

## Risks

| 위험 | 완화 |
|---|---|
| 넓힌 패턴이 엔진을 잡아 SIGINT | 패턴이 `soak run`을 요구하고 토큰 경계를 지킨다. 배제 테스트가 `engine run` 명령줄을 명시적으로 고정한다 |
| 남의 프로필 서베이에 시그널 | 소유 판정(D2)이 기록 경로로 거른다 |
| 설치된 스크립트가 두 번째 서베이를 시작 | 서로 다른 기록이라 경합 없음. rate budget 공유는 `issues.md` I3에 기록하고 스크립트 은퇴를 권고 |
| 콘솔이 띄운 서베이가 새 기록을 시작 | 콘솔은 지금도 그 경로를 보며 "아직 기록이 없다"라고 말한다. 기대와 일치한다 |
| soak이 조회 전용성을 잃음 | 이 change는 argv 두 플래그와 발견만 바꾼다. `internal/soak`의 import-graph 테스트가 조회 전용을 계속 고정한다 |

## Function Logic Map 대상

기존 함수 내부를 바꾸므로 편집 **전에** 작성한다.

- `cmd/tossctl/soakproc.go:restartSoak`
- `cmd/tossctl/soakproc.go:pgrepSoak`
- `cmd/tossctl/soakproc.go:spawnDetachedSoak`
- `cmd/tossctl/engineproc.go:enginePIDsForJournal` (공유 헬퍼로 추출, 동작 불변)
