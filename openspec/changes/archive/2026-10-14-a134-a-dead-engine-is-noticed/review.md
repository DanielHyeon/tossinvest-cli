# a134 review

## Pre-Edit Gate (2026-10-14, 구현 팀메이트 Opus)

```text
Pre-Edit Gate:
- change id / task id: a134-a-dead-engine-is-noticed / 1.2
- 대상 심볼(패키지.함수): main.newEngineCmd(cmd/tossctl/engine.go — AddCommand 한 문장 추가). 그 밖은 새 기호:
  enginewatch.{Observe, Check.Run, loadState, saveState, lostNotification, restoredNotification}(새 패키지),
  engine.NotificationPublisher(새 파일 a134_notification_publisher.go), main.newEngineHeartbeatCheckCmd · runEngineHeartbeatCheck ·
  describeHeartbeatOutcome(새 파일 engine_heartbeat_check.go). internal/obs/heartbeat.go 는 파일 · 타입 주석만(함수 무편집).
- CodeGraph definition/callers/callees/impact: analysis/code-context/codegraph-baseline.md (codegraph 1.6.0, raw/) —
  newEngineCmd 생산 호출 1(newRootCmd) · impact 38 · resolveNotificationPublisher 생산 호출 1(NewContext) · MarkerPath 호출 20 ·
  affected --filter 1070
- CodeGraphContext 후보와 evidence reconciliation: analysis/code-context/{codegraphcontext-context,evidence-reconciliation}.md (R1~R6, 불일치 R5 = 문서 전제)
- 기존 동작 파악 근거: base 421bf53c HEAD 직접 읽기 — enginelock.Hold/write/Release(마커 원자 쓰기 · 1분 갱신 · 정상 정지 삭제) ·
  runEngineRun 6단계 · AuxiliaryExecutor 계약 · resolveNotificationPublisher · obs.Ntfy.Publish(a133 경계) · a133 구조 핀 ①⑤ ·
  help_convention_test.go(표지 규약)
- Function Logic Map / Branch Test Map: analysis/function-logic/cmd-tossctl--newenginecmd/(분기 0, 편집 전 sha 446e27a3… · 편집 뒤 f311f5ad…)
- upstream 상속 테스트 영향: no — 새 패키지 · 새 파일. newEngineCmd 는 TossOS 명령 묶음. cmd/tossctl 전체 go test 로 확인(receipts)
- 실패 테스트 선행 작성: yes — analysis/red/(새 패키지라 RED = 빌드 실패: 기호 미정의)
- 설정·DB·journal 변경과 rollback: 없음 — 새 설정 키 0 · 원장 스키마 0 · 엔진이 읽는 파일 0. 감시자 상태 파일은 감시자만 읽음.
  롤백 = 호스트 단위 비활성(`systemctl --user disable --now tossos-heartbeat.timer`) + 상태 파일 삭제
- 안전 불변식 §0 위반 여부 검토: 통과 — 엔진 프로세스 코드 무편집(런타임 · 루프 · 보조 실행자 · 마커 무변경), 주문 · 손절 · 청산 ·
  대사 · 체결 · Guardian · outbox 무편집. 감시 명령은 저널 · 엔진 잠금 · 브로커 · 세션을 열지 않음(시험 고정). mutating=false 이고
  쓰는 것은 자기 상태 파일과 알림 한 건. 실 ntfy 송신 시험 0(Publisher 주입 · httptest)
```

## Function Logic Map 적용 범위

- 적용: `cmd/tossctl/engine.go:newEngineCmd` — 게이트가 계산하는 수정된 기존 함수 하나. 번들 `analysis/function-logic/cmd-tossctl--newenginecmd/`.
- Function Logic Map: not-applicable — `internal/enginewatch/*`(새 leaf 패키지), `cmd/tossctl/engine_heartbeat_check.go`(새 파일),
  `internal/app/engine/a134_notification_publisher.go`(새 파일 — 기존 `resolveNotificationPublisher` 를 부르기만 함, 그 함수 무편집),
  `internal/obs/heartbeat.go`(파일 · 타입 주석만 — 함수 범위와 겹치는 줄 0). 모두 base 에 같은 qualified 함수가 없는 새 기호이거나 함수 밖 주석.
- 비례 원칙(docs/WORKFLOW.md): 알림 · 생존 관측 경로이며 High-risk 경로 · 게이트 판정을 바꾸지 않음 → 변이 원장 · 다중 적대 리뷰 비의무.
  경량 변이 4 를 receipts 에 남김(4/4 CAUGHT).

## 공유 체크아웃 메모

- `cmd/tossctl/engine.go` 를 병행 change a135 가 같은 시간대에 편집했다(엔진 조립 `engineRuntime` 의 `Auxiliary` 줄). a134 편집 시점에는 그 파일이
  깨끗했고(`git status --short` 확인), 커밋 직전 diff 에는 a134 hunk 하나만 있음을 다시 확인한 뒤 커밋했다.
