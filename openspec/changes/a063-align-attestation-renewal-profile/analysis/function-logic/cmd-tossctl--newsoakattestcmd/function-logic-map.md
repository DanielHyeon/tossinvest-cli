# Function Logic Map: `newSoakAttestCmd`

- Source: `cmd/tossctl/soak.go`; E-based current S snapshot, SHA-256 `5602b691caadfdf4694fabe6f475e4a8fbc5e5195a4a4cd87ad63132c783a4c2` (retrospective, not pre-edit evidence).
- AST evidence: `ast.json`; it contains no B-id control-flow node.

## Inputs and invariants

`root` supplies profile/path resolution later in `runSoakAttest`; `opts` is the shared soak flag state. Construction creates the local `soak attest` Cobra command, binds `RunE` to `runSoakAttest`, and binds record, minimum-days, validity, output, verifier, notes, renewal-status, and repeatable supervised-proof-record flags. Its annotations set `source=local`; no `mutating=true` annotation is installed. The help text states that unmet qualification writes nothing and that read-only soak evidence cannot satisfy the live-only engine requirements.

## Branches and early returns

There is no conditional, switch, range, or early return: the sole return is the constructed `*cobra.Command`.

## Calls and live bindings

The calls create/bind flags (`StringVar`, `IntVar`, `DurationVar`, `BoolVar`, `StringSliceVar`) and the callback delegates all errors to `runSoakAttest`; there is no timeout/retry or live-service call in this constructor.

## State mutations and fallbacks

Flag binding mutates only the supplied options and command definition, not an account, engine, or profile artifact. The callback may later write local attestation/status files, but construction itself has no fallback or durable mutation.

## Safety conclusion

It neither places nor cancels an order and does not enable a runtime gate.
