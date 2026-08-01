# Function Logic Map: `runConsole`

- Source: `cmd/tossctl/console.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| root/options/context | resolved local profile paths only | CLI + profile config | fatal resolution errors return before serve |
| commander seam | authenticated loopback client only; no journal opener/dependency | a044 design D8/D10 | nil/unwired if engine endpoint unavailable |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B7 | context and evidence/path resolution | local reads only | return or documented unwired fallback | cmd console tests |
| B8-B18 | updater/container/lock/runtime optional wiring | local capability construction | warning/unwired or return | update/console tests |
| B19-B25 | autostart/seam/restart/openapi/listen outcomes | existing explicit capabilities | propagated error | cmd console tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `consoleJournalPath` | resolve read-only dashboard journal identity | no creation | cmd path tests |
| `positionpolicyrpc.Dial` | connect to engine-published loopback endpoint | fixed descriptor, bearer auth, no journal dependency | a044 transport tests |
| `console.ListenAndServe` | inject narrow commander only | blocks until shutdown | console integration tests |

## State mutations and fallbacks

- Existing broker reads, updater, engine process controls and config seams remain unchanged.
- The commander client owns only HTTP transport. It cannot open, create, migrate, or directly mutate a journal.

## Safety conclusion

- Safe edit boundary: dial the engine-owned descriptor only after optional autostart; never derive a writable service from `journalPath`.
- High-risk impact: yes — constructor indirection must not reintroduce journal authority; dial failure leaves policy UI unwired.
