## Deployment failure review

### Evidence

- Both `tossos` and `httpapi` restarted with exit 126.
- tini logged `exec /usr/local/bin/tossos-entrypoint failed: Permission denied`.
- repository mode and Git index mode were `0644`/`100644`.

### Scope decision

The fix is restricted to deterministic image packaging, a static regression test and the
operator runbook. It does not add a mutation route, touch configuration/journal data, start
the engine, or change any trading authorization.

Function Logic Map: not-applicable — no existing Go function is modified; the only Go
change is a new static test function and production behavior changes in one Dockerfile COPY.

### Review status

- Independent security review: PASS, P0=0/P1=0/P2=0 after no-cache image inspection
  confirmed `0755 root:root`, non-root uid/gid 10001 execution and no writable/setid
  expansion.
- Independent test/maintainability review: PASS, P0=0/P1=0/P2=0; the exact
  Dockerfile instruction test, NTFS/Git 0644 source condition, no-cache build,
  image mode inspection and non-root `test -x` evidence cover the regression.
