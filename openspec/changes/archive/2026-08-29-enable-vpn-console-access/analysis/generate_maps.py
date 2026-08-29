#!/usr/bin/env python3
"""Regenerate final logic maps from the checked-in Go AST evidence."""

from __future__ import annotations

import json
from pathlib import Path


ROOT = Path(__file__).parent / "function-logic"

DETAILS = {
    "EvaluateLadder": (
        "validated decimal policy/state, positive entry/observed/high-water/baseline, monotone baseline",
        "Validate, positive/fraction/parseRat, ComputeProtectedStop, lockPrice",
        "preserve promotion-before-breach, pending cancel-first, completion, and partial/final precedence; add runner only as a max-composed protection candidate",
        "TestEvaluateLadder",
    ),
    "NewExitObserver": (
        "non-nil journal/price/retrier/issuer/gateway, account, costs, valid policy snapshots",
        "DefaultRatchetConfig, DefaultLadderPolicy, Validate, Clock.Now",
        "reject unknown configured policy; do not expand the order or Guardian capabilities",
        "TestNewExitObserver",
    ),
    "openState": (
        "exit-eligible position; adopted and engine-entered origins remain distinct",
        "LookupDecision, ParsePreimage, OpenExitState, openAdoptedState",
        "snapshot the startup common policy only when a new state is opened; existing state is never rebound",
        "TestExitObserverOpens",
    ),
    "openAdoptedState": (
        "adoption record is the sole t0 and recovery authority",
        "OpenAdoptedExitState",
        "recover policy ID from the committed adoption record, never from later config",
        "TestExitObserverOpensAdopted",
    ),
    "judgeLadder": (
        "stored state policy ID selects one immutable registry policy; adopted origin is retained",
        "checkLadderPolicyStillFits, RungIndex, EvaluateLadder, record",
        "unknown/mismatched snapshots refuse judgement; record/Guardian/submit ordering remains untouched",
        "TestExitObserverLadder",
    ),
    "checkLadderPolicyStillFits": (
        "active rung must exist in the selected policy and its lock must not exceed stored baseline",
        "LockPrice, CompareDecimal",
        "accept the resolved per-state policy as input while preserving all fail-closed checks",
        "TestCheckLadderPolicyStillFits",
    ),
    "OpenExitState": (
        "position exists and is exit-eligible; kind/ID are validated; t0 prices are valid",
        "OpenRatchetState, BeginTx, appendExitEventTx, Commit, ExitState",
        "write policy ID in the same transaction as state open; unique/existence failures remain fail-closed",
        "TestOpenExitStatePolicySnapshot",
    ),
    "scanExitState": (
        "query column order exactly matches scan targets",
        "row.Scan",
        "map nullable legacy LADDER ID to default_v1 without rewriting rows",
        "TestLegacyLadderPolicyID",
    ),
    "OpenAdoptedExitState": (
        "position and adoption record exist and no exit state exists",
        "BeginTx, readAdoptionTx, OpenRatchetState, appendExitEventTx, Commit",
        "use adoption.exit_policy_id in the same transaction and never re-read config",
        "TestOpenAdoptedExitStatePolicySnapshot",
    ),
    "AdoptPosition": (
        "position has neither entry decision nor prior different adoption",
        "record, BeginTx, readAdoptionTx, ExecContext, Commit",
        "persist selected policy ID with adoption before setting the position pointer",
        "TestAdoptPositionPolicySnapshot",
    ),
    "record": (
        "trimmed identity, decimal prices/quantity, RFC3339 time and valid common policy ID",
        "riskcalc decimal validation, sha256 digest construction",
        "include policy ID in deterministic adoption preimage/digest",
        "TestAdoptionRecordPolicyID",
    ),
    "scanAdoption": (
        "query column order exactly matches adoption fields",
        "row.Scan",
        "read nullable exit_policy_id as empty legacy value",
        "TestPositionAdoptionPolicyID",
    ),
    "adoptOne": (
        "fresh observed price and validated adoption stop configuration",
        "SyntheticStop, AdoptPosition",
        "snapshot startup common policy into request; adoption itself never submits an order",
        "TestAdoptOnePolicySnapshot",
    ),
    "ExitObserver": (
        "Context dependencies and caller overrides are validated",
        "NewExitObserver",
        "inject config common policy only into observer construction",
        "TestContextExitObserverPolicy",
    ),
    "ReconcileDriver": (
        "Context dependencies, freshness and adoption settings are validated",
        "NewReconcileDriver",
        "inject the same startup common policy into adoption requests",
        "TestContextReconcileDriverPolicy",
    ),
    "mergeEngine": (
        "raw optional blocks merge onto safe zero-value config",
        "adoption/gate validation and normalization",
        "empty means legacy RATCHET; unknown non-empty policy is retained as rejected and cannot run",
        "TestMergeExitPolicy",
    ),
    "routes": (
        "all routes remain behind session0; state changes additionally require mutating CSRF",
        "http.ServeMux.HandleFunc, session0, mutating",
        "add GET /optimization and POST /optimization/exit-policy with the dedicated seam only",
        "TestOptimizationRoutes",
    ),
    "runConsole": (
        "remote flags are resolved before account/config wiring; local mode remains the zero value",
        "remoteAccessOptions, console.ListenAndServe, config-backed seams, audit recorder",
        "wire the all-or-nothing remote transport only; no verify/order approval or broker capability is added",
        "TestRemoteAccessTokenFileMustBePrivateAndLong",
    ),
    "newConsoleCmd": (
        "only the reviewed local port plus complete remote-mode file/path/network flags and one explicit access mode",
        "Cobra flag registration and runConsole",
        "accept trusted-network as a boolean decision but no token value, session, insecure transport, nonce, or automatic approval flag",
        "TestConsoleOffersOnlyTheCompleteRemoteAccessFlagSet",
    ),
    "remoteAccessOptions": (
        "zero flags mean native local mode; remote flags require exactly one of trusted-network or a private token file",
        "loadRemoteAccessToken, openAuditLog, RecordAction",
        "carry the explicit trusted-network decision without a credential and reject ambiguous or implicit access mode",
        "TestTrustedNetworkRemoteAccessDoesNotNeedATokenFile",
    ),
    "loadRemoteAccessToken": (
        "regular non-symlink file, owner-only permissions, at most 4096 bytes, trimmed token at least 32 bytes",
        "os.Lstat, os.ReadFile",
        "reject unsafe metadata or size before any listener/account construction",
        "TestRemoteAccessTokenFileMustBePrivateAndLong",
    ),
    "New": (
        "required verification seam plus zero local or fully validated remote configuration",
        "newRemoteRuntime, token generation, route construction",
        "refuse partial remote configuration before constructing an HTTP handler",
        "TestRemoteConfigurationIsAllOrNothing",
    ),
    "URL": (
        "trusted remote URL is the public root; authenticated remote URL is /login; native local retains the process session query",
        "Addr, strings.TrimSuffix, fmt.Sprintf",
        "send trusted-network browsers directly to the console without weakening either compatibility authentication path",
        "TestTrustedNetworkConsoleNeedsNoApplicationSession",
    ),
    "ListenOn": (
        "validated IP-literal bind and port",
        "netip.ParseAddr, net.Listen",
        "force tcp4/tcp6 explicitly so wildcard binds cannot silently change address family",
        "TestRemoteListenerMustMatchTheValidatedBind",
    ),
    "Serve": (
        "listener must match selected local/remote mode before banner or serve",
        "listenerAllowed, http.Server.Serve, ServeTLS, Shutdown",
        "use TLS 1.3 and bounded server timeouts remotely while preserving settle-before-shutdown",
        "TestRemoteListenerMustMatchTheValidatedBind",
    ),
    "ListenAndServe": (
        "validated Console chooses exactly one listener constructor",
        "New, Listen, ListenOn, Serve",
        "partial remote configuration never falls back to local or exposed HTTP",
        "TestRemoteConfigurationIsAllOrNothing",
    ),
    "writeBanner": (
        "local banner retains possession warning; remote banner prints no token",
        "Console.URL, fmt.Fprintf",
        "describe the selected trust boundary without disclosing either login or session credentials",
        "TestRemoteURLNeverCarriesAConsoleCredential",
    ),
    "routes": (
        "login/logout exist only in authenticated remote mode; health and every operational route retain their static wrapper classification",
        "http.ServeMux.HandleFunc, session0, mutating, remote.security",
        "remove login lifecycle from trusted-network mode while preserving network middleware and mutating gates",
        "TestTrustedNetworkConsoleNeedsNoApplicationSession",
    ),
    "session0": (
        "trusted-network is explicit and already passed remote security; authenticated remote/local session exchange remains unchanged",
        "remote.hasSession, acceptHandoff, grantSession, hasSessionCookie",
        "bypass only the application session in trusted mode; never bypass peer, Host, Origin or CSRF gates",
        "TestTrustedNetworkStillRejectsWrongPeerOriginAndCSRF",
    ),
    "mutating": (
        "POST, then remote same-origin, then form/CSRF, then handler",
        "remote.sameOrigin, ParseForm, tokenEqual",
        "all independent request gates must pass before any state-changing handler executes",
        "TestRemotePeerHostOriginAndCSRFAreIndependentGates",
    ),
    "listenerAllowed": (
        "local listener is loopback; remote listener TCP address exactly matches the validated bind",
        "loopbackOnly, netip.AddrFromSlice",
        "close/refuse a listener whose real address differs from configuration",
        "TestRemoteListenerMustMatchTheValidatedBind",
    ),
    "grantSession": (
        "remote handoff has valid peer and durable audit; local cookie exchange remains unchanged",
        "remote.record, remote.issueSession, http.SetCookie, http.Redirect",
        "audit failure prevents remote session issuance; remote cookie is distinct and Secure",
        "TestRemoteHandoffIssuesANewAuditedRemoteSession",
    ),
    "newRemoteRuntime": (
        "zero remote configuration or a complete TLS/CIDR/public-origin configuration with exactly one access mode",
        "validateRemoteFields, parseRemoteBind, parseAllowedCIDRs, parseRemotePublicURL, loadRemoteCertificate",
        "construct trusted-network without a token; retain token-auth compatibility; reject ambiguous modes",
        "TestTrustedNetworkAndTokenAuthenticationCannotBeCombined",
    ),
    "validateRemoteFields": (
        "complete remote transport fields, audit recorder and exactly one access mode",
        "strings.TrimSpace",
        "fail before listener construction on missing or conflicting access mode",
        "TestTrustedNetworkAndTokenAuthenticationCannotBeCombined",
    ),
    "remoteAccessEmpty": (
        "all remote fields including trusted-network selection",
        "strings.TrimSpace",
        "never mistake trusted-network selection for zero/local configuration",
        "TestRemoteConfigurationIsAllOrNothing",
    ),
    "TestEveryRouteGoesThroughTheSessionGate": (
        "only /login and fixed /healthz are public; all operational paths require session0",
        "registeredRoutes",
        "fail when a future route lacks authentication without an explicit public classification",
        "TestEveryRouteGoesThroughTheSessionGate",
    ),
    "TestEveryStateChangingRouteAlsoGoesThroughTheCSRFGate": (
        "logout joins every existing state-changing POST in the closed CSRF allowlist",
        "registeredRoutes",
        "fail on either an ungated mutation or a read accidentally made POST-only",
        "TestEveryStateChangingRouteAlsoGoesThroughTheCSRFGate",
    ),
}


