# Function Logic Map: `Notifier.logLeaseLost`

- Source: `internal/obs/notifier.go`
- AST evidence: `ast.json` — **편집 뒤**, :677–718, 분기 7 · 반환 1 · 호출 14, source_sha256 `fbdfd9e0218b…`, 추출 커밋 `b910173a`. 편집 전 번들은 `analysis/pre-edit/r26b/internal-obs--notifier.logleaselost.json(AST)`에 보존.
- Risk scan: `risk-pattern-report.md`
- 편집: (b910173a, codex 재확인 R3) 결과 enum 으로 분류 — 모르는 결과는 `EventAlertUndelivered` 오류(선점 이름 금지). 기록만, 판정 없음.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `res.Outcome` | Applied 외 전부(호출자가 Applied 를 먼저 거름) | 원장 | B3~B7 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | if (:678) | 로거 없음 → 반환 | — | (미실행) |
| B2 | switch (:681) | 결과 분기 | — | `TestA092AReleaseAfterTheEpochReadIsHonoured`, `TestA092AReleaseBeforeTheEpochReadRelatches` |
| B3 | case (:682) | 행 없음 → 오류 줄 | — | `TestA092AReleaseAfterTheEpochReadIsHonoured`, `TestA092AReleaseBeforeTheEpochReadRelatches` |
| B4 | case (:691) | 이미 정산 → 선점 기록 | — | `TestA092AnAcknowledgementPreemptsASendInFlight`, `TestA092TheLeaseLossLogClassifiesByOutcome` |
| B5 | case (:696) | LeaseLost · 토큰 없음 → 자기 반납 재확인 | — | `TestA092TheLeaseLossLogClassifiesByOutcome` |
| B6 | case (:703) | 모르는 결과 → 원장 이상 오류 줄 | — | `TestA092TheLeaseLossLogClassifiesByOutcome` |
| B7 | case (:711) | LeaseLost · 남의 토큰 → 선점 경고 | — | `TestA092TheLeaseLossLogClassifiesByOutcome`, `TestASenderThatLosesTheLeaseStopsAtOnce` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `Log.Error` · `Log.Event` · `Log.Warn` | 기록 | 없음 | AST |

## State mutations and fallbacks

- 로그 줄 하나.

## Safety conclusion

- Safe edit boundary: 로그 분류만 — 차단은 호출자가 이미 판정.
- High-risk impact: no(관측).
