# Branch Test Map: `Journal.OpenExitState`

편집 전 좌표(AST 20 분기, `internal/journal/exit_state.go` 117–204).

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | PositionID 없음 | 미측정 | n/a | no |
| B2 | kind 빈 값 | 미측정 | n/a | no |
| B3 | kind 미지 | 미측정 | n/a | no |
| B4 | switch | — | n/a | no |
| B5 | RATCHET | 미측정 | n/a | no |
| B6 | RATCHET + ID | 미측정 | n/a | no |
| B7 | LADDER | `TestOpenExitStatePreservesPinnedRuntimePolicyIdentity` | n/a | no |
| B8 | LADDER 빈 ID | 미측정 | n/a | no |
| B9 | else | — | n/a | no |
| B10 | 미지 정책 → 거부 | `TestA208JournalRefusesOtherUnregisteredPolicies` | n/a | no |
| B11 | t0 산술 실패 | 미측정 | n/a | no |
| B12 | BeginTx 실패 | 미측정 | n/a | no |
| B13 | 포지션 없음 | 미측정 | n/a | no |
| B14 | 포지션 읽기 실패 | 미측정 | n/a | no |
| B15 | exit 대상 아님 | 미측정 | n/a | no |
| B16 | 정체성 실패 | `TestOpenExitStateRejectsASeedWhosePolicyMeaningDoesNotMatchItsID` · `TestA208FloorOnlyCopyNeedsAnAdoptedPosition` | n/a | no |
| B17 | INSERT 실패 | — | n/a | no |
| B18 | 이미 있음 | 미측정 | n/a | no |
| B19 | 이벤트 실패 | 미측정 | n/a | no |
| B20 | Commit 실패 | 미측정 | n/a | no |

a208 F3: 사본 ID 로 편입 포지션 개시 `TestA208FloorOnlyAdoptionRecordsTheHybridCopy`, 비편입 포지션은 B16 거부 `TestA208FloorOnlyCopyNeedsAnAdoptedPosition`.
