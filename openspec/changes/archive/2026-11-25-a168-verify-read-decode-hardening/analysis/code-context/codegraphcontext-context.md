# a168 CodeGraphContext (보조 문맥, tasks 0.3)

- `make sdd-sync`(2026-11-25, 이 worktree, real 30 m 51 s) rc=2:
  - CodeGraph: 새 색인 3,329 파일 · 57,545 노드 · 212,162 엣지, `.sdd/index-state.json` fingerprint 기록(head `e57d06cd6`).
  - CodeGraphContext: `incomplete` — `codegraphcontext stats timed out`(a164 R5 와 같은 환경 사안).
  - GBrain: `gbrain_project.py sync … --no-embed` 1800 s 시간 초과(advisory). HEAD 는 움직이지 않음(sync 전후 `e57d06cd6`).
- 그래서 CodeGraphContext 질의 결과는 **없다**(「안 봤음」 — 도구가 끝나지 않음). 누락 후보 탐색은 대신 HEAD grep(`rtk proxy grep`)으로
  네 리더·두 해독 함수·`reconcileObject`·`WithAttemptObserver` 의 모든 참조를 셌다(design §4·§2, evidence-reconciliation).
- 이 change 의 분석 파일을 추가했으므로 fingerprint 는 다시 낡았다 — 2단계 착수 때와 gate 전 `make sdd-sync` 재실행 필요.
