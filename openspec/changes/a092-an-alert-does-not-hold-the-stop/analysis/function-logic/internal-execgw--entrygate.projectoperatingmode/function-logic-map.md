# Function Logic Map: `EntryGate.ProjectOperatingMode`

- Source: `internal/execgw/modegate.go`
- AST evidence: `ast.json` — **편집 뒤**, :35–67, 분기 6 · 반환 2 · 호출 4, source_sha256 `8cb275362b7d…`, 추출 커밋 `2714e393`. 편집 전 번들은 `analysis/pre-edit/unit4/`.
- Risk scan: `risk-pattern-report.md`
- 편집(착지 단위 ④ `2714e393`): 한 `g.mu` 구간 안의 교체(AC2) · `Seq <= modeSeq` 이면 무시(울타리, C5) · 모드 사유 존재가 바뀔 때만 revision(C20) · `g.Block` 호출 제거.
- 재번호: difflib 정렬(편집 전 `22db26e7` 소스 대비): 삭제 B1; B2~B3 → B1~B2(같은 분기, 줄 이동); 새 분기 B3(:49), B4(:54), B5(:55), B6(:64).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 함수 인자 · 수신자 | 편집 전 계약 + 위 편집 | 호출자 | 아래 분기 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if rec.Actor != ""` (:37) | — | — | `TestA092AProjectionReplacesTheModeLatch`, `TestA092AStaleProjectionIsNotApplied` |
| B2 | `if rec.Cause != ""` (:40) | — | — | `TestA092AProjectionReplacesTheModeLatch`, `TestA092AStaleProjectionIsNotApplied` |
| B3 | `if rec.Seq <= g.modeSeq` (:49) | — | — | `TestA092AStaleProjectionIsNotApplied` |
| B4 | `if !rec.BlocksEntry()` (:54) | — | — | `TestA092AStaleProjectionIsNotApplied`, `TestA092TheModeRevisionMovesOnlyWhenPresenceChanges` |
| B5 | `if had` (:55) | — | — | `TestA092AStaleProjectionIsNotApplied`, `TestA092TheModeRevisionMovesOnlyWhenPresenceChanges` |
| B6 | `if !had` (:64) | — | — | `TestA092AProjectionReplacesTheModeLatch`, `TestA092AStaleProjectionIsNotApplied` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `delete`, `g.mu.Lock`, `g.mu.Unlock`, `rec.BlocksEntry` | 편집 뒤 호출 | 위 편집 참조 | AST |

## State mutations and fallbacks

- 위 편집 요약이 전부 — 나머지는 편집 전과 같음(편집 전 번들 대조).

## Safety conclusion

- Safe edit boundary: 위 편집만.
- High-risk impact: High-risk(진입 게이트).
