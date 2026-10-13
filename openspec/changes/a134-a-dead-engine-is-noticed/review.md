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
