# Function Logic Map: `(*officialfx.ProductionAuthorityService).collectUS`

- Source: `internal/officialfx/production.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

```text
secure canonical pinned manifest bytes
  -> exact Ed25519 body verification
  -> durable trusted-time/generation CAS under process-shared lock
  -> private newHaircutPolicy
  -> existing ReadOfficial
  -> durable floor/generation commit
  -> opaque Evidence
```

## Branches and early returns

| Branch | Side effect | Result | Test |
| --- | --- | --- | --- |
| absent/unsafe/digest mismatch manifest | none | US refusal before FX read | manifest matrix |
| invalid canonical/schema/scope/decimal/time/signature | none | US refusal before FX read | manifest matrix |
| trusted-time/generation rollback or substitution | state read only | US refusal before FX read | restart rollback tests |
| official FX fails or has no window intersection | read attempt, no state advance | US refusal | official failure tests |
| exact same-generation replay | state floor advances monotonically | valid evidence | replay/race tests |
| higher generation | atomic durable state replace | valid evidence | generation test |

No multiplier default or fallback exists. KR collection does not depend on any branch in this map.

## Calls and live bindings

| Callee | Purpose | Error/retry contract |
| --- | --- | --- |
| `readProductionFile` | owner/mode/symlink-safe exact manifest/state read | fail closed, no fallback |
| `decodeRiskPolicyManifest` / `verifyRiskPolicyManifest` | canonical scope/time/signature and private policy mint | typed refusal before FX |
| `loadRiskPolicyState` | trusted-time/generation precondition | corruption/rollback refusal |
| `productionOfficialReader.readOfficial` | production adapter delegates to existing `ReadOfficial` | one context-bound official read |
| `storeRiskPolicyState` / lock marker | durable state advance after evidence validation | failure discards evidence |

## State mutations and fallbacks

- Process-shared flock serializes state read, official read and atomic state replacement.
- State advances only after the opaque evidence validates at the frozen collection clock.
- Missing/invalid policy never falls back to a numeric constant, identity conversion or activation
  attestation.

## Safety conclusion

- `ProductionAuthorityService.collectUS` in `internal/officialfx/production.go` has no broker or
  journal capability and refuses before returning evidence on any incomplete authority chain.
- KR collection is outside this state path and remains independently observable.
