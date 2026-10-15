# a141 additional gstack review

Date: 2026-10-15. Reviewer: `/root/gstack_compatibility_reviewer`, independent Codex context, separate from Terra implementation and `/root/a141_adversarial_reviewer`. Skill: `/home/daniel/.agents/skills/gstack/review/SKILL.md` and its review checklist. Scope uses the immutable change baseline `6765858af1beb472268cb699efa1964addb84d60`, not the unrelated upstream origin/main comparison. Production/test snapshot: `399ca8cd`; subsequent evidence-only commits through `62dc50d7` were inspected.

## Verdict

**PASS REVIEW — code and reported documentation findings resolved.** No additional demonstrated production defect in the four changed Go files. P2/P3 closure is recorded below. This is not full gate, deployment, archive, or implementation-completion approval. Tests were inspected, not executed by this reviewer; execution belongs to Terra.

## Findings

- **P2, confidence 10/10 — final logic maps describe old branches.** `analysis/function-logic/cmd-tossctl--newmcpcmd/function-logic-map.md:23` says `B4 | credentials are present`, and line 28 assigns backend resolution to B9. Current AST assigns resolver error B2 and eligibility B5. The call table repeats old IDs. `internal-mcp--toolresult/function-logic-map.md:22` says `current default number decoding rounds >2^53`; current source uses UseNumber. `internal-mcp--trim/function-logic-map.md:20` says `current assertion accepts only float64`; current source accepts checked json.Number/int/float64 through omittedCount. Helper and unchanged-test maps also contain unrelated omission-marker boilerplate. Correct final FLM/BTM against existing current AST, or explicitly separate historical baseline descriptions from final maps. A source-hash match alone does not validate semantic evidence. Sent to Manager/Terra; no production fix required.
- **P3, confidence 9/10 — README credential-only startup promise stale.** `README.md:504` says `둘 중 하나만 있어도 MCP 서버는 뜨고`. Official credentials alone no longer suffice when OpenAPI is disabled or effective backend is WTS. Document that the selected backend policy must allow at least one credential, including the enabled/non-WTS rule for official capability. Sent to Manager; informational documentation correction.

## Review coverage

- Scope matches approved a141 proposal/design: constructor eligibility and numeric integrity only. Changed Go files are `cmd/tossctl/mcp.go`, `cmd/tossctl/mcp_test.go`, `internal/mcp/server.go`, `internal/mcp/server_test.go`.
- Reviewed full scoped diff, actual constructor/callers, CLI predicate in `internal/app/app.go:131`, resolver delegation, shared `internal/ops/ops.go:174` catalog gate, response truncation and prior independent review findings. Existing current AST/FLM artifacts were read; semantic mismatch is reported above.
- Critical checklist: no new SQL, persistent writes, concurrency state, shell interpolation, enum, or LLM-to-execution boundary. The MCP direct dependency restriction aligns with the existing CLI predicate. Existing trading service/lineage setup remains within that restriction; no journal/risk/shared-token/engine implementation or operating configuration changed.
- Testing inspection: actual command negative-backend checks, guarded subprocess proxy, enabled/disabled matrix, exact json.Number type/token checks, cumulative omissions, malformed/nonfinite numeric counts. The matrix runs through the parent guarded test; its standalone test name is not execution proof. Final mutation receipt was read but no reviewer test rerun is claimed.
- Independent additional specialist `/root/gstack_compatibility_reviewer/a141_gstack_specialist` applied testing, maintainability, security, performance and API-contract checklists and returned NO FINDINGS. These were combined in one fresh context due the four-agent concurrency limit, not five independent specialists.
- Separate fresh red-team `/root/gstack_compatibility_reviewer/a141_redteam` returned NO FINDINGS after checking production/callers. Its fixture/test review was summary-only; primary reviewer inspected changed test source. No cross-model agreement is claimed.

## Workflow adaptations and limits

