# a063 CEO proposal review

Date: 2026-09-05. Mode: SELECTIVE_EXPANSION with approved a063 scope retained.
Reviewer: independent Codex agent, inherited host model; no Claude/Codex CLI outside voice was run. The Manager requested in-context phases with truthful unavailable cross-model disclosure. Consensus is therefore N/A, never two-model agreement. Scope and non-live implementation authorization came from the session; installation/survey activation remains outside that authorization.

## Intake and premises

Inputs: `restore.md`, `input-digests.json`, current D5, `../codegraph-hard-evidence.md`, `../sanitized-operational-evidence-2026-09-05.md`, and both generated AST/Function Logic/Branch Test bundles. Shared preambles of all four skills were compared against the fully read autoplan preamble; differences and all review methodology/carved section files were read. Source inspection covered the issuer, shared path resolvers, console reader/template, relevant existing tests, operations instructions, README and scoped git history. Other active changes were not read.

P1: renewal evidence must share an explicit operating profile. Accepted: existing CodeGraph resolver bindings establish why defaults can differ. P2: current production is already expired. Rejected: sanitized current evidence reports expiry 2026-10-05, so the August incident is historical only. P3: a service exit of zero proves renewal. Rejected: the observed shell fallback masks refusal; the fix must preserve the original nonzero result. P4: diagnostics may decide startup eligibility. Rejected: diagnostics are advisory and must not alter `Usable`, qualification or engine controls.

The real outcome is that an operator sees a refused renewal early enough to recover evidence without discovering it through a blocked restart. Doing nothing leaves the observed masking defect and no durable bounded last-attempt diagnostic. This is an operational correctness repair, not a market-positioning or new-product problem; competitor strategy is not a useful acceptance criterion.

## What already exists

| Subproblem | Existing code | Reuse |
|---|---|---|
| Record profile | `resolveSoakRecord` | Explicit config root |
| Issuer/console output path | `resolveSoakAttestationPath`, `runConsole` | One full-path authority |
| Qualification | `soak.BuildAttestation`, summary evaluation | Do not duplicate policy |
| Last good publication | `attest.Save` | Preserve existing issuance |
| Operator view | `Console.readAttestation`, capability-attestation template | Separate advisory fields |
| Operational guidance | `docs/operations.md` timer subsection | Extend in place |

## Alternatives and trajectory

| Approach | Effort human / agent estimate | Risk | Pros | Cons |
|---|---|---|---|---|
| Unit-only exit fix | 1h / 10m | Medium | Smallest change; immediate true exit | No normal-surface history or expiry warning; incomplete |
| Unit + opt-in bounded status + existing console | 1d / 1-2h | Low with tests | Complete requested behavior; no new service dependency | Adds local status schema and parser |
| Systemd DBus/journal query inside console | 2d / 2-3h | Medium | Direct service state | Platform coupling, raw-log sanitization, weaker isolated testing |

Recommend the middle approach, completeness 10/10 at plan level, over 4/10 unit-only coverage. Explicit data and existing interfaces are easier to inspect than shell-log parsing. Estimates exclude the independently required operational evidence period and final workflow blocker remediation.

```text
CURRENT                       THIS PLAN                       12-MONTH IDEAL
profile now explicit          tracked unit + true exit        reproducible routine renewal
shell masks refusal    --->   bounded attempt diagnostics --> operational evidence remains
no last-attempt view           advisory expiry/age in console  traceable without unsafe actions
```

Selective expansion scan: a 10x version would centralize all operational health checks, but that is unrelated infrastructure. Five nearby candidates were considered: exact warning boundary, attempt age, fixed recovery messages, status overwrite on successful retry, and custom-path collision tests. These are accepted as completeness within the existing blast radius, not new product features. Remote notifications, automatic survey restart, qualification rewrites, dashboard redesign and evidence migration are rejected from this change; no new follow-up TODO is warranted solely to make this narrow repair bigger.

Temporal interrogation: hour 1 needs exact CLI/path contract; hour 2-3 needs size/timestamp/enum and atomic-write semantics; hour 4-5 needs view/expiry precedence; hour 6+ needs missing/hostile-file fixtures and OFF compatibility. D5 now specifies these decisions before coding. The original frozen base is preserved: a proposed re-freeze was withdrawn after Manager/adversary identified immutable baseline rules, and inherited map failures remain explicit final-gate blockers.

## 1. Architecture

The issuer remains the sole qualification/publish path, with an optional diagnostic outcome after each attempt. The console reads the diagnostic from the resolved full attestation filename with a suffix and cannot issue, restart or toggle anything. This avoids new service coupling and prevents two different attestation basenames in one directory from sharing diagnostic state.

```text
explicit profile -> survey record -> existing qualifier -> last good attestation -> startup interlock
                                   | outcome                       |
                                   v                               v
                           optional bounded status -----> console read-only view
                           true process exit              advisory + current expiry
```

## 2. Error and rescue registry

