# Function Logic Map: `Console.handleSettings`

- Source: `internal/console/settings.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
| --- | --- | --- | --- |
| HTTP request | authenticated local console GET | console router/auth middleware | renderer shows read errors; handler does not mutate settings |
| optional seams | nil or injected read/write adapters | command console assembly | each section independently renders unwired/error state |
| strategy descriptor (a047) | server-owned fixed values and blockers only | future `strategy-runtime` provider | absent/invalid means read-only `not_configured/OFF` |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
| --- | --- | --- | --- | --- |
| B1 | each settings/limits/trading/autostart seam is present | performs read and records wired/value/error fields | continues rendering all sections | settings seam tests |
| B2 | seam absent or read fails | no write; marks unwired or load error | continues rendering | nil/error seam tests |
| B3 | updater/stager seams present | reads candidate/receipt state only | continues rendering | system update view tests |
| B4 | page assembled | template render | HTTP response | console settings page tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
| --- | --- | --- | --- |
| `engineRunning`, `engineNoteNow` | advisory process/autostart status | read failure becomes display state | CodeGraph + AST |
| injected `Load` methods | read effective configuration | no retry and no write in GET | CodeGraph + AST |
| `SystemUpdater.Inspect`, `signedReleaseReceipt` | read update provenance | independent from trading controls | CodeGraph + AST |
| `render` | escape and render canonical template | standard console error handling | CodeGraph + AST |

## State mutations and fallbacks

- GET performs no config, order, engine, or toggle mutation.
- Gate state reuses the one limit read rather than consulting a second source.
- a047 card must display fixed server values/provenance/blockers only; a050 owns
  later action endpoints and still forbids arbitrary input.

## Safety conclusion

- Safe edit boundary: add one optional read-only descriptor field/card and keep
  every other section independent. Invalid descriptor fails to OFF, never zero.
- High-risk impact: no mutation, but security-sensitive operator surface.
