# Branch Test Map: `LegacyLadderPolicyIdentity`

편집 전 좌표(AST 8 분기, `internal/exitpolicy/identity.go` 51–76).

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 빈 ID → default_v1 | 미측정 | n/a | no |
| B2 | switch | — | n/a | no |
| B3 | default_v1 | `TestLegacyPolicyIdentityIsExactAndUnknownIsFailClosed` | n/a | no |
| B4 | BALANCED | 미측정 | n/a | no |
| B5 | RUNNER | `TestApprovedPolicyTablesMatchPinnedSemanticDigests`(간접) | n/a | no |
| B6 | 편입 RUNNER | `TestA208OnWithRunnerKeepsTheExistingAdoptedRunnerCopy` | n/a | no |
| B7 | HYBRID_50 | `TestA208AdoptedHybridRowsKeepTheirMeaning` | n/a | no |
| B8 | 미지 ID → 오류 | `TestLegacyPolicyIdentityIsExactAndUnknownIsFailClosed` | n/a | no |

a208 F3 새 case: `TestA208FloorOnlyCopyIsAdoptionOnly`(비편입 거부) · `TestA208FloorOnlyAdoptionRecordsTheHybridCopy`(편입 시드 정체성).
