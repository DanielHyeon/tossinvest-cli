# Function Logic Map: `ExitObserver.record`

- Source: `internal/app/engine/exitloop.go` (1177-1303)
- AST evidence: `ast.json` — AST branches 16.

## Inputs and invariants

A usable quote produces a durable judgement before any submission. In-process order clearing may withhold liquidation; broker protection cancellation must not share that blocking path.

## Branches and early returns

| Branch | Result |
|---|---|
| B1 | unusable quote is no-op |
| B2-B4 | source/time provenance selected |
| B5-B10 | clear-before-arm, rejudge and delay paths |
| B11-B12 | proposal and intent built only when orderable |
| B13-B15 | journal error classified/pending/quarantine handled |
| B16 | unarmed result does not submit |

## Calls and live bindings

Calls quote guard, `clearTheSymbol`, judgement record, quarantine notification, then `submit` only after arming.

## State mutations and fallbacks

Updates judgement/delay state; only provenance has a fallback.

## Safety conclusion

Do not couple conditional-protection cancellation failure to B8/B9, because those withhold a protective sell.
