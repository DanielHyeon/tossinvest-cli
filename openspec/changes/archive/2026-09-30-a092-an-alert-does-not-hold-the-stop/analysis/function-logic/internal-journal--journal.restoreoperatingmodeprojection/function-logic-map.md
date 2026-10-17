# Function Logic Map: `Journal.RestoreOperatingModeProjection`

- Source: `internal/journal/operating_mode.go`
- AST evidence: `ast.json` — **편집 뒤**, :589–605, 분기 2 · 반환 2 · 호출 3, source_sha256 `093341ae5773…`, 추출 커밋 `2714e393`. 편집 전 번들은 `analysis/pre-edit/unit4/`.
- Risk scan: `risk-pattern-report.md`
- 편집(착지 단위 ④ `2714e393`): 투영 레코드에 `Seq: snapshot.Seq`(울타리).
- 재번호: difflib 정렬(편집 전 `22db26e7` 소스 대비): B1~B2 → B1~B2(같은 분기, 줄 이동).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 함수 인자 · 수신자 | 편집 전 계약 + 위 편집 | 호출자 | 아래 분기 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if err != nil` (:591) | — | — | (미실행) |
| B2 | `if p := j.modeProjectorRef(); p != nil` (:594) | — | — | `TestA092TransitionsCarryTheirCommitSequence`, `TestTheModeIsRestoredAfterARestart` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `j.CurrentOperatingMode`, `j.modeProjectorRef`, `p.ProjectOperatingMode` | 편집 뒤 호출 | 위 편집 참조 | AST |

## State mutations and fallbacks

- 위 편집 요약이 전부 — 나머지는 편집 전과 같음(편집 전 번들 대조).

## Safety conclusion

- Safe edit boundary: 위 편집만.
- High-risk impact: High-risk.
