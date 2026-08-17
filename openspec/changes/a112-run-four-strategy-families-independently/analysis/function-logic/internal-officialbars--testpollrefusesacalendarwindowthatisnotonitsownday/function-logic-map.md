# Function Logic Map: `TestPollRefusesACalendarWindowThatIsNotOnItsOwnDay`

- Source: `internal/officialbars/producer_test.go`
- **Revision: `base`.** The AST and every coordinate below describe the file **at the frozen comparison base `d9d9f71f`**, not the working tree. Read the body with `git show d9d9f71f:internal/officialbars/producer_test.go`.
- Source SHA-256: `1d36e945513968599fdb553b18c1b8160436965aeb68ca345be054d25f12afde` (the base file; `git show d9d9f71f:internal/officialbars/producer_test.go | sha256sum` verified equal to `ast.json` `source_sha256`, 2026-08-18)
- Signature: `TestPollRefusesACalendarWindowThatIsNotOnItsOwnDay(t *testing.T)`
- Source range: `1473:1`–`1497:2` at the base (ast.json `start`/`end`); the same 25 lines sit at `1492:1`–`1516:2` in the working tree
- AST evidence: `ast.json` regenerated 2026-08-18, revision `base`; branches 3, returns 0, calls 18, assignments 8, defers 0, go statements 0.
- Disposition: **Test helper / test function caught by the gate because it existed at the frozen base `d9d9f71f` and the decision-30 correction edited it. Enumerated for gate completeness: these carry no production branch.**
- Why this bundle is base-revision, measured rather than assumed: the function's 25 lines are **byte-identical** between `d9d9f71f` and the working tree (verified line by line; it moved +19). What the correction appended is the block of four new decision-30 tests **below** it, and `git diff --unified=0 d9d9f71f` reports that block as `@@ -1497,0 +1517,153 @@` — an insertion at base line 1497, which `check_analysis`'s `intersects()` counts as touching `start..end+1` on the base side while the new-side lines 1517–1669 fall past the current range 1492–1516. Hence the base revision. Nothing in the body constructs a bar, so no broker label ever crosses this test.
- Risk scan: `risk-pattern-report.md` (no configured risk pattern matched).

## Inputs and invariants

- The dedicated regression for ruling 24's second half (`PollClosedBars` B8/B9): a calendar snapshot whose `Today.Regular` window sits on **another** market-local day must be refused as `CALENDAR_INVALID`, before any request is sent.
- The snapshot is deliberately hand-built and then corrupted: `usCalendarAt(t, usOpen−1h)` produces a valid one (base line 1480), and base line 1481 replaces `Today.Regular` with `{Open: usOpen.AddDate(0,0,-1), Close: usClose.AddDate(0,0,-1)}`. The base-side comment at 1478–1479 states why: a snapshot that came through `AdaptOfficialCalendar` already satisfies the invariant, so only a caller who assembled a calendar by hand can reach this arm.
- `pollAt` is `usOpen + 10m` — an ordinary in-session instant. The refusal therefore has to come from the **window's** day, not from the poll instant's day; the latter is `PollClosedBars` B11 and has its own subtest elsewhere.
- The reader is `readerOf(usPage(t, pollAt, "unused", ""))` (base 1484) — one page that must never be requested. The fixture label `"unused"` says so out loud, and B3 is what enforces it.
- The store is real (`openTestStore` on `marketclock.NewFake(pollAt+1s)`), so "nothing was written" is a checkable property rather than an assumption; the assertions below stop at the refusal, which is itself the proof that no append was reached.

## Branches and early returns

The AST reports 3 branches and **no return nodes** — the function ends by falling off the bottom, and every branch is a `t.Fatal`/`t.Fatalf` assertion guard.

| Branch | AST kind | Source location (base) | What the guard proves | Test disposition |
|---|---|---|---|---|
| B1 | if | 1488:2 | `err == nil` after the poll → `t.Fatal("a regular window belonging to another day was accepted")` | proves the poll **refuses at all**. Without it the two guards below could pass vacuously on an accepted poll that merely happened to store nothing |
| B2 | if | 1491:2 | `refusalOf(t, err).Reason != RefusalCalendarInvalid` → `t.Fatalf` | proves the refusal is the *right* one: `CALENDAR_INVALID` from the window-edge day check (B9), not `CALENDAR_DAY_MISMATCH` (B11, the poll instant's day), not `NO_REGULAR_SESSION` (B7), not a reader or store error. `refusalOf` unwraps to the typed `*PollRefusal`, so the reason is read from the type, never from a message substring |
| B3 | if | 1494:2 | `len(reader.calls) != 0` → `t.Fatalf` | proves the fail-closed ordering: an invalid calendar spends **zero** quota on the shared OpenAPI credential. This is the assertion `internal-officialbars--pollclosedbars`' safety conclusion cites for "an invalid calendar sends no request at all" |

## Calls and live bindings

| Callee expression | Source location (base) | Evidence |
|---|---|---|
| `t.Parallel`, `context.Background` | 1474, 1475 | ordinary harness; the test owns its store, so parallel execution is safe |
| `usOpen.Add(10 * time.Minute)` | 1476 | an unremarkable in-session poll instant, so the refusal cannot come from the poll's own day |
| `openTestStore(t, marketclock.NewFake(pollAt.Add(time.Second)))` | 1477 | a real `strategyevidence.Store` on a **fake** clock under `t.TempDir()` |
| `usCalendarAt(t, usOpen.Add(-time.Hour))` | 1480 | a valid snapshot, fetched an hour before the open |
| `usOpen.AddDate(0,0,-1)`, `usClose.AddDate(0,0,-1)` | 1482 | the corruption: both window edges moved one day back — exactly the shape `PollClosedBars` B9 exists to catch |
| `readerOf(usPage(t, pollAt, "unused", ""))` | 1484 | **the load-bearing fixture**: a `fakeReader` that records every call; the page is labelled `"unused"` because it must never be served |
| `PollClosedBars` | 1485 | the function under test |
| `refusalOf` | 1491 | unwraps the error to `*PollRefusal` so B2 asserts a typed reason |
| `len(reader.calls)` | 1494 | the zero-request assertion |

## State mutations and fallbacks

- Locals only (8 AST assignments, base lines 1475–1491): `ctx`, `pollAt`, `store`, `foreign`, `foreign.Today.Regular`, `reader`, `_`/`err`, `got`. No package-level state, no goroutines, no defers.
- Side effects are confined to the test: one SQLite store under `t.TempDir()` on a fake clock, and one in-memory `fakeReader`. No HTTP server, no credential, no network.
- No fallback: all three guards are `t.Fatal`/`t.Fatalf`, so the first violated property stops the test rather than letting a later assertion report a misleading second failure.

## Safety conclusion

- Safe edit boundary: a test function whose body the decision-30 correction did not change. Editing it is safe; deleting it is not, because it is the only test that drives `PollClosedBars` B9 and the only one that proves the window-day refusal spends no quota.
- High-risk impact: no. There is no production branch behind these rows — they are assertions over `PollClosedBars`, whose own enumeration lives in `internal-officialbars--pollclosedbars`.
- What it protects on the High-risk side: the shared-credential quota (token-war memory). B3 is the measured form of "fail closed **before** the request"; the same assertion shape appears in `TestPollRefusesInvalidInputBeforeAnyRequest`, and between them every pre-request refusal in `PollClosedBars` is pinned to zero reader calls.
- Honest note on the disposition line: the shared wording says the correction "edited it". For this function specifically, the edit was adjacency in the diff, not a change to its logic — recording that distinction is more useful than claiming a change it did not have.
