# a198 착지 뒤 merge · base 재고정 · 번들 재확인 (2026-11-24)

- merge: origin/main `7fec501f1`(a198 archive) → 병합 커밋 `603046a23`. 충돌은 `docs/pm/generated/` 2파일뿐(재생성 · `--check` 통과), Go 충돌 0.
- base 재고정: `e4165fa3d` → `603046a23`, 단독 커밋 `ae7764bc2`(영수증은 그 커밋 메시지). 귀속 실측 — 옛 base 이후 이 change 디렉터리를 만진 비병합 커밋 5개의 `.go` 변경 0.
- 옛 base 에서의 `check_analysis` 원문: `check-old-base.txt` — required 6, 전부 a198 몫(`ExitObserver.endExitOrderCycle` · `ExitObserver.clearTheSymbol` · `inDoubtSettle.sweep` · `Gateway.withCancelRefusal` · `Recovery.confirmAcked` · 시험 1).
- 새 base 에서: required 0, 「evidence complete or diff-proven exempt」.
- **a198 은 `Context.ExitObserver` 를 편집하지 않았다**(1단계의 겹침 예상은 a198 design D2/D3 의 안이었고, 착지한 구현은 `ExitObserver.clearTheSymbol` 쪽). 편집 대상 파일 전부(`interlock.go` · `engine.go` · `exitwiring.go` · `reconcileloop.go` · `attestation_watch.go` · `strategy_entry_supervisor.go` · engine `gateway.go` · execgw `failclosed.go` · `retry.go` · `gateway.go` · `reason.go` · cmd `engine.go` · `soak.go` · `attest.go` · a166 testdata)의 `git diff e4165fa3d 603046a23` 가 비어 있음 → pre-edit 번들 20개 · 사전 증거 줄 번호 · AST 전부 그대로 유효(달라진 곳 0).
