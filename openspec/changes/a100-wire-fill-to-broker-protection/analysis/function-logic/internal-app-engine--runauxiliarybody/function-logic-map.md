# Function Logic Map: `runAuxiliaryBody`

- Source: `internal/app/engine/auxiliary.go` (123-130)
- AST evidence: `ast.json` — AST branches 1.

## Inputs and invariants

This is the one recover boundary for every auxiliary `Run`. A panic becomes `ErrAuxiliaryPanicked`; it must not terminate the process and thereby stop protective loops.

## Branches and early returns

| Branch | Result | Existing / planned test |
|---|---|---|
| B1 | recovered panic becomes typed auxiliary stop error | `TestAnAuxiliaryExecutorThatReturnsDoesNotStopTheEngine` |

## Calls and live bindings

Calls only `aux.Run` under `defer recover`; `runAuxiliary` owns classification and optional stop callback.

## State mutations and fallbacks

No mutation; panic conversion is the sole fallback.

## Safety conclusion

If A100 uses auxiliary composition, it must preserve this recovery boundary and avoid turning its typed error into a self-blocking authority event.
