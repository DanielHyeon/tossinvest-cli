# Review — a060-soak-restart-uses-this-profile

- 날짜: 2026-08-03
- 보이스: Manager 셀프리뷰 + 적대적 Eng 관점 (프로세스 시그널 경로이므로 가중;
  대상은 조회 전용 서베이라 주문 위험은 없다)
- 근거: 프로덕션 파일 실측, 변이 검증 4종, a059 회귀 9건

## 무엇을 바꿨나

콘솔의 [soak 재시작] 버튼 하나. 세 조각이 함께 움직인다.

```
변경 전   exec.Command(binary, "soak", "run")      ← 자식이 기본 프로필로 뜬다
          pgrep -f "tossctl soak run"              ← 플래그가 없어서 우연히 맞았다
          소유 판정 없음

변경 후   exec.Command(binary, soakArgs(root)...)  ← 콘솔의 프로필을 물려받는다
          pgrep -a -f "tossctl( .*)? soak run"     ← 그 명령줄에 맞는다
          pidsOwnedBy(… soakRecordForConfigDir)    ← 기록 경로로 소유를 판정한다
```

## 이건 예측이 아니라 이미 벌어진 일이었다

a059 `issues.md` I1은 "soak spawn에 `--config-dir` 전달을 추가하면 같은 결함이 재현된다"
고 적었다. 확인해 보니 전달을 **안 해서** 이미 깨져 있었다.

```
$ docker exec tossos-tossos-1 tail /var/lib/tossos/config/soak.log
Error: soak: no Open API credentials — … Run `tossctl openapi login` …   (반복)

$ docker exec tossos-tossos-1 ls /var/lib/tossos/config/openapi-credentials.json
-rw-------  147  Jul 26 12:27        ← 자격증명은 있다
$ docker exec tossos-tossos-1 ls /var/lib/tossos/config/capability-soak.jsonl
No such file or directory            ← 콘솔이 가리키는 기록은 없다
```

콘솔 화면은 "아직 기록이 없다. `tossctl soak run`으로 시작하라
(/var/lib/tossos/config/capability-soak.jsonl)"라고 말한다. 정직한 문장인데, 그 아래
버튼을 눌러도 그 기록은 생기지 않는다. 자식이 자격증명을 못 찾고 즉시 죽기 때문이다.
컨테이너에서 이 버튼은 **한 번도 성공한 적이 없다.**

## 세 조각을 나눌 수 없는 이유 — 변이가 증명한다

프로필 상속만 넣고 패턴을 그대로 두면 무엇이 깨지는지 변이 M2가 그대로 보여 준다.
상속을 유지한 채 패턴만 옛 리터럴로 되돌리자 세 테스트가 빨개졌다.

| 변이 | RED가 된 테스트 | 되돌림 |
|---|---|---|
| M1 argv를 `"soak","run"`으로 복원 | `TestTheSoakSpawnCarriesThisProfile` | yes |
| M2 패턴만 `tossctl soak run`으로 복원 (상속 유지) | `TestTheSoakPatternMatchesWhatTheConsoleSpawns`, `TestOnlyThisRecordsSoakIsFound`, `TestTheRestartDoesNotSignalAnotherRecordsSoak` | yes |
| M3 공유 소유 판정 약화 (`\|\|` → `&&`) | 엔진 2건 + soak 2건 | yes |
| M4 공유 헬퍼가 매처를 안 봄 | `TestUnparsableProcessLinesAreDropped`, `TestOnlyThisRecordsSoakIsFound` | yes |

M3이 공유 헬퍼의 값을 보여 준다. 한 줄을 약화했더니 엔진과 soak 양쪽이 함께 빨개졌다 —
판정이 하나라는 뜻이다.

## 왜 안전 방향인가

- **조회 전용 경로다.** `internal/soak`은 구조적으로 계좌를 바꿀 수 없고
  import-graph 테스트가 그것을 고정한다. 이 change는 argv 두 플래그와 발견만 건드린다.
- **시그널은 좁아진다.** 패턴을 넓히면 이 콘솔이 모든 프로필의 서베이를 보게 되는데,
  호스트는 컨테이너와 PID namespace를 공유한다. 소유 판정이 기록 경로로 거른다.
  기준을 기록으로 잡은 것은 soakproc.go 자신이 "두 서베이가 한 기록에 append하는 것"을
  피해야 할 상태로 적고 있기 때문이다.
- **"안 죽으면 spawn 안 함" 규칙은 그대로다.** 이 change는 그 판단에 들어가는 목록만
  정확하게 만든다.
- **엔진은 한 글자도 달라지지 않는다.** a059가 만든 엔진 소유 판정 테스트 9건이 추출
  뒤에도 손대지 않은 채 green이다 (tasks 5.4). 하나라도 고쳐야 했다면 그것은 추출이
  아니라 동작 변경이었을 것이다.

## 구현 중 발견

### F1. soak의 drift 테스트에는 대조할 상대가 없었다

```go
if soakProcessPattern != "tossctl soak run" { … }   // 삭제함
```

`tools/soak-autostart.sh`는 이 저장소에 없다. 엔진 쪽은 `autostartScript(t)`로 실제
파일을 읽어 대조하지만 soak 쪽은 상수를 자기가 적은 리터럴과 비교할 뿐이었다. 값이
바뀌었다는 것만 알려 주고 **그 값이 맞는지는 영원히 말하지 못하는** 테스트다.

`soakArgs`가 만드는 argv에 패턴을 묶는 계약으로 교체했다. 설치된 스크립트와의 관계는
테스트가 아니라 `issues.md` I3에 사실로 남겼다.

### F2. `--record`를 자식에게 넘기지 않기로 했다

넘기면 기록 경로가 두 곳에서 정해진다(부모가 계산한 값, 자식이 `--config-dir`에서
유도한 값). 두 값이 어긋날 수 있는 자리를 새로 만드는 셈이라 `--config-dir` 하나만
넘긴다. `--record` override는 CLI 전용 통로로 남겼다.

### F3. `enginePIDsForJournal`은 wrapper로 남겼다

시그니처를 그대로 둔 덕분에 a059의 테스트 7건을 **한 줄도 고치지 않고** 회귀 증거로
쓸 수 있었다. 추출이 추출임을 증명하는 가장 싼 방법이다.

## Function Logic Map

`restartSoak`·`pgrepSoak`·`spawnDetachedSoak`·`enginePIDsForJournal`의 map과 branch
map을 **편집 전에** 작성했다 (tasks 1.1). 편집 후 AST를 갱신했고, 신규 leaf 함수와
테스트 함수는 생성 산출물로 채웠다. 대상 26건, 그중 둘
(`engineCommandConfigDir` — 이름이 바뀌어 사라짐, `soakLogPath` — 주변 편집에 밀림)은
base revision 증거다.

## 컨테이너 실측

tasks 5.7 — **아직 하지 않았다.** 배포가 엔진을 재시작하는데 지금은 KRX 장중이다.
측정 항목과 조건은 tasks.md에 적어 두었고, 배포는 사람 승인 아래 폐장 후에 한다.
