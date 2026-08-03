# Function Logic Map: `ApplyFillRisk`

- Source: `internal/weeklyvaluelane/risk.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| state/plan | matching sealed plan; private canonical-sealed numeric balances, latches and fill map | a066/risk replay | fail-closed refusal |
| fill event | bounded identity, exact campaign/ordinal/FX, known fees | authoritative fill port | preserve fill+latch on unknown |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | duplicate exact fill | none | duplicate | retry test |
| B2 | conflicting/unknown identity | clone, preserve fingerprint, latch unknown | applied/refusal | missing-risk test |
| B3 | invalid FX/time/arithmetic | clone, preserve fill, latch unknown | applied/refusal | FX test |
| B4 | known risk | clone, transfer held to filled | applied | conservative floor test |
| B5 | actual above transferred | clone, preserve fill, overage latch | applied | overage test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| actualFillRisk | checked exact monetary risk and FX ceil | fail closed | CodeGraph + AST |
| cloneRiskState | copy-on-write transition | original unchanged | CodeGraph + AST |

## State mutations and fallbacks

- Positive fill evidence is always retained; invalid actual risk latches instead of estimating zero. Any post-construction scalar/map mutation invalidates the state seal and is never re-sealed by a failed transition.

## Safety conclusion

- Safe edit boundary: pure state transition only.
- High-risk impact: yes; actual-risk preservation and overage gate.
