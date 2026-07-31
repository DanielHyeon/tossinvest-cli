## Context

The remote security wrapper, `Console.render`, and the shared/restart HTML
templates each set `Referrer-Policy: no-referrer`. The console requires
explicit `Origin` evidence before any state-changing handler. A real
Chrome form submission from the canonical HTTPS console was captured with an
intentionally invalid CSRF token. Chrome sent `Origin: null` and no `Referer`;
the request therefore stopped at the origin gate. Repeating the same safe probe
after applying a document-level `same-origin` policy produced the canonical
`Origin` and `Referer` and reached the CSRF gate.

The request-origin evaluator is intentionally fail-closed. An opaque `null`
origin can come from sandboxed or otherwise opaque contexts, so accepting it as
equivalent to a missing header would weaken the boundary.

## Goals / Non-Goals

**Goals:**

- Make canonical HTTPS console forms supply canonical origin evidence in Chrome.
- Preserve strict rejection of explicit `null`, malformed, multiple, and
  cross-origin header values.
- Preserve peer CIDR, exact Host, TLS, CSRF, action audit, and operational
  handler ordering.
- Verify the deployed browser path without invoking a restart or another
  operational action.

**Non-Goals:**

- Changing `sameOrigin` or `sameOriginForMutation`.
- Allowing opaque origins.
- Changing authentication, CSP, networking, Docker port publication, engine,
  orders, risk controls, settings, or operating toggles.
- Allowing Chrome DevTools' optional `/.well-known/appspecific/` connection
  probe through the deny-by-default CSP.

## Decisions

### Use `Referrer-Policy: same-origin`

The console will replace `no-referrer` with `same-origin` in all three policy
surfaces: the remote response wrapper, the rendered-page response header, and
the shared/restart HTML meta declarations. This retains request origin and
referrer evidence only when navigating within the same origin and does not
disclose referrer data to cross-origin destinations. Keeping every setter in
sync prevents a later response header or document meta policy from silently
overriding the wrapper contract.

Alternative considered: treat explicit `Origin: null` as absent and use the
TLS+Host fallback. Rejected because `null` is an explicit opaque origin and the
existing fail-closed contract correctly distinguishes it from header absence.

Alternative considered: loosen the origin gate after validating CSRF. Rejected
because it would reorder independent safety gates and broaden the effect of a
CSRF bug or token leak.

### Keep CSP unchanged

`default-src 'none'`, `form-action 'self'`, and the remaining security headers
stay unchanged. The Chrome DevTools well-known probe warning is a browser-tool
diagnostic, not an application request needed by the console.

### Verify with a non-mutating browser probe

The browser-level proof replaces the rendered CSRF value with an intentionally
invalid token before submitting `/restart`. A correct policy reaches the CSRF
refusal page; the restart handler cannot run. Unit coverage asserts the exact
remote and rendered response headers, both HTML meta declarations, and explicit
`Origin: null` rejection through the full mutation gate. The opaque-origin test
uses canonical direct TLS and Host plus a valid CSRF token, requires the origin
refusal body, and asserts that the wrapped handler is never invoked. This proves
an explicit opaque header cannot activate the headerless fallback or reach an
operational handler.

## Risks / Trade-offs

- [Same-origin pages receive full same-origin referrer URLs] → The console has
  no third-party resources, cross-origin destinations receive no referrer, and
  the benefit is the canonical evidence required by the write gate.
- [A future header change can silently recreate `Origin: null`] → Pin the exact
  policy in a regression test and retain the safe browser probe in verification.
- [A stale page can still hold an old process-local CSRF token after restart] →
  Existing CSRF refusal text instructs the operator to reload; this change does
  not weaken or persist the token.

## Migration Plan

1. Add failing response-header, rendered-document, and opaque-origin contracts.
2. Change only the two response header values and two HTML meta values, then
   run focused, race, and full validation.
3. Rebuild and force-recreate the Compose service.
4. Run the invalid-CSRF Chrome form probe and require canonical `Origin` plus a
   CSRF refusal response.

Rollback is replacement with the previous verified image and restoration of
all four `Referrer-Policy: no-referrer` policy values in `remote.go`, `pages.go`,
and `templates.go`. No data or configuration migration is involved.

## Open Questions

None.
