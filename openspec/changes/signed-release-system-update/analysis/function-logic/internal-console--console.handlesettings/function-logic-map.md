# Function Logic Map: `Console.handleSettings`

- Source: `internal/console/settings.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| query notice | untrusted text rendered escaped | request URL + html/template | displayed only |
| settings seams | nil or injected | `Console.Options` | each section independently unwired |
| local candidate | fixed inspection | `SystemUpdater.Inspect` | reason displayed |
| signed release seam | nil or injected | `ReleaseDownloader` | download action hidden/explained |
| process provenance | present only after successful in-process download | console memory | restart labels provenance unknown |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | each settings seam nil | no read | unwired UI | existing tests |
| B2 | each seam injected | read current state | section model populated | existing tests |
| B3 | release downloader injected | no network on GET | stage action rendered | new settings test |
| B4 | candidate exists after restart without receipt | inspect only | provenance unknown label | new settings test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `SystemUpdater.Inspect` | non-executing candidate metadata | local errors become reason | CodeGraph + AST |
| `render` | escaped settings HTML | template error handled centrally | CodeGraph + AST |

## State mutations and fallbacks

- GET remains read-only and performs no release network request.

## Safety conclusion

- Safe edit boundary: populate wiring/provenance presentation fields only.
- High-risk impact: no account risk; update trust wording is security-sensitive.
