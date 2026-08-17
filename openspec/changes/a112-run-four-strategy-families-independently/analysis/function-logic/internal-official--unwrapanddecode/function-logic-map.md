# Function Logic Map: `unwrapAndDecode`

- Source: `internal/official/client.go`
- Source SHA-256: `6d145916a6797f3a32842561346c8767585a951d8883ebd81bf8af4632106639` (current worktree; verified with `sha256sum` 2026-08-17, equal to `source_sha256` in `ast.json`)
- Signature: `unwrapAndDecode(body []byte, out any) error` (`ast.json`: `unwrapAndDecode(params=2, results=1)`)
- Source range: `213:1`–`228:2`
- AST counts: branches 4, returns 5, calls 5, defers 0, go statements 0 (`ast.json` generated 2026-08-17 by `go run ./tools/logic-map`).
- Risk scan: `risk-pattern-report.md`.
- Citation-only bundle: this function is NOT edited by a112; its branch enumeration is evidence for the L1b brief (official raw reader + bar producer). Any later body edit requires a fresh RED/BTM.

## Inputs and invariants

- `body` is whatever `(*Client).doRequest` read off the wire — the whole response, with no byte cap (see the `internal-official--client.dorequest` bundle). `out` is the caller's typed destination and is allowed to be nil; the only caller is `(*Client).send` at client.go:365, reached from `get`/`getWithHeaders`/`post`/`postWithHeaders`/`deleteAcct`.
- The envelope type is `apiEnvelope`: one field, `Result json.RawMessage`, carrying the struct tag for key `result` (client.go:179–181). Decoding is plain `encoding/json` twice — envelope, then payload — with no `json.Decoder`, so the standard library's permissive rules apply and are NOT tightened here: unknown top-level keys are ignored, and duplicate keys resolve last-wins. Measured on this exact struct (2026-08-17): `{"result":1,"result":2}` yields `Result` = `2`; `{}` and `{"x":1}` yield `Result == nil`; `{"result":null}` yields `Result` = the four bytes `null`, which is NOT nil.
- That permissiveness is the reason a112's own reader does not reuse this path. `a112MBUSStrictJSONObject` re-reads the body under duplicate-key and UTF-8 rules of its own, and `a112_mbus_static_test.go` structurally forbids the a112 reader from calling `.send(`/`.get(` at all, so no a112 evidence is minted through this function.
- Invariant that matters to the brief: a missing `result` key is a refusal, but a present `"result": null` is not — it reaches B4 and unmarshals as JSON null, which leaves `out` untouched and returns nil.

## Branches and early returns

Exact AST return nodes: `215, 219, 222, 225, 227`.

| Branch | AST kind | Source location | Meaning (one short clause) | Test disposition |
|---|---|---|---|---|
| B1 | if | 214:2 | `out == nil` → return nil immediately; the body is discarded and never parsed | `TestCancelConditionalOrderIntegration`, `TestModifyConditionalOrderStillToleratesANullResult` (both answer 2xx with a body while the caller passes a nil `out`, and both must succeed) |
| B2 | if | 218:2 | envelope `json.Unmarshal` failed → `ErrServer: decoding envelope` | `TestMalformedPublicDiscoveryDoesNotPrimeTheSequence` (200 with the truncated body `{"result":[{…},`) |
| B3 | if | 221:2 | `env.Result == nil` — the `result` key was absent → `ErrServer: response has no 'result' key` | not-applicable: no existing test answers a 2xx without a `result` key through `send`; the sibling refusal on the a112 raw path is pinned by `TestA112MBUSOrderbookAndCalendarRejectMissingOrNullResult`, but that reader does not call this function |
| B4 | if | 224:2 | payload `json.Unmarshal(env.Result, out)` failed → `ErrServer: decoding result payload` | `TestTypedMarketCalendarRejectsMalformedTime` (`"startTime":"local-time-without-zone"` cannot decode into `time.Time`) |

## Calls and live bindings

| Callee expression | Source location | Evidence |
|---|---|---|
| `json.Unmarshal(body, &env)` | 218 | envelope decode; permissive by construction (unknown keys ignored, duplicate keys last-wins) |
| `fmt.Errorf("%w: decoding envelope: %s", ErrServer, err)` | 219 | every refusal here is an `ErrServer`, which `ShouldFallback` accepts (errors.go:71, `TestShouldFallback`) |
| `fmt.Errorf("%w: response has no 'result' key", ErrServer)` | 222 | absent-key refusal; carries no response body |
| `json.Unmarshal(env.Result, out)` | 224 | payload decode into the caller's type |
| `fmt.Errorf("%w: decoding result payload: %s", ErrServer, err)` | 225 | payload refusal; wraps the decoder message, not the body |

## State mutations and fallbacks

- No shared state. The only local is `env` (2 AST assignments, both local `:=`); the only write outside the frame is through the caller's `out` pointer, performed by `json.Unmarshal` at 224. Nothing touches the `*Client`, the token cache, or the rate-budget store.
- No fallback and no partial decode: a refusal returns before `out` is written by this function, so the caller sees its own zero value together with a non-nil error. The one silent path is B1 — with `out == nil` a 2xx body is accepted unexamined, which is exactly why a112 mints evidence from a raw reader instead.

## Safety conclusion

- Safe edit boundary: the function has no side effects beyond `out`, so a body edit is contained; but it is the single decode boundary for every 2xx official response, order confirmations included (`orders_write.go:192/229/270` post through `send`). Loosening B3 or B4 would let a shape-drifted broker response arrive as a zero value that reads like data.
- High-risk impact: yes by adjacency — auth/execution responses pass through here — though the function itself performs no order, sizing, or protection decision. It fails closed: three of the four branches return `ErrServer`, and `ErrServer` is a `ShouldFallback` class, so a decode refusal routes the caller to its fallback rather than to fabricated data.
- Untested branch: B3 only (absent `result` key). B1, B2 and B4 are pinned by the tests named above; the package suite is green (`go test ./internal/official -count=1`, 2026-08-17).
