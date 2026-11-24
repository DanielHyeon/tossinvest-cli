1. **Expiry warning promises exit-only startup for files that will actually prevent startup.** [attestation_watch.go:437](/home/daniel/tossos-work/a199-impl/internal/app/engine/attestation_watch.go:437) unconditionally says restarting after expiry preserves stops and exits.

   Reproducer: matching account, every required endpoint, expiry one hour ahead, and `format_version: 0`. The approved compatibility path accepts this file before expiry because [exit_only_start.go:70](/home/daniel/tossos-work/a199-impl/internal/app/engine/exit_only_start.go:70) returns immediately when `Verify` succeeds. The watch therefore emits `imminent` with the exit-only promise. At expiry, the identical file reaches the format guard at [exit_only_start.go:81](/home/daniel/tossos-work/a199-impl/internal/app/engine/exit_only_start.go:81), returns `ErrMalformed`, and prevents startup entirely. A complete endpoint list containing an additional blank item produces the same contradiction through the guard at line 85.

   An operator relying on that warning can restart after expiry and lose automatic stop-loss execution. Preserve approved pre-expiry acceptance, but derive the forecast from the actual post-expiry classification. Add boundary cases covering both invalid version and extra blank endpoint.

   Severity: Medium  
   Blocks: yes

Static review at `6c30b91b3`; tests not executed.

Recommendation: fix the expiry forecast before merging because it promises continued stop-loss execution for files that the post-expiry interlock rejects entirely.
VERDICT: FINDINGS