# Function Logic Map: `Console.handleSoakRestart`

- Source: `internal/console/restart.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `w`, `r` | gated POST request | `routes` + `session0` + `mutating` | gate refuses before handler |
| `Options.RestartSoak` | nil or narrow process seam | `cmd/tossctl.runConsole` | nil redirects with unwired notice |
| returned note | arbitrary operator text, blank allowed | restart seam | error redirects; blank gets fixed success text |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | restart seam nil | dashboard redirect only | early return | unwired restart test |
| B2 | restart seam errors | dashboard redirect only | early return with failure | failed restart test |
| B3 | note trims blank | replace local note | fixed success redirect | blank-note test |
| B4 | saved credential generation is incomplete | none | HTTPS setup retry redirect | pending-generation recovery test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `Options.RestartSoak` | process restart | one call; error is displayed; no retry | CodeGraph + AST |
| `redirectDashboard` | PRG redirect | always 303 | current HEAD |

## State mutations and fallbacks

- No account/config mutation occurs in this package.
- Current fallback converts a blank successful seam note into a fixed notice.
- Edit boundary adds serialized credential preflight before the process seam;
  all non-ready states must return without spawning, and an incomplete saved
  generation reopens the setup form instead of becoming an ordinary restart.

## Safety conclusion

- Safe edit boundary: preserve the gated POST and existing restart seam; add only
  credential-state branching and login redirect.
- High-risk impact: yes, credential/authentication path; no LIVE order path.
