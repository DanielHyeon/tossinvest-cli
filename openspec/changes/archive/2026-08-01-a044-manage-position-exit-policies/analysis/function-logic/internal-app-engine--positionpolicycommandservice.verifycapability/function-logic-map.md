# Function Logic Map: `PositionPolicyCommandService.verifyCapability`

- Source: `internal/app/engine/position_policy_command.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| capability token | 32-byte base64url random value | prior Preview | invalid without mutation |
| grant time | no clock rollback; READOPT age <=15s inclusive; global TTL active | engine clock | stale/expired grant consumed |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | malformed token | none | invalid | invalid capability tests |
| B2-B3 | constant-time digest scan | record matching index only | invalid if absent | replay tests |
| B4-B5 | missing or wrong instance/domain | consume if found | invalid | cross-engine/domain tests |
| B6 | clock precedes issue time | consume | stale | rollback test |
| B7 | READOPT exceeds 15-second inclusive freshness | consume | stale | boundary test |
| B8 | danger delay not reached | retain grant | too early | danger delay test |
| B9 | global TTL reached | consume | expired | expiry test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `ConstantTimeCompare` | prevent token/digest/domain timing shortcuts | full scan; no early match return | AST |
| engine clock | enforce issue/freshness/delay/expiry boundaries | rollback is stale, not reusable | AST |

## State mutations and fallbacks

- READOPT uses the one quote-derived Preview state only; verification never re-observes or extends it.

## Safety conclusion

- Safe edit boundary: keep freshness inclusive at 15s, consume rollback/stale/expired grants, and retain only too-early grants.
- High-risk impact: yes
