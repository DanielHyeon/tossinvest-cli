# Function Logic Map: `Context.ExitObserver`

- Source: `internal/app/engine/exitwiring.go:319-364`
- AST evidence: `ast.json` (Pre-Edit HEAD `a284a2b0a`)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| Context | nonnil, verified automation | engine startup | existing unavailable error |
| Guardian | ReductionIssuer | verified injection | existing unavailable error |
| lower-limit board | Context-owned immutable board | `NewContext` | caller cannot supply/replace board |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | nil Context | none | unavailable error | existing test |
| B2 | gate unverified | none | unavailable error | existing test |
| B3 | Guardian not issuer | none | unavailable error | existing test |
| B4 | Names nil | default names | continue | existing test |
| B5 | Notifier present | overwrite record-only sinks | continue | existing test |
| B6 | Floor nil | default floor | continue | a087 board overwrite test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `exitSideRetrier` | existing observation retry discipline | no a087 change | AST setup |
| Context fields | authoritative production wiring | no caller authority | AST assignments |
| `NewExitObserver` | validate final observer | returns existing validation error | AST return |

## State mutations and fallbacks

- `opts.LowerLimitBoard` is always overwritten from `Context`, following Retrier/Issuer/Submit authority.
- caller-supplied board or admission cannot affect production observer.

## Safety conclusion

- Safe edit boundary: one production override assignment.
- High-risk impact: yes — prevents a human/fixture-provided board from bypassing dormant wiring.
