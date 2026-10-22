# Branch Test Map: `newMatcher`

- Source: `internal/execgw/indoubt.go` (652-679)
- AST branches 3 — 실행 칸은 편집 전 패키지 스위트 coverprofile 측정(시험 이름 귀속은 GREEN 재추출 때)

| Branch | Coordinate | Scenario | Baseline executed | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `if at 654:2` | if err != nil { | 미실행 — 패키지 스위트 count=0 | 해당 없음 — 편집 전 기준선(0.4) | 해당 없음 — 편집 전 기준선(0.4) |
| B2 | `if at 670:2` | if stamp == "" { | 실행됨 — 패키지 스위트 count=1 | 해당 없음 — 편집 전 기준선(0.4) | 해당 없음 — 편집 전 기준선(0.4) |
| B3 | `if at 673:2` | if dispatched, err := time.Parse(time.RFC3339, stamp); err == nil { | 실행됨 — 패키지 스위트 count=1 | 해당 없음 — 편집 전 기준선(0.4) | 해당 없음 — 편집 전 기준선(0.4) |
