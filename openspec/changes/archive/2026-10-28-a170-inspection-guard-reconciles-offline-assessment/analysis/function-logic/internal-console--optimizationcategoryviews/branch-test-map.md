# Branch Test Map: `optimizationCategoryViews`

- Source: `internal/console/optimization_view.go` (203-230) — 편집 뒤. AST branches 10.
- 측정: 커버리지 시험 바이너리(`go test -c -cover`)로 top-level 시험 607 개를 하나씩 `-test.run '^<Test>$' -test.coverprofile` 실행, 분기 본문 블록 count>0 인 시험만 인용(2026-10-25, a173 close). n 은 그 블록을 지나는 시험 수.

| Branch | Anchor | Scenario | Test | 측정 n | a173 |
|---|---|---|---|---|---|
| B1 | `205:2` | 카테고리 순회 | `TestOperatorBodyCarriesNoDeveloperIdentifiers` | 41 | 무변경 |
| B2 | `207:3` | 카테고리 ID switch | `TestOperatorBodyCarriesNoDeveloperIdentifiers` | 41 | 무변경 |
| B3 | `208:3` | overview | `TestOperatorBodyCarriesNoDeveloperIdentifiers` | 41 | 무변경 |
| B4 | `210:3` | exit-protection → Available=writable | `TestOperatorBodyCarriesNoDeveloperIdentifiers` | 41 | 무변경 |
| B5 | `212:4` | writable → 연결됨 | `TestOptimizationCategoryOrderMobileTouchFocusAndCSP` | 14 | 무변경 |
| B6 | `214:11` | !writable → 미배선 · 읽기 전용 | `TestOperatorBodyCarriesNoDeveloperIdentifiers` | 27 | 무변경 |
| B7 | `217:3` | position-management | `TestOperatorBodyCarriesNoDeveloperIdentifiers` | 41 | a173: Status 에서 ID 분리 → Tech |
| B8 | `219:3` | candidate-filters | `TestOperatorBodyCarriesNoDeveloperIdentifiers` | 41 | a173: Status 에서 ID 분리 → Tech |
| B9 | `221:3` | strategy-runtime | `TestOperatorBodyCarriesNoDeveloperIdentifiers` | 41 | a173: Status 에서 ID 분리 → Tech |
| B10 | `223:3` | performance-history | `TestOperatorBodyCarriesNoDeveloperIdentifiers` | 41 | a173: Status 에서 ID 분리 → Tech |
