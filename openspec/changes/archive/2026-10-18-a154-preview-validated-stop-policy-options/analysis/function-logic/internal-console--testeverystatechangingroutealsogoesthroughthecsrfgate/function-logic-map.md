# Function Logic Map: `TestEveryStateChangingRouteAlsoGoesThroughTheCSRFGate`

- Source: `internal/console/static_test.go`; fresh AST: `ast.json`; risk report: `risk-pattern-report.md`.

## Inputs and invariants

The scanner reads package source AST and reports opaque registrations or missing security gates; no runtime state is mutated. Existing guard and negative cases must remain enforced.

## Branches and early returns

| Branch | AST shape | Location | Contract |
|---|---|---|---|
| B1 | range | line 416 | Preserve existing scanner/error semantics; compute-only exception is limited to the declared path. |
| B2 | switch | line 418 | Preserve existing scanner/error semantics; compute-only exception is limited to the declared path. |
| B3 | case | line 419 | Preserve existing scanner/error semantics; compute-only exception is limited to the declared path. |
| B4 | case | line 421 | Preserve existing scanner/error semantics; compute-only exception is limited to the declared path. |
| B5 | case | line 423 | Preserve existing scanner/error semantics; compute-only exception is limited to the declared path. |
| B6 | range | line 427 | Preserve existing scanner/error semantics; compute-only exception is limited to the declared path. |
| B7 | if | line 428 | Preserve existing scanner/error semantics; compute-only exception is limited to the declared path. |

## Calls and live bindings

Parser, AST inspection and testing errors only; no network, broker, DB, settings or live bindings. See AST for exact calls.

## State mutations and fallbacks

Local route list and test failure reporting only.

## Safety conclusion

Session gate is recognized through one explicit no-store wrapper; state-changing routes retain existing CSRF checks.
