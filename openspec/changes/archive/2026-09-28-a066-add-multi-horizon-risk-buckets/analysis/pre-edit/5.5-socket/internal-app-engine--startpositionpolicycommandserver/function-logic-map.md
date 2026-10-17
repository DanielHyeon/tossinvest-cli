# Function Logic Map (pre-edit): `StartPositionPolicyCommandServer`

- Source: `internal/app/engine/position_policy_transport.go`
- AST evidence: `ast.json` (pre-edit, extracted 2026-09-29 at HEAD `09465697`, 50–167, 16 branches)

## Branches and early returns (pre-edit, from `ast.json`)

| Branch | Line | Condition | Effect |
|---|---|---|---|
| B1 | 52 | `commands == nil` | error, nothing started |
| B2 | 56 | engine directory blank | error |
| B3 | 59 | engine directory not private | error |
| B4 | 64 / B5 else 68 / B6 65 | control directory create; exists is fine, other error returns | error or `createdControlDir` |
| B7 | 71 | control directory validation fails | remove the directory it created, error |
| B8 | 72 | (inside B7) created by this call | remove |
| B9 | 78 | cleanup helper: created by this call | remove |
| B10 | 107 | loopback listen fails | cleanup, error |
| B11 | 112 | token generation fails | close listener, cleanup, error |
| B12 | 123 | `/v1/health` non-GET | 405 |
| B13 | 130 | `/v1/positions` non-GET | 405 |
| B14 | 135 | `List` error | mapped RPC error |
| B15 | 148 | `commands` implements `exitQuarantineCommands` (a079) | register the three quarantine routes on the same mux |
| B16 | 158 | descriptor publication fails | close listener, cleanup, error |

## Planned edit (a066 5.5, Manager ruling 2026-09-29 Q1)

Right after B15, the same discovery for a066: `if relaxations, ok := commands.(riskRelaxationCommands); ok { registerRiskRelaxationRoutes(mux, server, token, relaxations) }`. This is the a079 precedent verbatim: same listener, same bearer token, same private descriptor. The trust boundary is identical, and an engine build without the capability serves the old route set. One new branch lands between B15 and B16, so old B16 becomes B17. Nothing before B15 moves. The descriptor is still published last, so the new routes exist before any client can read the address.