| Path / error class | Rescue | Operator sees | Planned proof |
|---|---|---|---|
| record input / stage error | return nonzero, bounded failed code | fixed failed diagnostic | isolated CLI input error |
| qualification refusal | preserve last good file; return nonzero | refused + closed reason codes | incomplete evidence |
| attestation publication / filesystem error | return nonzero | failed, no invented issuance | write error fixture |
| status publication / filesystem error | return nonzero even after issuance | systemd failure; unknown/stale prior diagnostic | good attestation preserved |
| missing status | degrade to unknown | diagnostics unavailable | missing fixture |
| malformed/oversize/hostile status | reject to unknown | fixed warning, no raw text | parser/permission fixtures |
| malformed attestation | existing denial with fixed text | unusable, no parser leakage | malformed attestation fixture |
| old/future status | stale/invalid to unknown | age or unavailable timestamp | frozen-clock boundaries |

No catch-all success rescue is accepted. Typed stage/evaluation classification must not parse or persist free-form error text. The diagnostic write cannot erase successful attestation publication merely to make files agree.

## 3. Security

The new attack surface is a local diagnostic file rendered in an existing page. Closed fields/codes, 4096-byte input bound, 16 unique reasons, owner-only regular files and rejection of symlinks constrain injection, leakage and resource use. No new credential, network endpoint or live capability is introduced; approval of the template digest/profile remains a distinct deployment operation.

## 4. Data and interaction edge cases

```text
status -> type/size/owner/schema/time checks -> fixed view -> existing page
missing -> unknown                     malformed -> unknown
empty   -> unknown                     current expiry wins over status expiry
fresh issued -> issued                 old issued -> stale/unknown
fresh refused/failed -> warning         later success -> replaces earlier failure
```

There are no new submit/double-click/navigate-away mutations. Reloading the page must reevaluate age and expiry with its read clock, including absent attestation. Concurrent replacement must give either a complete old document or complete new document; inconsistent issued-expiry data degrades to unknown rather than reassuring the operator.

## 5. Code quality

The plan reuses existing qualification, full-path resolution and presentation patterns. A new leaf helper/schema is preferable to adding policy branches or shell parsing inside the issuer. The existing issuer has 12 AST branches, so its edit must preserve the full Branch Test Map instead of treating the opt-in hook as an exemption.

## 6. Tests

The Friday-night acceptance case is a refused attempt with a still-valid attestation: process fails, last good file survives, console warns, startup decision stays identical. Hostile cases are oversize/symlink/unknown-field/free-form diagnostic text and different configured attestation filenames in the same directory. The test plan in the Eng phase must pair OFF/on invocations and use frozen clock fixtures; live broker and production survey are never test doubles.

## 7. Performance

New status reads have a hard 4096-byte bound and at most 16 codes, so per-refresh work is bounded. The plan adds no network query, polling service, DB pool, retry loop or broker request. Worst-case source soak loading already exists; this change must not add a second qualification scan solely to classify diagnostic reasons.

## 8. Observability

Systemd failure, fixed console reason, attempt age and current expiry jointly explain the repair's behavior. A recent issued diagnostic is not proof of engine usability and cannot suppress current expiry. No remote alert channel is required; the known limitation is that an unattended operator must still consult the existing console or systemd state.

## 9. Deployment and rollback

```text
review template+digest -> human approves target/ops -> backup old unit -> install/reload
 -> approved survey window -> actual qualifying days -> fresh same-profile attestation
rollback: restore backed-up unit -> reload only if approved -> preserve evidence files
```

Existing binaries do not know the new boolean flag, so deployment must first confirm the selected binary supports it before installing the new ExecStart. Verification must not restart the engine or change autostart/trading settings. The current healthy expiry does not replace approval and actual operational proof; final gate/archive remains blocked until all requirements, including inherited maps, are met.

## 10. Long-term trajectory

A tracked unit plus closed versioned diagnostics makes later maintenance reproducible. The file schema is additive and opt-in; rollback reversibility is 4/5 because external installation and the new flag require coordination. No schema migration or new durable trading state is introduced, and diagnostic corruption never becomes an authorization decision.

## 11. Design

The existing capability-attestation section is the correct location for an expiry warning because it already reports current issuance and usability. Startup denial reasons and renewal health must have distinct headings so a failed renewal does not imply a stopped engine. Fixed brief recovery text and visible age/expiry provide enough information without creating an action button or exposing raw paths.

## Failure modes registry and completion

All eight error-registry rows have planned handling, visible outcomes and tests. Critical silent gaps at plan level: zero after D5; implementation proof remains pending. System audit identified historical premise drift and masked systemd exits; scope retained; 1 path-collision issue was accepted and fixed; all 11 sections examined; 5 in-scope completeness decisions accepted; 5 expansions excluded; 0 new TODOs; 6 diagram types present (architecture, data, state, error, deployment, rollback); no touched-source diagram conflict identified.

CEO dual voice table: premises, right problem, scope, alternatives, market applicability, trajectory each have primary-review findings above; Claude=N/A, outside Codex=N/A, consensus=N/A for all six. No cross-model claim is made. Remaining limitations: no outside-model pass, implementation/runtime evidence pending, original-base gate blockage, operational approval and acceptance outstanding.
