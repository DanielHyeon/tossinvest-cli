# Proposal-freeze review

- Change: `wire-production-risk-guardian`
- Base: `676f8aef0f45140bd57ab16d435b5b3c2178aa02`
- Review date: 2026-07-30
- Reviewer: independent adversarial Eng review (`proposal_review`)
- Initial verdict: **HOLD**

## Findings and decisions

### P1 — CLI test could not inject the allowlisted filesystem probe

Accepted. A normal exported `engine.Options` field would weaken the production
durability guard. The design now uses a dedicated `tossos_testseams` build tag:
the helper exists only in the tagged test build, while the ordinary CLI wrapper
passes no decorator and exposes no runtime input.

### P1 — Same-journal ownership was asserted structurally but not behaviorally

Accepted. The command regression must issue a reduce-only decision through the
published production Guardian, read it through `Context.Journal`, count exactly
one construction, and prove context close closes the shared handle.

### P1 — High-risk crash-path gate was not addressed

Accepted with a scoped not-applicable decision. This change adds no persistent
write or multi-step replacement during Guardian construction. It adds an
abrupt-return cleanup test and runs the existing journal crash suite. A new
child-process crash test would only duplicate journal-owned durability behavior.

## Freeze verdict

**GO after strict validation of these accepted changes.** No P0 finding remains.

## Implementation evidence

- `go test ./internal/app/engine -count=1`: pass
- `go test -tags=tossos_testseams ./cmd/tossctl -count=1`: pass; the actual CLI
  assembler constructs one production Guardian with shipped `UNWIRED`
  protection and the real isolated journal
- `go test -race ./internal/app/engine ./cmd/tossctl`: pass
- `go test -race -tags=tossos_testseams ./cmd/tossctl -run
  TestAssembleEngineWiresOneProductionGuardianToTheEngineJournalAndExitObserver`:
  pass
- `go test ./internal/journal -run 'Crash|crash' -count=1`: all three existing
  crash/restart cases pass
- `make test`, `make lint`, `make validate`: pass
- both post-edit Function Logic Map checks: pass

No new subprocess crash test applies: Guardian construction adds no persistence
transition. It either returns the assembled authority or closes the already
opened journal and refuses startup; journal commit/crash semantics remain owned
by the three existing crash tests above.

## Independent implementation review

- Review date: 2026-07-30
- Reviewer: independent `proposal_review` context
- Verdict: **GO**
- P0/P1 findings: none

The reviewer independently traced production assembly from official account
resolution through real journal open, exactly-one Guardian construction, startup
interlock publication, `Context.Guardian`, and `BuildExitObserver`. The review
confirmed all configured USD bounds, official-only gateway construction,
`UNWIRED` exposure-raising refusal, reduce-only behavior, shared-journal cleanup,
and the tagged actual CLI assembly regression.

Independent reruns passed the uncached affected-package race suite, tagged CLI
assembly race regression, journal crash suite, and both strict OpenSpec
validations.
