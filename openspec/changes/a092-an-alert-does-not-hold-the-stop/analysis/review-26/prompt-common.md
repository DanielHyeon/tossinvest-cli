# a092 26라운드 — 구현 리뷰 (착지 단위 ② ~ ⑤ + 25.6 전체)

읽기 전용 리뷰다. **파일을 만들거나 고치지 마라.** git 상태를 바꾸는 명령(commit·checkout·stash·reset·add·init·config)을 실행하지 마라.
LIVE 주문·운영 토글·엔진 기동·`mutating: true` 명령(`tossctl engine mode-release` · `alerts ack` 포함) 실행 금지. 운영 원장(`~/.config/tossctl/`)과 자격 증명 파일(`~/.codex` 포함)은 열지 마라.
시험을 돌릴 때는 대상 트리의 **사본**에서만(아래 트리를 `/tmp/claude-1000/a092-r26-<보이스>-<pid>` 로 복사, 만들기 전 `set -euo pipefail`, 사본에서 `git -C <사본> rev-parse` 가 **실패**함을 단언 — 실제 저장소가 아님). `GOCACHE`/`TMPDIR` 은 `/tmp/claude-1000/` 아래. journal 패키지 전체 판은 `-timeout 40m`(기본 600s 를 넘는 스위트). 끝나면 사본을 지워라.
대상 트리: `/tmp/claude-1000/a092-r26-tree`(git archive, git 저장소 아님). `openspec/changes/a092-an-alert-does-not-hold-the-stop/review.md` 와 `analysis/review-*/` 의 다른 리뷰 원문은 읽지 마라(독립성).

## 스펙

frozen 24판: `openspec/changes/a092-an-alert-does-not-hold-the-stop/specs/engine-safety/spec.md`(델타), `design.md` D0.3e~D0.3i, `tasks.md` §21~§25.
정본: `openspec/specs/engine-safety/spec.md`, `openspec/specs/risk-management/spec.md`.

## 구현 (base `721d0338` 이후, 코드 커밋)

- ② `c6e2e3ac`: `Journal.RecordAlert`, `obs.RecordOnly`, 모드 통지 키 `rec.ID`, exit 주입 지점별 기록 전용 배선.
- 25.6 `0e4f26af`: a066 완화 통지 → `Notifier.RecordCritical`. 로그 줄 필드 제거 `f48e7865`.
- ③ `fbc6df5f` + 25라운드 수리 `55963f29`: `claimAndDeliver` 잠금 범위, `deliver` 원칙 E(반납 행 없음 포함), `judge`, `escalate` 반환값, `logClaimHeld` INFO.
- ④ `2714e393` · `c9c93a7b`: 모드 커밋 순서(rowid) · `ProjectOperatingMode` 원자 교체 + 울타리 · `buildGateway` 투영 배선 + 복원 · `tossctl engine mode-release` + 모드 전용 소켓 · `alerts ack` mutating · a124 경계 핀 갱신.
- ⑤ `e55102f0`: `obs.NormalRelay`(exit 일반 등급 유계 이관) · 런타임 보조 실행자 · `AuxiliaryExecutor.StopEvent` · `ExitObserver` 무조건 기록 전용 · `ErrModeAnnouncementFailed` 호출자 둘 · 인수 줄 `claim_stolen` · 구조 핀 파일 `internal/app/engine/a092_structure_pins_test.go`.
- 변이 하네스 · 원장: `analysis/harness/mutate_unit2.py`, `analysis/mutation-*/`. FLM: `analysis/function-logic/`.

## 저자가 스스로 세운 주장 (공격 대상)

1. **k3(청산 수량 상한 조회의 401 강화 통지)는 행동 시험이 없다 — 「floor 가 exit Retrier 를 쓴다(배선 시험)」·「Retrier 401 경로가 기록만 한다(k2 행동 시험)」·「ConfirmedFloor 의 브로커 호출은 전부 retrier.Query 클로저 안(AST 핀)」의 합성으로만 선다.** 합성의 빈틈을 코드로 찾아라.
2. exit 관측 goroutine 에서 원격 전송에 닿는 도달 경로가 **하나도** 없다(critical · 모드 통지 · 일반 등급 모두). 동기 전송을 기다리는 잠금에도 줄 서지 않는다.
3. 운영자 승인(Acknowledge)이 어느 순간에 끼어들어도 게이트 오개방(미전달 critical 행이 있는데 전달 실패 사유 해제)이 없다. 입구 밖 기록자는 `parkAlert` 하나다.
4. 모드 투영: 겹친 전이의 역순 도착 · 교체 중 점검 · 벽시계 되감김 · 재시작 복원 어디서도 산 게이트가 커밋 순서의 최신보다 **덜 보수적인** 모드를 보지 않는다.
5. `mode-release` 는 원장 판정(OPERATOR · 승인 · audit 가 commit 앞)을 우회할 경로가 없고, 결과는 다시 읽은 값이며, 원장만 고친 완화는 산 게이트를 못 연다.
6. 새 보조 실행자와 모드 제어 엔드포인트의 어떤 실패도 엔진(손절 루프)을 멈추지 않는다.
7. **exit 일반 등급 이관(C8)**: `NormalRelay` 의 버림 기록(`EventNormalAlertDropped`, 유형 + 키)이 델타 「무엇을 버렸는지는 기록되어야 한다(SHALL — 기록 없는 유실은 전송된 것과 구별되지 않는다)」를 **모든** 버림 경로(가득 참 · 이관 없음 · 종료 배수 · 실행자 정지 뒤 남은 것 · 발행 실패)에서 만족한다. 버퍼 64 유계 · 비차단 넘김은 「이관은 유실을 허용하는 형태여도 된다」 문장 안이다.
