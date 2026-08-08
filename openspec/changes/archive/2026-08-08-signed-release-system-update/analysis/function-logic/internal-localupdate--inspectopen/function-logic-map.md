# Function Logic Map: `inspectOpen`

- Source: `internal/localupdate/updater.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| open file | already no-follow descriptor | `inspectPath`/caller | stat/read error |
| expected module/command/platform | exact constructor constants | updater | mismatch error |
| VCS settings | revision and modified flag from Go build info | `debug/buildinfo` | returned for release binding |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | non-regular/non-executable | none | error | existing refusal tests |
| B2 | hash/build info unreadable | descriptor read only | error | existing script test |
| B3 | module/command/platform mismatch | none | error | existing identity tests |
| B4 | all identity checks pass | none | metadata including VCS fields | metadata test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `buildinfo.Read` | parse without execution | malformed input fails closed | CodeGraph + AST |
| `io.Copy`/SHA-256 | bind exact descriptor bytes | read error fails | CodeGraph + AST |

## State mutations and fallbacks

- Adds metadata extraction only; release staging performs expected revision and
  modified-tree refusal.

## Safety conclusion

- Safe edit boundary: expose `vcs.revision` and `vcs.modified` from the already
  parsed settings map.
- High-risk impact: yes; provenance-to-binary binding depends on these fields.
