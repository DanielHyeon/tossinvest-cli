## Context

`tossctl openapi login` is not an OAuth/browser handshake. It accepts an Open
API key and secret, then stores them in the existing protected credential file.
The detached soak loads that file at startup and exits immediately when it is
absent. Today `restartSoak` reports success after process creation without
checking this prerequisite, which produces a false dashboard notice.

The console already owns its configured access mode, exact HTTPS origin/host and
peer checks, CSRF gate, no-store response policy, and action audit. The Docker
deployment bind-mounts the CLI config directory read-write, so the existing
credential file is already the correct persistent store. The official client
already renews short-lived access tokens from the persisted key and secret.

This is an authentication-path change and is therefore High-risk under
`docs/WORKFLOW.md`, even though validation is read-only and the soak cannot
mutate an account.

## Goals / Non-Goals

**Goals:**

- Make the normal saved-credential path a single soak-restart click.
- In configured HTTPS mode, route missing or authentication-rejected credentials
  to a clear setup page before any detached process is spawned.
- Validate and persist a submitted key/secret, then start the soak in the same
  POST flow so the operator does not click restart again.
- Preserve existing token renewal and container persistence.
- Keep secret values out of every response, URL, log, audit detail, test output,
  and retained development evidence.

**Non-Goals:**

- Issuing or recovering Open API keys from the Toss developer portal.
- Adding OAuth, QR, WTS session login, or a credential-reading endpoint.
- Starting a real soak, engine, changing an operating toggle, or placing a LIVE
  order during tests.
- Changing the official credential JSON schema, token renewal algorithm, CSP,
  VPN trust boundary, or Docker mount layout.

## Decisions

### 1. Use a console-owned onboarding state machine, not a shell subprocess

The console receives narrow seams for checking official credentials and
validating/saving a submitted pair. `internal/console` never imports Cobra,
executes `tossctl`, resolves credential paths, or touches the credential file
directly. One console mutex serializes the ordinary restart-preflight path and
the setup-save-then-restart path from the first credential check through the
final restart result, so concurrent clicks cannot validate one credential
generation and start with another. `cmd/tossctl` implements the seams with the same
`resolveOpenAPIPaths`, `validateOpenAPICredentials`, and
`saveOpenAPICredentials` functions used by the CLI.

Alternative: run `tossctl openapi login` as a child process. Rejected because
command-line secrets can be exposed through process listings, the command does
not validate before saving, and subprocess output would create another secret
handling surface.

### 2. Preflight every explicit soak restart, but do not poll on dashboard loads

One soak-restart click loads the saved credentials and performs the existing
official `Accounts` read-only probe. Ready credentials proceed directly to
`RestartSoak`. Missing credentials or an authentication rejection redirect to
the setup page without spawning. IP allow-list, rate-limit, server, and
transport failures stop the restart with their specific guidance and do not
misclassify the key as absent.

This costs one official read per explicit restart, not per dashboard refresh,
and prevents known first-cycle failures from being reported as success.

Alternative: check only whether the file exists. Rejected because an expired or
revoked key would still spawn a child that exits immediately.

### 3. Successful credential submission continues directly into soak start

`GET /openapi/login` renders an empty key/secret form. The form submits to the
separately enumerated `POST /openapi/login/save` route so static route guards
can prove the write path has the mutation gate. That POST
validates the submitted pair with the official read-only probe using an
isolated, newly named token-cache path. This prevents a still-valid token from
the currently stored credential pair from falsely authenticating a replacement
pair. The temporary token cache is created only with 0600 protection and is
removed after validation.

After validation succeeds, the handler atomically replaces the credential file
through the existing store and verifies that the result is a regular 0600 file,
invalidates the normal token cache, records a fixed-detail
`openapi.credentials.saved` audit event, and only then calls `RestartSoak`.
Success means every preceding step succeeded and `RestartSoak` itself returned
success; it uses the existing 303 dashboard redirect with that truthful result.
Validation, persistence, token-cache invalidation, audit, or restart failure
never produces a "started" notice. A persistence failure has no success audit.
If cache invalidation or audit later fails, the response says
"자격증명은 저장됨, soak은 시작되지 않음" with fixed retry guidance; the
protected credentials remain saved. A fixed, secret-free 0600 pending-generation
marker is written before persistence and removed only after normal-token
invalidation and the save-success audit both complete. While it remains,
ordinary restart preflight cannot start soak and reopens the blank setup form;
a valid resubmission retries the complete sequence and clears the marker only
after success.

`RestartSoak` first signals and waits for every old soak. After the old
processes have exited and immediately before detached spawn, it invokes a
second normal-token invalidation fence. This closes the cross-process window in
which an old soak could recreate a token for the previous credential
generation after setup preflight removed the cache. Fence failure prevents
spawn and returns a truthful restart error; the next explicit restart retries
the same fence.

Once credential persistence has been attempted, even an error retains the
marker because a store may have partially changed before returning that error.
Marker read or removal failure also remains fail-closed. Its canonical path is
`<config-dir>/openapi-onboarding.pending`, which is
`/var/lib/tossos/config/openapi-onboarding.pending` in Compose and therefore
shares the credential file's persistent bind mount.

If credential setup completes but `RestartSoak` itself fails, the credential
generation remains complete and the existing truthful restart error is shown;
the next ordinary restart may reuse that validated generation.

The form never repopulates either field after failure. Access-token expiry does
not enable this screen by itself because the official client renews access
tokens from the saved long-lived credentials.

