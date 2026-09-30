# Function Logic Map: `Notifier.escalate`

- Source: `internal/obs/notifier.go`
- AST evidence: `ast.json` — **편집 전**, :378–399, 분기 4 · 반환 1 · 호출 5, source_sha256 `0bc75668ff17…`, 추출 HEAD `b3f14925`.
- Risk scan: `risk-pattern-report.md`
- 편집 목적: 23.3 K2 수단(design D0.3h 3): 반환값 없음(:378 — 실패는 로그만)을 **승격 포함 여부와 실패**를 돌려주게 바꾼다. 호출자가 「승격 포함 판정의 승격 쓰기 실패 → 무조건 차단」을 한다. 분기와 로그는 그대로.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `n.Journal` · `n.AccountRef` | 둘 중 하나라도 없으면 승격 미포함 | 조립 | B1 조기 반환 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `n.Journal == nil || AccountRef 빈 값` (:379) | — | 반환(:380) — 편집 뒤 「미포함」 | `TestADeadTransportIsStillFoundAfterASuccessfulDelivery` |
| B2 | `switch` (:384) | — | — | 전부 |
| B3 | `err != nil && n.Log != nil` (:385) | 오류 로그 | — | `TestAClaimThatFailsAttemptsTheDurableBlock` |
| B4 | `changed && n.Log != nil` (:391) | 경고 로그 | — | `TestAnUndeliverableCriticalAlertTightensTheOperatingMode` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `n.Journal.EscalateOperatingMode(ctx, acct, CRITICAL_ALERT_UNDELIVERED, nil)` :382 | 모드 승격 · 통지 없음 | 오류 → 편집 뒤 반환 | AST |

## State mutations and fallbacks

- 원장 모드 행 하나(변화 있을 때). 통지자 nil — 전달 실패 강화는 통지하지 않음(M1).
- **B3 는 `err != nil && n.Log != nil` 이 한 조건**이라 `n.Log == nil` 이면 실패가 B4 로 새지 않고 switch 를 빠져나감 — 편집 뒤 반환값은 로그 유무와 무관하게 오류를 싣는다.

## Safety conclusion

- Safe edit boundary: 분기 · 로그 불변, 반환값 추가.
- High-risk impact: yes — 모드 승격.
