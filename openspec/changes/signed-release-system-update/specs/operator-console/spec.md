## ADDED Requirements

### Requirement: Settings stages only an authenticated signed release
The authenticated settings page SHALL display a “check and stage signed
release” action when production release download is wired.
`POST /settings/system-update/download` SHALL require the normal console session
and CSRF token and SHALL ignore request fields that attempt to choose a
repository, URL, tag, asset, destination, trust root, or signer. The action
SHALL invoke the fixed latest-release verifier and candidate publisher, then
redirect to settings with either an actionable refusal or the verified tag,
asset, archive SHA-256, signer workflow, and staged candidate SHA-256.

Download and signature verification SHALL NOT install an executable, request a
restart, stop an engine or verification, access an account or credentials, or
execute candidate code. Only the final local candidate publication window SHALL
share the existing in-process update/start exclusion; network work SHALL NOT
hold the engine/verification flock. The full signed-release operation SHALL use
a separate single-flight mutex so discovery/TUF verification cannot run
concurrently and an older slow result cannot overwrite a newer result.
Installation remains the separate existing operator-reviewed action.

#### Scenario: Unauthenticated download request
- **WHEN** a client posts the download route without the console session and matching CSRF token
- **THEN** the request is rejected and no release or candidate operation starts

#### Scenario: Request supplies another repository and path
- **WHEN** an authenticated request includes repository, URL, tag, path, command, or destination fields
- **THEN** those fields cannot change the constructor-bound release source or sibling candidate destination

#### Scenario: Signed release is staged
- **WHEN** the latest platform archive passes provenance, extraction, and executable validation
- **THEN** settings reports the verified identities and renders the existing separate candidate install review

#### Scenario: Verification fails
- **WHEN** release discovery, download, provenance verification, extraction, or publication fails
- **THEN** settings explains the failure, preserves the running executable and previous candidate, and requests no relaunch

#### Scenario: Download runs while engine is active
- **WHEN** an engine already holds its execution flock and the operator checks and stages a release
- **THEN** bounded network and verification work may proceed without stopping the engine, but no executable installation occurs

#### Scenario: Install commit races final staging
- **WHEN** an install commit or process relaunch becomes committed before a verified download can publish
- **THEN** candidate publication is refused and the old console cannot stage bytes after committing its own replacement

#### Scenario: Two release checks overlap
- **WHEN** one signed-release request is still discovering or verifying and another request arrives
- **THEN** the second request waits on the release-only single-flight lock without blocking engine start, and reverse completion cannot publish an older tag last

#### Scenario: Candidate survived a console restart
- **WHEN** settings renders a local candidate without process-local verified release provenance
- **THEN** it labels the provenance as unknown and does not call the candidate signed until a new signed download succeeds
