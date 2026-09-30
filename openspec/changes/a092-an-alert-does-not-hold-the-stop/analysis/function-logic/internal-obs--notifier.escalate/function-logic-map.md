# Function Logic Map: `Notifier.escalate`

- Source: `internal/obs/notifier.go`
- AST evidence: `ast.json` — **편집 뒤**, :422–444, 분기 4 · 반환 2 · 호출 5, source_sha256 `6878b8f1df55…`, 추출 커밋 `55963f29`(25라운드 수리 뒤 재추출). 편집 전 번들은 `analysis/pre-edit/unit3/`에 보존.
- Risk scan: `risk-pattern-report.md`
- 편집(착지 단위 ③ — `fbc6df5f`): 반환값 `(included bool, err error)` 추가 — B1 조기 반환은 `false, nil`, 끝은 `true, err`. 분기 · 로그 불변.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `n.Journal` · `n.AccountRef` | 둘 다 있어야 승격 포함 | 조립 | B1 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | if (:420) | 승격 미포함 → `(false, nil)` | — | `TestA092AFailedEscalationLatchesUnconditionally`, `TestA092AReleaseBeforeTheEvidenceDoesNotChangeTheVerdict` |
| B2 | switch (:425) | 결과 분기 | — | `TestA092AFailedEscalationLatchesUnconditionally`, `TestA092AReleaseAfterTheEpochReadIsHonoured` |
| B3 | case (:426) | 승격 실패 로그 | — | `TestA092AFailedEscalationLatchesUnconditionally`, `TestA092RecordOnlyFailureLatchesAndEscalates` |
| B4 | case (:432) | 승격 됨 로그 | — | `TestA092AReleaseAfterTheEpochReadIsHonoured`, `TestA092AReleaseBeforeTheEpochReadRelatches` |
| 종단 | — | — | `true, err` — 변이 L09 · L10 | `TestA092AFailedEscalationLatchesUnconditionally` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `EscalateOperatingMode(…, nil)` | 승격 · 통지 없음 | 오류 반환 | AST |

## State mutations and fallbacks

- 모드 행 하나(변화 시).

## Safety conclusion

- Safe edit boundary: 분기 · 로그 불변, 반환값만.
- High-risk impact: yes — 모드 승격.
