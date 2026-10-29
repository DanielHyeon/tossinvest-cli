# Branch Test Map: `buildFlattenWiring`

- Source: `cmd/tossctl/flatten.go` (199-273; GREEN 뒤 AST 재추출 2026-10-29)
- AST branches 8 — 실행 칸은 편집 전 패키지 스위트 coverprofile 측정(시험 이름 귀속은 GREEN 재추출 때)

| Branch | Coordinate | Scenario | Baseline executed | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `if at 201:2` | if err != nil { | 미측정 — 이 패키지 coverprofile 없음 | 해당 없음 — 편집 전 기준선(0.4) | 해당 없음 — 편집 전 기준선(0.4) |
| B2 | `if at 206:2` | if err != nil { | 미측정 — 이 패키지 coverprofile 없음 | 해당 없음 — 편집 전 기준선(0.4) | 해당 없음 — 편집 전 기준선(0.4) |
| B3 | `range at 210:2` | for _, a := range accounts { | 미측정 — 이 패키지 coverprofile 없음 | 해당 없음 — 편집 전 기준선(0.4) | 해당 없음 — 편집 전 기준선(0.4) |
| B4 | `if at 211:3` | if strings.TrimSpace(a.DisplayName) != "" { | 미측정 — 이 패키지 coverprofile 없음 | 해당 없음 — 편집 전 기준선(0.4) | 해당 없음 — 편집 전 기준선(0.4) |
| B5 | `if at 216:2` | if accountRef == "" { | 미측정 — 이 패키지 coverprofile 없음 | 해당 없음 — 편집 전 기준선(0.4) | 해당 없음 — 편집 전 기준선(0.4) |
| B6 | `if at 221:2` | if root.configDir != "" { | 미측정 — 이 패키지 coverprofile 없음 | 해당 없음 — 편집 전 기준선(0.4) | 해당 없음 — 편집 전 기준선(0.4) |
| B7 | `if at 227:2` | if err != nil { | 미측정 — 이 패키지 coverprofile 없음 | 해당 없음 — 편집 전 기준선(0.4) | 해당 없음 — 편집 전 기준선(0.4) |
| B8 | `if at 242:2` | if err != nil { | 미측정 — 이 패키지 coverprofile 없음 | 해당 없음 — 편집 전 기준선(0.4) | 해당 없음 — 편집 전 기준선(0.4) |
