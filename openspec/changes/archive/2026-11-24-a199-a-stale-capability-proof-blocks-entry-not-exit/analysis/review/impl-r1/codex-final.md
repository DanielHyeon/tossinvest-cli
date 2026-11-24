codex
1. **P1 — Invalid chronology reaches exit-only (Q2).** `internal/attest/attest.go:211–237` never checks `IssuedAt`; `internal/app/engine/exit_only_start.go:75–96` accepts expiry/scope failures after checking version, endpoint items and account. Thus `issued_at > expires_at`—or missing issuance—can become exit-only. This is a gap in the design’s explicit predicates too, despite its “only stale” intent.
   **Repro:** Matching account, version 1, valid endpoints, expiry yesterday, issuance tomorrow; startup succeeds exit-only.

2. **P0 consequence — Exit-only audit failure prevents stop-loss runtime (Q4).** `internal/app/engine/exit_only_start.go:118–126` sets the latch, then returns an audit error; `internal/app/engine/engine.go:589–591` closes the journal and returns no Context. No exit/reconcile loops start. This follows the existing acceptance-audit policy (`interlock.go:426–433`), rather than introducing an entry bypass.
   **Repro:** Allow settings audit, then inject an append/fsync failure specifically for `automation_gate.exit_only`; observe startup error and zero stop-loss submissions.

3. **P2 — New runtime regression test hangs on early runtime failure.** `internal/app/engine/a199_exit_only_runtime_test.go:198–199` consumes the sole `done` result before `Fatalf`; deferred cleanup at lines 190–192 receives from that channel again forever. The test masks its useful failure with a suite timeout.
   **Repro:** Make recovery return an error before the first sell; test blocks in deferred cleanup.

4. **Q1 — No reachable exposure bypass found in current production assembly.** Buy/raising-amend classification: `internal/execgw/gateway.go:381,451`; latch checks: lines 592,669,859–868. The same gate is constructed/passed at `internal/app/engine/gateway.go:265,328,375` and `engine.go:586,610`; latch installation precedes Context return (`exit_only_start.go:118`, `engine.go:602`) and loop/auxiliary launch (`runtime.go:294–324`). Strategy workers still require `Verified` (`strategy_entry_supervisor.go:469`).
   `Clear` can technically remove any named latch (`internal/execgw/retry.go:569–579`), but no production caller clears this reason. Symbol clearing touches only symbol blocks (`symbolgate.go:106`); reconcile rebuild removes only its two reasons (line 184); mode projection removes only its reason (`modegate.go:57`); alert ack clears only alert-undelivered (`internal/obs/notifier.go:981`). Replay bypasses normal submission but is disabled by `Attested:nil` (`internal/app/engine/gateway.go:338`, `internal/execgw/replay.go:260`).
   **Probe:** With protection wired, submit buy/quantity-up/price-up amendments after recovery, reconciliation, mode release and alert ack; expect zero broker calls.

5. **Q2 — Remaining requested input boundaries behave as designed.** Expired foreign/masked accounts fail the shared digit comparison (`exit_only_start.go:92`, `internal/attest/account_match.go:10`, `attest.go:296`). Blank endpoint items and stale version 0/negative fail (`exit_only_start.go:81–88`); too-new versions fail Load (`attest.go:174`); zero expiry/empty lists fail Verify (lines 215–219). Fresh files passing Verify still bypass the extra guards, intentionally (`exit_only_start.go:70–73`).
   No normal issuer-shaped false refusal found: issuer writes version/full account/positive validity/endpoints (`internal/soak/attest.go:261–267`); expiry, missing GET/POST, both together, and hyphen-only account differences reach exit-only.
   **Probe:** Table-drive those shapes through startup, including raw JSON version 0 rather than `Save`, which replaces zero.

6. **Q3 — No skipped clause found.** `internal/app/engine/interlock.go:521–537` preserves Guardian, limits and trading-policy refusals; lines 568–585 preserve limit-source, daily-buy, open-position, gateway and transport checks. Staleness is committed only afterward (line 603). Errors retain the clause sentinel plus existing wrappers, without a stale sentinel. Missing/invalid 7b/7c limits use `ErrLimitsRequired`; declaration mismatches use `ErrGuardianLimitsMismatch` (`daily_buy_interlock.go:23–36`, `open_positions_interlock.go:21–33`). Ordinary non-stale verification order/messages/audit actions remain; structured logs add `gate_exit_only` (`interlock.go:481`).
   **Probe:** Combine stale with each isolated clause failure and assert its sentinel, absent stale sentinel, and no exit-only audit.

7. **Q4 — Restart/rewrite behavior matches frozen startup semantics.** Each construction creates a fresh gate and reclassifies the file (`gateway.go:265`, `interlock.go:541,561`): stale restart re-latches; repaired restart takes normal acceptance. Watch snapshots the boot marker (`attestation_watch.go:237`) and independently rereads the file (line 331); replacement changes notifications, not the boot verdict/latch. Stale→healthy therefore remains exit-only until restart; healthy→stale does not dynamically demote a running verified engine, as D6 specifies.
   **Probe:** Replace the file between interlock and first watch read in both directions; assert unchanged boot status and latch.

