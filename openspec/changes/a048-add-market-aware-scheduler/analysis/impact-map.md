# a048 implementation impact map

## Hard evidence

- Base: `0ab482fe2e51748fe3e3e0b039ff4748b8daf8da`.
- CodeGraph was synchronised against this worktree before implementation.
- `internal/clock.Market.RegularSession` already owns KR/US IANA locations and
  half-open regular-session/DST primitives. It deliberately has no holiday
  table; authoritative holidays and early closes belong to the official
  calendar adapter added by this change.
- `internal/official.Client.MarketCalendar` is currently an untyped operational
  read. a048 adds a parallel typed read so existing MCP/CLI callers and their
  JSON shape do not change.
- `internal/official.Client.RateBudget` already records per-normalised-path
  remaining/reset/observed-at. a048 exposes its exact reset derivation as a pure
  read-only helper reused by the coordinator; request and retry behaviour do not
  change.
- `cmd/tossctl.candidateIntervals` is used only by the standalone read-only
  candidate scan/watch commands. The a047 entry runtime does not exist at this
  base, so gating these commands would break OFF/upstream behaviour and would
  not gate a strategy entry loop. a048 therefore exposes a dormant candidate /
  entry budget seam for a047 instead.
- `internal/app/engine.Runtime.Run` supervises reconcile, exit and fill-detect
  loops supplied by production wiring. Scheduler decisions are not inserted
  into this loop set; a closed market can gate only a future entry loop.
- a047 contains planning artifacts only. There is no activation-manifest
  package or runtime symbol to import. Auto-resume therefore uses a sealed
  verifier interface whose method is package-private; production has zero
  implementations and zero activation mint paths, so nil/unavailable
  verification is fail-closed. When a047 lands, it must add a scheduler-internal
  trusted manifest adapter/constructor rather than exporting a boolean or a
  caller-constructible claim.

## Owned implementation surface

| Area | Existing/new surface | a048 decision |
|---|---|---|
| Session authority | `internal/official`, new typed response | Preserve legacy map read; add strict typed method. |
| Decision/calendar/budget | new `internal/scheduler` | Pure typed decision, `sha256:` canonical digest, freshness, reserve, desired-state and sealed manifest seam. |
| Desired persistence | new scheduler file store | Missing file is revision-0 OFF/OFF/none/regular; expected-revision CAS under a cross-process flock, atomic 0600 write, and no approval synthesis. A committed OFF cannot be overwritten by a stale ON. |
| Entry cadence | new scheduler budget coordinator plus official reset parser helper | Missing/stale provenance grants zero candidate/entry/analytics polls; analytics may consume at most half of discretionary capacity so candidate/entry retain their higher-priority share; safety classes remain outside this deny path. Low-priority commitments and request-start observation cycles are opaque, one-shot and generation-scoped. Reconciliation uses a monotonic completion watermark rather than wall timestamps; issued commitment memory has an absolute 256-per-endpoint/generation cap. Manual evidence cannot reset generation or either issuance cap even when commitments are empty. Reset trust reuses official's exact threshold, plausibility, kind and overflow-safe derivation semantics. |
| Engine | tests only | Prove closed-market decisions do not stop reconcile/exit/filldetect. No a047 entry loop is invented. |
| Console | `Console.routes`, new read-only status page | `strategy-runtime > 시장·일정`; server-defined labels/chips only; no free-form controls or write action. Calendar provenance is accepted only as an official canonical digest plus fetched-at. |
| CLI wiring | `runConsole` | The sole `internal/console` importer injects a desired-state reader and a typed calendar read through the already-shared official client. Market-none performs no fetch; selected KR/US fetches provenance on page read. There is no save, activation or toggle seam. |

The canonical page path is `/strategy-runtime/market-schedule`; no legacy alias
is registered, so there is one authorization surface. The current console has no
six-category navigation model yet, so the page temporarily reuses the
`optimization` nav highlight. a050 must move that one canonical path into its
`strategy-runtime > 시장·일정` navigation owner rather than add a second route.

UI evidence fixes that route to authenticated GET/HEAD only; POST, PUT, PATCH
and DELETE return 405. The rendered DOM contains no form, input, textarea,
select, button or contenteditable node, exposes no raw reader error/path or
arbitrary state/reason/provenance string, and keeps server-defined enum labels
only. Tests cover the 360px responsive CSS contract, native keyboard links and
focus indicator, semantic main/h1/table landmarks, ARIA label targets, CSP and
DOM structure. StockOS lane-console information architecture is reference only;
it does not widen this page's value or control surface.

## Safety boundary

- No LIVE order, Guardian, exit judgement or broker mutation path changes.
- No existing candidate/discovery cadence changes while scheduler is OFF.
- Calendar or budget uncertainty can only reduce future entry polling.
- Unknown poll/reset enum values and equal-time budget corrections fail closed;
  equal-time evidence can only reduce remaining capacity.
- Initial/manual budget observations can seed or conservatively refresh evidence
  but cannot reconcile a commitment, advance a generation, or clear commitment /
  observation-cycle issuance memory even when no commitment remains. Only an
  opaque request cycle begun after completion can do so, including across
  wall-clock rollback or held responses.
- Reset trust is parser-equivalent to `internal/official`: exact epoch threshold,
  inclusive -1m/+24h plausibility, raw-kind-derived equality, overflow-safe
  delta conversion/addition, and no absolute-value negation of `MinInt` duration.
- Delta reset derivation tolerates at most one second around a fixed first
  anchor, never walks that anchor transitively, exposes the earliest compatible
  deadline, and waits through the latest plausible prior boundary before a new
  generation. Epoch reset identity remains exact.
- The console page is read-only and cannot create an approval or activation
  manifest.
- a048 is dormant, not fully activated: `Decision` can reach
  `ENTRY_ALLOWED` only with the opaque capability returned by a successful
  `Restore`, but no production verifier can currently mint that capability.
