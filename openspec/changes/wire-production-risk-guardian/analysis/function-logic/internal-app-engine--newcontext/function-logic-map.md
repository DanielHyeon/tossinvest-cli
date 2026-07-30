# Function Logic Map: `NewContext`

- Source: `internal/app/engine/engine.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `ctx` | caller context; cancellation bounds account and startup reads | caller | propagate/refuse before context publication |
| `opts` credentials/config | official credentials and parseable config | `NewOrderPath` | return without opening journal |
| automation gate | OFF, or ON with five positive limits, currency and attestation | config + `runInterlock` | fail closed, no context/loop |
| account reference | non-empty official account | `resolveAccountRef` | audited startup refusal |
| journal | one writable, durable, account-scoped engine journal | `openEngineJournal` | audited refusal; no Guardian |
| Guardian | injected test instance, or exactly one production `RiskGuardian` when gate ON | engine wiring | close journal and refuse on construction/interlock error |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `NewOrderPath` fails | none | propagate | existing credentials/config tests |
| B2 | audit open fails | none | propagate | existing audit tests |
| B3 | no clock injected | choose system clock | continue | existing default-clock test |
| B4 | gate settings cannot be audited | attempted audit | refuse | existing audit failure test |
| B5 | account resolution fails | refusal audit | `ErrAccountUnresolved` | existing account test |
| B6 | journal open fails | refusal audit | refuse | existing durability test |
| B7 | apply-hook bind fails | close journal | refuse | existing hook test |
| B8 | gateway build fails | close journal | refuse | existing gateway test |
| B9-new | gate ON, no injected Guardian | construct production Guardian from same journal/account | close/refuse on error | production construction + cleanup tests |
| B10 | interlock fails | close journal | return interlock error | existing + policy mismatch tests |
| B11 | interlock verifies | publish same local Guardian | context | identity/exit-observer test |
| B12 | gate OFF | publish nil Guardian | unverified context; CLI later starts no loop | gate-off non-construction test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `NewOrderPath` | load config/credentials and build official-only order path | no WTS fallback | CodeGraph + AST |
| `resolveAccountRef` | establish durable account scope | caller context; fail closed | CodeGraph + AST |
| `openEngineJournal` | open sole writable journal | filesystem/integrity guards | CodeGraph + AST |
| `buildGateway` | bind journaled mutation path and observation wiring | close journal on failure | CodeGraph + AST |
| `newProductionGuardian` | derive risk authority from gate/account/journal | validates full policy and costs | new helper contract |
| `runInterlock` | compare Guardian limits and all startup clauses | audited refusal; no loops | CodeGraph + AST |

## State mutations and fallbacks

- Audit setting/refusal writes precede context publication.
- Journal handle is acquired once and transferred to `Context`; every later
  error closes it.
- The new edit may only introduce one local `guardian` value after gateway
  assembly and before interlock; that value must feed both `gateFacts` and
  `Context.Guardian`.
- Explicit injected Guardians remain unchanged for focused tests.

## Safety conclusion

- Safe edit boundary: Guardian selection/construction between gateway success
  and `runInterlock`, plus replacing the two `opts.Guardian` reads with the one
  local value. Do not reorder account, journal, gateway, or interlock.
- High-risk impact: yes — Guardian, order authorization, and exit issuance
  identity. Requires policy equivalence, ownership, race and cleanup evidence.
