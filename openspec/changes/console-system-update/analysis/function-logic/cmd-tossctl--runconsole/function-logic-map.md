# Function Logic Map: `runConsole`

- Source: `cmd/tossctl/console.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| command context/root/options | loopback console invocation | Cobra | path errors propagate |
| executable | absolute current `tossctl` path | `binstamp.SelfPath` | updater disabled with warning |
| engine exclusion | resolved journal directory | `enginelock` | update route refuses when unwired/busy |
| verify activity | strict marker reading | `strictVerifyActivity` | any unclassifiable evidence refuses update |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B5 | nil context or evidence-path resolution failures | background fallback or no server | propagate path errors | command path tests |
| B6-B10 | journal/engine/updater resolution outcomes | warn and disable affected read/update seam | console still serves | wiring tests |
| B11-B12 | local updater invalid or valid | warn/leave nil or inject fixed updater | continue | fixed-updater static test |
| B13-B14 | engine directory exists and flock acquisition outcome | inject real acquire/release closure | handler receives exact lock error | console + enginelock tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `binstamp.SelfPath`, `localupdate.New` | bind fixed current/candidate/rollback paths | never accepts HTTP path | AST + wiring test |
| `enginelock.Acquire` | hold actual engine exclusion across commit | error refuses | AST |
| `strictVerifyActivity` | fail-closed external verification check | checked early and at commit | AST + unit test |
| `console.ListenAndServe` | inject all capabilities and serve | propagates server error | CodeGraph + AST |

## State mutations and fallbacks

- Missing dashboard paths remain nonfatal; missing safety seams make only installation unavailable.
- The updater, engine lock and verification checker cross the package boundary as narrow capabilities.

## Safety conclusion

- Safe edit boundary: add fixed installer and two activity guards without changing live verification approval.
- High-risk impact: yes — this assembly controls executable replacement and exclusion from account work.
