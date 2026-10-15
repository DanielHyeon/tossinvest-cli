## Context

Intake AST distinguishes general adaptCandles from strict raw candle ingestion. Only general chart consumers need chronological adaptation. Strict readers preserve broker close-time timestamps and newest-first cursor progression; officialbars owns conversion to open time. These contracts must be tested separately.

## Goals / Non-Goals

**Goals:** Chronological general charts, correct domestic read-symbol lexical acceptance, reproducible public contract evidence.

**Non-Goals:** New strategy symbols or activation, new order capabilities, conditional-expiry validation, new endpoints/WebSocket clients, workflow schedule edits, deployed settings, token or execution changes.

## Decisions

1. Reverse only the general adapted candle slice from official newest-first wire order. Preserve timestamp and every price/volume value. Empty and singleton responses remain valid. Do not sort/reverse strict raw pages and do not shift their time labels.
2. Change only the domestic strict read lexical predicate to six ASCII uppercase alphanumeric characters, matching the published contract and including 0101N0. Preserve case-sensitive request handling, market selection, exact request-symbol propagation into returned Page.Symbol, duplicate/time/price validation and US lexical policy. Lowercase and surrounding whitespace remain invalid. The wire response does not independently attest symbol identity; do not claim response identity verification or introduce normalization. This does not add symbols to any strategy universe or authorize orders. Because this reader feeds engine evidence, classify that predicate as high risk and map/test all affected refusal branches.
3. Refresh REST and add AsyncAPI public JSON snapshots with metadata: official URL, retrieval date, document version, SHA-256 of exact jq -S . output including its trailing newline (the existing monitor canonicalization), path/channel counts. Preserve the existing REST monitor metadata schema so its next comparison remains valid. Record exact differences separately from assumptions about breaking changes. Fetch only public documentation URLs; no account APIs.
4. Leave .github/workflows/daily-monitor.yml unchanged. AsyncAPI automated monitoring is a later change; adding the snapshot alone is not a claim that monitoring or WebSocket support exists.
5. Keep conditional maximum-one-year validation separate: it affects protection create/replace/recovery and requires its own clock/calendar contract and approval scope.

## Risks / Trade-offs

- General reversal accidentally leaks into engine data → explicit strict newest-first/raw timestamp and officialbars close-to-open regression tests.
- Lexical expansion affects trusted evidence → high-risk maps, invalid character/length/request-symbol propagation tests; no strategy or order validator modifications.
- Large specification diff masks code changes → separate snapshot evidence and counts; no generated production API client.

## Migration Plan

No runtime/data migration or deployment. Source and public documentation only; revert reviewed code/docs to roll back.

## Open Questions

None pending; if the live official version differs from intake, record the actual retrieved version and source hash rather than falsely labeling it 1.2.24/1.2.2.
