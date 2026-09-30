# Function Logic Map: `Notifier.logClaimHeld`

- Source: `internal/obs/notifier.go`
- AST evidence: `ast.json` — **편집 전**, :636–646, 분기 1 · 반환 1 · 호출 4, source_sha256 `0bc75668ff17…`, 추출 HEAD `b3f14925`.
- Risk scan: `risk-pattern-report.md`
- 편집 목적: 21.4 GREEN 목록의 「`logClaimHeld` 등급과 주석」. 주석 :626-635 의 전제(*"the only way to arrive here is a lease left behind by a sender that died"*)는 a098 이후 거짓이고, 단위 ③ 뒤에는 **같은 조건의 동시 동기 발송**도 이 갈래에 닿는다. 등급 결정은 a099 r4 시험(`TestAHeldRowIsNotWhispered` — WARN 단언)과 충돌하므로 **Manager 확인 뒤 편집**(review §24.6). 이 번들은 편집 전 증거.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `claim` | `ClaimHeldElsewhere` 결과(보유자 · 나이 · 만료) | `claimAndDeliver` B6 · `Flush` | — |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `n.Log == nil` (:637) | — | 반환(:638) | (미실행) |
| 종단 | — | `n.Log.Warn(EventAlertClaimHeld, …)` | — | `TestAHeldRowIsNotWhispered` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `n.Log.Warn` :640 | 보유자 · 나이 · 만료 한 줄 | — | AST |

## State mutations and fallbacks

- 로그 한 줄만.

## Safety conclusion

- Safe edit boundary: 등급 · 주석만(결정 대기).
- High-risk impact: 낮음 — 관측 등급.