Alternative: save then ask the operator to press soak restart again. Rejected
because the requested outcome is one continuous onboarding action.

### 4. Treat the form as a bounded secret-ingress route

Both credential routes additionally require a direct TLS request. The GET route
is protected by `session0`, which means an application session in token mode and
the already approved VPN-membership boundary in trusted-network mode. The POST
route adds the mutation gate, preserving exact HTTPS origin/host, allowed peer,
CSRF, and method checks without reintroducing an application login in the
deployed trusted-network configuration.
Because the existing mutation middleware calls `ParseForm`, its
backward-compatible optional body-limit parameter installs `MaxBytesReader`
before that parse for this route only. All existing routes omit the parameter
and preserve their behavior; peer, exact Host/origin, configured access mode,
method, CSRF, and authorization checks remain mandatory. Oversize requests deterministically
return 413. The key and secret exist only
in the bounded request body and the narrow setup seam. Errors exposed to HTML
are controlled classifications and fixed guidance, never raw request values.
Audit details are fixed strings.

The legacy `127.0.0.1` plaintext console keeps serving its existing non-secret
operations but refuses both credential routes with fixed HTTPS guidance. This
feature's guided first-run form is therefore a configured HTTPS-console
capability; operators deliberately running the legacy plaintext mode retain the
existing terminal `tossctl openapi login` fallback.

Alternative: include the API key in an error form or redirect for convenience.
Rejected because even the key is credential material and URLs/browser history
are not an approved store.

### 5. Preserve the current credential and container persistence contract

The existing credential file remains the only file-persisted source and retains
0600 permissions. In Docker it remains under `/var/lib/tossos/config`, backed by
the existing host config bind mount. Environment credentials keep their
existing precedence. If rejected credentials came from the environment, the
console does not offer an ineffective file replacement or claim success; it
shows fixed guidance naming `TOSSCTL_OPENAPI_KEY` and
`TOSSCTL_OPENAPI_SECRET` (never their values) and instructing the operator to
update/recreate the container environment. No database,
journal, config schema, or environment-variable contract changes.

The pending marker belongs only to a file-managed credential generation. A
complete environment pair remains authoritative exactly as
`official.LoadCredentials` defines, so file onboarding is neither offered nor
used while that pair is present. The marker remains dormant rather than being
rewritten or deleted; if the environment pair is later removed, file-mode
preflight sees the marker and requires setup completion before starting.
Environment preflight validates through an isolated temporary token cache and,
only after that pair passes, invalidates the normal cache before process spawn.
This prevents a valid token issued for a previous environment generation from
authenticating the replacement pair or being reused by the child.

## Risks / Trade-offs

- **[Credential form expands the console attack surface]** → Keep the route
  behind all existing remote/access-mode/mutation gates, cap its body, never echo
  inputs, use password fields, and add route/static/security regression tests.
- **[Official validation spends rate budget]** → Run exactly once per explicit
  restart or login submission; never during dashboard polling.
- **[Temporary official outage could block a requested restart]** → Fail closed
  with the classified outage message. A false success and immediately dead soak
  is worse than an explicit retry.
- **[Persistence succeeds but success audit fails]** → Report the distinct
  "saved, not started" partial state, do not start soak, retain the 0600
  pending-generation marker across restarts, and let only a later valid
  submission complete and clear it. Never record a save-success event before
  persistence succeeds.
- **[Existing token authenticates a replacement pair]** → Validate submitted
  credentials with an isolated temporary token cache, remove it after use, and
  invalidate the normal cache only after protected persistence succeeds.
  Validate authoritative environment credentials through the same isolation
  boundary and invalidate the normal cache after successful preflight. After
  the old soak exits, invalidate once more immediately before spawn so that
  the prior process cannot recreate a stale generation in the intervening
  window.
- **[Existing credential file has permissive mode]** → Write to a new 0600
  temporary file, fsync it, atomically replace the target, verify the target is
  a regular 0600 file, and fail closed before marker completion or spawn on any
  error.
- **[Rejected credentials are environment-managed]** → Do not save a file that
  cannot take precedence. Show container-environment repair guidance and keep
  soak stopped.
- **[Concurrent restart/setup clicks cross credential generations]** →
  Serialize restart preflight and setup continuation under one console mutex
  through the final restart result.
- **[Token refresh fails]** → Treat the official probe result as the existing
  fixed auth/transient classification, keep soak stopped, and never include
  credential-derived response content in operator text or logs.
- **[A child may fail for a reason after successful credential preflight]** →
  This change removes the observed credential-startup false positive. Other
  post-spawn lifecycle readiness is separate process-supervision scope and is
  not presented as solved.

## Migration Plan

1. Deploy the console routes and seams without changing credential paths.
2. Existing valid credential files continue to work and make soak restart a
   single click.
3. Operators without a credential file are redirected to the setup page on
   their next soak restart when using the configured HTTPS console.
4. Before rollback, confirm `openapi-onboarding.pending` is absent. If present,
   complete the setup with the current image; if completion is impossible, stop
   the service and quarantine the pending marker, credential file, and token
   cache together so an older image cannot ignore an incomplete generation.
5. Deploy the previous image. A completed existing credential file is compatible
   and remains protected; no schema migration is required.

## Open Questions

None. The operator explicitly selected persistent credentials with guided setup
only when missing/rejected, followed by automatic soak start.
