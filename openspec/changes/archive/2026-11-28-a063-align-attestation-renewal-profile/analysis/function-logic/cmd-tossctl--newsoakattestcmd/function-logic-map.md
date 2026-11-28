# Function Logic Map: `newSoakAttestCmd`

- Source: `cmd/tossctl/soak.go`; current snapshot re-extracted 2026-11-28 (file hash moved by same-file edits; function body unchanged by a063 D6), SHA-256 `65a048a5cbc10cc8ebf7513d8b8785a58106878436d19112e32939d595dcde8c` (retrospective, not pre-edit evidence).
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
