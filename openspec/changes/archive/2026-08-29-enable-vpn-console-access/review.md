## Proposal-freeze review

Date: 2026-07-31

Voices: Manager architecture review, adversarial Security/Eng review, operator
usability review.

### Findings and decisions

1. **Reject: ordinary `ports: 37085:37085` plus the current URL token.**
   The current credential is printed for a loopback browser and remains valid for
   the process. VPN routing or Docker publish would turn terminal possession into
   a remotely replayable bearer URL. The design requires an independent login
   token that never enters URL/log/cookie and issues a bounded session.
2. **Reject: “VPN traffic is already encrypted, so HTTP is sufficient.”**
   VPN membership does not prove application identity, split-tunnel/firewall
   mistakes happen, and browser cookies/actions cross another trust boundary.
   TLS 1.3, hostname validation and Secure cookies are mandatory.
3. **Reject: host wildcard publish with documentation saying to use a firewall.**
   `TOSSOS_VPN_BIND_IP` has no fallback and Compose rendering fails when absent.
   The app also checks peer CIDR and ignores forwarding headers.
4. **Accept with mitigation: container-internal wildcard bind.**
   It is required for bridge networking but is paired with exact host VPN-IP
   publish, required allowed CIDR, TLS/login and minimal container privileges.
5. **Accept: one random application token rather than a user/password database.**
   This remains a single-operator tool. VPN access and a 32-byte independent token
   form separate boundaries; the token is expected to live in a password manager
   and a 0600 file. Multi-user identity/RBAC remains a separate change.
6. **Resolve repudiation gap.**
   Successful login is fail-closed on append-only audit fsync; bounded failures,
   rate-limit and logout are recorded without credential values. Existing
   state-change audits remain authoritative for what was changed.
7. **Resolve session replay gap.**
   The credential is never itself the cookie. Sessions are random, process-local,
   IP/UA-bound, idle/absolute expiring, logout-revocable and replaced after a valid
   restart handoff.
8. **Preserve account safety.**
   Remote auth grants exactly the existing console capability graph. It adds no
   direct broker/order route, changes no verify typed nonce, and bypasses no engine
   startup interlock or LIVE order approval.

### DREAD gate

Every tailored finding with average DREAD ≥7 in `analysis/threat-model.md` has a
named implementation or operator owner. No unresolved P0/P1 design finding remains.

Verdict: proposal approved for RED implementation.

## Implementation review

Date: 2026-07-31

Scope: final Go diff, route/static guards, Dockerfile/Compose render, built image,
dummy-secret TLS health/login smoke, production-file secret scan.

### Findings resolved

1. **High — wildcard address-family ambiguity.** A container smoke showed that
   `net.Listen("tcp", "0.0.0.0:…")` became an IPv6 unspecified listener on this
   host, correctly tripping exact-bind validation. `ListenOn` now selects
   `tcp4`/`tcp6` from the validated literal, and a listener mismatch test covers
   both the accepted and refused paths.
2. **High — Compose secret ownership.** Image UID 10001 could not read a host
   0600 file owned by UID 1000. Compose now requires an explicit non-root host
   UID/GID, the entrypoint refuses UID 0, and its private tmpfs uses the same
   identity. The image default remains 10001:10001.
3. **Medium — rate-limit audit amplification and in-memory growth.** Repeated
   blocked attempts initially could append one audit line per request. The
   implementation now emits one rate-limit event per peer/window and bounds both
   peer-attempt and active-session maps with expiry/oldest eviction.
4. **Medium — remote auth implementation complexity.** Automated Go quality
   review rated the first `remote.go` draft C because configuration and login
   validation were concentrated in two functions. Parsing/certificate validation
   and POST login were separated; the final score is A (92), with no high finding.
5. **Medium — logout audit failure.** Session revocation used to ignore an audit
   write error. It now always revokes/clears the cookie but returns 503 instead of
   claiming a successful audited logout when fsync is unavailable.

### Verification conclusion

- Remote production files, deployment files, `.env.example`, and operations
  documentation contain zero high/critical secret-scan findings. The repository
  wide heuristic scanner reports false positives in historical AST hashes,
  lockfiles and explicit fake test credentials; none is a production credential.
- Built-image inspection confirms non-root user, healthcheck and no source/VCS
  material in the runtime stage. Compose inspection confirms read-only root,
  `cap_drop: ALL`, no-new-privileges and bounded PID/CPU/memory/logging.
