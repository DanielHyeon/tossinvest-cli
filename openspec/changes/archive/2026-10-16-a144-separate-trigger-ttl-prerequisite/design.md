## Context

A100 made M0 trigger-only mode a strict resume/redo path: prior `Outstanding` and M0 checkpoints HOLD, and every procedure step except `conditional-trigger` must have latest `PASS`. `idempotency-ttl-edge` is intentionally opt-in because it can create a second live order. Its result is not read by the trigger, which instead uses current sellable quantity and bid/last/tick values. The all-PASS rule therefore blocks an otherwise isolated trigger measurement when TTL is skipped, failed, refused, missing, or nonterminal.

This follow-up supersedes A100's all-PASS wording only for the TTL-edge prerequisite. It does not edit A100 historical artifacts.

## Goals / Non-Goals

**Goals:**

- Exclude exactly `StepIdempotencyTTLEdge` from M0 prior-PASS admission.
- Preserve all prior-record lines and their verdicts.
- Keep M0 trigger mode unable to execute the TTL step body, issue a TTL broker mutation, or add a TTL planned mutation. Normal preflight, `Plan.Excluded` metadata, and the ordinary no-opt-in skipped append remain permitted.
- Preserve every other M0 safety boundary.

**Non-Goals:**

- No TTL measurement, replay, record repair, aggregate PASS, or verdict conversion.
- No new flag, configuration, capability/dependency framework, broker API, engine, order scope, receipt, recovery, or approval-flow change.
- No relaxation for failed, refused, skipped, missing, or nonterminal verdicts of any step other than TTL-edge.

## Decisions

### A single fixed exception in the existing prerequisite predicate

Change `M0ExactPrerequisites` to skip only `StepIdempotencyTTLEdge` in addition to `StepConditionalTrigger`. Keep iteration over `Steps()` for all remaining current and future steps.

This is narrower than a verdict policy, a mutating/non-mutating classification, or an allowlist supplied by flags. A TTL verdict of PASS, skipped, failed, refused, missing, or nonterminal has identical M0 admission treatment: it is not a prerequisite. Existing TTL lines are retained unchanged; ordinary Runner flow may append its normal skipped line for a missing or nonterminal TTL entry because the opt-in is absent, but it must not dispatch the TTL body or synthesize PASS. Every other step requires latest `VerdictPass`.

### Preserve M0 mode boundaries outside the predicate

Keep CLI `validateM0TriggerMode` and direct `verifylive.New` requirements unchanged: include-trigger requires confirm-each, resume, exactly one redo equal to conditional-trigger, a usable receipt, no include-ttl-edge, no outstanding artifact, settled M0 state, and official raw-read source. Existing execution lock remains outside this change. An outstanding artifact or a non-`pending-create` M0 checkpoint remains a pre-broker HOLD. A `pending-create` checkpoint retains its existing narrow exception: it may construct the broker only for read-only recovery, then ends HOLD with zero confirmation, cleanup, or mutation.

Because redo is exactly conditional-trigger, the Runner's settled logic keeps a terminal TTL record settled and does not dispatch it. If TTL is missing or nonterminal, its opt-in is false, so normal Runner flow may append the ordinary skipped state and `Plan.Excluded` metadata rather than dispatching the TTL body. The implementation must demonstrate zero TTL order mutation, TTL step-body dispatch, and TTL planned mutation for each TTL verdict class in M0 mode.

### Tests prove behavior at both boundaries

Add focused tests for CLI and direct Runner admission. They must show TTL PASS/skipped/failed/refused/missing/nonterminal permits normal preflight and allowed skipped/excluded metadata, but zero TTL step-body dispatch, broker mutation, and planned mutation when all other prerequisites pass. They must show a missing/failed/refused/non-PASS non-TTL step rejects before broker factory, confirmation, request, plan mutation, or cleanup when no `pending-create` checkpoint is present. Tests must separately retain pre-broker HOLD for `Outstanding` and non-pending M0 states, plus `pending-create` read-only recovery followed by HOLD with zero confirmation, cleanup, or mutation.

## Risks / Trade-offs

- [A skipped or failed TTL measurement remains unmeasured.] → It stays explicitly recorded and is not converted to PASS; this exception applies only to M0 trigger admission.
- [A broad edit could admit unrelated incomplete work.] → Use one fixed step-ID exception and negative tests over all other steps.
- [A regression could replay TTL or another mutation during a trigger run.] → Assert strict redo set plus fake-broker request logs, TTL step-body dispatch set, and planned mutation set.
- [Existing record semantics could be rewritten.] → Do not rewrite or aggregate existing verdicts; only the ordinary no-opt-in skipped append for missing/nonterminal TTL is permitted.
