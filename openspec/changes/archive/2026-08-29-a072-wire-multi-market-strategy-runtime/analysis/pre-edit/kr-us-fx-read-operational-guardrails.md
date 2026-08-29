# Spec: paired KR/US FX read operational guardrails

**Author:** Codex backend teammate  
**Date:** 2026-08-04  
**Status:** Approved  
**Reviewers:** a072 manager  
**Related contract:** a072 account-base currency wave and production FX authority service

## 2026-08-04 deployment correction

Production deployment proved that adding the strategy-only exchange-rate read to the global startup
attestation invalidates the existing engine evidence and prevents every safety-class loop from starting.
That violates FR-8 and the repository invariant that an unavailable market-entry dependency may close entry
but must not stop reconciliation, protection, exit, fill detection or emergency reduction. The correction
keeps the official FX contract probe, retry budget, immutable origin and market-local authority checks, but
removes the strategy-only GET from the legacy global engine-start endpoint set.

## Context

The paired FX authority service depends on two distinct read domains. KR verifies the selected
account through `GET /api/v1/accounts` before minting same-currency identity evidence. US reads
`GET /api/v1/exchange-rate` from the immutable official Open API origin before minting USD-to-base
evidence. The current engine attestation names the account read but not the exchange-rate read, and
the execution retry catalog has no market-scoped query names or rate budget for either authority
read.

`internal/monitor.Run` is intentionally a WTS session-cookie runner. It cannot authenticate the
official Open API OAuth endpoint. This slice therefore adds an official-source contract/schema
probe that fake tests can execute, but does not place that probe in the WTS `Probes()` registry.
The probe remains operationally inactive until a separate OAuth runner is designed and wired.

## Functional Requirements

- **FR-1:** Engine `RequiredEndpoints()` MUST retain `GET /api/v1/accounts` but MUST NOT include the
  strategy-only `GET /api/v1/exchange-rate`. The exchange-rate read is authorized and refused inside the
  US market-local strategy FX authority path so missing evidence cannot stop safety-class loops.
- **FR-2:** Execgw MUST publish distinct `RequiredQuery` values for KR account identity and US
  exchange rate. It MUST NOT represent them as one combined KR+US query.
- **FR-3:** The KR budget MUST bind market `KR`, evidence source `same-currency`, official Open API
  endpoint `GET /api/v1/accounts`, maximum freshness age five minutes, one soak request per cycle,
  and the bounded read retry policy.
- **FR-4:** The US budget MUST bind market `US`, evidence source `official-fx`, official Open API
  endpoint `GET /api/v1/exchange-rate`, maximum freshness age fifteen seconds, one soak request per
  cycle, and the bounded read retry policy.
- **FR-5:** Both runtime read budgets MUST remain bounded to three physical attempts and eight
  seconds total, with the existing capped backoff and rate-limit handling. Permanent/auth/canceled
  failures MUST NOT receive blind retries.
- **FR-6:** The official FX schema probe MUST bind only
  `https://openapi.tossinvest.com/api/v1/exchange-rate?baseCurrency=USD&quoteCurrency=KRW`, method
  GET, one request per run, and status 200 plus exact critical string fields under `result`.
- **FR-7:** `monitor.Probes()` MUST remain the WTS-session registry and MUST NOT execute the official
  OAuth probe.
- **FR-8:** A US FX read failure MUST remain a US market-local evaluation failure. It MUST NOT cancel
  a successful KR identity evaluation or fill, reconciliation, protection, exit, and emergency
  reduction loops. The symmetric KR failure rule also applies.

## Non-Functional Requirements

- **NFR-1:** Tests MUST use fakes and schema fixtures only; no real network, `tossctl monitor`, cron,
  LIVE order, toggle, or deployment action is permitted.
- **NFR-2:** Probe errors MUST NOT include response bodies or account/rate values.
- **NFR-3:** Returned budget catalogs MUST be caller-independent values and MUST expose no raw rate,
  haircut, digest, token, account reference, or authority constructor.
- **NFR-4:** Existing mutation, exit, and safety-loop retry behavior MUST remain unchanged.

## Acceptance Criteria

### AC-1: required endpoint and bounded budgets (FR-1 through FR-5)

Given the default a072 runtime definitions  
When engine endpoints and execgw read budgets are inspected  
Then the official exchange endpoint is absent from the global startup attestation set
And KR identity and US FX have distinct exact source, endpoint, freshness, retry, and soak budgets.

### AC-2: official schema probe is isolated (FR-6, FR-7, NFR-1, NFR-2)

Given fake valid and invalid exchange-rate bodies  
When the official contract probe checker runs  
Then only status 200 and the critical string fields pass  
And the WTS probe registry contains no `openapi.tossinvest.com` URL.

### AC-3: paired market failure isolation (FR-8, NFR-4)

Given a successful fake KR selected-account identity check and a failing fake US exchange read  
When both strategy workers evaluate in the same runtime  
Then US alone emits the typed market-local fault after bounded retry  
And KR continues evaluating while every safety loop stays alive.

## Edge Cases

- **EC-1:** Missing/wrong base currency, quote currency, rate, mid-rate, validity start, or validity end
  fails the official schema probe without returning body fragments.
- **EC-2:** A 401/403, permanent 4xx, canceled context, exhausted 429 budget, or transient exhaustion
  returns failure; it never mints a fresh observation.
- **EC-3:** Mutating a caller-owned returned budget map cannot change a later catalog result.
- **EC-4:** The inactive official contract probe cannot be mistaken for executed soak evidence. Its absence
  closes US strategy entry locally and cannot refuse engine startup or stop safety-class loops.

## API Contracts

```go
const (
    QueryAccountIdentity RequiredQuery = "strategy_account_identity"
    QueryExchangeRate    RequiredQuery = "strategy_exchange_rate"
)

type StrategyReadBudget struct {
    Market                  string
    EvidenceSource          string
    EndpointSource          string
    Endpoint                string
    Query                   RequiredQuery
    StaleAfter              time.Duration
    Retry                   RetryPolicy
    SoakMaxRequestsPerCycle int
}

func StrategyReadBudgets() map[RequiredQuery]StrategyReadBudget
func OfficialReadContractProbes() []monitor.Probe
```

## Data Models

| Entity | Field | Constraint |
| --- | --- | --- |
| strategy read budget | market | exact `KR` or `US`; no combined scope |
| strategy read budget | evidence source | exact `same-currency` or `official-fx` |
| strategy read budget | endpoint source | exact `official-open-api` |
| strategy read budget | endpoint | exact read-only method and path |
| strategy read budget | staleness | positive and bounded per FR-3/FR-4 |
| strategy read budget | retry | existing default read policy; no mutation policy |
| strategy read budget | soak requests | exact one logical request per cycle |
| official probe | URL | immutable official host, exact USD/KRW query |

## Out of Scope

- OAuth credential loading or a live official-source monitor runner.
- Editing `internal/official`, `internal/officialfx`, journal, broker, activation, config, or toggles.
- Executing a soak, monitor command, cron job, network request, LIVE order, or deployment.
- Adding China or any currency pair other than the current USD/KRW production profile.
