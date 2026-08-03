# Function Logic Map: `joinPositions`

- Source: `internal/console/portfolio.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`
- Base commit: `openspec/changes/a061-screens-show-what-they-already-know/base-commit.txt`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `broker []domain.Position` | the holdings cache reading, possibly empty or absent | `holdingsCache` | empty -> the ledger half renders alone |
| `managed []journal.PositionExit` | live positions of every account | `ReadOnly.LivePositionExits` | empty -> the broker half renders alone |
| `journalReadable bool` | whether the ledger answered at all | `journalView.Readable` | false -> rows are unknown, never "unmanaged" |
| `m.Quarantine` (a061) | active quarantine of this position's current generation, or nil | `ReadOnly.accountActiveQuarantines` | nil -> `Quarantined()` false, the pre-a061 rendering |

**Invariant**: a join failure is not a screen failure. Each side renders what it
has, and a symbol present on only one side keeps its own row.

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `for _, h := range broker` | one row per holding, indexed by `(market, symbol)` | none | existing join tests |
| B2 | `for _, m := range managed` | merges the ledger half | none | existing join tests |
| B3 | no broker row claimed this key | appends a ledger-only row | falls through | existing broker-missing test |
| B4 | `else if rows[at].InJournal` | a second live instance gets its own row | falls through | existing duplicate-instance test |
| B5 | (the `else if` condition itself) | same as B4 | falls through | same |
| B6 | sort comparator: markets differ | orders by market then symbol | none | existing ordering tests |

**a061 changes exactly one line** inside B2's body: `row.Quarantine = m.Quarantine`,
beside the existing `row.HasExit` / `row.Exit` copies. No branch was added,
removed or reordered.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `symbolKey` | the `(market, symbol)` join key | pure | AST |
| `attest.Mask` | account reference masking for the screen | pure | AST |
| `journal.Position.ExitEligible`, `Adopted` | eligibility and provenance verdicts | pure | AST |
| `sort.SliceStable` | stable market/symbol ordering | pure | AST |

No I/O of any kind.

## State mutations and fallbacks

- Builds and returns a fresh slice; mutates nothing the caller owns.
- The quarantine pointer is copied, not dereferenced: a nil stays nil and the
  row behaves exactly as it did before a061.

## Safety conclusion

- Safe edit boundary: one field copy in a pure function.
- High-risk impact: no.
- Safety invariant 0.3 untouched.
