# Branch Test Map: `LegacyLadderPolicyIdentity`

편집 뒤 좌표(a208 F3 재추출 · AST 9 분기, `internal/exitpolicy/identity.go` 51–79). B8 이 새 case, 옛 B8(default) → B9.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 빈 ID → default_v1 | 미측정 | n/a | 미측정 |
| B2 | switch | — | n/a | — |
| B3 | default_v1 | `TestLegacyPolicyIdentityIsExactAndUnknownIsFailClosed` | n/a | yes — 선택 시험 통과(F3) |
| B4 | BALANCED | 미측정 | n/a | 미측정 |
| B5 | RUNNER | `TestApprovedPolicyTablesMatchPinnedSemanticDigests`(간접) | n/a | yes — 선택 시험 통과(F3) |
| B6 | 편입 RUNNER | `TestA208OnWithRunnerKeepsTheExistingAdoptedRunnerCopy` | n/a(base 에서도 통과 기대) | yes — 선택 시험 통과(F3) |
| B7 | HYBRID_50(편입 · 비편입 같은 정체성) | `TestA208AdoptedHybridRowsKeepTheirMeaning` | n/a(가드 — base 에서도 통과 기대) | yes — 선택 시험 통과(F3) |
| B8 | 편입 보호선 사본 · 비편입 거부 | `TestA208FloorOnlyCopyKeepsTheHybridFloors` · `TestA208FloorOnlyCopyIsAdoptionOnly` | yes — red-f3.txt | yes — 선택 시험 통과(F3) |
| B9 | 미지 ID → 오류 | `TestLegacyPolicyIdentityIsExactAndUnknownIsFailClosed` | n/a | yes — 선택 시험 통과(F3) |
