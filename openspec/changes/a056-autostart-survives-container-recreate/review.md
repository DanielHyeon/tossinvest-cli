# Review — a056-autostart-survives-container-recreate

## 무엇을 바꿨나

`startEngine`의 사전 확인 한 칸. 자문 마커가 신선하다는 것 **하나만으로** 기동을 거부하던
분기를, 다른 관측이 동의할 때만 거부하도록 바꿨다.

```
변경 전   if status.Running { return 거부 }          ← 아래 프로세스 검사에 도달 불가
변경 후   if markerRefusesStart(status.Running, observed, findErr != nil) { return 거부 }
```

여섯 칸 진리표 중 **한 칸**이 움직인다: 마커 fresh + 프로세스 미관측 → 거부에서 진행으로.
나머지 다섯 칸은 그대로다.

## 왜 안전 방향인가

- 배타는 여전히 journal flock이다. 이 change는 flock에 **도달하게** 만들 뿐 flock을
  건드리지 않는다. journal 단일 writer 불변식은 유지된다.
- 거부를 없애지 않았다. 근거를 자문 파일에서 관측된 프로세스로 옮겼고, 안내 문구는
  마커의 PID·갱신 시각을 계속 쓴다(`TestAFreshMarkerWithALiveProcessStillRefuses`가
  그 두 조각이 문구에 남아 있는지 고정한다).
- 열거 실패는 부재가 아니다. 증명할 수 없는 부재는 부재로 읽지 않고 기존 거부를 유지한다.
- 손절 즉시성 방향: 엔진이 떠 있어야 exit 루프가 돈다. 유령 마커로 인한 감시 공백을
  없애는 것은 즉시성을 **강화**한다.

## 완화 방향 위험 하나 — 그리고 실측이 그것을 더 크게 만들었다

엔진이 실제로 살아 있는데 pgrep이 못 보면 두 번째 spawn이 일어난다. 그 spawn은 flock에서
죽으므로 결과는 올바르고 잃는 것은 안내 문구의 품질이다.

**실측 결과 그 경우는 예외가 아니라 컨테이너의 기본값이었다** (issues.md I2).
`engineProcessPattern`은 `"tossctl engine run"`인데 컨테이너 엔진의 cmdline은
`/usr/local/bin/tossctl --config-dir … --session-file … engine run`이라 그 연속 문자열이
없다. `pgrep -f 'tossctl engine run'` → exit 1, `pgrep -f 'engine run'` → 16.

그래서 컨테이너 모드에서는 `observed`가 언제나 false이고, a056의 거부 분기는 도달하지
못한다. 배타는 전적으로 flock이며 이는 spec이 정한 정본 배타이므로 정합성은 유지된다.
다만 이 change가 약속한 "관측되면 안내하며 거부한다"는 **컨테이너 밖에서만 참**이다.
review 문구를 그렇게 정정한다.

같은 결함이 `stopEngine`에도 있다 — 도는 엔진을 못 찾아 "실행 중인 엔진을 찾지 못했다"를
반환한다. a056과 무관한 기존 결함이고 더 심각하며, 패턴·autostart 스크립트·drift 테스트가
함께 움직여야 해서 별도 change로 남겼다.

## 구현 중 발견

### F1. 소스 규칙을 표현식 모양으로 쓰면 정당한 사용까지 잡는다

`TestNoPathRefusesOnMarkerFreshnessAlone`의 첫 초안은 파일 전체에서
`enginelock.Read(...); status.Running {` 형태를 금지했다. 그 즉시 `stopEngine`이 걸렸다.

`stopEngine`의 그 코드는 거부가 아니라 **보고**다 — "종료시켰지만 활성 마커가 아직
신선하다, 최대 StaleAfter 뒤 사라진다". 자문 신호를 의도대로 쓰는 자리다. 상태를
*이름 붙이는* 것과 상태를 *판정하는* 것은 다르고, 이 change가 고정하려는 규칙은 후자에만
해당한다. 그래서 검사를 `startEngine` 본문으로 좁혔고, 왜 좁혔는지를 테스트 주석에 남겼다.

**이 발견은 테스트가 코드보다 먼저 만들어 낸 것이다.** 규칙을 코드로 쓰지 않았다면
`stopEngine`의 그 줄을 "같은 결함"으로 오해하고 함께 고쳤을 것이다.

### F2. `engineFindProcesses`를 두 번 부르지 않는다

결합 판정과 그 아래 프로세스 거부가 각자 열거하면 한 결정 안에서 두 답이 나올 수 있다.
한 번 부르고 두 곳이 같은 관측을 공유하게 했다.

### F3. `markerRefusesStart`는 인라인 조건이 아니라 이름 있는 함수다

진리표 여섯 칸을 프로세스 spawn 없이 직접 고정할 수 있고(`TestMarkerRefusesStart‑
OnlyWithCorroboration`), 다음 편집자가 조건과 함께 이유를 읽는다. 되돌리기 쉬운 모양은
되돌아간다.

## 변이 검증

| 변이 | RED가 된 테스트 | 되돌림 |
|---|---|---|
| B3 조건을 `status.Running` 단독으로 복원 | `TestAGhostMarkerDoesNotRefuseAStart`, `TestNoPathRefusesOnMarkerFreshnessAlone` | yes |
| `observed`를 항상 false로 | `TestAFreshMarkerWithALiveProcessStillRefuses`, `TestStartingIsRefusedWhenAProcessIsAlreadyThere` | yes |
| 열거 오류를 부재로 취급 | `TestEnumerationFailureKeepsTheRefusal` | yes |

세 변이 모두 RED, 모두 되돌렸고, 되돌린 뒤 전체 green.

## Function Logic Map

`cmd/tossctl/engineproc.go:startEngine`의 map과 branch map을 **편집 전에** 작성했다
(tasks 1.1). 엔진 기동 경로이므로 면제하지 않았다. 대상 10건, 그중 하나
(`TestStartingIsRefusedWhileTheMarkerIsFresh`)는 삭제된 함수라 base revision 증거다.

## 남은 것

- tasks 5.3/5.4의 컨테이너 실측은 아직 하지 않았다. 재배포가 필요하고, 휴장 시간과 사람
  승인을 조건으로 걸어 두었다.
