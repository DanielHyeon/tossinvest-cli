## Why

Chrome serializes a same-origin form POST as `Origin: null` when the console
document uses `Referrer-Policy: no-referrer`. The strict origin gate correctly
rejects that opaque origin, so every console write, including `/restart`, fails
before CSRF validation despite being submitted from the canonical HTTPS page.

## What Changes

- Serve console documents with `Referrer-Policy: same-origin`.
- Keep the strict request-origin evaluator unchanged: explicit opaque,
  malformed, or cross-origin evidence remains rejected.
- Add a regression contract proving the response policy supports canonical
  same-origin form submissions without weakening CSP, CSRF, Host, TLS, or peer
  checks.
- Record a browser-level verification using an intentionally invalid CSRF token
  so the probe can prove origin acceptance without invoking a restart or any
  operational handler.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `console-request-origin`: Require console documents to retain same-origin
  request evidence for form POSTs while withholding referrer data from
  cross-origin destinations.

## Impact

- Affected code: `internal/console/remote.go` remote security response headers,
  `internal/console/pages.go` rendered-page headers, and the shared/restart
  document policies in `internal/console/templates.go`.
- Affected tests: console security-header and browser-origin regression tests.
- No API, dependency, configuration, database, journal, engine, order, or
  operating-toggle change.
- The deny-by-default CSP remains unchanged; Chrome DevTools' blocked
  `/.well-known/appspecific/com.chrome.devtools.json` probe remains harmless and
  out of scope.
