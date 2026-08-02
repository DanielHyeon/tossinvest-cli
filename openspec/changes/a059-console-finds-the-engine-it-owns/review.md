# Review — a059-console-finds-the-engine-it-owns

- 날짜: 2026-08-03
- 보이스: Manager 셀프리뷰 + 적대적 Eng 관점 (엔진 종료 시그널 경로이므로 위험 등급 가중,
  `docs/WORKFLOW.md` 리뷰 게이트)
- 근거: a056 `issues.md` I2의 실측, 프로덕션 컨테이너 재측정, 변이 검증 4종

## 무엇을 바꿨나

프로세스 발견 한 곳. 그리고 그것을 고치는 순간 열리는 문을 같은 change에서 닫았다.

```
변경 전   pgrep -f "tossctl engine run"          ← 콘솔이 띄운 엔진과 매칭되지 않음
변경 후   pgrep -a -f "tossctl( .*)? engine run" ← 후보 선별
          + journal 디렉터리 일치 판정           ← 소유 판정 (Go에만 있다)
```

## 왜 두 개가 한 change인가

패턴만 넓히면 결함 하나를 더 나쁜 결함으로 바꾼다.

호스트는 컨테이너와 PID namespace를 공유한다. 이 저장소가 도는 머신에서 지금
`pgrep -af tossctl`은 컨테이너의 엔진(pid 1411526 = 컨테이너 안의 16)을 그대로 보여
준다. 기본 프로필로 뜬 호스트 콘솔이 패턴만 넓힌 코드로 정지 버튼을 누르면, OPEN
포지션의 exit 루프를 돌리고 있는 컨테이너 엔진에 SIGTERM이 간다.

`docs/WORKFLOW.md` §0.3 — "손절·비상 청산의 즉시성을 약화·지연하는 변경은 금지한다".
소유 판정 없는 패턴 확장은 그 금지에 걸린다. 그래서 둘은 나눌 수 없다.

## 왜 안전 방향인가

- **시그널은 좁아진다.** 변경 전 컨테이너에서 이 경로는 0건을 시그널했다(우리 엔진을 못
  찾아서). 변경 후에는 우리 엔진 1건을 시그널하고 남의 엔진 0건을 유지한다. "우리 것"의
  정의는 spec이 인스턴스 배타에 쓰는 것과 같은 journal 디렉터리다.
- **증명 불가는 전부 배제 방향이다.** 파싱 안 되는 줄, 명령줄 없는 pid, 해석 안 되는
  기본 경로 — 전부 목록에서 빠진다. SIGTERM 대상 목록에서 "모르겠다"는 "내 것 아님"이어야
  한다.
- **열거 실패는 여전히 부재가 아니다.** a056 D3 그대로다. 빈 목록과 오류는 다른 것이고,
  `pgrep -a` 미지원 환경에서도 오류로 떨어져 기동 거부가 유지된다.
- **거부가 늘어나는 쪽은 의도다.** a056의 거부 분기가 컨테이너에서 처음으로 도달한다.
  늘어나는 거부는 전부 "엔진이 실제로 돌고 있을 때"이며, `엔진 런타임 수명주기`의
  `두 번째 인스턴스 기동` scenario가 이미 그 결과를 정하고 있다.
- **손절 즉시성**: 정지 버튼이 동작하는 것 자체는 운영자의 명시적 요청이다. 이 change가
  새로 만드는 능력이 아니라 원래 있어야 했던 능력이다.

## 구현 중 발견

### F1. drift 테스트 세 개가 있었는데 아무도 이 결함을 잡지 못했다

`TestTheEngineProcessPatternMatchesTheAutostartScript`,
`TestTheRestartCapMatchesTheAutostartScript`,
`TestTheStalenessWindowMatchesTheAutostartScript` — 셋 다 **Go 상수와 셸 스크립트**를
대조한다. 양쪽이 **같은 값으로 틀려 있으면** 전부 통과한다. 정확히 그 상태였다.

빠져 있던 대조는 상수와 **실제 argv**다. 그래서 새 테스트는 명령줄을 손으로 적지 않고
`engineArgs`를 호출해 만든다 — spawn 경로가 바뀌면 테스트가 따라 움직인다.

이 발견은 이 change보다 넓다. `soakProcessPattern`도 같은 종류의 대조만 갖고 있으며,
지금은 우연히 맞다(`spawnDetachedSoak`이 플래그를 붙이지 않는다). `issues.md` I1.

### F2. `pgrep -a`는 BusyBox에도 있다 — 확인하고 썼다

컨테이너는 procps-ng가 아니라 BusyBox 1.37이다. `-a`(Show command line too)가 있고
`PID CMDLINE` 형식이 procps-ng와 같다는 것을 실측으로 확인했다. 없는 환경에서는 pgrep이
오류를 반환하고 그것은 열거 실패로 처리된다 — 부재로 읽지 않는다.

### F3. `stopEngine`의 journal 디렉터리 해석이 앞으로 왔고, 오류가 더 이상 삼켜지지 않는다

변경 전에는 함수 **맨 끝**에서 마커 문장을 만들 때만 구했고 `derr`을 무시했다. 소유
판정에는 그 값이 필요하므로 맨 앞으로 옮겼고, 실패하면 오류를 반환한다.

동작 변화가 하나 있다: journal 경로를 해석하지 못하는 환경에서 예전에는 시그널을
시도하고 마커 보고만 생략했지만, 이제는 아무것도 하지 않고 오류를 돌려준다. 소유를
판정할 수 없을 때 SIGTERM을 보내지 않는 것이 이 change의 규칙이므로 일관된다.
부수적으로 `engineJournalDir` 중복 호출 하나가 사라졌다.

### F4. `--config-dir=X` 표기도 읽는다

`engineArgs`는 공백 표기만 만들지만 cobra는 둘 다 받는다. 사람이 `=`로 띄운 엔진을
"남의 것"으로 오판하면 정지 버튼이 다시 조용해진다.

## 변이 검증

| 변이 | RED가 된 테스트 | 되돌림 |
|---|---|---|
| 패턴을 `tossctl engine run`으로 복원 | `TestTheProcessPatternMatchesWhatTheConsoleSpawns`, `TestStoppingFindsTheEngineTheConsoleStarted`, `TestStartingIsRefusedWhenOurOwnEngineIsAlreadyRunning`, `TestTheEngineProcessPatternMatchesTheAutostartScript` | yes |
| 소유 판정 약화 (`\|\|` → `&&`) | `TestOnlyThisProfilesEngineIsFound`, `TestTheDefaultProfileMatchesAnEngineStartedWithoutFlags`, `TestStoppingDoesNotSignalAnotherProfilesEngine` | yes |
| 토큰 경계 제거 (`tossctl.*engine run`) | `TestTheProcessPatternIgnoresTheOtherSubcommands` | yes |
| lookup에 빈 journal 디렉터리 전달 | `TestBothButtonsAskAboutThisProfilesJournal`, `TestStartingIsRefusedWhenOurOwnEngineIsAlreadyRunning` | yes |

네 변이 모두 RED, 모두 되돌렸고, 되돌린 뒤 `cmd/tossctl` 433건 green.

## Function Logic Map

`pgrepEngine`·`startEngine`·`stopEngine`의 map과 branch map을 **편집 전에** 작성했다
(tasks 1.1). 엔진 기동·종료 경로이므로 면제하지 않았다. 편집 후 AST와 세 branch map을
갱신했고, 신규 leaf 함수와 테스트 함수 19건은 생성 산출물로 채웠다. 대상 22건.

## 컨테이너 실측

tasks 5.4/5.5 — 아래 "실측" 절 참조.
