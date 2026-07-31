# Verification

## RED

- Date: 2026-07-31
- Command:
  `go test ./internal/console -run 'TestRemoteSameOriginEvidencePrecedence|TestHeaderlessRemoteLoginRemainsStrict|TestHeaderlessCanonicalTLSPostReachesCSRFAndHandlerGates|TestRemoteMutationOriginFallbackRejectsIndirectEvidence' -count=1`
- Result: expected failure.
- Observed failures:
  - explicitly empty, whitespace, and multiple Origin values fell through or
    were accepted;
  - multiple Referer values were accepted;
  - a headerless canonical direct-TLS `/restart` POST was rejected by the origin
    gate before reaching the CSRF/handler gates.
- Safety: tests used `httptest` and a counter-only handler. No settings,
  restart, engine, order, journal, or LIVE side effect was invoked.

## GREEN / VERIFY

- Focused GREEN:
  `go test ./internal/console -run 'TestRemoteSameOriginEvidencePrecedence|TestHeaderlessRemoteLoginRemainsStrict|TestHeaderlessCanonicalTLSPostReachesCSRFAndHandlerGates|TestRemoteMutationOriginFallbackRejectsIndirectEvidence' -count=1`
- Result: PASS.
- Post-GREEN AST and risk reports were regenerated for both modified existing
  functions.

## Full verification

- `go test ./internal/console -count=1`: PASS.
- `go test -race ./internal/console -count=1`: PASS.
- `make test`: PASS.
- `make vet`: PASS.
- `make validate`: PASS, 45 OpenSpec items.
- `make sdd-sync`: PASS; CodeGraph hard-evidence fingerprint updated.
- `make sdd-check`: PASS.
- `python3 tools/logic-map/check_analysis.py --change fix-console-origin-fallback`:
  PASS for all modified existing Go functions.
- `python3 tools/pm/generate_master_tracker.py --check`: PASS.
- Advisory note: GBrain was held by its existing single-writer process, so
  synchronization retained its prior freshness and emitted the workflow's
  allowed `advisory busy` warning. CodeGraph and all hard gates were current.
- Independent implementation/security review: APPROVE, no P0/P1 blockers.

## Post-gate delivery procedure

1. Commit and push only this Story/OpenSpec, console predicate, regression test,
   and generated PM tracker scope.
2. Build the existing Compose service and recreate it without changing config,
   engine state, operating toggles, or order state.
3. Verify container health and restart policy.
4. Fetch the deployed HTTPS form, then submit `/restart` without Origin or
   Referer but with a deliberately wrong CSRF token. Reaching the CSRF refusal
   rather than the origin refusal proves the deployed origin gate passed while
   guaranteeing that restart and every downstream mutation remain uncalled.
5. Archive the change, update the Story to the dated archive path, regenerate
   PM trackers, retain the credential-free investigation learning, and rerun
   PM/SDD checks.
