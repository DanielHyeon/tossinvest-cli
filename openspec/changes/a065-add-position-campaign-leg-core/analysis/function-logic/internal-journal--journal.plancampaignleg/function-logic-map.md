# Function Logic Map: `Journal.PlanCampaignLeg`

- Source: `internal/journal/position_campaign.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| requested quantity/sequence | positive canonical decimal, next contiguous sequence | request + campaign legs | invalid request |
| expected campaign version | exact | campaign projection | version conflict |
| exposure admission | campaign unblocked, latest Position not CLOSING, no unresolved SELL intent | journal local state | `ErrExposureBlocked` |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | invalid quantity/identity | none | invalid request | zero/sequence tests |
| B2 | deterministic retry | none | existing leg | command retry test |
| B3 | stale version | none | version conflict | command test |
| B4 | EXIT FIRST block | none | exposure blocked | Position/pending SELL tests |
| B5 | valid next leg | leg+campaign+command+event atomically | record | plan tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `campaignExposureBlockedInTx` | local EXIT FIRST query | no broker/wait; SQL error fails closed | AST |
| `TransitionCampaign` | D4 admission | invalid transition refuses | AST |

## State mutations and fallbacks

- Only journal planning evidence mutates; no intent, attempt, broker or toggle is created.

## Safety conclusion

- Safe edit boundary: exposure-raising admission and ordered plan persistence.
- High-risk impact: yes; EXIT FIRST is fail closed without touching safety paths.