- Dummy credential smoke reached healthy `/healthz`, completed remote login with
  a distinct session, opened the console, and wrote a 0600 audit line without
  loading a real account or invoking a broker request.

Verdict: no unresolved P0/P1 implementation finding.

## Requirement-change review: trusted-network access

Date: 2026-07-31
Decision authority: user/operator explicit approval in the active session
Verdict: **APPROVED WITH NETWORK-BOUNDARY CONDITIONS**

The operator clarified that the host loopback console is single-user and that VPN
membership is already authenticated. The requirement therefore changes from
VPN plus application token to loopback/VPN trusted-network access without an
application login. This intentionally supersedes the earlier review decision that
VPN membership was not authentication.

The approval does not extend beyond the configured loopback/VPN network boundary.
Public or wildcard host publish remains prohibited. TLS, actual peer CIDR, exact
Host, same-origin, CSRF, action audit, engine startup interlocks, verification
approval and every LIVE-order invariant remain mandatory.

### Revised threat decision

| Risk | DREAD average | Decision / owner |
|---|---:|---|
| Stolen VPN account/device obtains full console authority | 8.4 | Accepted by the single operator; VPN administrator must revoke the account/device |
| Public/wildcard or wrong-interface publish | 8.2 | Block with required host bind, CIDR, explicit trusted-network flag and Compose validation |
| DNS rebinding or cross-origin POST | 7.6 | Retain exact Host, Origin/Referer and CSRF checks |
| TLS key or broker session disclosure | 7.8 | Retain 0600/read-only secret mounts and secret scanning |
| Accidental trusted-network activation | 8.0 | Reject implicit mode and token/trusted-mode conflicts before opening the socket |

### Pre-Edit Gate for the revision

```text
Pre-Edit Gate:
- change id / task id: enable-vpn-console-access / 6.2-6.4
- target symbols: newRemoteRuntime, validateRemoteFields, remoteAccessEmpty,
  Console.URL, Console.routes, Console.session0, newConsoleCmd, remoteAccessOptions
- CodeGraph definition/callers/callees/impact: complete
- CodeGraphContext candidates and evidence reconciliation: complete; current HEAD wins
- existing behavior: local query-token cookie exchange; remote token login with server session
- Function Logic Map / Branch Test Map: existing change-owned maps; refresh required after implementation
- upstream regression impact: yes; authentication expectations change only for approved local/trusted-network access
- failing tests first: required by task 6.2
- config/DB/journal rollback: no schema change; remove trusted-network flag to return to authenticated compatibility mode
- safety invariant review: network access broadens, LIVE/order/toggle authority does not
- decision: edit allowed after RED evidence
```

## Trusted-network revision implementation review

Date: 2026-07-31

Scope: trusted-network configuration validation, route/session boundary, CLI
wiring, Compose command, operations documentation, and focused regression tests.

### Findings and disposition

1. **Resolved — stale authenticated-mode banner.** The trusted-network runtime
   initially still printed the compatibility token-mode warning. The banner now
   states that there is no application login while explicitly retaining TLS,
   CIDR, exact Host/Origin, CSRF, and action-audit requirements.
2. **Verified — login lifecycle is absent only in trusted-network mode.**
   `/login` and `/logout` are registered only for compatibility token mode;
   trusted-network reaches operational handlers directly after the outer remote
   security middleware. Native local and token-auth compatibility behavior are
   unchanged.
3. **Verified — state-changing routes remain gated.** Removing `session0`
   authentication in trusted-network mode does not remove `mutating`: POST,
   same-origin, form parsing, and CSRF all run before an operational handler.
   Actual peer CIDR and exact Host remain in the outer middleware.
4. **Verified — deployment cannot select ambiguous access modes.** CLI
   construction rejects missing mode and `--trusted-network` plus
   `--remote-token-file`. Compose contains the explicit trusted-network flag
   and no remote-token flag or secret.
5. **Automated code-quality review.** The `code-reviewer` Go quality checker
   reported zero issues for `internal/console`. Manual universal/Go review found
   no new unbounded collection, goroutine, resource, unchecked-error, insecure
   TLS, broad CORS/CSP, or credential-in-source issue in the revision.

Verdict: the trusted-network revision has no unresolved P0/P1 implementation
finding. The accepted residual risk remains compromise of the operator's VPN
account/device.
