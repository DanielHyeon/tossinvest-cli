# Function Logic Map: `httpAPIReader.readPositions`

- Source: `cmd/tossctl/httpapi_reader.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| Reader dependencies | holdings/account required; journal optional | injected read-only seams | missing required seam returns typed error |
| Journal/lifecycle rows | readable/unavailable, KR/US | journal `LivePositionExits` + `PositionPolicies` | unavailable rows remain unknown |
| Runtime policy | effective-known or unavailable | engine runtime snapshot | desired settings never substitute |
| Exit evidence | fresh/stale/legacy and lifecycle generation | persisted journal | mismatch clears prices and identities |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | reader/holdings/account seam missing | none | error | production reader tests |
| B2 | account reference read fails | none | error | reader failure tests |
| B3 | broker holdings read fails | none | error | reader failure tests |
| B4 | journal readable | SELECT-only reads | continues | journal projection tests |
| B5 | live positions read fails | none | error | journal error tests |
| B6 | policy states read fails | none | error | lifecycle tests |
| B7 | each policy state | exact position-ID map | continues | lifecycle matrix |
| B8 | each journal row | exact market+symbol map | continues | mixed rows |
| B9 | each broker row | local projection | continues | broker/journal join tests |
| B10 | matching stored row | adds stored/lifecycle/reference evidence | continues | legacy and mismatch tests |
| B11 | no stored row | unknown exit plus candidate plan if safe | continues | US candidate test |
| B12 | matched row released | clears actionable exit | continues | released test |
| B13 | journal unavailable in broker-only branch | changes unknown reason only | continues | unavailable test |
| B14 | each remaining journal row | journal-only projection | continues | journal-only tests |
| B15 | row already seen from broker | skips duplicate | continues | join tests |
| B16 | journal-only row released | clears actionable exit | continues | released test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `LivePositionExits` / `PositionPolicies` | obtain exact journal and lifecycle evidence | errors abort the read | CodeGraph + AST |
| `applyStoredPosition` | canonical fail-closed snapshot projection | pure local mutation | package tests |
| `applyManagementProjection` | shared management status | runtime unknown remains unknown | positionpolicy tests |
| `applyExitLineReference` | shared legacy/plan/mismatch reference | clears mismatch evidence | a053 focused tests |

## State mutations and fallbacks

- Mutates only the response-local `httpapi.Position` values and cache result.
- Uses exact market+symbol and position-ID joins; it neither writes the journal nor invokes reconcile/order commands.
- Cross-generation values and identities are cleared after the existing stored projection.

## Safety conclusion

- Safe edit boundary: read-only transport projection after existing canonical and management adapters.
- High-risk impact: no direct order side effect; exit evidence remains fail-closed and non-actionable unless canonical.
