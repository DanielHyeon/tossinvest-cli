# Function Logic Map: `readProtectionArtifact`

- Source: `internal/attest/protection_signature.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| artifact + direct parent | exact basename, owner, modes, regular/single-link and stable identities | filesystem | any metadata absence/change fails closed |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B8 | validate parent/file before open, fstat identity, bounded read, then revalidate opened file, path and parent snapshots | read-only file descriptor | typed sentinel | symlink/hardlink/mode/owner/file+parent TOCTOU |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| lstat/open/fstat/read/restat | complete mediation around read | no retry | CodeGraph + AST |

## State mutations and fallbacks

- None; callback exists only for deterministic same-package race tests.

## Safety conclusion

- Safe edit boundary: add post-read parent identity and policy checks without relaxing file checks.
- High-risk impact: yes.
