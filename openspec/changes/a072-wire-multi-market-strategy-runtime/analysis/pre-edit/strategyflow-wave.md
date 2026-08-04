# a072 strategyflow wave — pre-edit evidence and branch matrix

## Scope and frozen base

- Implementation base captured for this wave: `171739a4c25153bc304020f0e76acbad9346215b`.
- The shared worktree advanced while parallel a072 work was running. This wave owns only the new
  `internal/strategyflow` pure composition package, the minimum read-only weekly plan accessor, and
  this a072 analysis/status material. Journal, Gateway, protection, engine, command and deployment
  files are explicitly out of scope.
- Function Logic Map: not-applicable. Production composition is implemented only in new functions.
  The sole existing-file Go change adds a read-only leaf accessor and changes no existing function
  body, branch, mutation, fallback, callback or side effect.

## Hard evidence

`codegraph sync .` completed before implementation. Direct CodeGraph queries and current source
inspection established:

| Symbol | Current location | Current caller evidence | Authority boundary |
| --- | --- | --- | --- |
| `strategyrouter.Route` | `internal/strategyrouter/router.go:195` | no production caller before this wave | pure route result; zero mutations |
| `continuationlane.EvaluateKR/US` | `internal/continuationlane/evaluator.go:179/183` | lane tests only before this wave | pure lane decision/refusal |
| `reversallane.EvaluateKR/US` | `internal/reversallane/evaluate.go:10/15` | lane tests only before this wave | pure lane decision/refusal |
| `weeklyvaluelane.EvaluateKR/US` | `internal/weeklyvaluelane/evaluate.go:96/100` | lane tests only before this wave | pure lane decision/refusal |
| `strategydispatch.Dispatch` | `internal/strategydispatch/dispatch.go:251` | no production caller reported | imports journal and official Gateway capability |

The source digests inspected before composition were router
`21fde4fa78acf9988c95e479b375dbb75308f4cb0ccb08435185ff9339c5d6fc`, continuation evaluator
`f72455b6b02c4c781036d2167d6b2d704277605dc47efab6aaec62bd34146269`, reversal evaluator
`efba87285b101307111b9792804df64eef71e5361586191e0512627f6453ef46`, weekly evaluator
`1c19eba0315bb8cb4a2d8eaf543964a8746fbf47e968ab4106665d75b0261125`, and dispatch
`43b611261cd8e13ced3490f883c9d1201e1fcf4a128faf3bde39c5d7f27d58eb`.

CodeGraph showed that `strategydispatch.Dispatch` closes over `internal/journal`, `internal/execgw`
and `internal/strategyengine`. Therefore it cannot be used as the approved-candidate-to-lane seam.
The composition boundary is frozen immediately before Guardian/dispatch, with a static dependency
test rejecting journal, Gateway, trading, config, engine and operating-writer packages.
CodeGraphContext refresh completed as supporting evidence; it did not expand authority or override
the direct source/CodeGraph findings. File memory and GBrain recall returned no prior result.

The router and lane packages intentionally expose no production authority-minting constructor. Real
cross-package acceptance is therefore tested only under the repository's explicit
`tossos_testseams` build tag. The tagged bridge uses the real private seals to create deterministic
KR/US fixtures; normal builds omit the bridge. The integration test then calls public production
`strategyflow.Evaluate`, which fixes `strategyrouter.Route` and `defaultRegistry`, and reaches all six
real evaluators: KR/US continuation, KR/US reversal and KR/US weekly value. This closes the acceptance
gap without adding a normal-build activation, cap or evidence-authorization constructor.

## Current logic map (unchanged functions)

1. `ApprovedSnapshot.Valid` and opaque accessors establish the exact approved-candidate value.
2. `strategyrouter.Route` validates owner scope and market lifecycle. An existing active owner pins
   horizon/lane/version/campaign; otherwise it selects one enabled eligible candidate or returns a
   typed invalid/disabled/ambiguous/scope/version refusal.
3. The selected market/lane binding invokes exactly one market-specific pure evaluator.
4. Each evaluator validates its sealed plan, evidence/config scope, risk cap and leg state, then
   returns either the first native refusal or an entry/add decision with campaign/leg/risk lineage.
5. `strategydispatch.Dispatch` is a later authority boundary and is not reachable from strategyflow.

## Frozen branch-to-scenario matrix for this wave

| Branch | KR scenario | US scenario | Expected result | Authority count |
| --- | --- | --- | --- | --- |
| paired registry | continuation, reversal, weekly present | continuation, reversal, weekly present | exact six descriptors; both markets `OFF/UNOBSERVED` | Guardian/broker/mutation 0 |
| approved candidate invalid/scope mismatch | reject KR before route | reject US before route | typed candidate/scope refusal | lane/Guardian/broker 0 |
| router refusal | market/horizon not selectable | market/horizon not selectable | preserve first native router code; skip lane | lane/Guardian/broker 0 |
| new campaign route | exact KR candidate evidence/config | exact US candidate evidence/config | exact market lane evaluates once | Guardian/broker/mutation 0 |
| existing owner route | campaign pinned to current KR owner | campaign pinned to current US owner | lane campaign must equal router campaign | Guardian/broker/mutation 0 |
| lane input substitution | US input presented to KR binding | KR input presented to US binding | typed unsupported binding; skip lane | Guardian/broker/mutation 0 |
| lane refusal | native KR code retained | native US code retained | sealed partial candidate/router/lane lineage | Guardian/broker/mutation 0 |
| accepted lineage mismatch | candidate/evidence/config/campaign substituted | same | typed lineage mismatch; incomplete | Guardian/broker/mutation 0 |
| accepted lineage incomplete | campaign/leg/ceiling/risk absent | same | typed lineage incomplete | Guardian/broker/mutation 0 |
| accepted complete | candidate→router→lane→campaign/leg/risk | same | sealed complete immutable lineage | Guardian/broker/mutation 0 |
| one market waits/refuses | no effect on US binding/state | no effect on KR binding/state | independent result; no combined market scope | Guardian/broker/mutation 0 |

The broader a072 matrix for leases, A→B→A generation drift, fencing, broker idempotency, restart,
worker isolation, central fallback and safety-loop continuity remains owned by the isolated-runtime,
dispatch and engine integration waves. Strategyflow intentionally has no representation capable of
granting those authorities.
