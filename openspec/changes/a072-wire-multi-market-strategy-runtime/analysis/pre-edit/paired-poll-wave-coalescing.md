# Paired poll-wave coalescing — pre-edit map

## Function Logic Map

1. KR and US pollers remain independent dispatch consumers.
2. The context owns one mutex-protected paired authority refresh cache.
3. The first market in a poll wave collects the full KR/US graph once.
4. Its peer reuses that immutable assembly only within a bounded one-second coalescing window.
5. Later waves recollect; failures are never cached as successful authority.
6. The original running supervisor remains the projection's latch source; disposable refresh supervisors never replace it.

## Branch Test Map

| Branch | Expected |
|---|---|
| Concurrent KR/US refresh | one paired authority collection |
| Call after coalescing window | new paired collection |
| First refresh failure | peer retries/fails safely; no success cached |
| Runtime supervisor latch | projection follows original running supervisor |
