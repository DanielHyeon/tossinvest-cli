# a092 25라운드 — 구현 리뷰 (착지 단위 ② · 25.6 · ③)

읽기 전용 리뷰다. **파일을 만들거나 고치지 마라.** git 상태를 바꾸는 명령(commit·checkout·stash·reset·add·init·config)을 실행하지 마라.
LIVE 주문·운영 토글·엔진 기동·`mutating: true` 명령 금지. 운영 원장(`~/.config/tossctl/`)과 자격 증명 파일(`~/.codex` 포함)은 열지 마라.
시험을 돌려도 되지만(`go test`, 읽기 전용 트리 밖 쓰기 필요 시 `GOCACHE`/`TMPDIR` 은 `/tmp/claude-1000/` 아래) 트리의 파일은 바꾸지 마라.
대상 트리: `/tmp/claude-1000/a092-r25-tree`(git archive `22db26e7`, git 저장소 아님). `openspec/changes/a092-an-alert-does-not-hold-the-stop/review.md` 와 `analysis/review-*/` 의 다른 리뷰 원문은 읽지 마라(독립성).

## 스펙

frozen 24판: `openspec/changes/a092-an-alert-does-not-hold-the-stop/specs/engine-safety/spec.md`(델타), `design.md` D0.3e~D0.3i, `tasks.md` §21~§25.
정본: `openspec/specs/engine-safety/spec.md`, `openspec/specs/risk-management/spec.md`.

## 구현 (base `721d0338` 이후, 커밋 순)

- 단위 ② `c6e2e3ac`: `internal/journal/record_alert.go`(`Journal.RecordAlert` — 임차 없는 기록 · 기록자별 재알림 창) · `internal/obs/record_only.go`(`RecordOnly{N}` 의 `Notify` · `AnnounceOperatingMode`, `recordCritical` — `n.mu` 아래 기록, 실패 시 래치 · 승격) · `internal/obs/mode.go`(모드 통지 키 `rec.ID`) · `internal/app/engine/exitwiring.go` `Context.ExitObserver` + `exit_record_only.go`(주입 지점별 기록 전용: Alerts · Announcer · Retrier 값 복사본 · Floor 복사본) · `cmd/tossctl/engine.go` `engineRuntime`(Announcer 제거).
- 25.6 `0e4f26af`: `Notifier.RecordCritical`(공개 입구) · `internal/app/engine/risk_relaxation_command.go` `notifyRelaxation` → 입구 · `riskRelaxationRepository` 에서 `EnqueueAlert` 제거 · `position_policy_command.go` `notices`.
- 단위 ③ `fbc6df5f`: `internal/obs/notifier.go` — `claimAndDeliver` 잠금 범위(claim 만 잠금 안), `deliver` 세 래치 자리의 원칙 E(`readVerdict` · vanished 자리 `BlockUnlessClearedSince` · 나머지 둘은 판정 반환), `notifyCritical.judge`(조건부 차단 → 승격 → 승격 포함 && 실패면 무조건 차단), `escalate` 반환값, `logClaimHeld` INFO. 시험 훅 `deliveryHook`(`export_test.go`).
- 시험: `internal/journal/a092_record_alert_test.go`, `internal/obs/a092_record_only_test.go`, `internal/obs/a092_lock_scope_test.go`, `internal/app/engine/a092_*_test.go`, `cmd/tossctl/a092_exit_options_pin_test.go`, 수정된 교차 시험 `internal/obs/a096_one_send_per_condition_test.go` `TestAcknowledgeCannotClearTheGateMidSend` · `internal/obs/a099_round4_test.go` `TestAHeldRowIsNotWhispered` · `internal/app/engine/a066_risk_relaxation_test.go` · `cmd/tossctl/engine_risk_relaxation_test.go` · `cmd/tossctl/engine_test.go`.
- 변이 하네스와 원장: `openspec/changes/a092-an-alert-does-not-hold-the-stop/analysis/harness/mutate_unit2.py`, `analysis/mutation-unit2/` · `mutation-25.6/` · `mutation-unit3/`.

## 저자가 스스로 세운 주장 (공격 대상)

1. **k3(청산 수량 상한 조회의 401 강화 통지)는 따로 행동 시험이 없다 — 「floor 가 exit Retrier 를 쓴다(배선 시험 · 변이)」와 「Retrier 401 경로가 기록만 한다(k2 행동 시험)」의 합성으로만 선다.** 이 합성이 참인지, 합성의 빈틈(예: floor 가 Retrier 를 거치지 않는 조회 경로, 다른 announcer 로 가는 경로)이 있는지 코드로 판정하라.
2. 잠금을 좁힌 뒤 한 조건의 이중 발송을 막는 것은 원장 임차뿐이다(변이 L19 귀속).
3. 운영자 승인이 전송 중에 끼어들어도 게이트 오개방이 없다(`deliver` FLM 의 「Acknowledge 와의 겹침 표」: `analysis/function-logic/internal-obs--notifier.deliver/function-logic-map.md`).
4. 모드 통지 키에 `rec.ID` 를 붙여도 통지 수는 상태를 바꾼 전이 수를 넘지 않는다.
5. a066 완화 통지 이행 뒤 행 모양(키 · 제목 · 본문 · payload 바이트)이 이행 전과 같다.
6. exit goroutine 에서 알림 경로에 닿는 **모든** 도달 경로가 기록 전용이다(알림 · 관측 두절 강화 · 가격 조회 401 · 청산 상한 조회 401). 일반 등급(C8)은 단위 ⑤까지 동기 최선 발송으로 남는다(알려진 잔여).
