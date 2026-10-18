# Function Logic Map: `registeredRoutes`

- Source: `internal/console/static_test.go`; fresh AST: `ast.json`; risk report: `risk-pattern-report.md`.

## Inputs and invariants

The scanner reads package source AST and reports opaque registrations or missing security gates; no runtime state is mutated. Existing guard and negative cases must remain enforced.

## Branches and early returns

| Branch | AST shape | Location | Contract |
|---|---|---|---|
| B1 | range | line 131 | Preserve existing scanner/error semantics; compute-only exception is limited to the declared path. |
| B2 | range | line 137 | Preserve existing scanner/error semantics; compute-only exception is limited to the declared path. |
| B3 | if | line 145 | Preserve existing scanner/error semantics; compute-only exception is limited to the declared path. |
| B4 | if | line 152 | Preserve existing scanner/error semantics; compute-only exception is limited to the declared path. |
| B5 | if | line 156 | Preserve existing scanner/error semantics; compute-only exception is limited to the declared path. |
| B6 | if | line 160 | Preserve existing scanner/error semantics; compute-only exception is limited to the declared path. |
| B7 | if | line 165 | Preserve existing scanner/error semantics; compute-only exception is limited to the declared path. |
| B8 | if | line 170 | Preserve existing scanner/error semantics; compute-only exception is limited to the declared path. |
| B9 | range | line 175 | Preserve existing scanner/error semantics; compute-only exception is limited to the declared path. |
| B10 | if | line 176 | Preserve existing scanner/error semantics; compute-only exception is limited to the declared path. |
| B11 | if | line 183 | Preserve existing scanner/error semantics; compute-only exception is limited to the declared path. |
| B12 | if | line 184 | Preserve existing scanner/error semantics; compute-only exception is limited to the declared path. |
| B13 | switch | line 185 | Preserve existing scanner/error semantics; compute-only exception is limited to the declared path. |
| B14 | case | line 186 | Preserve existing scanner/error semantics; compute-only exception is limited to the declared path. |
| B15 | case | line 188 | Preserve existing scanner/error semantics; compute-only exception is limited to the declared path. |
| B16 | if | line 200 | Preserve existing scanner/error semantics; compute-only exception is limited to the declared path. |

## Calls and live bindings

Parser, AST inspection and testing errors only; no network, broker, DB, settings or live bindings. See AST for exact calls.

## State mutations and fallbacks

Local route list and test failure reporting only.

## Safety conclusion

Session gate is recognized through one explicit no-store wrapper; state-changing routes retain existing CSRF checks.
