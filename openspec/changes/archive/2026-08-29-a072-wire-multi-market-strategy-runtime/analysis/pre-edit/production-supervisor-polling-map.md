# Function Logic Map: paired production supervisor polling

## `engine.Context.NewPairedStrategyEntryProductionAssembly`

Current function collects KR/US authority in one frozen wave but always calls the value-only dormant
constructor. New branch: for each market, require verified automation plus complete schedule,
candidate, route, FX, proposal, five-bucket, account, Guardian, Gateway, sealed protection and current
entry-gate observations. Only then create an effective market worker. Its cycle recollects the full
paired wave before selecting the target market, so no startup snapshot becomes order authority.
Incomplete market authority creates a dormant worker and does not suppress the peer.

## `engine.StrategyEntrySupervisor.Run` / `runMarket`

Current Run starts two queue-driven children; only external `Trigger` can enqueue. Production has no
trigger caller. Add an opt-in bounded poll interval to the worker descriptor. Run starts one poller
per effective opted-in market behind the same start barrier, immediately enqueues one cycle, then
uses the injected clock's cancellable Sleep. Existing workers with zero interval remain exactly
queue-driven. The poller never calls a broker; it only invokes the already bounded worker cycle.

An opted-in production cycle declares that it recollects authority. `evaluationState` therefore does
not expire the construction snapshot, while the cycle itself must revalidate all opaque authorities
and Gateway gates. Ordinary workers retain the current fixed-expiry behavior.

## `engine.New` Context literal

Add one private central-owner coordinator. Every fresh KR/US recollection shares it, preventing a
new assembly or peer market from fencing the other on each poll. It mints no owner until an admitted
first leg reaches dispatch issuance.

# Branch Test Map

| Branch | Expected |
|---|---|
| gate OFF / automation unverified | both workers dormant, no poller |
| KR and US complete | both immediate cycles start concurrently |
| KR incomplete, US complete | KR dormant; US cycles; safety loops unaffected |
| US incomplete, KR complete | inverse isolation |
| poll interval zero | existing explicit-trigger semantics unchanged |
| recollect returns no fresh proposal | nil/no-op, no lease and no market latch |
| sealed protection or entry gate refuses | zero lease/transport during assembly |
| same proposal replay | no duplicate submission; consumed lease treated as no-op |
| context cancellation | pollers and children drain before Run returns |
