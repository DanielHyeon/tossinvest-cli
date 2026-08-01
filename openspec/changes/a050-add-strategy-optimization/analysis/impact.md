# a050 implementation impact

## Integration baseline

- Implementation comparison base: `948e7211a1109a6d8d68204959cd83a2f1b957f1`.
- This is the a050 worktree merge that includes the independently gated a041–a049
  integration line. The base was deliberately refreshed after that merge so
  Function Logic Map checking does not misclassify already reviewed upstream
  functions as a050 edits.

## Hard evidence and ownership

| Boundary | Current owner / API | a050 use | Failure rule |
|---|---|---|---|
| writable setting metadata | a041 `exitpolicy.CommonPolicyFieldDescriptor` through transport-neutral `settingmeta.FieldDescriptor` | exact key/owner/category manifest; finite option IDs only | missing, duplicate, unexpected, or drifted coverage refuses registry construction |
| a044 position management | `positionpolicy.Descriptor` plus current-row command seam | read-only deep link and status; no settingmeta adapter is invented | absent engine commander stays unwired/read-only |
| a046 candidate filters | `candidate.CandidateFilterDescriptors` | dormant read-only evidence/status projection | unapproved remains non-numeric and non-writable |
| a047 runtime | `strategyengine.DormantRuntimeDescriptor` and read-only console projection | shows lane/autostart/gate/LIVE as separate authority states | no activation manifest or authority is minted |
| a049 performance | narrow `performance.ReadOnly.Dashboard` over an existing clean checkpoint | fixed 30-day/all-market/all-lane evidence adapter; no create, migration, WAL writer or collector authority | active WAL/change race, query mismatch, source age over 72h, link missing, missing metric, insufficient sample, or read failure blocks evidence-backed candidates |
| trading journal | `journal.ReadOnly` and `performancejournal.New(*journal.ReadOnly)` | no writable journal handle crosses console/httpapi boundaries | engine remains the sole trading-state writer |
| optimization writes | actor-bound `optimization.Commander` | immutable HMAC-bound candidate, preview capability, version CAS, audit, append-only rollback | actor mismatch, metadata/evidence corruption, stale base, invalid option, expired capability, trigger drift, or audit failure aborts the transaction |

Direct import and AST inspection at this baseline confirms that
`internal/optimization` has no broker, order, journal, lane, gate, kill-switch,
process-control, or position-mutation dependency. Production composition lives in
`cmd/tossctl`; HTML is an adapter in `internal/console` and receives only the
narrow commander/read interfaces.

## Function risk set

The following existing functions are high-risk and have Function Logic Map and
Branch Test Map evidence before their post-integration edits:

- `optimization.Open`, `Store.init`, `Store.Apply`, `Store.snapshot`,
  `Store.history`, `Store.audit`, and `Store.Close`;
- `optimization.BuildRegistry` and `CoreRegistry`;
- `main.newConsoleOptimizationCommander` and `main.runConsole`;
- `console.handleOptimizationSave` and `optimizationFieldViews`.

New evidence adapters, canonical digest helpers, migrations, trigger verification,
and actor-bound wrappers have their own Function Logic/Branch Test Maps and focused
tamper, cross-actor, freshness, fault, migration, and race tests.

## UI authority and input boundary

The write path has one owner-approved writable field at this HEAD. The operator
chooses a server preset, reviews a read-only before/after page, waits three seconds
for contextual-risk changes, checks one explicit confirmation box, and presses
apply. There is no text, number, range, textarea, contenteditable, symbol, reason,
or typed-confirmation surface. a044/a046/a047/a049 remain separate read-only views
until their owner exposes an approved `settingmeta` write provider.

Every mutation accepts only `application/x-www-form-urlencoded` with an exact
server-known field set. Multipart, duplicate, missing, or unexpected fields are
rejected before the commander is called.

## Deferred activation boundary

a050 records canonical desired settings but does not install or approve an
activation manifest. Consequently contextual/new-position changes clear effective
entry and cannot become LIVE authority. Engine consumption that could activate a
new manifest remains a separately reviewed activation concern; a050 provides no
route or fallback that enables it.
