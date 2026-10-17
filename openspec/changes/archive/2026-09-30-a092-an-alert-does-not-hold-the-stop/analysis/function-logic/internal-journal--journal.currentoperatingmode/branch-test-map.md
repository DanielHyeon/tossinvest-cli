# Branch Test Map: `Journal.CurrentOperatingMode`

- Source: `internal/journal/operating_mode.go`; **편집 뒤** 측정 — `analysis/harness/``coverage-post-unit4-journal.json`(`./internal/journal` 시험 30개), 연결 워크트리 `2714e393`.
- 재번호: difflib 정렬(편집 전 `22db26e7` 소스 대비): B1~B1 → B1~B1(같은 분기, 줄 이동).

| Branch | AST anchor | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | if at 570:2 | `if account == ""` | `TestCurrentOperatingModeNeedsAnAccount` | 편집 전 RED: `analysis/mutation-unit4/red-*.log` · 변이 M01 | 블록 570.19-573.3을 시험 1개가 실행, PASS |
