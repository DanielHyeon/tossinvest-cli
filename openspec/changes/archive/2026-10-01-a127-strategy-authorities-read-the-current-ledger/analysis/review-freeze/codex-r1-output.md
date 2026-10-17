DECLARATION: I did not read or write anything under ~/.codex and made no file changes.

Reviewed HEAD `65fe66e7`. Paths below abbreviate `openspec/changes/a127-strategy-authorities-read-the-current-ledger/` as `a127/`.

1. **P1 — Missing-column contract exceeds what unchanged route SQL guarantees.**  
   `a127/specs/strategy-runtime/spec.md:34–35`: “버전은 같지만 적재기가 읽는 열 하나가 원장에 없다” → “fail-closed”.  
   However, `internal/strategyrouter/production.go:652–653` returns successfully when no active owner exists, **before** preparing the `position_campaigns` query at `:665`. A version-matching database missing `position_campaigns.entry_blocked` can therefore produce a route authority for an owner-free scope. SQLite validates columns only in statements actually prepared. Either require preparation of every read-set query before acceptance, or explicitly narrow the contract to executed queries and justify that boundary. Add an empty-owner missing-column case.

2. **P2 — Direction-specific route errors will be erased unless the wrapper changes.**  
   `a127/design.md:65`: “문구만 방향을 말한다”; `tasks.md:26` requires directional messages.  
   `internal/strategyrouter/production.go:350–352` discards the opener’s error:
   `fmt.Errorf("%w: owner snapshot", ErrProductionRouteUnavailable)`.  
   Changing only `openProductionRouteSnapshot` cannot satisfy the public loader’s diagnostic requirement. Explicitly include propagation through `LoadProductionRouteAuthorityBatch`, preserve `errors.Is(..., ErrProductionRouteUnavailable)`, and assert newer/older wording at both exported loader boundaries.

3. **P2 — S8 is not a distinguishing mutation; its recorded control does not establish the promised guarantee.**  
   `a127/design.md:99`: “읽기 집합 열 하나를 `COALESCE(…,'')` 등으로 우회”.  
   `COALESCE(missing_column,'')` still fails name resolution. So do missing-column references within `LEFT JOIN` and `EXISTS`. This mutation cannot create the claimed bypass. `analysis/measurements/readset-probe.log:7–8` records a quoting error producing an empty column name, not a valid-schema versus deleted-column comparison; the probe source is absent. Replace S8 with a real bypass, such as replacing a selected column with a constant or suppressing a query error. Preserve the reproducible harness and cover the conditional campaign query.

4. **P2 — Proposal contains accidentally embedded tool instructions and a duplicate specification.**  
   `a127/proposal.md:56–97` includes `</invoke>`, `<invoke name="Bash">`, and “`sed -i ... && ... cat > specs/strategy-runtime/spec.md`”, followed by a second requirement copy. `design.md:110–111` also contains closing tool tags. These are committed document contents, not executed here. Remove them before freeze: they contaminate the review artifact and introduce another specification copy that can diverge.

5. **P3 — Census and coordinates are inaccurate despite matching AST evidence.**  
   `a127/proposal.md:33–35` claims six files “journal 패키지 밖” while including journal itself, then describes journal as a separate store. `internal/journal/readonly.go:154` opens the engine ledger; production callers include `cmd/tossctl/httpapi.go:623` and `performance_project.go:109`. State the narrower defensible claim: **two non-journal-package direct read-only openers of the engine ledger**.  
   `proposal.md:20` and `design.md:12` identify route rejection at `:614`; rejection is `production.go:612`, while `:614` returns success. `design.md:63` locates the scope-refusal comment at `:41–45`; its relevant text is `production_snapshot_authority.go:38–41`. All five AST source hashes and stated branch counts do match HEAD.

The central constant-injection approach is sound for the normal production wiring: engine startup calls `journal.Open`, migration reaches `SchemaVersion`, and the supervisor supplies that handle’s path to both loaders. I found no alternate non-test caller supplying these loaders an independently selected ledger. “Always the same migrated file” nevertheless assumes the pathname is not replaced after opening; reopening a path is not inode binding.

The v28–v35 steps do not remove the queried columns. V34 changes per-reservation policy binding, but this loader still reads the bucket-level `p.record_digest`; acceptance evidence should distinguish legacy reservations from v34 reservations with differing policy records. Runtime-read version injection would defeat the check; D2 and S7 correctly prohibit it, but that protection depends on the planned structural assertion.

No P0 established. Static review; no production execution or file changes.

VERDICT: FAIL