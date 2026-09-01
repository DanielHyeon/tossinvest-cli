# Function Logic Map: `validateTypedPayload`

- Source: `internal/strategyevidence/model.go`
- Post-L1a source SHA-256: `a67ab059e4cba377f0faaeb80f1c88821de2198707779b7175f901adc4b1819a` (frozen-base SHA was `c49652af…`; L1a inserted B3–B5 only)
- Signature: `validateTypedPayload(kind EvidenceKind, canonical []byte) error`
- Source range: `398:1`–`438:2`
- AST evidence: `ast.json`, regenerated 2026-08-17 after the L1a GREEN edit (12 branches; the frozen base had 9 — B3/B4/B5 are the inserted kind switch and its two cases; former B3–B9 shift to B6–B12 with unchanged semantics).
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- Called from `NewEnvelope` (model.go:127) after `canonicalJSON`, so `canonical` is a canonical top-level object with numbers already normalised (`1.9080e4` → `19080`; `1.50` → `15e-1`). It is the only typed hook re-run on `Store.Append`, `scanEnvelope` replay and `Snapshot.Valid`.
- Order is load-bearing: decode (B1) → `rejectSecretFields` (B2) → **new kind dispatch (B3–B5, returns for the two breakout kinds)** → legacy generic map (B6–B12) untouched for every other kind. Legacy canonical bytes/digest vector `7e30f2af…` is pinned by `TestLegacyKindCanonicalBytesAndDigestAreUnchanged`.

## Branches and early returns

- Exact AST return nodes: `403:3, 406:3, 413:3, 416:3, 434:4, 437:2`.

| Branch | AST kind | Source location | L1a disposition |
|---|---|---|---|
| B1 | if decode error → return err | 402:2 | unchanged; exercised by every envelope test |
| B2 | if secret-like field → return err | 405:2 | unchanged; `TestClosedBarRejectsSecretLikeField` proves the breakout kinds still pass through it first |
| B3 | switch kind (new) | 410:2 | **new** — dispatch placement guarded by `TestClosedBarDispatchRunsBeforeTheLegacyTypeMap` (mutant "dispatch after legacy map" killed) |
| B4 | case KindOfficialClosedBar1m → strict decoder | 411:2 | **new** — all `TestClosedBar*` decoder tests (unknown field/enum, float, raw↔minor, session day, finality, successor, …) |
| B5 | case KindOfficialQuoteL1 → strict decoder | 414:2 | **new** — `TestQuoteL1RejectsEveryContractViolation`, `TestQuoteL1RejectsUnknownEnumValues`, `TestQuoteL1EnvelopeAcceptsCanonicalQuote` |
| B6 | range legacy type map | 419:2 | unchanged (legacy) |
| B7 | if field absent → continue | 421:3 | unchanged (legacy) |
| B8 | switch expected type | 425:3 | unchanged (legacy) |
| B9 | case bool | 426:3 | unchanged (legacy) |
| B10 | case string | 428:3 | unchanged (legacy) |
| B11 | case number | 430:3 | unchanged (legacy) |
| B12 | if !valid → return error | 433:3 | unchanged (legacy); `TestClosedBarDispatchRunsBeforeTheLegacyTypeMap` asserts a breakout payload never reaches this message |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `json.NewDecoder` | 399:13 |
| `bytes.NewReader` | 399:29 |
| `decoder.UseNumber` | 400:2 |
| `decoder.Decode` | 402:12 |
| `rejectSecretFields` | 405:12 |
| `decodeClosedBar1mObject` | 412:13 |
| `decodeQuoteL1Object` | 415:13 |
| `fmt.Errorf` | 434:11 |

### 손으로 쓴 주석 — 완전성 주장이 아니다

위 표가 `ast.json` 의 호출 전부이고 `tools/logic-map/role_check.py` 가 1:1 로 대조한다.
아래는 그 자리에 있던 손으로 쓴 분석이다. 줄 번호만 적거나 한 줄이 호출 여럿을 묶어서
기계가 읽지 못했고, 그래서 잘려 있어도 게이트가 조용했다(a112 4차 리뷰가 센 39 개 중 하나).
근거로서의 값은 남으므로 지우지 않는다. **좌표는 위 표가 정본이다** — 아래 산문의
줄 번호는 그때 손으로 읽은 값이고, 어긋나면 위 표가 맞다.

| Callee (hand-written note) | Source location | Evidence |
|---|---|---|
| `json.NewDecoder(...).Decode` | 399–402 | unchanged |
| `rejectSecretFields(object)` | 405 | unchanged; reused by the strict decoders through this hook |
| `decodeClosedBar1mObject(object)` | 412 | new (breakout_bar.go); returns typed `ValidationError` on any contract violation |
| `decodeQuoteL1Object(object)` | 415 | new (breakout_bar.go) |

## State mutations and fallbacks

- None. No assignments outside locals, no defers/goroutines (AST). No fallback: a breakout kind either passes its strict decoder or the envelope is refused.

## Safety conclusion

- The inserted switch returns before the legacy map, so legacy kinds are byte-for-byte unaffected (pinned vector, whole existing suite and four consumer packages green). Both reviewers confirmed the diff is exactly this insertion plus the `kindSupportsMarket` case list.
