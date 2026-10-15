# a141 proposal-freeze review

Date: 2026-10-15. Reviewer: `/root/a141_adversarial_reviewer`, independent Codex review context; not Terra implementer. Baseline inspected: `6765858af1beb472268cb699efa1964addb84d60`. Method: gstack autoplan CEO → Design applicability → adversarial Engineering → DX, using review-specific sections loaded from installed skills. This is one actual reviewer applying four perspectives, **not four independent models**. Additional Claude/Codex subprocess voices were not launched: delegation assigns this reviewer a bounded independent review, concurrent slots belong to implementation/preparation, and there is no authorized Claude provider in this task. Cross-model consensus is N/A. Manager's later independent review and additional gstack implementation review remain required.

No production edits, test execution, runtime commands, external telemetry, credential access, deployment or operational setting changes were performed. Existing user scope approval supplies the premise; no new scope approval is inferred. The only owned file is this review. Skill preferences for external restore/test-plan files, plan edits and global telemetry were not executed; review/test-plan evidence is contained here to respect assigned file ownership. Source-specific branch statements below follow reading Terra-produced ASTs in `/tmp/tossos-a141-intake/`.

## Verdict

**CHANGES REQUESTED before freeze**: F1 explicit eligibility contract; F2 omission-counter regression coverage; F3 executable gate task wording. No request to expand runtime or trading scope. These are review decisions for Manager incorporation, not edits to the spec by the reviewer.

## Findings

### F1 — P1: distinguish backend selection from authorization (confidence 10/10)

Design decision 1 says effective WTS **or disabled OpenAPI** must not construct official dependencies, but the requirement weakens disabled behavior to “disabled and the existing resolver does not authorize official access.” `internal/app/app.go:185-186` has `if flag == "" { return cfg.Prefer, nil }`; the resolver selects a preference and does not inspect `Enabled`. The actual existing CLI authorization is separate at `internal/app/app.go:131`: `if creds != nil && cfg.OpenAPI.Enabled && prefer != "wts" {`. ASTs: `internal-app-app-go--ResolveBackend.ast.json`, `internal-app-app-go--New.ast.json` (B8 at line 131). A disabled configuration plus explicit `--backend openapi` still resolves to openapi but must not construct official clients. The current conditional scenario can be read as allowing that bypass.

Required correction: state the eligibility predicate explicitly, with `enabled=false` rejecting official access for **all** valid overrides, including openapi/official. Keep normalization/invalid-backend errors in existing resolver. Add disabled + explicit official override to offline regression matrix. Existing `internal/app/app_test.go:TestNewOfficialClientTruthTable` is the reference behavior, not authorization supplied by resolver itself.

### F2 — P2: exact-number decoding also changes omission metadata types (confidence 10/10)

Design promises existing omission metadata compatibility but only says “minimal decoding changes.” `internal/mcp/server.go:327` currently reads existing omission count using `.(float64)`; `:342` emits `_omitted_items: dropped` where `dropped` is an int. AST `internal-mcp-server-go--trim.ast.json` enumerates marker B1/B2. A UseNumber decoder changes an incoming retained marker to `json.Number`; repeated shrinking also encounters the int it generated. Merely switching the decoder can silently lose the cumulative omission count while still producing valid JSON with a plausible marker. Current `TestToolResultCapsOversizedPayload` only checks marker presence, not count conservation.

Required correction: explicitly test count conservation (`retained real items + total omitted = original count`) across repeated trimming, and decoded pre-existing markers; preserve or minimally correct handling of generated int / decoded json.Number / existing float64 as applicable. Do not add paging, limits, or unrelated schema features. A malformed marker policy need not become a new feature. This is a test requirement within already promised omission metadata integrity; implementation may demonstrate why a supported path is unreachable, with evidence, rather than overbuild conversions.

### F3 — P1: gate checkbox is circular (confidence 10/10)

Task 3.3 says pass repository gate, while the footer says mark it only after actual gate succeeds using the paired-gate convention. `tools/gate.sh:215` counts this change's own unchecked tasks: `OPEN_COUNT=$(grep -c "$OPEN_TASK_RE" "$TASKS_FILE" || true)`; lines 219-221 reject any. The exception at lines 271-288 applies only to a different paired change. This change has no deploy pair, and own task 3.3 is still counted even if one were added. Therefore honest execution cannot satisfy the stated task order.

