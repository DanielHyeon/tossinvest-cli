# a134 evidence reconciliation

기준 commit `421bf53c` · codegraph 1.6.0 · HEAD 직접 읽기.

| # | 주장 | CodeGraph | HEAD | 판정 |
|---|---|---|---|---|
| R1 | `obs.Heartbeat` 생산 생성자 0 | query 결과 정의만 | `grep -rn "obs.Heartbeat{" cmd internal` 비시험 0 · 시험 `obs_test.go` 3 | 일치 |
| R2 | 엔진이 1분마다 원자적으로 쓰는 생존 파일이 이미 있음 | `MarkerPath` 호출자에 `runEngineRun` | `enginelock.Hold` 의 ticker(`RefreshEvery` = 1m) · `write` 가 임시 파일 + `Chtimes` + rename · `Release` 가 지움 | 일치 |
| R3 | 마커에 계좌 칸 없음 | — | `enginelock.Marker` 칸: pid · started_at · refreshed_at · binary · note · ready_at · proc_instance | 일치 — 시험이 고정(design D5) |
| R4 | ntfy 해석 규칙의 생산 호출자는 하나 | `NewContext` 하나 | engine.go:463 | 일치 |
| R5 | 「ntfy 가 박동 부재를 스스로 경보」 | — | 저장소 안 ntfy 문서(`docs/operations.md` 643 · `AGENTS.md` 77 · README)에 수신 측 경보 기능 언급 0 | **주석 · spec 문장이 근거 없음** — design D0 |
| R6 | 하위 명령 등록이 바꾸는 것 | `impact newEngineCmd` 38 = 명령 트리 시험 | leaf 하나 추가 | 일치 — 주석 규약 시험(help_convention_test.go)이 새 leaf 를 자동 포함 |

불일치는 R5 하나이고 코드가 아닌 문서의 전제다. 편집 차단 사유 없음.
