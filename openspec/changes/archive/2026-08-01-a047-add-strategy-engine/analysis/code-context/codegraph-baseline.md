# a047 CodeGraph hard-evidence baseline

Captured from worktree `feat/p1-a047-strategy-engine` at `7cbb36e7` after
`make sdd-sync` on 2026-07-31. CodeGraph is authoritative for graph structure;
the CodeGraphContext report is advisory and was reconciled against these files
and the current tests.

## Production boundaries

| Boundary | Definition | Direct graph evidence | Implementation consequence |
| --- | --- | --- | --- |
| Runtime assembly | `cmd/tossctl/engine.go:engineRuntime` | Calls `engineFillDetector`, `Context.ReconcileDriver`, `Context.ExitObserver`, `Context.Recovery`, and `engine.NewRuntime`. | Do not add an entry loop until every a045/a046/a048/source blocker is wired. Existing reconcile, exit, and fill-detect loops must remain unchanged while the lane is OFF. |
| Runtime supervision | `internal/app/engine/runtime.go:NewRuntime`, `Runtime.Run` | `NewRuntime` has 12 direct callers and `Runtime.Run` impacts 19 symbols, including `runEngineRun` and the supervisor health path. | Prefer a separately constructed dormant entry component. Any eventual supervised loop must preserve cancellation, critical-return, and degradation semantics. |
| Guardian entry authority | `internal/execgw/riskguardian.go:RiskGuardian.IssueEntry` | Calls `scopedIntent`, the pure risk chain, `EntryExposureValue`, `RecordDecisionAndReserveWithRecollection`, and snapshot collection. Impact reaches 28 symbols and Guardian assembly tests. | Strategy orchestration may only present an already-bounded `risk.Intent`; it must not duplicate sizing, reservation, or limit decisions. |
| Official mutation gateway | `internal/execgw/gateway.go:Gateway.Place` | Calls fail-closed preflight, wire serialization, journal dispatch, and official submission. Impact reaches 93 symbols. | The strategy package cannot import the gateway. A narrow orchestrator owns the single Guardian-issued reference and official gateway call. No paper/shadow/canary submitter is introduced. |
| Durable decision | `internal/journal/decision.go:Journal.RecordDecision` and `internal/journal/issuance.go` | `RecordDecision` has 19 direct callers and an 86-symbol impact; Guardian entry uses atomic decision+reservation issuance rather than bare recording. | Candidate/lane/manifest provenance must be bound durably without weakening the existing atomic issuance/reservation contract. |
| Existing end-to-end entry seam | `internal/app/engine/tracer.go:Tracer.submitEntry` | Called only by `Tracer.Run`; calls `EntryIssuer.IssueEntry` and the gateway submitter. | Use only as a behavioral reference. The tracer accepts operator parameters and is not a strategy/runtime consumer. |
| Console operating surface | `internal/console/settings.go:Console.handleSettings` | Routed only from `/settings`; reads config/attestation/engine state. | a047 adds a read-only `strategy-runtime` descriptor/card. Mutating lane/autostart/LIVE controls remain a050-owned server actions. No text, number, textarea, contenteditable, free symbol/reason, or typed-confirmation control. |
| Approved verdict | `internal/candidate/thresholdset.go:ApprovedCandidate` | Production consumers are intentionally absent and enforced by `approved_consumer_guard_test.go`. | `internal/strategy` becomes the first allowlisted pure value-only consumer. It cannot import journal, gateway, broker, callback, channel, pointer, mutable collection, or another authority root. |

## Required regression bindings

- `internal/candidate/approved_consumer_guard_test.go`: permit only the explicit
  pure strategy boundary and keep reverse/transitive authority laundering
  rejected.
- `internal/execgw/riskguardian_race_test.go` and reservation tests: preserve
  submit-time tightening and concurrent aggregate-limit refusal.
- `internal/app/engine/runtime_test.go`: preserve graceful cancellation,
  failure supervision, recovery-before-loop, and degradation behavior.
- `internal/app/engine/seal_test.go` plus dependency guards: keep an
  unauthorized order mutation unspellable from pure lane code.
- console DOM/static tests: the a047 state card is read-only and contains none
  of the arbitrary-input controls forbidden by the change.

## Supporting-context reconciliation

CodeGraphContext highlighted the engine/exit runtime as high fan-in and the
journal exit-state code as high complexity. That supports isolation, but it does
not expand scope: a047 will avoid changing exit-state logic and will keep the
existing exit/reconcile loop set reachable while entry is not configured or
OFF. Ambiguous inferred edges in the supporting report are not used as evidence.
