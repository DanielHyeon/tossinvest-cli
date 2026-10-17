# Branch Test Map: `unwrapAndDecode`

- Source: `internal/official/client.go`, SHA-256 `6d145916a6797f3a32842561346c8767585a951d8883ebd81bf8af4632106639`; branch IDs follow `ast.json` (generated 2026-08-17).
- AST counts: branches 4, returns 5, calls 5, defers 0, go statements 0. Source range `213:1`–`228:2`. Signature `unwrapAndDecode(body []byte, out any) error`.
- Citation-only bundle: this function is NOT edited by a112; its branch enumeration is evidence for the L1b brief (official raw reader + bar producer). Any later body edit requires a fresh RED/BTM.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if at 214:2 — 2xx with a body, caller passes a nil `out`; the body must be discarded without decoding | `TestCancelConditionalOrderIntegration`, `TestModifyConditionalOrderStillToleratesANullResult` | n/a (not edited) | existing suite green |
| B2 | if at 218:2 — 200 whose body is truncated JSON; the read must refuse and prime nothing | `TestMalformedPublicDiscoveryDoesNotPrimeTheSequence` | n/a (not edited) | existing suite green |
| B3 | if at 221:2 — 2xx whose body has no `result` key, with a non-nil `out` | not-applicable: no existing test drives that body through `send`; the a112 raw reader pins the equivalent refusal in `TestA112MBUSOrderbookAndCalendarRejectMissingOrNullResult` but never calls this function | n/a (not edited) | not-applicable |
| B4 | if at 224:2 — `result` present but undecodable into the caller's type (timezone-naive time string) | `TestTypedMarketCalendarRejectsMalformedTime` | n/a (not edited) | existing suite green |

Non-branch properties the L1b brief cites (calls, not branches): permissive `encoding/json` semantics — unknown keys ignored, duplicate keys last-wins, `"result": null` decoding to a non-nil `RawMessage` — are measured on the `apiEnvelope` struct, not asserted by any test in this package. The a112 reader supplies its own strict object parser instead.

Verification: `go test ./internal/official -count=1` green on 2026-08-17 (721 tests passed across the four packages read by this change; exit 0). No RED round applies — a112 does not edit this function.
