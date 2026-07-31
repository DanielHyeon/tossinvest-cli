# openapi-onboarding Specification

## Purpose
TBD - created by archiving change automate-soak-openapi-onboarding. Update Purpose after archive.
## Requirements
### Requirement: one-action soak onboarding

In configured HTTPS mode, the console SHALL make an explicit soak restart one
continuous operator action.
It SHALL validate persisted official Open API credentials with one read-only
probe before spawning. Ready credentials SHALL proceed directly to soak
restart. Missing or authentication-rejected credentials SHALL route to the
HTTPS Open API setup screen without spawning. A successful setup submission
SHALL validate and persist the submitted credentials and then start the soak
without requiring a second restart click. The ordinary restart-preflight path
and setup-save-continuation path SHALL be mutually exclusive from the first
credential check through the final restart result.

#### Scenario: saved credentials are ready
- **WHEN** the operator clicks soak restart and the persisted credentials pass the official read-only probe
- **THEN** the console invokes soak restart immediately and returns its truthful result without another operator action

#### Scenario: credentials are missing
- **WHEN** the operator clicks soak restart and no persisted or environment credentials are available
- **THEN** no soak process is spawned and the operator is redirected to an HTTPS setup page that explains the required key and secret

#### Scenario: credentials are rejected
- **WHEN** the operator clicks soak restart and the official probe classifies the persisted credentials as authentication-rejected
- **THEN** no soak process is spawned and the operator is redirected to the setup page with replacement guidance

#### Scenario: rejected credentials are environment-managed
- **WHEN** restart preflight rejects credentials supplied by the container environment
- **THEN** no soak process is spawned, no credential file replacement is offered, and the operator receives fixed guidance to update or recreate the container environment

#### Scenario: transient validation failure
- **WHEN** the explicit restart preflight encounters an IP allow-list, rate-limit, server, or transport failure
- **THEN** no soak process is spawned, the failure is classified for the operator, and the key is not misreported as missing

#### Scenario: successful first-time setup
- **WHEN** the operator submits a key and secret that pass the official read-only probe and protected persistence succeeds
- **THEN** the console invalidates the normal token cache, records a secret-free save-success audit event, waits for the old soak to exit, invalidates the normal cache again immediately before spawn, starts the soak in the same POST flow, and redirects to the dashboard only with the successful restart result

#### Scenario: old soak recreates a token during replacement
- **WHEN** an old soak can refresh and recreate the shared normal token cache after credential preflight or save has invalidated it
- **THEN** restart waits for that old process to exit and applies a second token-cache invalidation fence immediately before detached spawn; fence failure starts no new soak and returns a truthful retryable restart error

#### Scenario: concurrent setup and restart requests
- **WHEN** setup and restart requests overlap
- **THEN** one completes its credential check/save/restart sequence before the other begins, so a request cannot validate one credential generation and restart with another

#### Scenario: persistence or audit fails
- **WHEN** submitted credentials validate but protected persistence or the required save-success audit fails
- **THEN** the console reports the classified failure, does not start soak, does not claim success, and does not record save success before persistence; once persistence is attempted it retains a secret-free 0600 pending-generation marker across restarts even if the store returns an error, blocks ordinary restart preflight, reopens the blank setup screen, and clears the marker only after a later valid submission completes persistence, token invalidation, audit, and marker removal

#### Scenario: pending marker cannot be read or removed
- **WHEN** restart preflight cannot read the file-generation marker or completed setup cannot remove it
- **THEN** the console remains fail-closed, starts no soak, and reports fixed retry guidance without exposing credential-derived content

### Requirement: persistent credential and token lifecycle

The console SHALL persist a validated key and secret through the existing
protected Open API credential store and SHALL reuse that store across console
and container restarts. Short-lived access-token expiry SHALL be handled by the
existing official client renewal behavior and SHALL NOT require the operator to
re-enter a stored valid key and secret. The setup path SHALL be offered when
credentials are missing or authentication-rejected, not merely because the
cached access token expired. An incomplete-generation marker SHALL use the same
persistent config mount as the credential file at
`<config-dir>/openapi-onboarding.pending`.

