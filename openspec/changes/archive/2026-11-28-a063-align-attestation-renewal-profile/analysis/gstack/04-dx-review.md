# a063 Developer-experience proposal review

2026-09-05. Follows CEO → Design → Eng. Product: CLI tool/operator documentation, DX POLISH. Persona: maintainer operating an already installed TossOS profile; expects an explicit command, truthful status, isolated tests and reproducible installation. Hall-of-fame reference sections 1-8 were read individually; their marketing statistics and competitor times were not independently measured or reused as facts.

## Developer perspective

I arrive at the README and see tossinvest-cli described as a CLI and MCP interface. The quick-start material is about installing and connecting an account, but I already operate an existing profile. For this task I move to the attestation timer subsection of `docs/operations.md`. It explains why `--config-dir` matters, and I want to see a checked-in service definition instead of reconstructing a command from historical incident prose. I need the exact new flag and the selected binary's help output before changing anything installed.

When renewal refuses, a systemd success result would send me down the wrong path. I want the console to say the last renewal failed while still showing the expiry of the attestation I already have. If the status file is missing or old, “unknown” is more useful than a reassuring green label. I should not need to paste logs containing account information into a support conversation. I also do not want a recovery instruction to restart a running engine or silently activate survey behavior.

Before approving installation I want a template digest, target profile and backup instructions. After installation I can inspect true service exit state and approved evidence dates. A passing fixture test helps me trust the code but cannot stand in for that operational proof.

## Journey and timing

| Stage | Action / source | Friction and disposition |
|---|---|---|
| Discover | README CLI identity | Existing install onboarding unchanged |
| Evaluate | Operations timer subsection | Historical/default-profile wording must reflect configured locations |
| Install | Review repository template | Human-approved; verify binary flag support first |
| Hello world | Isolated tests + help | Target under 5 minutes after build, estimate only |
| Integrate | Exact explicit profile command | One flag, no record/out override |
| Debug | Console health + systemd exit | Fixed reason + age + expiry; raw transcript excluded |
| Upgrade | Binary before template | Old binary rejecting new flag must be documented |
| Scale | Existing six-hour timer | No new broker load or scheduler |
| Migrate/rollback | Restore backed-up unit | Preserve original evidence and settings |

Time-to-hello-world for this change means inspecting help and an isolated diagnostic fixture, not issuing production evidence. It is unmeasured; under five minutes after a built binary is a documentation target. The three qualifying days are a safety acceptance condition and must never be “optimized” to meet a CLI onboarding benchmark. Competitive-market benchmarks are not applicable to this internal correctness repair; no external competitive research was performed.

First-time walkthrough (inferred, not a performed user study): T+0 read operations timer heading; T+30s locate source template; T+1m check new help flag and exact config root; T+2m distinguish current expiry from last attempt; T+3m locate approved installation/rollback instructions. The useful moment is understanding a refusal without an engine restart, delivered by existing console text and a reproducible command, not a hosted playground.

## Eight passes

1. **Getting started, 8/10:** Reuse existing installation and add one exact template/flag example to operations docs. Separate safe inspection from approved installation so readers know what they can run now. A measured isolated walkthrough would raise confidence beyond a plan score.
2. **CLI design, 8/10:** `--record-renewal-status` is descriptive and default-off. Recording mode rejects explicit record/output overrides, and the full resolved attestation path selects the diagnostic. This gives one explainable golden path without a second status-path flag.
3. **Errors/debugging, 8/10:** Three traced cases are qualification refusal, diagnostic-write failure after successful issue, and invalid/stale status. They respectively require fixed unmet-criteria guidance, a nonzero service result while preserving the good attestation, and unknown health with age/current expiry. Recovery messages should point to the timer/runbook and approved evidence collection rather than raw stderr or restart actions.
4. **Documentation, 8/10:** Extend the existing operations subsection with template location, explicit config root, flag, actual exit behavior, thresholds, fixed reason meanings and safe inspection. Correct the absolute claim that attestation always lands under the default config directory because the shared resolver honors a configured attestation file. The Manager owns these documentation updates.
5. **Upgrade, 8/10:** Default-off behavior preserves old callers. An old binary will reject the new unit flag, so confirm the selected binary before installation and retain a backup unit. No evidence relabeling, database migration, codemod or silent toggle update is needed.
6. **Environment/tooling, 8/10:** Go temporary-profile/httptest fixtures can test the behavior without user systemd or broker access. Unit parsing/drift checks are deterministic and do not activate the timer. Linux systemd deployment is explicit; portable code must not accidentally acquire a systemd runtime dependency.
7. **Community/ecosystem, 8/10 within scope:** Existing repository issue/review workflows and MIT project context are adequate. This change creates no SDK, public API, billing or plugin ecosystem commitment. Avoid publishing sensitive diagnostic transcripts in support examples.
8. **Measurement, 8/10:** Test the displayed state, true exit result, status age and expiry boundaries. Operational acceptance records actual evidence dates and a fresh attestation after approved deployment. No telemetry installation or uploads are needed to measure this repair.

## Checklist and completion

Required: exact CLI help/example; template path and binary-before-unit order; fixed problem/cause/recovery wording; local test fixture walkthrough; profile-aware operations text; approval/rollback instructions; no test/live-evidence conflation. N/A: new SDK types, free tier, codemods, public pricing, hosted sandbox, new community channels. All eight dimensions score 8/10 at plan level, no prior comparable scoped scores, timing unmeasured, no outside voices, six DX consensus dimensions all N/A.

No new TODO expansion or unresolved implementation design decision remains. Real UX timing, rendered browser QA, engineering test execution, inherited original-base gate remediation and operational acceptance remain unperformed or blocked as separately recorded; these are not hidden by the score.