Required correction: make checkbox 3.3 cover actual constituent offline checks and Manager's pre-gate evidence/diff verification. Put aggregate `make gate CHANGE=a141-mcp-respects-backend-and-exact-numbers`, final Manager verdict, strict archive and PM sync in explicit non-checkbox completion conditions; record their real results after execution. Do not alter gate code or falsely mark gate complete in advance.

## CEO perspective — scope and value

Premise challenge: stored credentials alone should not override a deliberately disabled backend; correct. MCP output numbers should not change value because the payload is large; correct. Upstream parity alone is insufficient justification for merging unrelated commands; plan correctly selects two defects. A larger upstream merge would introduce order types, token behavior and unknown operating changes beyond user authorization.

Existing leverage: `app.New` and resolver truth-table tests supply the eligibility policy; `hybrid.Policy` supplies read-routing preference; `mcp.NewServer` already takes explicit dependencies; Go JSON decoder exact-number support supplies the numeric fix. No additional platform, service or configuration is needed.

```text
CURRENT: MCP credential presence can expose direct official dependency; truncation decodes float64
PLAN:    one eligibility boundary + exact retained numeric tokens + count-preserving omission evidence
IDEAL:   adapters consistently apply existing policy, documented/tested without broad upstream replacement
```

| Alternative | Effort/risk | Decision |
|---|---|---|
| Local eligibility guard and exact-number decoding | Small; needs regression matrix and marker checks | Accept |
| Rebuild MCP around a generic client/service factory | Larger abstraction and initialization behavior change | Reject |
| Replace local MCP with upstream wholesale | Unbounded safety/compatibility changes | Reject |

Temporal review: first collect fresh evidence and freeze contract; next Terra writes RED tests before minimum fix; then independent adversarial and additional gstack review; last run real gates and archive. No reason to consume credentials or touch operating settings at any stage. Six-month regret would be policy drift between CLI and MCP; exact predicate tests mitigate it without a new policy engine.

1. Architecture: explicit dependencies are sufficient; no new coupling required. F1 prevents the resolver from accidentally becoming a false authorization proof.
2. Error/rescue: keep resolver errors, missing-credential behavior, MCP isError and serialization failures observable; no retries introduced. See registry below.
3. Security: reduced capability exposure is desired; no new inputs, secrets or privilege. Disabled and WTS cases need direct-read and mutation-service absence assertions, not only router labels.
4. Data flow: small and large outputs must agree on number semantics. F2 covers partial-list truth, with explicit retained versus omitted items.
5. Code quality: reuse normalization and constructors. A new factory abstraction would increase the blast radius without supporting a new requirement.
6. Tests: new constructor/integration regression coverage must accompany pure serialization tests. Existing routing tests alone do not instantiate the MCP wiring boundary.
7. Performance: same full marshal/decode/shrink algorithm; exact-number tokens add no extra network operations. Huge non-array strings remain an existing limit limitation, outside this fix.
8. Observability: existing errors, auth snapshot and omission metadata are sufficient. No credential logs, telemetry or background health probes should be added.
9. Deployment: no deployment authorized; branch changes plus offline evidence only. Code revert is rollback; no data migration or backfill.
10. Trajectory: adapter-local fix is reversible and supports later reviewed compatibility work. No whole-project rewrite or trading algorithm changes justified.
11. Visual design: no UI rendering, layout, accessibility or input interaction changes; examined scope and skipped visual-design phase per autoplan applicability rule.

CEO completion: scope accepted with F1/F2/F3 corrections. No product challenge or taste decision requires new user input. No market/competitor assertion used; relevant comparison is existing local CLI behavior.

## Design perspective

No visual UI scope. MCP response shape is the user-facing interaction contract: preserve content/text, isError, valid JSON number types, small-response bytes and partial-list markers. Loading/empty/error states already belong to existing operations; no new screens or presentation layers. Design score: N/A (no visual-design change); contract clarity blocked by F1/F2 until incorporated.

## Adversarial Engineering perspective

```text
config.Enabled + credentials + resolver(config.Prefer, override)
                  |
                  v
      eligibility: credentials && Enabled && effective != wts
           / false                         \ true
 official=nil; trading=nil        existing official client + gated service
           \                                /
       WTS optional -> hybrid -> MCP server/catalog -> existing operation gates

payload -> MarshalIndent -> <=30000: existing text unchanged
                  |
               >30000
                  v
          exact-number decode -> shrink arrays + cumulative count -> MarshalIndent
```

