# Branch Test Map: `Console.handleReport`

- Source: `internal/console/pages.go` (355-368). AST branches 2.
- 측정(2026-10-28, HEAD `eaf713ff9`): `go test ./internal/console -count=1 -coverprofile` 전체 실행에서 B1 블록 `pages.go:359.16,361.3` count 0, B2 블록 `361.8,366.3` count>0. 아래 두 시험만 `-run` 으로 돌린 프로파일에서도 B2 count 1. a173 의 시험별 측정(607 개, B1 n=0 · B2 n=21)과 같은 결론.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | report 오류 → `page.Error` | 측정상 없음(기존 공백 — a172·a173 무관) | no | 미실행(시험 0) |
| B2 | 성공 → 텍스트 렌더 · Record 분리 | `TestOperatorBodyCarriesNoDeveloperIdentifiers` · `TestEachUnmeasuredCauseIsExplainedOnce` | no(사후 작성) | yes |
