# KR/US production FX engine loader — pre-edit specification

## Scope

Wire the existing `officialfx.ProductionAuthorityService` into the production strategy assembly for
KR identity and US quote-to-account-base conversion in the same implementation wave. This slice
reads official authority and the already-specified signed US haircut policy; it does not size an
order, reserve a bucket, issue a first leg, write an activation, or call a broker mutation.

## Trust configuration

The engine derives account ID from its already resolved official account and account base currency
from the verified automation-gate `limit_currency`. US policy trust is pinned by:

- `TOSSOS_FX_RISK_POLICY_MANIFEST_SHA256`
- `TOSSOS_FX_RISK_POLICY_KEY_ID`
- `TOSSOS_FX_RISK_POLICY_PUBLIC_KEY_BASE64`

The digest and strict base64 Ed25519 public key are passed to
`officialfx.NewProductionAuthorityService`; no private/signing key, multiplier, rate, evidence
digest, freshness value or activation writer is accepted from the engine loader.

## Paired contract

1. Reuse the scheduler/candidate frozen UTC observation. `ProductionAuthorityConfig.Now` always
   returns that exact value.
2. Start KR and US collection independently. KR re-verifies the account and mints only same-currency
   identity evidence. US verifies the owner-only signed, monotonic haircut policy and reads only the
   official quote-to-base endpoint.
3. A schedule/candidate-not-ready market performs no FX/account read. Failure or panic in one market
   becomes only that market's typed refusal and never cancels the peer.
4. Validate each opaque evidence at the frozen instant and exact pair (`KRW→base` for KR,
   `USD→base` for US) before marking it ready.
5. Keep opaque `officialfx.Evidence` private. The command boundary receives only market, ready,
   typed reason, quote/base currency and digest observations.
6. Do not promote either worker until risk buckets, lane inputs, first-leg admission and Gateway
   cycle assembly are complete.

## Required paired RED scenarios

- KR and US begin in the same wave and both preserve the one frozen timestamp.
- US official/policy failure leaves KR identity ready; KR account re-verification failure leaves US
  ready.
- Schedule/candidate OFF for either market makes zero collector calls for that market.
- A market collector panic is contained locally.
- Invalid/missing US trust pins refuse US without weakening or synthesizing policy; KR behavior is
  independent of those US-only pins.
- Public snapshots and command assembly expose no `officialfx.Evidence`, collector, client, Gateway,
  journal, order or writer.

## Completion boundary

Completion is one KR/US test checkpoint. It is not operational activation and grants no LIVE order
approval.
