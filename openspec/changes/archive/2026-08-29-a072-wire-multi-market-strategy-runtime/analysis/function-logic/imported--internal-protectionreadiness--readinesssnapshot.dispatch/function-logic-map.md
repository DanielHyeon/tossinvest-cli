# Function Logic Map: `ReadinessSnapshot.Dispatch`

- Source: `internal/protectionreadiness/dispatch.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| sealed snapshot | release plus independent KR/US market seals | `Assess`/`DefaultSnapshot` | corrupt selected market returns `state_corrupt`; peer market remains usable |
| dispatch scope | exact account/profile/market and attested order/session/quantity/trigger/replace/capability contract | persisted entry plan plus sealed supervisor contract | any missing/substituted field returns typed refusal before transport |
| current time | non-zero UTC instant inside attestation lifetime | gateway clock | invalid/future/expired returns typed refusal |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | release mismatch | none | `state_corrupt` | corrupt snapshot test |
| B2 | invalid dispatch identity/contract | none | `invalid_attestation` | scope substitution matrix |
| B3 | selected KR or US market seal mismatch | none | `state_corrupt` | market-isolation corruption test |
| B4 | verdict is not exactly WIRED/none | none | stored typed refusal | missing/expired evidence tests |
| B5 | provenance or exact contract mismatch | none | `scope_mismatch` | order/session/quantity/trigger/replace/capability substitution tests |
| B6 | issued/expiry window invalid | none | invalid/expired | expiry tests |
| B7 | all checks pass | none | allowed with sealed snapshot ID | valid dispatch test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `marketVerdictSeal` | authenticate the selected immutable market verdict | pure/no retry | CodeGraph + AST |
| `validDigest` | validate build/evidence/supervisor/capability digests | pure/no retry | current HEAD |

## State mutations and fallbacks

- Pure decision function; no file, durable-state, toggle or broker mutation.
- A corrupt KR sub-snapshot must not invalidate a separately sealed US sub-snapshot and vice versa.

## Safety conclusion

- Safe edit boundary: extend the already sealed dispatch preimage; do not add fallback/default contract values.
- High-risk impact: yes — this is the final exposure-raising protection gate.
