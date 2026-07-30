## Context

`engine.NewContext` already owns the only safe startup sequence:

1. load config and official-only broker,
2. resolve the broker account,
3. open the writable journal and bind fill hooks,
4. build the gateway,
5. run the automation interlock,
6. publish a verified context used to construct reconciliation, exit observation,
   fill detection, and recovery.

The production caller in `cmd/tossctl` supplies no Guardian. Package tests hide
that omission by constructing a separate journal and injecting a ready-made
Guardian before `NewContext` is called. That test shape cannot be the production
shape because the Guardian must write decisions and reservations into the exact
journal owned by the engine context.

## Goals / Non-Goals

**Goals:**

- Construct one real `execgw.RiskGuardian` from the resolved account, the opened
  engine journal, the startup clock, configured gate limits, and the configured
  cost model.
- Make the configured gate the single source for the five audited bounds and
  their currency.
- Pass the same interface value through the interlock and `Context.Guardian`, so
  `Context.ExitObserver` necessarily consumes the same `ReductionIssuer`.
- Preserve explicit Guardian injection for focused package tests.
- Exercise the production CLI assembler with isolated files and an `httptest`
  official broker.

**Non-Goals:**

- Enabling the gate, changing limits, weakening interlock clauses, or changing
  live order approval.
- Adding a new entry strategy or changing stop/target sizing.
- Adding a notification transport.
- Contacting a live Toss endpoint in tests.

## Decisions

### D1. The engine profile constructs the production Guardian

`engine.NewContext` will resolve a local variable named `guardian` after account
resolution, journal opening, apply-hook binding, and gateway construction, but
before `runInterlock`.

If `Options.Guardian` is non-nil, it is an explicit injected authority and is
used unchanged. If it is nil and the gate is ON, the engine profile constructs
the real Guardian. If the gate is OFF it does not construct one, because an
unset gate has zero limits and the command will not build a loop set.

Alternative rejected: construct the Guardian in `cmd/tossctl` before
`engine.NewContext`. The CLI does not yet own the resolved account or the
journal, so it would necessarily open a second journal or duplicate the startup
sequence.

### D2. Gate limits are transcribed once into a full risk policy

A small pure function starts from `risk.DefaultPolicy()` to retain the existing
minimum reward:risk and re-entry rules, then replaces:

- max order quantity,
- max order notional,
- max open exposure,
- max daily loss,
- max daily loss ratio,
- every money currency.

`RiskBudget` is set equal to the configured maximum daily loss. This is the same
conservative derivation documented by `risk.DefaultPolicy`: it adds no loss
permission beyond the daily ceiling. Float config values are formatted with
Go's shortest round-tripping decimal representation before entering the
string/decimal risk domain.

The policy version is the stable identifier
`engine.automation_gate/risk-policy-v1`.

Alternative rejected: use `risk.DefaultPolicy()` unchanged. It is KRW and would
make a correctly configured USD gate fail the single-source interlock check.

### D3. Identity is structural, not inferred

The local `guardian` variable is passed to `runInterlock`, and only after a
verified result is stored in `Context.Guardian`. `Context.ExitObserver` already
type-asserts that field to `ReductionIssuer`; it will not receive another
constructor or factory.

Tests cover identity and ownership at two levels:

- an engine-package spy proves the object inspected by the interlock is the
  object later used by the exit observer;
- a command-package regression invokes the real assembler with no Guardian
  injection, counts exactly one construction, issues one isolated reduce-only
  decision through the published Guardian, and proves the decision is readable
  through `Context.Journal`. Closing the context must close that shared handle.

### D4. The CLI regression uses build-tagged test seams only

Production `engineAssemble` remains the zero-extra-options wrapper used by
`runEngineRun`. A private helper accepts `official.Option` values and an
unexported `engine.Options` decorator. The production wrapper passes no
decorator.

The command regression is compiled with the dedicated `tossos_testseams` build
tag. Under that tag only, the engine package exposes a helper that lets the
command-package test set the fixed ext4 `journal.FSProber` and count Guardian
construction. Protection remains the shipped `UNWIRED` constant: the regression
issues only a reduction, so overriding it would weaken the fact being tested.
The helper does not exist in the normal binary or normal package API. There is
no flag, environment variable, config key, or production caller that can reach
it.

Alternative rejected: export the journal prober on `engine.Options`. That would
let production callers bypass a durability interlock in order to make one test
portable.

### D5. This change adds no crash-sensitive persistence transition

Guardian construction happens after the existing journal open and before any
loop or decision. It writes no row and performs no rename/replace sequence.
Construction failure follows the existing abrupt-return cleanup shape and closes
the journal. Therefore a child-process crash test would exercise SQLite/journal
durability already owned by `internal/journal`, not a new transition introduced
here. The high-risk crash gate is satisfied by:

- a construction-failure cleanup test proving the opened handle is closed,
- the existing journal crash suite,
- an explicit reviewed not-applicable decision for a new subprocess crash test.

## Risks / Trade-offs

- [A float gate value could be transcribed differently from the audited
  snapshot] → use one round-tripping formatter and table-test all five fields,
  USD, fractional ratios, invalid and non-finite inputs.
- [Injected test Guardians could mask production construction again] → the
  command regression passes no Guardian and asserts the concrete real type.
- [Guardian construction after the journal opens can fail] → close the journal
  on every failure, record a startup refusal, and start no loop.
- [A future exit observer could bypass `Context.Guardian`] → identity regression
  test and existing `ReductionIssuer` construction refusal remain.
- [A command regression could silently bypass the production filesystem guard]
  → compile its durability seam only under `tossos_testseams`, retain shipped
  `UNWIRED` protection, and keep the zero-option production wrapper as the path
  under ordinary builds.

## Migration Plan

No data migration is needed. Deploy the binary with the automation gate OFF,
run isolated and package tests, then let the operator start it under the existing
human-approved gate and limits. Rollback is the previous binary; journal schema
and stored decisions are unchanged.

## Open Questions

None.
