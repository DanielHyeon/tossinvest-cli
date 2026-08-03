# Function Logic Map: `NewRouter`

- Source: `internal/httpapi/router.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| primary Reader | non-nil existing seven-resource reader | Options | constructor error |
| strategy runtime reader | nil or narrow read-only projection reader | Options | nil means explicit dormant projection at read time |
| mutation route map | exact two approved optimization paths only | hard-coded allowlist | constructor error before serving |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | primary reader nil | none | error | existing constructor test |
| B2 | clock nil | assign `time.Now` | continue | existing constructor test |
| B3 | mutation path not exact or handler nil | none | error | mutation path guard test |
| B4 | valid options | clone route map only | router | strategy runtime route test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| none | constructor copies capability references only | no I/O/retry | CodeGraph + AST |

## State mutations and fallbacks

- Adding the projection reader does not add to `allowedMutationRoutes`; it is stored as a separate read-only interface.

## Safety conclusion

- Safe edit boundary: retain all validation and copy one optional read-only seam.
- High-risk impact: no; no request is executed in the constructor.
