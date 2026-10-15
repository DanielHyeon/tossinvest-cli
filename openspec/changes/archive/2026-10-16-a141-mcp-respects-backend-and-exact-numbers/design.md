## Context

This is a targeted compatibility fix in the CLI/MCP adapter boundary. The exact branch inventory and callers must be extracted from current source into analysis/code-context and AST bundles before implementation. No upstream wholesale merge is authorized.

## Goals / Non-Goals

**Goals:** Backend resolution governs every MCP dependency; truncated tool results preserve exact numeric values; existing enabled official behavior and errors remain compatible.

**Non-Goals:** New order types or tools, engine execution changes, shared-token redesign, runtime configuration edits, broader argument-schema redesign, production deployment.

## Decisions

1. Reuse existing CLI eligibility: credentials exist AND cfg.OpenAPI.Enabled AND resolved preference != wts (internal/app/app.go, corroborated by /tmp/tossos-a141-intake AST). Backend resolution alone is not authorization. Effective WTS or disabled OpenAPI must not expose a direct official client or construct an official trading client as a bypass. Auto/official routing and existing validation errors remain defined by the resolver. An explicitly conflicting override must follow existing resolver semantics, documented by a regression test; do not invent precedence.
2. Gate MCP construction at the CLI-equivalent boundary; leave the hybrid accessor unchanged unless a remaining bypass is demonstrated and Manager approves the scope extension. Keep engine construction and official-only enforcement unchanged. Prefer the smallest existing constructor seam; avoid a new general factory abstraction.
3. Decode truncated JSON using exact-number preservation. Re-encoding must preserve numeric tokens, including values above 2^53 and long decimals, without converting them to strings. Preserve existing truncation metadata and valid JSON behavior, including cumulative _omitted_items counts when decoding json.Number, pre-existing omission markers and repeated trim passes. Malformed marker values must not panic or overflow. Do not expand response-limit behavior beyond a defect demonstrated by tests.
4. Use offline injected dependencies/HTTP fixtures only. Test forbidden official construction/calls, not merely final object labels. Retain current permission, confirmation and gateway checks for trading tools.

## Risks / Trade-offs

- Backend dependency changes affect financial tool reachability → high-risk AST/branch maps, complete effective-backend test matrix, adversarial review.
- Truncation can hide a precision regression → test nested and retained array/object values across the truncation threshold with exact JSON assertions.
- Shared working history can contaminate evidence → single implementation writer, immutable baseline, commit-scoped gate evidence.

## Migration Plan

No data or operating configuration migration. Commit reviewed code and offline evidence only. A code revert is the rollback; deployment is outside this authorization.

## Open Questions

No user decisions pending. Any additional affected callsite or precedence ambiguity must be reported to Manager and reflected in this design before scope expansion.
