# Evidence reconciliation — a131

- 기준: base `233d5d9c`. 확인한 파일·심볼: `internal/obs/log.go`(`NewLogger` · `Logger.emit`) · `internal/obs/notifier.go`(`claimed_by` · `notificationFor`) ·
  `internal/official/reads.go`(`accountsLocked`) · `internal/app/engine/interlock.go`(`resolveAccountRef` · 게이트 수락 audit `:418` · `:441`) ·
  `internal/app/engine/{reconcileloop,runtime,exitloop,exitwiring,adoption}.go`(FieldAccount 자리) · `cmd/tossctl/engine.go`(`runEngineRun`) · `cmd/tossctl/main.go` ·
  `cmd/tossctl/engineproc.go`(`spawnDetachedEngine` — engine.log = 자식 stdout/stderr) · `tools/engine-autostart.sh:121`(같음) · `cmd/tossctl/soakproc.go` ·
  `internal/audit/audit.go` · `internal/soak/{record,summary,attest}.go` · `internal/verifylive/{record,reconcile,reconcile_record,reconcile_check,endpoints}.go` ·
  `internal/attest/attest.go`(`SameAccountMasked` · `Mask` · `sameAccount` · `Verify`) · `cmd/tossctl/soak.go`(`runSoakAttest` · `supervisedProofs`) · `internal/journal/position_projection.go`(`PositionID`).

| # | 불일치 | 현재 HEAD 로 내린 결론 |
|---|---|---|
| R1 | `callers Record` 이 이름 충돌로 11 자리 | audit 직접 호출은 `RecordChange`·`RecordAction` 둘(HEAD grep) — 치환은 `Record` 한 곳 |
| R2 | CodeGraph 는 `obs.FieldAccount` 를 **값으로** 쓰는 자리를 호출 관계로 안 보임 | go/parser census(`analysis/census/census-base.tsv`) + `git grep FieldAccount` 20 자리 — design D1 L1 |
| R3 | `affected` 기본 실행 시험 0~1 | 알려진 판별식 결함(CLAUDE.md 하네스 주석). `--filter` 결과를 씀 |
| R4 | CodeGraphContext 색인 0 | advisory — census 와 HEAD grep 으로 대신. sdd-sync 뒤 재확인 |
| R5 | engine.log 생산자 | CodeGraph 질의 없이 HEAD 로: `spawnDetachedEngine`·`engine-autostart.sh` 가 자식 stdout/stderr 를 파일에 붙임 → 계좌 제거는 **자식 안**(로거 핸들러 · `runEngineRun` Writer · `main` 오류 줄)에서 해야 함. 부모 파이프 방식은 부모가 죽으면 엔진이 SIGPIPE 로 죽어 기각 |
| R6 | 운영 파일 실측 vs census | 실측 키(`account` · `claimed_by` · `error` · 비JSON 69 · `pos-` 3,405)가 census 자리(L1·L2·L3·L5·L4)와 전부 대응. census `print` 12 중 로그 싱크로 가는 것은 `runEngineRun`·`runSoakAttest` 둘(나머지는 대화형 출력·다이제스트 입력) |
