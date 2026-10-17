# a143 independent proposal review

Date: 2026-10-15. Reviewer: `/root/a143_adversarial_reviewer`, separate from Terra implementer. Verdict: **CHANGES REQUESTED** (one contract blocker below). No implementation or tests executed.

## Method and evidence

Read AGENTS.md, `.claude/CLAUDE.md` SDD_SHARED, WORKFLOW READY/evidence/review requirements, proposal/design/spec/tasks, and gstack autoplan methodology. Applied its sequential CEO → Design → Engineering → DX lenses under the assigned read-only proposal-review scope. This is one independent reviewer using four perspectives, **not** four independent models, a full gstack autoplan execution, or the later mandatory code review. No nested Claude/Codex reviewers, skill onboarding, telemetry, operating changes, or test execution performed.

Intake source revision differs from worktree, but relevant file SHA-256 values match current files and AST pins:

- candle_reads.go: `c22b069c909ba4052a7d24e05683baeb82b61f3f31ed63868097d9efc83c43e2`
- strict_minute_candles.go: `d32181a939f298db306f492b488468b5925ac0ba97dad3f82cb1cb3286254ced`
- officialbars/producer.go: `8d45ca93b090cfe9e10a93e5a658991ed3376820b56dfa05e49b809171c16772`

Evidence: `/tmp/tossos-a143-intake/ast/{adaptCandles,strict-minute-candles,strict-minute-symbol-validation,officialbars-adopt-page}.json`; current source inspected after pin verification. Public downloaded REST evidence `/tmp/toss-openapi-live-20261015.json` documents descending candles and KR example `0101N0`. This review did not freshly retrieve public URLs.

## 1. CEO: problem, scope, delivery

The proposed separation solves a real representation incompatibility without replacing TossOS execution. General Chart ordering and strict candle evidence serve different consumers and must stay separate. Six-character KR lexical acceptance is a compatibility change, not permission to add strategy symbols. Snapshot refresh is useful evidence but introduces no WebSocket support. Deferring one-year conditional expiration is justified because protective create/replace/recovery semantics need a separate approved contract. No scope expansion recommended.

## 2. Design: consumer contract

No interactive UI changes; visual design review is not applicable. Output semantics do apply: latest candle last, unchanged timestamp and OHLCV, empty/singleton stable. The plan explicitly protects raw newest-first engine pages and close-to-open conversion. Accept reversal rather than sorting: ordering is an upstream contract and unrelated data sanitization is outside this change.

## 3. Engineering: contract blocker and risks

**P1 — Claimed existing identity guarantee does not exist in strict transport reader.** Design decision 2 says “existing normalization” and “requested/returned identity matching”; spec says a mismatched response identity must retain strict refusal. `StrictMinuteCandles` performs lexical validation on the supplied symbol, sends it unchanged, decodes candles/nextBefore, and injects the request symbol into `StrictMinutePage.Symbol`. It does not normalize the symbol or authenticate a response-symbol field. These claims would either require a new validation contract beyond lexical-only scope or produce an untestable acceptance criterion.

Required plan correction: preserve exact request-symbol propagation and existing lexical/data/cursor validation; explicitly preserve lowercase/whitespace rejection. If identity checks in officialbars are part of invariant coverage, name that layer and demonstrate them separately, rather than claiming the broker response has an identity guarantee. Do not add fictional response fields or silently broaden implementation.

**P2 — Specify canonical bytes for snapshot hashes.** Existing daily-monitor computes full hash over `jq -S .` output, including final newline, and path hash over `jq -S '.paths | keys'` output. A generic sorted-JSON serializer can produce different whitespace/Unicode bytes and permanently trigger false changes. Record/use the exact existing algorithm; no workflow edit required. AsyncAPI metadata can separately record its own explicit canonicalization.

Required test matrix: multi/empty/singleton general charts and unchanged values/timestamps; canonical `0101N0` and numeric KR; lowercase, whitespace, punctuation, non-ASCII and wrong length refusal; unchanged US lexical policy; newest-first strict pages/cursors and officialbars close-time conversion. High-risk lexical predicate requires fresh function/branch maps and mutation evidence as assigned. Required checks and immutable baseline remain implementation prerequisites; review does not certify READY before predecessor and evidence gates pass.

Data boundary:

```text
public official response → general adapter → chronological domain.Chart
public official response → strict raw reader → newest-first close-time page
                                             → officialbars → open-time bars
public schema documents → pinned JSON + metadata (no execution integration)
```

## 4. DX: reproducibility and interpretation

Existing signatures and command naming remain stable. Store source URL, retrieval date, actual version, counts, exact hash algorithm and verification receipt so a maintainer can independently reproduce the baseline. Do not copy stale HTTP status or claim fresh live retrieval when only a local downloaded document was reused. Keep AsyncAPI snapshot-only status and deferred automated monitoring prominent. Snapshot bulk should be reviewed separately from the small production diff.

## Disposition

Manager must correct P1 before freezing; P2 is a concrete compatibility acceptance refinement. After corrections, this reviewer can recheck the plan. Implementation review, Terra test receipts, additional gstack code review, full gates and Manager acceptance remain outstanding. No execution/settings/token/journal/risk path change approved by this review.

## Re-review after Manager corrections

P1 resolved: design decision 2 and specification now explicitly preserve exact request-symbol propagation, reject lowercase/whitespace, and deny independent response-symbol attestation. P2 resolved: design/spec pin exact `jq -S .` bytes with trailing newline and retain existing metadata compatibility. Minor editorial follow-up sent to Manager: replace residual “identity tests” in design Risks with “request-symbol propagation tests”; it does not override the corrected normative requirements.

Final proposal-freeze verdict: **PASS** for the corrected scope. This is proposal acceptance only. Predecessor completion, immutable baseline, CodeGraph/function/branch evidence, implementation, test execution, independent adversarial code review and additional gstack review remain separate mandatory gates. No test or deployment success claimed.