- `cmd/tossctl` 전체 시험의 실패 1(`TestProductionRuntimeStartsExactlyOneAlertDeliverer`)은 a135 의 미커밋 시험 편집 원인 — receipts.

## 구현 독립 리뷰 (2026-10-14, 분리된 컨텍스트 — code-reviewer 서브에이전트, 85752cd6 대상, 버리는 worktree 에서 시험 · 변이, 공유 체크아웃 무편집 · 네트워크 0)

판정: **APPROVE-WITH-FIXES** — P0 0 · P1 1 · P2 3 · P3 5. 확인: 판정표 D2 = 코드 · 중복 억제 D3(전송 성공 뒤 이동 · 손상 상태는 억제 안 함 · oneshot 은
겹치지 않음) · 전송 D4(규칙 사본 0 · `*obs.Ntfy` → a133 경계) · 계좌 칸 0 · 단위 파일 · 마커 단일 writer(`enginelock.Hold` 생산 호출 1) ·
거래 경로 무편집. 리뷰어 변이 4 중 2 생존(아래 P3-4).

| # | 등급 | 발견 | 처리 |
|---|---|---|---|
| P1-1 | P1 | 수단은 있는데 전달이 안 되는 경우(호스트 토큰 없음 · 401 · 토픽 오류 · DNS)는 상태가 안 바뀌는 동안 매 분 종료 0 — 첫 전송이 끊김 때 | **수용** — `--send-test`(`enginewatch.SendTest`, normal 1건, 실패 = 비정상 종료, 상태 무변경) + 설치 절차 4 에 단위와 같은 환경으로 필수 실행. 시험 `TestA134TheSendTestProvesDelivery` · `TestA134HeartbeatCheckSendTestReachesTheChannelAndNamesFailure`(401 거절 포함) |
| P2-1 | P2 | 복구가 마커 신선도만 봄 — 재기동 루프의 몇 초짜리 마커 · Recover 에서 멈춘 엔진이 거짓 복구 | **수용** — 끊김 → 정상 이동은 `ready_at` 필수(해석 불가 = 준비 안 됨). 시험 `TestA134RecoveryNeedsAnEngineThatFinishedItsRecovery` · 실제 `Hold/Ready/Release` 수명 `TestA134TheRealMarkerLifecycleIsWhatTheWatcherReads`. 변이 M5 CAUGHT |
| P2-2 | P2 | autostart OFF 에서 `rt.Run` 전 오류 종료 · 주 goroutine 패닉도 마커를 지워 IDLE | **수용(잔재 명시)** — design D1. 닫으려면 엔진 쪽 묘비가 필요 — 엔진 무편집 범위 밖 |
| P2-3 | P2 | 첫 마커 쓰기 실패(비활성 핸들) · 갱신 실패(디스크)는 거짓/오해 알림 | **수용(잔재 명시 + 문구)** — 크게 실패하는 쪽. 끊김 문구에 「마커 쓰기가 실패하는 중」 가능성 추가 |
| P3-1 | P3 | 호스트 suspend 뒤 거짓 끊김 · 복구 1쌍 | **수용(명시)** — design D1, 노트북 호스트에서만 |
| P3-2 | P3 | 「세션을 열지 않음」 과장 — 루트 공통 전처리가 세션 · 자격 증명 파일을 읽음 | **수용** — spec · help · 파일 주석을 「엔진 · 저널 · 엔진 잠금을 열지 않고 브로커를 부르지 않음」으로 좁힘, design D6 에 범위 |
| P3-3 | P3 | oneshot 시작 시간 상한 없음 | **수용** — `TimeoutStartSec=50s` |
| P3-4 | P3 | 시험 공백: 복구 → IDLE, cmd autostart OFF, 실제 Hold/Release 수명 | **수용** — `TestA134RecoveryToIdleSaysTheWatchEnded` · `TestA134HeartbeatCheckWithAutostartOffLeavesAStoppedEngineAlone` · 수명 시험. 리뷰어 생존 변이 둘을 M6 · M7 로 재현 → CAUGHT |
| P3-5 | P3 | 호스트 바이너리가 낡거나 없으면 단위가 매 분 실패 — 감시자 자체는 아무도 안 봄 | **수용(명시)** — design D7 의 선택적 외부 박동이 그 답. 설치 절차 4 가 바이너리를 실측 |

## 사람 승인 base 재고정 (2026-10-14)

- 승인: 사용자(2026-10-14, Manager 제시안 「1. 둘 다 승인」 선택).
- 옛 base `421bf53c` → 새 base `ecf9fb3fdeaa`(이 change 첫 Go 커밋의 부모).
- 귀속 실측: 옛 base..새 base 에서 이 change 디렉터리를 만지고 `.go` 를 고친 비병합 커밋 0.
- 옛 base 에서 required 26 함수 — 거의 전부 형제 착지 몫(a128 1.6·1.7(7c16a776·ecf9fb3f)). 새 base 격리 워크트리 실측: check_analysis rc 0, evidence complete.
