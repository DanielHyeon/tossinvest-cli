# a134 CodeGraph baseline

- 도구: codegraph **1.6.0**(`codegraph --version`), 기준 commit `421bf53c`(base-commit.txt), 원문은 `raw/`.
- 질의와 결론:

| 질의 | 결과(raw) | 결론 |
|---|---|---|
| `callers newEngineCmd` | 4 — `newRootCmd`(root.go:52) + 시험 3 | 하위 명령 등록 한 줄의 생산 호출자는 루트 트리 하나. 기존 시험은 트리를 걸으므로 새 leaf 의 주석 규약(`source` · `mutating`) 시험이 자동으로 새 명령을 덮음 |
| `impact newEngineCmd` | 38 기호 | 명령 트리 시험군. 동작 변화 = leaf 하나 추가뿐 |
| `callers resolveNotificationPublisher` | 13 — 생산은 `NewContext`(engine.go:463) 하나 + a074 시험 12 | 해석 규칙의 생산 호출자는 엔진 조립 하나. 감시자가 같은 함수를 부르도록 내보내기 입구를 새 파일에 둠(규칙 사본 없음) |
| `callers MarkerPath` | 20 — 생산 `startEngine` · `stopEngine` · `candidateEngineProbe` · `runEngineRun` 등 + a102/a111 시험 | 마커 경로의 철자는 `enginelock.MarkerPath` 하나. 감시자도 이것을 씀 |
| `query Heartbeat` | `obs.Heartbeat` 구조체(heartbeat.go:41) — 생산 생성자 0 | 같은 결론을 `grep -rn "obs.Heartbeat{" cmd internal`(비시험 0)이 확인 |
| `affected cmd/tossctl/engine.go internal/obs/heartbeat.go --filter '*_test.go'` | 1070 시험 파일 | 기본 판별식은 Go 시험을 못 찾음(.claude/CLAUDE.md 하네스 주석) — `--filter` 로 잼. 넓이는 cmd/tossctl · obs 의존 폐포 |
