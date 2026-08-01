# Function Logic Map: `ExitObserver.workingSet`

- Source: `internal/app/engine/exitloop.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| account positions + per-row exit-state results | closed/zero skipped; eligible only; corrupt generation quarantined | journal projection and a042 quarantine contract | storage errors stop cycle; semantic errors isolate one generation |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B3 | position/state storage reads | none | global error | existing working-set failures |
| B4-B7 | skip closed/zero; unmanaged; missing state open | open/alert existing behavior | continue per position | existing exitloop tests |
| B8-B10 | semantic corrupt state, quarantine, valid managed append | durable quarantine; valid-first ordering | corrupt skipped/alerted after emergency-capable rows | isolation/blocking alert test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `OpenExitStateResults`, `QuarantineExitSnapshot` | preserve driver errors but isolate semantic rows | journal-only, no notifier | CodeGraph + AST |
| `openState`, `managedPolicyIdentity` | establish/validate policy | existing fail-closed contract | CodeGraph + AST |

## State mutations and fallbacks

- Valid managed rows are returned before quarantined notices, so synchronous notifier retries cannot precede another position's emergency judgement.

## Safety conclusion

- Safe edit boundary: working-set classification only; order path unchanged.
- High-risk impact: yes; emergency ordering test required.
