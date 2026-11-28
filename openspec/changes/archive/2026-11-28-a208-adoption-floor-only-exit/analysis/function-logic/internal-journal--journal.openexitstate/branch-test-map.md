# Branch Test Map: `Journal.OpenExitState`

편집 뒤 좌표(a208 F3 재추출 · AST 20 분기, `internal/journal/exit_state.go` 117–205). 분기 수 무변(B10 조건 확장, 조건이 두 줄이 되어 B11 이후 +1 줄).

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | PositionID 없음 | 미측정 | n/a | 미측정 |
| B2 | kind 빈 값 | 미측정 | n/a | 미측정 |
| B3 | kind 미지 | 미측정 | n/a | 미측정 |
| B4 | switch kind | — | n/a | — |
| B5 | case RATCHET | 미측정 | n/a | 미측정 |
| B6 | RATCHET 에 정책 ID | 미측정 | n/a | 미측정 |
| B7 | case LADDER | `TestOpenExitStatePreservesPinnedRuntimePolicyIdentity` | n/a | yes — 선택 시험 통과(F3) |
| B8 | LADDER 빈 ID | 미측정 | n/a | 미측정 |
| B9 | else | — | n/a | — |
| B10 | `!ok && policyID != "default_v1" && !IsAdoptedFloorOnlyPolicyID(policyID)` — **a208 F3 조건 확장** | `TestA208FloorOnlyCopyNeedsAnAdoptedPosition`(사본 ID 는 통과해 B16 로) | yes — red-f3.txt | yes — 선택 시험 통과(F3) |
| B11 | OpenRatchetState 실패 | 미측정 | n/a | 미측정 |
| B12 | BeginTx 실패 | 미측정 | n/a | 미측정 |
| B13 | 포지션 없음 | 미측정 | n/a | 미측정 |
| B14 | 포지션 읽기 실패 | 미측정 | n/a | 미측정 |
| B15 | exit 대상 아님 | 미측정 | n/a | 미측정 |
| B16 | seedPolicyIdentity 실패 | `TestOpenExitStateRejectsASeedWhosePolicyMeaningDoesNotMatchItsID` · `TestA208FloorOnlyCopyNeedsAnAdoptedPosition` | yes — red-f3.txt | yes — 선택 시험 통과(F3) |
| B17 | INSERT 실패 | — | n/a | — |
| B18 | 고유 위반(이미 있음) | 미측정 | n/a | 미측정 |
| B19 | 이벤트 기록 실패 | 미측정 | n/a | 미측정 |
| B20 | Commit 실패 | 미측정 | n/a | 미측정 |

사본 ID 로 편입 포지션 개시: `TestA208JournalRefusesOtherUnregisteredPolicies`(수용 쪽) · 엔진 `TestA208FloorOnlyAdoptionRecordsTheHybridCopy`.