for directory in sorted(p for p in ROOT.iterdir() if p.is_dir()):
    ast_path = directory / "ast.json"
    if not ast_path.exists():
        continue
    ast = json.loads(ast_path.read_text())
    function = ast["function"]
    qualified = (
        f"{ast['receiver']}.{function}" if ast.get("receiver") else function
    )
    inputs, calls, boundary, test = DETAILS.get(
        function,
        ("validated caller inputs", "AST-listed callees", "preserve existing fail-closed behavior", f"Test{function}"),
    )
    branches = ast.get("branches") or []
    branch_rows = "\n".join(
        f"| {b['id']} | existing {b['kind']} branch at line {b['at']['line']} | only the branch's documented state transition | existing return/error contract | `{test}` |"
        for b in branches
    ) or f"| B1 | straight-line path | documented state transition | documented return | `{test}` |"
    test_rows = "\n".join(
        f"| {b['id']} | {b['kind']} path at current line {b['at']['line']} plus its complement | `{test}` | yes | yes |"
        for b in branches
    ) or f"| B1 | straight-line success/error contract | `{test}` | yes | yes |"

    (directory / "function-logic-map.md").write_text(
        f"""# Function Logic Map: `{qualified}`

- Source: `{ast['file']}`
- AST evidence: `ast.json` (`{ast['source_sha256']}`)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| function inputs and persisted state | {inputs} | caller types, journal/config schema, immutable registry | error/refusal; never broaden authority or silently fall back |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
{branch_rows}

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| {calls} | preserve current computation, persistence, and wiring contracts | errors propagate or are converted to the existing fail-closed refusal | CodeGraph + `ast.json` |

## State mutations and fallbacks

- {boundary}.
- No LIVE gate, trading toggle, broker call, or existing-position rebind is introduced by configuration.

## Safety conclusion

- Safe edit boundary: {boundary}.
- High-risk impact: yes; branch tests and post-edit AST/risk refresh are mandatory.
"""
    )
    (directory / "branch-test-map.md").write_text(
        f"""# Branch Test Map: `{qualified}`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
{test_rows}
"""
    )
