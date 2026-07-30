# Function Logic Map: `runConsole`

- Source: `cmd/tossctl/console.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| command context | nil or cancellable context | Cobra | nil becomes background; interrupt cancels server |
| root/options | parsed CLI config/remote options | Cobra flags | invalid remote config returns before listener |
| path resolvers | host/container absolute paths | app path functions | required verify paths fail; journal/marker are degraded visibly |
| concrete seams | nil or least-capability adapters | cmd/tossctl wiring | unavailable section is rendered unwired |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | command context nil | choose background | continue | command tests |
| B2-B6 | remote/verify/soak/attestation resolution error | none | return exact error | existing console tests |
| B7 | journal path error | stderr warning | continue with journal unwired | existing path tests |
| B8-B9 | engine dir resolves/fails | set marker or warn | console remains available | engine console tests |
| B10-B20 | container/system-update resolution combinations | wire updater or warn | console remains available | system-update tests |
| B21-B22 | engine dir exists / lock acquisition fails | install update lock seam | error returned by seam | update serialization tests |

The change adds a pre-serve autostart load/start decision. OFF and read error
must never call `startEngine`; ON calls it once and carries its note into the
console.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `remoteAccessOptions` | validate HTTPS remote contract | fail before listen | AST |
| path resolvers | bind records/session/journal | required paths return; optional paths warn | AST |
| `console*Seam` adapters | least-capability injection | nil means unwired | CodeGraph + AST |
| `startEngine` | the only engine spawn path | marker/process + engine interlock; error is not bypassed | CodeGraph impact |
| `console.ListenAndServe` | own server lifetime | returns listener/server error | AST |

## State mutations and fallbacks

- Reads config and may spawn only after an explicit persisted autostart=true.
- No order method is exposed here; the child engine owns all interlocks.
- Autostart read/start failure is a visible engine note, not a console outage.

## Safety conclusion

- Safe edit boundary: construct the autostart seam once, call a small testable helper, inject seam and initial note.
- High-risk impact: yes — a persisted human approval can cause a LIVE-capable process attempt.
