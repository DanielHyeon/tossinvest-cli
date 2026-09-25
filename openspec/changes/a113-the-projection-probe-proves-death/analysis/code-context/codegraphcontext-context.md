# CodeGraphContext supporting context — a113

- 날짜: 2026-09-25 · `codegraphcontext find name projectionSocketAccepts` / `reclaimStaleControlDirectory`
  → **"No code elements found"** (rc 0). `codegraphcontext list` 에는 이 워크트리
  (`/mnt/D/Axipient/workspace/TossOS`)가 Project 로 등록돼 있으나 두 Go 심볼을 돌려주지 않는다.
- 판정: 보조 문맥 **없음**(not-applicable — 색인이 이 심볼을 모른다). 편집 권한 근거로 쓰지 않는다.
- GBrain: `python3 tools/sdd/gbrain_project.py search "projectionSocketAccepts"` → exit 75
  `[gbrain-project] busy: owner pid=1482463` — 다른 세션이 소유(WORKFLOW 의 busy 예외). not-applicable.
- 대체 보조 문맥(수동, HEAD grep): 패키지 import 자 7 파일 — `cmd/tossctl/{console.go, engine.go,
  httpapi.go}` + cmd 테스트 3 + `internal/console/strategy_runtime_integration_unix_test.go`.
  소비자는 전부 `Dial`/`Start` 를 거친다.
- 선례: `internal/positionpolicyrpc/private_staging_unix.go` `privateSocketAccepts`(a109 §1-fix F1·G7)
  — 이식 원형.
