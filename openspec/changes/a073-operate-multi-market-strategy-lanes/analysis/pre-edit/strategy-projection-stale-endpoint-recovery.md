# Strategy projection RPC stale endpoint recovery pre-edit map

## Safety invariant

A prior engine killed without cleanup must not prevent safety loops from starting, while a live projection owner must never have its endpoint removed by a second process.

## Existing branch

`Start` requires an exclusive control-directory create and aborts on any pre-existing directory.

## Required branches

- Inspect only the exact control directory, descriptor, and socket paths without following symlinks.
- Refuse cleanup when the descriptor PID is alive or the socket answers.
- For a dead/malformed stale owner, remove only validated regular descriptor/socket entries and then the empty exact directory.
- Retry exclusive creation once after stale cleanup; concurrent creators remain mutually exclusive.

## Branch tests

- SIGKILL-shaped stale descriptor/socket directory is reclaimed.
- A live owner causes a typed refusal and remains reachable.
- Symlink, unexpected entry, permissive mode, and concurrent replacement cases fail closed without broad deletion.
