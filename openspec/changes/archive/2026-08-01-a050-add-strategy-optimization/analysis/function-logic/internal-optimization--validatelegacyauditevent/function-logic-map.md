# Function Logic Map: `validateLegacyAuditEvent`

- Source: `internal/optimization/store.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| legacy event | structurally valid v2 event | migration buffer | must exactly match its verified snapshot metadata and one candidate change |
| snapshot match | audit ID, actor, reason, created time identical | immutable snapshot | mismatch aborts migration |
| candidate match | key, before and after option identical to one decoded change | immutable candidate payload | missing/corrupt/mismatch aborts migration |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | snapshot missing/corrupt | none | wrapped event error | corrupt legacy test |
| B2 | event metadata differs from snapshot | none | mismatch error | corrupt legacy test |
| B3 | candidate row missing/unreadable | none | wrapped error | corrupt legacy test |
| B4 | candidate changes JSON invalid | none | invalid changes error | corrupt legacy test |
| B5 | iterate candidate changes | none | compare all | migration test |
| B6 | exact key/before/after match found | none | success; otherwise final mismatch error | migration/corrupt tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `scanSnapshot` | authenticates historical snapshot | full digest required | migration tests |
| candidate query/JSON decode | corroborates event intent | exact one candidate; no fallback | corrupt legacy test |

## State mutations and fallbacks

- Read-only validation inside migration transaction. It never edits a legacy row or accepts a merely plausible standalone event.

## Safety conclusion

- Safe edit boundary: legacy event corroboration before first digest.
- High-risk impact: yes; migration must not sign forged history.
