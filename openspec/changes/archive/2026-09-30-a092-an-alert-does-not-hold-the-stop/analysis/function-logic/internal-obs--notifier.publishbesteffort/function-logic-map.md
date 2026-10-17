# Function Logic Map: `Notifier.publishBestEffort`

- Source: `internal/obs/notifier.go`
- AST evidence: `ast.json` — **편집 뒤**, :179–192, 분기 2 · 반환 1 · 호출 6, source_sha256 `d705f78d68c1…`, 추출 커밋 `15b64676`. 편집 전 번들은 `analysis/pre-edit/r26b/internal-obs--notifier.publishbesteffort.json(AST)`에 보존.
- Risk scan: `risk-pattern-report.md`
- 편집: (b910173a, 보이스 A#4 · B#4) 발행 실패 줄의 원래 사건 유형 · 등급 키를 `trigger_event` · `trigger_severity` 로 — emit 이 쓰는 줄 자신의 event · severity 를 가리지 않음. 분기 불변.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `n.Publisher` | nil 이면 B1 | 조립 | 반환 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | if (:180) | 발행기 없음 → 반환 | — | `TestA092DropAndNoJournalLinesKeepTheirOwnEventKey`, `TestTheTransitionLogLineIsCountable` |
| B2 | if (:183) | 발행 실패 → 경고 줄 | — | `TestA092DropAndNoJournalLinesKeepTheirOwnEventKey`, `TestObservationFailureAlertNeverReachesTheGateOrTheMode` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `Publisher.Publish` | 최선 발송 | 오류는 로그만 | AST |

## State mutations and fallbacks

- 로그 줄 하나.

## Safety conclusion

- Safe edit boundary: 로그 키만.
- High-risk impact: no(일반 등급 관측).
