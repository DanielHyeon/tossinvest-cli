## Why

The console currently reports that a detached soak started as soon as the child
process is spawned, even when the child immediately exits because Open API
credentials are absent. Operators must discover the real prerequisite in
`soak.log` and then know an undocumented CLI command, so the normal console
workflow cannot recover from first-run credential setup.

## What Changes

- Make one **soak restart** click reuse persisted Open API credentials and start
  or restart the soak without another operator action.
- Detect missing Open API credentials before spawning a detached soak and send
  the operator to a console credential setup screen instead of displaying a
  false success notice.
- Add a narrow form to the configured HTTPS console for API key and secret
  submission. Validate them with the official read-only account probe, persist
  them with the existing protected credential store, and automatically start
  the soak after a successful submission. The legacy plaintext loopback console
  never accepts credential ingress.
- Keep saved credentials across console/container restarts. The existing
  official client remains responsible for access-token refresh, so access-token
  expiry does not ask the operator to enter the key and secret again.
- Never render, redirect, log, audit, retain, or otherwise disclose the submitted
  secret. Credential writes remain behind the existing peer, exact HTTPS
  host/origin, access-mode, request-size, CSRF, and audit gates. In the deployed
  trusted-network mode, authenticated VPN membership remains the access mode and
  no application login is reintroduced.
- Do not add a LIVE order path, change an operating toggle, weaken CSP, or change
  engine/order/risk behavior.

## Capabilities

### New Capabilities

- `openapi-onboarding`: Guided, persistent Open API credential onboarding tied
  to the one-click soak restart workflow.

### Modified Capabilities

- `operator-console`: Replace the blanket prohibition on every credential route
  with one write-only, gated Open API credential setup route whose request
  values can never be read back through the console.

## Impact

- `internal/console`: new credential page/POST handler, route gates, dashboard
  guidance, and deterministic soak restart continuation.
- `cmd/tossctl`: console seams that resolve, validate, save, and audit official
  Open API credentials before starting the soak.
- `internal/official`: existing credential file format and 0600 persistence are
  reused; no schema change is planned.
- Docker deployment: the existing read-write config bind mount preserves the
  saved credential file and secret-free pending-generation marker across
  image/container replacement.
- Security: this adds a credential-writing surface but no credential-reading
  response, broker mutation, LIVE order, operating-toggle, or engine-start
  capability.
