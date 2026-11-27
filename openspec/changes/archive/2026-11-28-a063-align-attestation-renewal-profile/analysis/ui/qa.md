# Isolated UI QA status

Synthetic fixtures were exported with a temporary `go -overlay` test source
that reused the real console harness; shared source was unchanged. The real
gstack browser rendered `file:///tmp/a063-fixture/*.html` only.

- `refused-desktop.png`, `refused-mobile.png`, and `refused-tablet.png` show a
  refused renewal, bounded streak reason, and 72-hour / 48-hour expiry warning.
- `unknown-desktop.png` shows `renewal 진단: unknown` with no `0001` timestamp.
- Browser text inspection found no console errors. `browse stop` returned exit
  1 after screenshots because its local browser server crashed during shutdown;
  screenshots were already written. No production endpoint was accessed.

Actual rendered-template proof remains isolated Go `httptest` coverage:

- `TestDashboardShowsBoundedRenewalDiagnosticWithoutChangingAttestationUsability`
  renders refused renewal diagnostics and the 72-hour warning.
- `TestDashboardRedactsMalformedAttestationContents` renders invalid
  attestation input without raw content or a zero timestamp.
- `TestDashboardMismatchedRenewalStatusAndBlankPathAreUnknownWithoutZeroTimestamp`
  renders `unknown` without `0001` and checks blank-path state.

Latest focused console test run completed successfully before this note. No
production console, service, survey, order, or network broker was used.