Read-only source review; no tests or live operations executed. ai-memory retrieval unavailable because local-server access was denied.
hook: Stop
hook: Stop
hook: Stop Completed
hook: Stop Failed
tokens used
153,947
1. **P1 — Invalid chronology reaches exit-only (Q2).** `internal/attest/attest.go:211–237` never checks `IssuedAt`; `internal/app/engine/exit_only_start.go:75–96` accepts expiry/scope failures after checking version, endpoint items and account. Thus `issued_at > expires_at`—or missing issuance—can become exit-only. This is a gap in the design’s explicit predicates too, despite its “only stale” intent.
   **Repro:** Matching account, version 1, valid endpoints, expiry yesterday, issuance tomorrow; startup succeeds exit-only.

2. **P0 consequence — Exit-only audit failure prevents stop-loss runtime (Q4).** `internal/app/engine/exit_only_start.go:118–126` sets the latch, then returns an audit error; `internal/app/engine/engine.go:589–591` closes the journal and returns no Context. No exit/reconcile loops start. This follows the existing acceptance-audit policy (`interlock.go:426–433`), rather than introducing an entry bypass.
   **Repro:** Allow settings audit, then inject an append/fsync failure specifically for `automation_gate.exit_only`; observe startup error and zero stop-loss submissions.

3. **P2 — New runtime regression test hangs on early runtime failure.** `internal/app/engine/a199_exit_only_runtime_test.go:198–199` consumes the sole `done` result before `Fatalf`; deferred cleanup at lines 190–192 receives from that channel again forever. The test masks its useful failure with a suite timeout.
   **Repro:** Make recovery return an error before the first sell; test blocks in deferred cleanup.

4. **Q1 — No reachable exposure bypass found in current production assembly.** Buy/raising-amend classification: `internal/execgw/gateway.go:381,451`; latch checks: lines 592,669,859–868. The same gate is constructed/passed at `internal/app/engine/gateway.go:265,328,375` and `engine.go:586,610`; latch installation precedes Context return (`exit_only_start.go:118`, `engine.go:602`) and loop/auxiliary launch (`runtime.go:294–324`). Strategy workers still require `Verified` (`strategy_entry_supervisor.go:469`).
   `Clear` can technically remove any named latch (`internal/execgw/retry.go:569–579`), but no production caller clears this reason. Symbol clearing touches only symbol blocks (`symbolgate.go:106`); reconcile rebuild removes only its two reasons (line 184); mode projection removes only its reason (`modegate.go:57`); alert ack clears only alert-undelivered (`internal/obs/notifier.go:981`). Replay bypasses normal submission but is disabled by `Attested:nil` (`internal/app/engine/gateway.go:338`, `internal/execgw/replay.go:260`).
   **Probe:** With protection wired, submit buy/quantity-up/price-up amendments after recovery, reconciliation, mode release and alert ack; expect zero broker calls.

5. **Q2 — Remaining requested input boundaries behave as designed.** Expired foreign/masked accounts fail the shared digit comparison (`exit_only_start.go:92`, `internal/attest/account_match.go:10`, `attest.go:296`). Blank endpoint items and stale version 0/negative fail (`exit_only_start.go:81–88`); too-new versions fail Load (`attest.go:174`); zero expiry/empty lists fail Verify (lines 215–219). Fresh files passing Verify still bypass the extra guards, intentionally (`exit_only_start.go:70–73`).
   No normal issuer-shaped false refusal found: issuer writes version/full account/positive validity/endpoints (`internal/soak/attest.go:261–267`); expiry, missing GET/POST, both together, and hyphen-only account differences reach exit-only.
   **Probe:** Table-drive those shapes through startup, including raw JSON version 0 rather than `Save`, which replaces zero.

6. **Q3 — No skipped clause found.** `internal/app/engine/interlock.go:521–537` preserves Guardian, limits and trading-policy refusals; lines 568–585 preserve limit-source, daily-buy, open-position, gateway and transport checks. Staleness is committed only afterward (line 603). Errors retain the clause sentinel plus existing wrappers, without a stale sentinel. Missing/invalid 7b/7c limits use `ErrLimitsRequired`; declaration mismatches use `ErrGuardianLimitsMismatch` (`daily_buy_interlock.go:23–36`, `open_positions_interlock.go:21–33`). Ordinary non-stale verification order/messages/audit actions remain; structured logs add `gate_exit_only` (`interlock.go:481`).
   **Probe:** Combine stale with each isolated clause failure and assert its sentinel, absent stale sentinel, and no exit-only audit.

7. **Q4 — Restart/rewrite behavior matches frozen startup semantics.** Each construction creates a fresh gate and reclassifies the file (`gateway.go:265`, `interlock.go:541,561`): stale restart re-latches; repaired restart takes normal acceptance. Watch snapshots the boot marker (`attestation_watch.go:237`) and independently rereads the file (line 331); replacement changes notifications, not the boot verdict/latch. Stale→healthy therefore remains exit-only until restart; healthy→stale does not dynamically demote a running verified engine, as D6 specifies.
   **Probe:** Replace the file between interlock and first watch read in both directions; assert unchanged boot status and latch.

Read-only source review; no tests or live operations executed. ai-memory retrieval unavailable because local-server access was denied.
