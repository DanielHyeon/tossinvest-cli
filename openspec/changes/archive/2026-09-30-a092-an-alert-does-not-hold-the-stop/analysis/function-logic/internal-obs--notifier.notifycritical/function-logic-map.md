# Function Logic Map: `Notifier.notifyCritical`

- Source: `internal/obs/notifier.go`
- AST evidence: `ast.json` — **편집 뒤**, :195–250, 분기 4 · 반환 3 · 호출 11, source_sha256 `d705f78d68c1…`, 추출 커밋 `15b64676`. 편집 전 번들은 `analysis/pre-edit/r26b/internal-obs--notifier.notifycritical/`에 보존.
- Risk scan: `risk-pattern-report.md`
- 편집: (b910173a, 보이스 A#4 · B#4) 원장 없음 경고 줄의 원래 사건 유형 키 `FieldEvent` → `FieldTriggerEvent`(줄 자신의 event 를 가리지 않음). 분기 불변.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `n.Journal` | 없으면 B1 | 조립 | 경고 + 최선 발행 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | if (:196) | 원장 없음 | — | `TestA092DropAndNoJournalLinesKeepTheirOwnEventKey`, `TestCriticalWithoutAJournalIsLoudRatherThanSilent` |
| B2 | if (:200) | 로그 | — | `TestA092DropAndNoJournalLinesKeepTheirOwnEventKey`, `TestCriticalWithoutAJournalIsLoudRatherThanSilent` |
| B3 | if (:220) | claim 실패 → 승격 | — | `TestAClaimThatFailsAttemptsTheDurableBlock`, `TestAClaimThatFailsBlocksNewEntries` |
| B4 | if (:242) | 미전달 → judge | — | `TestA092AFailedEscalationLatchKeepsTheAccountOut`, `TestA092AFailedEscalationLatchesUnconditionally` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `n.claimAndDeliver` | 기록 · 발송 | (sent, owed, verdict, err) | AST |
| `n.judge` | 원칙 E | — | AST |

## State mutations and fallbacks

- 로그 키 하나만 바뀜.

## Safety conclusion

- Safe edit boundary: 로그 키만.
- High-risk impact: no(관측 줄) — 판정 불변.