#### Scenario: environment pair while a file generation is pending
- **WHEN** both environment credential variables are non-empty while a file-managed pending marker exists
- **THEN** the environment pair remains authoritative exactly as the existing loader defines, file onboarding is not offered, and the dormant marker is neither rewritten nor cleared nor allowed to affect the environment generation; if the environment pair is later removed, the marker blocks file-mode restart until setup completes

#### Scenario: environment credentials replace a cached generation
- **WHEN** a complete environment credential pair is checked while the normal token cache may contain a valid token issued for an older pair
- **THEN** preflight validates the environment pair through an isolated temporary 0600 token cache, removes that temporary cache, invalidates the normal cache, and requires the post-exit pre-spawn invalidation fence so the child cannot reuse the old generation

#### Scenario: replacement target was permissive
- **WHEN** protected persistence replaces an existing credential path whose mode is more permissive than 0600
- **THEN** the store atomically replaces it with a verified regular 0600 file or reports failure while the pending marker remains and no soak starts

#### Scenario: container replacement
- **WHEN** a validated credential file was saved through the console and the Docker container is recreated with the existing config bind mount
- **THEN** the next soak restart reuses the saved credentials without showing the setup form

#### Scenario: cached access token expires
- **WHEN** the cached official access token is expired but the persisted key and secret remain valid
- **THEN** the official client renews the token during the restart preflight and the soak starts without credential re-entry

### Requirement: credential ingress is secret-safe

The Open API setup GET SHALL return blank credential fields and SHALL NOT expose
stored credential values. Credential submission SHALL be accepted only through
a bounded POST body over direct TLS and behind the configured HTTPS peer, exact Host/origin,
configured access mode, method, CSRF, and audit boundaries. Token-authenticated
mode SHALL require its application session. Trusted-network mode SHALL continue
to use authenticated VPN membership as the application access boundary and
SHALL NOT add a separate application login. The console SHALL NOT place the key
or secret in HTML responses, redirect URLs, error text, application logs, audit
details, retained memory, or test output. Failed validation or persistence SHALL
not start a soak and SHALL not persist a rejected pair.

#### Scenario: setup form render
- **WHEN** an authenticated operator opens the Open API setup page
- **THEN** the page contains blank password-type key and secret inputs and contains no stored credential material

#### Scenario: plaintext loopback request
- **WHEN** an otherwise authenticated request opens or submits the Open API setup route without direct TLS
- **THEN** the console returns 403 before parsing or exposing any credential field and performs no validation, persistence, audit, or soak restart

#### Scenario: missing CSRF
- **WHEN** a credential submission has valid field values but lacks the current CSRF token
- **THEN** the request is rejected before validation, persistence, audit, or soak restart

#### Scenario: cross-origin submission
- **WHEN** a credential submission originates outside the configured exact HTTPS origin
- **THEN** the request is rejected before credential parsing or any side effect

#### Scenario: oversized submission
- **WHEN** a credential submission exceeds the setup request-body limit
- **THEN** the route-specific body limit deterministically returns 413 before mutation middleware parses the form and without bypassing peer, Host/origin, configured access mode, method, CSRF, or authorization checks and without validation, persistence, audit, or soak restart

#### Scenario: deployed trusted-network submission
- **WHEN** an allowed VPN peer submits from the exact configured HTTPS origin in trusted-network mode with a valid CSRF token
- **THEN** the request reaches credential validation without an application login, while a wrong peer, Host/origin, or CSRF value remains rejected before the save seam

#### Scenario: invalid credentials
- **WHEN** submitted credentials are rejected by the official read-only probe
- **THEN** neither field is echoed or persisted, no soak starts, and the page shows fixed replacement guidance

#### Scenario: replacement credentials with an existing valid token
- **WHEN** the operator submits a replacement key and secret while the normal token cache still contains a valid token for the old pair
- **THEN** validation uses an isolated temporary 0600 token-cache path, the old token cannot authenticate the submitted pair, and the temporary cache is removed

#### Scenario: saved-token refresh fails
- **WHEN** access-token refresh fails with an authentication or transient official-client classification
- **THEN** soak remains stopped and only fixed classified guidance, with no credential-derived content, is returned

