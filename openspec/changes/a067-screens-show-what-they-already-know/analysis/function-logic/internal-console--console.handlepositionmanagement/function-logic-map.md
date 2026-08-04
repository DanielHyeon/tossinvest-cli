# Function Logic Map: `Console.handlePositionManagement`

- Source: `internal/console/position_policy.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`
- Base commit: `openspec/changes/a067-screens-show-what-they-already-know/base-commit.txt`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `r.URL.Query()["notice"]` | the redirect notice after an apply | this handler's own redirect | absent -> no notice |
| `c.opts.Settings` | nil means no desired-config seam | wiring | nil -> `DesiredErr` names the missing seam |
| `c.opts.PositionPolicies` | nil means no engine-owned commander | wiring | nil -> the page renders unwired and returns early |
| `PositionPolicies.Runtime` | running engine settings | engine | error -> `RuntimeErr`, effective stays unknown |
| `PositionPolicies.List` | one row per live position | engine | error -> `LoadErr` and an early return |
| `c.holdings` (a067) | the holdings cache this console has already filled | `/positions` | empty -> no names, which is the pre-a067 rendering |

**Invariant**: this screen is read-only over the broker. It made no broker call
before a067 and makes none after it -- `holdingNames` peeks the cache and never
refreshes it.

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `c.opts.Settings == nil` | `DesiredErr` names the unwired seam | falls through | existing settings tests |
| B2 | `else if` desired load failed | `DesiredErr` carries the error | falls through | existing |
| B3 | (the load-error condition) | same | falls through | existing |
| B4 | `else` desired load succeeded | fills `Desired`, `DesiredVerdict` | falls through | existing |
| B5 | `c.opts.PositionPolicies == nil` | renders the unwired page | **early return** | existing unwired test |
| B6 | `Runtime` returned an error | `RuntimeErr` | falls through | existing |
| B7 | `else` runtime known | fills `Effective`, `BlockSource` | falls through | existing |
| B8 | `for _, block := range runtime.Blocks` | appends account-level blocks | none | existing reconcile tests |
| B9 | `List` returned an error | `LoadErr` | **early return** | existing |
| B10 | `for _, state := range states` | one view row per position | none | existing |
| B11 | `management.Block != nil` | per-row reconcile block | none | existing |
| B12 | `state.Status == StatusManaged` | inherit/override actions | falls through | existing action tests |
| B13 | `else if` released and externally eligible | re-adopt action | falls through | existing |
| B14 | `for _, policy := range RegisteredCommonPolicies()` | one override action per policy | none | existing |
| B15 | `state.ExternalLifecycleEligible()` | release action | none | existing |
| B16 | (the released/eligible condition) | same as B13 | none | existing |

**a067 changes two lines**: one `names := c.holdingNames(c.now())` before B10, and
`Name:` in the `policyRowView` literal inside B10. No branch was added, removed or
reordered, and no action token, preview or apply path is touched.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `Settings.Load` | desired adoption config | error becomes `DesiredErr` | AST |
| `PositionPolicies.Runtime` | running engine settings | error becomes `RuntimeErr` | AST |
| `PositionPolicies.List` | lifecycle rows | error becomes `LoadErr` and returns | AST |
| `positionpolicy.ProjectManagement` | shared management verdict | pure | AST |
| `Console.holdingNames` (a067) | stock names already read | no error; peeks the cache, makes no broker call | `internal/console/protection_liveness.go` |
| `Console.policyAction` | signed opaque action tokens | unchanged | AST |

## State mutations and fallbacks

- Builds one page value and renders it. No config write, no journal write, no
  broker call, no engine command.
- An unknown name renders as no name. The handler never derives one from the
  symbol or from another market's holding.

## Safety conclusion

- Safe edit boundary: a display string on the view model. `positionpolicy.State`
  is untouched, so the engine gains no field it might trust.
- High-risk impact: no. The preview/apply capability path is not reached.
- Safety invariant 0.4 (rate budget) unchanged: zero broker calls before and
  after.
