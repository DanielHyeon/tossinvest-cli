# Function Logic Map: `checkProtectionFileInfo`

- Source: `internal/attest/protection_matrix.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| Function inputs | Opened or lstat `FileInfo` plus expected UID. | Current HEAD + OpenSpec | Fail closed with typed error/decision |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1+ | B1 non-regular/symlink → refuse; B2 mode !=0600 → refuse; B3 owner unavailable/mismatch → refuse; B4 otherwise accept. | No mutation. Existing gap: link count and direct-parent security are outside this function. | Typed refusal or validated result | See branch map |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| Current callees | `fileOwnerUID`; no side effects. | No implicit retry; errors propagate fail-closed | CodeGraph + AST |

## State mutations and fallbacks

- No mutation. Existing gap: link count and direct-parent security are outside this function.

## Safety conclusion

- Safe edit boundary: Add link-count validation and pair it with direct-parent and post-read identity checks in the loader.
- High-risk impact: yes; dormant logic only, no broker mutation or WIRED binding.