This is a local, explicitly scoped review with no PR or release action. Remote fetch/Greptile/version-queue checks, telemetry, external artifact synchronization, skill setup/config writes, automatic production fixes and persistent memory updates were not performed. No separate Claude or CLI Codex process was run; all actual reviewers are Codex agent contexts. Thus this is the gstack checklist/specialist/red-team workflow adapted to the authorized Manager/Terra separation, not a claim that the skill's literal Claude-plus-Codex CLI pipeline ran. Data migration and UI/design specialists are not applicable. Automated slop scan, build/test execution and full repository gates are outside this reviewer role; Terra receipts and Manager verification remain required.

## Reviewed SHA-256

```text
cmd/tossctl/mcp.go          5ce62e06dcc392d1701a197bca0ece3119a0cbf2da057f0d891753582beaf3d0
cmd/tossctl/mcp_test.go     4599da0635ccf209cb4a6ea7b2199481034dcbb117e094aef3943a4f1dd73998
internal/mcp/server.go      a621b9b36b040e9afdf193be32281b9950417bff998155d9c2a23dc31a1f876f
internal/mcp/server_test.go 49e2a2b0a6aa60b654b7a1d1c85012509aa3d34a9a03113575715a8a24bb94ba
```

## Documentation re-review at 6d0f342d

P3 CLOSED: `6d0f342d0d8de598ff701ba01b9524fc75cdadc2` changes README authentication wording to require a credential allowed by backend policy and explicitly states enabled/non-WTS eligibility. P2 remains OPEN: that commit changes README only, and inspected worktree FLM/BTM files still contain the old branch numbering and float64-only descriptions identified above. The four reviewed Go hashes remain identical. Code verdict remains PASS CODE; final documentation/gate acceptance is pending.

## Documentation re-review at 38d3d6c2

P2 PARTIALLY FIXED, still OPEN. `38d3d6c280a7d906dc8b6e9addcc08f3abbc1b5a` fixes the eligibility wording and exact-number/current marker-type descriptions. It does not align final branch IDs, call evidence references or BTM with the current AST. Actual newMCPCmd IDs are B1 config:47, B2 resolver:51, B3 paths:58, B4 credentials:62, B5 eligibility:67, B6 session-error:81, B7 session:85, B8 neither-credential:89, B9 routed-WTS:99, B10 cache:114, B11 update:119. Actual trim IDs are B1 keep-at-least-one:330, B2 nonnegative-keep:333, B3 marker-present:336, B4 validated-count-and-addition-bound:337. The eligibility helper FLM still describes numeric metadata and omission markers unrelated to its function. These precise corrections were sent to Manager; no new runtime finding or source change. PASS CODE remains, documentation closure pending.

## Final semantic re-review — Manager working-tree corrections

P2 CLOSED. Re-read final FLM/BTM after Manager corrections: newMCPCmd has separate input/branch/call tables with current AST IDs B1–B11; trim has B1–B4 in actual order; helper correctly states branches=null and maps condition coverage rather than inventing B1. toolResult accurately describes UseNumber and does not claim an unavailable decoder seam. BTM distinguishes actual focused test receipts from unchanged paths with no individual branch-hit proof. Unrelated omission-marker text was removed from the unchanged small-payload test map. An intermediate table-replacement defect was caught and repaired before this closure. P3 remains CLOSED at 6d0f342d.

Reviewed final FLM SHA-256 (paths relative to analysis/function-logic):

```text
cmd-tossctl--newmcpcmd/function-logic-map.md 38fa1b8408da91e98490a839a274df1bb6b91fc7d3fcb1c827c5052c1b2e1357
cmd-tossctl--mcpofficialeligible/function-logic-map.md 523cd4ba96871dae50f1648799609f658273125e12e75e87792699ba5902c6d9
internal-mcp--trim/function-logic-map.md 2e95dcffd46bab682b841c1deaf6cee9f11f77e4838958d758613680d4ff9019
internal-mcp--toolresult/function-logic-map.md efe806859ed403467ada134871db5c0192b2bf10672b1a5ebf7f01fea5d11439
```

Four Go source/test hashes above remain unchanged. Final verdict: PASS REVIEW with no unresolved finding from this additional gstack pass. Parser-format checks, full tests, index freshness, full gate and Manager completion acceptance remain separate; no successful result is inferred for those ongoing checks.
