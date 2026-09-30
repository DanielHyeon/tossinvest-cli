# 26라운드 보이스 A (동시성 · 잠금 순서) — 원문

- 검토 대상: `15cb8540` 읽기 전용 사본. `071e67c1`(codex #2 · #3 · #5 수리) · `d8769cfb`(codex P0 · #4 · #6 수리) **이전**이다 — 판정 표의 #2 · #3 · #5 는 `071e67c1` 에서 이미 닫힌 항목과 겹친다(review §24.11 대조).
- 실행: 사본에서만(`set -euo pipefail`, `git -C` rev-parse 실패 단언 뒤), 끝나고 사본 · GOCACHE · TMPDIR 삭제.

**판정: APPROVE** — P0 · 안전 불변식 위반 없음. P1 하나(mode-release 통지 유실)는 병합 전 수리 권고.

- `-race`: obs 전체, execgw · journal(`A092|A096|A099|A124|Mode|Relax`), app/engine(`A092|A124|A098|Mode|Relax|Auxiliary|Runtime`) 모두 ok, 경합 0.

## 잠금 순서 (코드로 확인)

```
n.mu ─▶ journal 연결(SetMaxOpenConns(1), BEGIN IMMEDIATE) ─해제─▶ g.mu(Block/Clear/BlockUnlessClearedSince)
      (recordCritical · claimAndDeliver 의 claim · Acknowledge — 잠금 안은 로컬 원장 + 게이트뿐, 원격 전송 없음)
TransitionOperatingMode: 연결(tx + audit 파일 쓰기) ─commit·해제─▶ j.modeMu(R, 참조만) ─▶ g.mu(Project) ─해제─▶
      announcer(RecordOnly): n.mu ─▶ 연결 ─▶ [실패 시 g.mu] ─n.mu 해제─▶ escalate: 연결 ─▶ g.mu
CheckEntryFor: g.mu(참조만) ─해제─▶ authorityRefresh(연결) ─▶ g.mu
```

- 역방향 간선 없음(g.mu 를 쥐고 n.mu · 연결을 잡는 경로, 연결을 쥐고 n.mu 를 잡는 경로 없음; apply hook 에 notifier 없음) → 교착 없음.
- n.mu 를 쥔 채 원격 전송하는 것은 `Notifier.Flush`(notifier.go:830-907) 하나 — 생산 호출자 0, `TestA092TheNotifierFlushHasNoProductionCaller` 가 고정.

## 발견

| # | 등급 | 파일:줄 | 무엇이 틀렸나 | 근거 | 제안 |
|---|---|---|---|---|---|
| 1 | P1 | `internal/app/engine/modeops.go:93-101,116`; `mode_control_transport_unix.go:156`; `internal/obs/record_only.go:130,146` | 사람의 완화가 커밋된 뒤 요청 문맥(`r.Context()`)이 끝나면 완화 통지가 영구히 사라짐 — 통지 기록 · 모드 승격 둘 다 죽은 ctx 로 시도해 실패, in-memory 래치만 남고 `alerts ack` 한 번이면 풀리며 재시작하면 흔적 없음. 같은 호출의 재읽기도 같은 ctx 라 실패해 커밋된 완화를 오류(500)로 보고. 델타 「완화도 운영자에게 통지되어야 한다(SHALL)」·「커밋된 완화를 실패로 보고하면…(SHALL)」 위반. a066 `notifyRelaxation`(risk_relaxation_command.go:178)은 이 이유로 `context.WithoutCancel` 을 쓰는데 모드 완화만 빠짐 — `releaseCtx` 주석(notifier.go:760-768)이 경고한 반패턴. | 사본 탐침: 커밋 뒤 · 통지 전 cancel 하는 announcer 래퍼로 `fx.ops.Release` → `NotifyError: …context canceled`, `err = re-reading the operating mode…: context canceled`, 원장 `NORMAL (seq 2)`, `pending outbox rows = 0`, `gate blocks = map[critical_alert_undelivered:…]` → `n.Acknowledge` 뒤 `gate blocks after ack = map[]`. 창은 통지의 BeginTx 가 단일 연결을 기다리는 동안 — 클라이언트 Ctrl-C · 30s 타임아웃이면 열림. | `Release` 안의 전이 · 재읽기를 `context.WithoutCancel(ctx)` 로. 일반적으로 recordCritical · notifyCritical 의 `escalate` 도 기록 실패 원인이 ctx 종료일 때 떨어진 ctx 로 시도(releaseCtx 규칙). RED 는 위 탐침 모양. |
| 2 | P2 | `internal/obs/normal_relay.go:47-62`; `internal/app/engine/auxiliary.go:130-137` | 주장 7 「실행자 정지 뒤 남은 것」 거짓 — 이관 실행자가 패닉으로 죽으면 발행 중 1건과 그 뒤 버퍼(64)에 들어가는 Offer 가 기록 없이 사라짐(버퍼가 찬 뒤의 것만 기록). | 사본 탐침(cap 4, panic publisher): 7건 중 드롭 줄 2건(`e`, `f`)뿐, `first` + `a..d` 5건 무기록. | 멈추면 닫힘 상태(이후 Offer 즉시 드롭 기록), 패닉 복구 시 잔여 기록, 발행 중 1건 기록. |
| 3 | P2 | `internal/app/engine/runtime.go:340-343`; `normal_relay.go:50-55,65-74` | 종료 배수 불완전 — `cancel()` 뒤 exit 루프와 이관이 동시에 내려가 배수 뒤 exit 의 마지막 Offer 가 버퍼에 남아 무기록. | 사본 탐침: 취소된 ctx 로 `Run` 반환 뒤 `Offer("late")` → drop lines = 0. 생산 빈도는 추측. | #2 와 같은 닫힘 플래그, 또는 wg.Wait 뒤 잔여 드롭 기록. |
| 4 | P2 | `internal/obs/normal_relay.go:81-84`; `record_only.go:88-90,113-115` | 새 버림 기록 줄이 `FieldEvent`(`"event"`)를 다시 넘겨 JSON 키 중복 — 리더는 마지막 값을 쓰므로 줄의 event 가 원래 사건(`exit.proposal_capped`)으로 읽히고 `event==engine.normal_alert_dropped` 규칙은 안 걸림. log.go:31-40 주석이 이미 「버그」로 적은 결함, K13 근거 무효화. | 탐침 로그 원문: `{"msg":"engine.normal_alert_dropped","event":"engine.normal_alert_dropped",…,"event":"exit.proposal_capped","alert_key":"e",…}` | `FieldTriggerEvent` 로. 같은 모양의 기존 줄(notifier.go:200)도. |
| 5 | P2 | `internal/obs/notifier.go:180-190`(NormalRelay 재사용) | 이관의 발행 실패가 `EventAlertUndelivered`(K13 이 critical 실행자 것이라 한 타입)로 나감; Publisher nil 이면 무기록. | 코드 인용 `if n.Publisher == nil { return }` / `n.Log.Warn(EventAlertUndelivered, …)` | 이관 전용 발행 함수(실패 · nil 발행기를 `EventNormalAlertDropped` 로). |
| 6 | P2 | `internal/app/engine/exitwiring.go:352-354` | Floor 만 `if opts.Floor == nil` 존재 검사 — Alerts · Announcer · Retrier 는 25라운드 B#4 로 무조건 덮는데 Floor 빠짐. 지금은 cmd 핀(`a092_exit_options_pin_test.go`)이 생산 리터럴의 Floor 키를 금지해 막음. | 코드 인용 | 무조건 `exitSideFloor(…)` 로 감싸거나 넘긴 Floor 의 retrier 교체. |
| 7 | P2(추측 포함) | `internal/app/engine/a092_structure_pins_test.go:452-510` | k3 AST 핀은 `X.official.M(...)` 호출만 셈 — 메서드 값 별칭 · 헬퍼 이동은 못 봄. 현재 코드엔 우회 없음(확인). | 핀 코드 | 본문 안 `official` 선택자 참조 전부를 세어 Query 리터럴 안에만 있게. |
| 8 | P2 | `internal/obs/notifier.go:947-973` | Acknowledge 가 n.mu 아래 backlog N 건을 행마다 autocommit(fsync N 회) — 그동안 exit 의 critical 기록이 n.mu 대기. 로컬이라 델타 문언 안이나 N 비례. | 코드 인용 | 필요하면 한 트랜잭션으로. |

## 저자 주장 판정

1. **참** — 합성 섬. exit Retrier 는 설정값뿐인 값 복사; `ConfirmedFloor` 의 `f.official.*` 두 호출은 `f.retrier.Query` 리터럴 안; 401 경로는 `r.Announcer`(RecordOnly) 하나. 남은 틈은 #7 · #6(둘 다 현재 막힘).
2. **참** — exit goroutine 도달 알림 경로 전부 RecordOnly; n.mu 보유 원격 전송은 생산 호출자 없는 Flush 뿐.
3. **참** — PENDING 생성 경로 셋(RecordAlert · ClaimAlertForDelivery 는 n.mu 아래, EnqueueAlert 는 parkAlert 만); 래치는 근거 확정 뒤 세대로 조건부, 틀리는 방향은 과잠금뿐(재승인으로 풀림). 단 #1 처럼 기록 자체가 실패하면 행 없이 래치만 서고 승인이 그것을 엶.
4. **참** — 교체는 g.mu 한 번 안; 울타리 `rec.Seq <= g.modeSeq`; rowid 는 삭제 · 수정 없는 표 + BEGIN IMMEDIATE 로 커밋 순서; 복원은 투영기 먼저 묶고 최신 rowid; 벽시계 제거. (투영기가 AccountRef 를 안 보는 것은 단일 계좌 프로세스 가정 — 기존 설계.)
5. **부분 참** — 원장 판정 우회 없음(계정 고정, OPERATOR · 승인 · audit 타입 있는 nil 거절). 다만 ctx 취소면 재읽기 실패로 커밋된 완화를 오류 보고(#1).
6. **참** — 보조 실행자 패닉 recover, 기동 실패는 강등, 핸들러 패닉은 net/http 회수. 일반 등급 무기록 유실(#2)은 있으나 루프는 안 멈춤.
7. **거짓** — 패닉 뒤(#2) · 배수 뒤(#3) 무기록, 발행 실패는 다른 타입 · nil 발행기 무기록(#5), 버림 줄 event 키 가림(#4).

(a) 이중 발송 없음 · (b) 게이트 오개방 없음(#1 은 행이 없는 경우) · (c) 영구 잠김 없음 · (d) 교착 없음 · (e) 모드 투영 역행 없음.

Recommendation: APPROVE, 단 #1 을 병합 전 RED→GREEN 으로 닫고 #2~#5 는 C8 후속으로 묶을 것.
