# a101 리뷰 기록

## 이 change가 어디서 나왔나

a100의 조건주문 probe를 배포하다 나왔다. 문서화된 절차대로 컨테이너를 재생성했더니
**capability soak이 죽은 채로 남았다**(2026-08-11 14:00Z 실측). a100 tasks 0.10b에 결함으로
적고 별도 change로 분리했다 — a100은 브로커 상주 보호이고 이것은 콘솔의 프로세스 수명주기다.

## 확인한 것과 그것이 설계를 바꾼 지점

### F1. a060이 폐기한 것을 되돌리는 게 아니다 — 아무도 물려받지 않은 절반이다

`soakproc.go:10-13`이 "`soak-autostart.sh`… 2026-08-03 (a060 I3)에 은퇴"라고 적고 있어
**먼저 그 폐기 근거를 읽었다.** archive `issues.md` I3의 근거는 "콘솔이 컨테이너로 옮겨가며
이 스크립트의 역할은 끝났다"였다.

**그 전제에 구멍이 있었다.** 콘솔은 *수동 재시작*은 물려받았지만(`RestartSoak` seam)
**"떠 있지 않으면 띄운다"는 감시자 역할은 물려받지 않았다.** a060의 판단은 옳았다 — 그것은
이미 만료된 죽은 산출물이었고, 플래그 없이 자식을 띄워 **기본 프로필에 두 번째 서베이**를
만들 수 있는 위험이 실제로 있었다.

⇒ 이 change는 그 위험을 재현하지 않는다. 콘솔이 이미 가진 `restartSoak` seam을 그대로 쓰므로
프로필 플래그(`soakArgs`)가 붙고 소유 판정(`soakFindProcesses(recordPath)`)이 이 기록에 붙은
서베이만 대상으로 한다. **`operator-console` spec:918이 이미 그 판정을 요구하고 있고, 그
조문은 "기본 경로를 명시한 콘솔과 생략한 autostart는 같은 인스턴스다"라며 autostart를 이미
상정한다.**

### F2. `runConsole`이 0.0%다 — 이 측정이 편집 형태를 결정했다

편집 전 측정: `cmd/tossctl` 49.6%, **`runConsole` 0.0%, 분기 41개 전부 미실행.**

콘솔 조립 함수는 어느 테스트도 부르지 않는다. 실제 포트와 브로커 client와 엔진 control
plane이 필요하기 때문이다. ⇒ **그 안에 판정을 넣으면 그 판정은 검증 불가능한 자리에 놓인다.**

그래서 `runConfiguredEngineAutostart`의 형태를 그대로 따랐다 — 그 함수의 주석이 이유를 이미
적어 뒀다: *"It receives functions so tests can prove every branch without constructing a
broker or spawning a process."* 판정은 전부 `runConfiguredSoakAutostart`(100.0%)와
`rememberSoakApproval`(100.0%)에 있고, `runConsole`에는 호출·출력·nil 검사만 들어갔다
(분기 41 → 44).

### F3. 승인 실패가 재시작을 되돌리면 안 된다

`rememberSoakApproval`을 별도 함수로 뽑은 이유다. 승인 기록 시점에 **서베이는 이미 돌고
있다.** 기록 실패를 재시작 실패로 보고하면 대시보드가 "재시작 실패"를 표시하고, 운영자의
자연스러운 반응(다시 누르기)이 **방금 선 서베이를 죽인다.** 실패는 반환 문자열에 덧붙인다.
`TestSoakApprovalFailureDoesNotUndoTheRestart`가 이것을 고정한다.

### F4. 새 화면을 만들지 않았다 (YAGNI)

엔진 autostart에는 설정 화면이 있다. soak에는 만들지 않았다 — **콘솔에는 애초에 soak 정지
버튼이 없다.** 시작만 가능한 표면에 중지 승인 UI를 새로 파는 것은 지금 필요 없는 일이고,
승인 기록은 이미 있는 재시작 버튼이 남긴다. 끄려면 config를 고친다.

## 이탈 (`not-applicable`이 아니라 이탈이다 — 숨기지 않는다)

- **`internal/config`의 RED을 관측하지 않았다.** `soak_io.go`를 먼저 쓰고 `soak_io_test.go`를
  나중에 썼다. 테스트 자체는 스펙에서 유도했고 11건 전부 통과하지만, **실패하는 것을 본 적이
  없으므로 그 테스트들이 무엇을 잡는지는 증명되지 않았다.** `cmd/tossctl` 쪽은 순서를
  지켰고 RED을 컴파일 실패로 관측했다(`undefined: runConfiguredSoakAutostart`).
- **Pre-Edit 통과 조건 (1)·(2)를 테스트로 고정하지 못했다.** `runConsole` 0.0%가 이유이며
  `pre-edit-gate.md`와 `branch-test-map.md`에 각각 적었다. 코드 리뷰 조건으로만 존재한다.
- ~~**`check_analysis --change a101`이 `internal/soak.Runner.RunCycle`을 미매핑으로
  보고한다.**~~ **해소(2026-08-12).** 원인은 진단대로 stacked change였다 — a101의 base가
  a100의 *문서* 커밋 `ce78b0db`였고 a100의 Go 변경이 미커밋이라 a101의 diff로 새어
  들어왔다. a100을 `4c6927ea`로 커밋하고 `base-commit.txt`를 그 커밋으로 다시 고정하니
  `evidence complete or diff-proven exempt`가 나온다.

  **재고정은 커밋 뒤에만 가능하다.** `capture_change_base.py`는 이미 잡힌 base를
  덮어쓰지 않으므로(`change base already captured`) 파일을 직접 고쳤다. 쌓은 change의
  base는 "내가 시작한 시점의 HEAD"가 아니라 **"내 아래 change가 커밋된 지점"**이다.

## 남은 위험

1. **첫 배포에서는 아직 켜져 있지 않다.** 키가 없으므로 false다. 운영자가 soak 재시작을 한 번
   눌러야 승인이 기록되고, **그 다음 배포부터** 자동으로 살아난다. 이번 배포 자체는
   이 결함을 고치지 못한다 — 고쳐지는 것은 그 다음부터다.
2. **`runConsole`은 여전히 0.0%다.** 이 change는 그 사실을 우회했을 뿐 고치지 않았다.
   콘솔 조립을 테스트 가능하게 만드는 일은 별도 change다.
3. **autostart가 엔진과 같은 rate budget을 쓴다.** 순서(엔진 다음)로만 다루었고, 두 프로세스가
   기동 직후 동시에 read를 내는 구간은 측정하지 않았다. soak의 `PauseWhile`은 verify에만
   양보하고 엔진에는 양보하지 않는다.
