# Branch Test Map: `Console.handleReport`

- Source: `internal/console/pages.go` (353-366) — 편집 뒤. AST branches 2.
- 측정: 커버리지 시험 바이너리(`go test -c -cover`)로 top-level 시험 607 개를 하나씩 `-test.run '^<Test>$' -test.coverprofile` 실행, 분기 본문 블록 count>0 인 시험만 인용(2026-10-25, a173 close). n 은 그 블록을 지나는 시험 수.

| Branch | Anchor | Scenario | Test | 측정 n | a173 |
|---|---|---|---|---|---|
| B1 | `357:2` | report 오류 → `page.Error` | 측정상 없음(기존 공백) | 0 | 무변경 — 측정 시험 0(기존 공백, a173 무관) |
| B2 | `359:9` | 성공 → 텍스트 렌더 · Record 분리 | `TestOperatorBodyCarriesNoDeveloperIdentifiers` · `TestEachUnmeasuredCauseIsExplainedOnce` | 21 | a173: Record 분리 · 본문에서 경로 줄 제거 |
