# Function Logic Map: `ExitObserver.submit`

- Source: `internal/app/engine/exitloop.go` (1343-1418)
- AST evidence: `ast.json` — AST branches 11.

## Inputs and invariants

Only a durable armed proposal enters. Floor/issuance/attach validation occurs before broker submission; in-doubt retains the arm to prevent duplicate sells.

## Branches and early returns

| Branch | Result |
|---|---|
| B1 | floor failure returns |
| B2 | zero floor releases |
| B3 | issuance refusal releases |
| B4 | attach failure stops submission |
| B5 | sell intent refusal releases |
| B6-B10 | confirmed/in-doubt/in-flight/default classification |
| B11 | diagnostic detail fallback |

## Calls and live bindings

Calls floor, issuer, journal attach, sell intent, and submit gateway.

## State mutations and fallbacks

Releases only when no confirmed or in-doubt live order exists.

## Safety conclusion

No protection convergence failure may become another submit-blocking condition.