Nil path: absent credentials must not construct official objects; absent WTS session must not gain WTS authorization from a placeholder client. Empty path: empty arrays and small payloads retain existing serialization. Error path: invalid backend normalization returns existing error; unsupported payload marshal remains RPC error; no mutation is executed by tests. Happy path: authorized official and authenticated WTS operations retain their supported routing and safety checks.

Architecture review: change touches adapter constructors and response serialization, without database schema, Guardian or engine gate coupling. Existing hybrid accessor may be changed only if actual callers/evidence justify it. No new service, dependency or artifact distribution pipeline is necessary.

Code quality review: use standard decoder exact-number support; avoid changing each financial payload struct to strings. Preserve existing constructor seams, and fix omission counter representation only as demanded by F2 regression. Do not duplicate the entire app startup stack into a new generic abstraction.

### Test plan and coverage diagram (Terra execution required)

| Flow/branch | Existing evidence | Required verification |
|---|---|---|
| enabled + creds + auto/openapi/official alias | CLI truth table, resolver tests | Offline MCP exposes eligible official dependencies; no real network |
| disabled + creds + auto | CLI truth table | MCP direct reads and mutation dependency absent; WTS remains usable |
| disabled + creds + explicit openapi/official | Resolver normalization only | Negative MCP construction/call test, no official bypass (F1) |
| enabled + creds + effective WTS; override beats config | Resolver/WTS router tests | MCP direct official/trading absent; auth snapshot agrees |
| no creds, WTS present / neither creds nor WTS | MCP missing-session tests | Existing fallback or startup error maintained; no fabricated auth |
| malformed config/credential/session and invalid override | Existing startup loaders/resolver | Keep existing error contract; test changed sequencing where observable |
| authorized official mutation guard | Existing trading-service safeguards | Offline tool refuses forbidden mutation; no guard/lineage removal |
| small numeric JSON | TestToolResultKeepsSmallPayload | Exact existing encoding unchanged |
| oversized nested + retained array JSON | Oversize cap test | Positive/negative >2^53, long decimal numeric tokens equal input; no string conversion |
| repeated truncation and prior omission marker | Existing test checks marker presence only | Count conservation and numeric-marker compatibility (F2) |
| marshal failure, nil/empty payload, untrimmable large scalar | Existing serializer behavior | Preserve compatibility; no new claimed universal hard cap |

No LLM prompt changes or new background jobs. Use Go testing + local httptest/injected transports; configure update-check seam/cache so an MCP constructor test cannot access the internet. Race and full repository checks are Terra responsibilities. No production timing/load claims are made; the test plan above is pending, not passing evidence.

| Failure mode | Error/rescue or mitigation | User-visible outcome | Gap |
|---|---|---|---|
| Disabled backend but saved official credential | Eligibility rejects object exposure | No unauthorized official operation | F1 contract clarification |
| Invalid override | Existing resolver error | Existing invalid-backend message | Regression must preserve |
| No eligible credentials/session | Existing startup/operation denial | Explicit failure, not success-empty | Coverage required |
| Large numeric payload | Exact-number decode/re-encode | Correct retained numeric value | RED/GREEN required |
| Multiple shrink passes | Preserve accumulated omitted count | Accurate partial-list indicator | F2 coverage |
| Unsupported payload serialization | Existing RPC encoding error | Explicit error | Preserve |
| Own gate waits on unchecked gate task | Non-checkbox aggregate completion condition | Honest execution record | F3 correction |

Performance review: no new request loops, pools, cache or persistent data structures; same response-size algorithm remains. Decode changes must not increase trim-loop iterations unexpectedly; count/termination tests suffice for this bounded fix. Existing unbounded huge scalar behavior is not newly promised resolved.

Engineering completion: 3 actionable plan findings; 0 tested implementation claims. No production edits permitted until Manager integrates corrections and all READY evidence is present.

## DX perspective

Persona: developer attaching a local MCP host to an already configured TossOS installation. Empathy narrative: “I selected WTS or disabled OpenAPI and expect all tool paths to honor that choice. If a result is shortened I must know what was omitted, and a financial identifier must remain exact.” No new installation or auth steps should be required; time-to-hello-world is unmeasured and unchanged by this plan, with zero additional steps as the target.

| Journey stage | Review result |
|---|---|
| Discover | Existing CLI/MCP docs retained |
| Evaluate | Scope is compatibility fix, no new product claims |
| Install | Existing binary workflow unchanged; no install performed |
| Authenticate | Existing shared credentials/tokens retained; no live auth performed |
| Configure | Respect existing selection and enabled switch (F1) |
| First call | Existing catalog/tool names and output shape preserved |
| Debug | Existing error contract retained; auth snapshot must match reachability |
| Integrate | JSON types/tokens preserved and omission count accurate (F2) |
| Upgrade/maintain | No data/config migration; final review/gate receipts required (F3) |

| DX pass | Scoped assessment |
|---|---|
| Getting started | 8/10 plan intent: no added steps; elapsed time not measured |
| API/CLI design | 7/10 until exact F1 eligibility specified; existing resolver precedence preserved |
| Errors/debugging | 7/10: preserve existing invalid backend/missing auth/encoding errors; disabled-policy wording may remain imperfect and is not a new blocker |
| Documentation/learning | 8/10: OpenSpec includes boundaries; no new API/tutorial needed |
| Upgrade/migration | 8/10: no data migration, reduced previously unauthorized capability is intentional |
| Environment/tooling | 8/10: offline Go fixtures required; full gate outstanding |
| Community/ecosystem | N/A for this internal compatibility correction; no outreach/licensing changes |
| Measurement/feedback | 7/10 pending regression receipts and actual final gate |

DX checklist: preserve names/types; retain structured isError; verify auth snapshot against eligible dependency; prevent accidental network in constructor tests; keep deployment explicitly outside scope. Scores assess this plan only, not a measured running product or overall project quality.

## Decision audit and completion

| ID | Phase | Decision | Class/principle | Rejected alternative |
|---|---|---|---|---|
| D1 | CEO/Eng | Keep two targeted defects; no broader merge | Mechanical, pragmatic | Wholesale upstream replacement |
| D2 | Eng/Security | Require exact eligibility predicate and disabled-override case | Mechanical, explicit | Resolver preference treated as authorization |
| D3 | Eng/DX | Test accurate accumulated omission count | Mechanical, completeness | Marker-presence assertion alone |
| D4 | Eng/Delivery | Remove self-referential aggregate gate checkbox | Mechanical, explicit | Advance falsely checked gate or gate-code edit |
| D5 | Design | Skip visual review, inspect MCP response contract | Applicability | Invented UI work |

Cross-perspective themes: capability truth must match configuration (CEO, Eng, DX); numeric and partial-list truth must match output (Eng, DX). No independent multi-model consensus claimed. Not in scope: order types, engine/risk/journal/token changes, new services, live tests, deployments, settings, broad schema redesign, paging, telemetry.

Manager follow-up required: incorporate F1-F3, record acceptance in review.md, verify fresh AST/CodeGraph/baseline and all READY conditions, then release Terra implementation. This review does not substitute for subsequent independent implementation review, additional gstack review, actual gates, archive or Manager completion verification.

## Re-review — proposal freeze approved

2026-10-15, same independent reviewer. Manager incorporated all three findings; re-read the actual revised design, delta spec and tasks. **APPROVED for proposal freeze; zero unresolved proposal findings.** This supersedes the initial changes-requested verdict above, retained for audit.

- F1 resolved: requirement and design now state `credentials && Enabled && preference != wts`. Disabled scenario explicitly includes openapi/official override; eligible scenario requires enabled + credentials + auto/openapi. Resolver remains preference normalization, not authorization.
- F2 resolved: design and explicit scenario require cumulative omission counts with json.Number, prior markers and repeated trimming; malformed values cannot panic or overflow. Terra must supply regression receipts before implementation approval.
- F3 resolved: task 3.3 is pre-gate checks and Manager evidence inspection. Actual aggregate gate is a separate mandatory completion condition, so tasks need not falsely claim gate execution. Repository workflow still requires archive/PM synchronization after gate.
- Hybrid accessor excluded by default; expansion requires demonstrated remaining bypass and Manager review. No changes to user safety scope.

Reviewed artifact SHA-256:

```text
proposal.md 56d2a7c7929d9195ca6db4b142dedb262a512169176df8432262d3c4c62372d4
design.md 6088eef1efba16d83dd98e5deb9f00532fba644df194fde3cf067339e4df373d
tasks.md a5cd35acd34259d18a995184c819b0cf812807b3837cb2c50eedf55f5863ab27
specs/mcp-backend-and-number-integrity/spec.md 9ca574497319927af718d3fae7c57515b52954d3926e28bf4e00fbbb1dab6365
```

Freeze approval is not a READY declaration: committed immutable baseline, current CodeGraph, pre-edit maps, strict validation and PM checks remain Manager/Terra evidence obligations. Implementation review and additional gstack review remain pending. Tests were not executed by this reviewer.
