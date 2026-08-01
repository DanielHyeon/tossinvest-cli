# Review: a047-add-strategy-engine

- Date: 2026-07-31
- Voices: Security, Test Architecture, Maintainability

## Security decision

pure lane/orchestrator와 immutable activation-manifest 기반 default-OFF 구현만 승인한다. exact source는 StockOS commit `d75113d3`, source-set digest `09260ac…`, KRX `parker_vwap_trend_v1` conservative profile로 고정했다. a045/a046 선행 gate 없이는 exposure-raising dispatch를 연결하지 않는다.

## Findings and decisions

1. manifest는 account/build/lane/threshold/settings/attestation/Guardian/reconciliation/protection/policy/scheduler와 개별 human approvals를 하나의 digest로 묶는다.
2. dispatch 직전에 모든 field를 재검증하며 mismatch/expiry/kill/reconcile degradation/high-risk config change는 effective entry를 OFF로 만든다. exit는 유지한다.
3. identity와 lineage는 candidate life, lane, threshold, settings, manifest digest를 끝까지 전달한다.
4. lane package는 broker/journal/config를 import하지 않는 pure domain이다.

## Verification evidence

- OpenSpec strict validation: pass.
- Source commit/digest/constants contract is frozen, but the sorted path/blob manifest that reproduces `09260ac…` is unavailable. The verifier therefore returns `not_configured`; protection activation evidence also remains absent and effective entry is OFF.

## 2026-08-01 dormant implementation evidence

- Pre-Edit Gate: journal/Guardian/gateway/runtime baselines and callers were mapped before edits. Released Guardian, gateway, engine run/submit and `RiskIntent.Canonical()` production functions were not changed.
- ApprovedCandidate crosses only `internal/strategy`'s value-only, type-checked boundary. The Parker evaluator has no broker, journal, config, clock or callback dependency.
- Activation comparison covers all 32 binding fields, expiry/attestation expiry, revocation, generation and decision-to-dispatch revalidation. No exported installer/minter or runtime wiring exists.
- Blocker paths call Guardian, durable-attempt and gateway zero times. The only successful orchestration fixture uses typed test spies and checks durable plan before the official-only gateway seam plus duplicate identity refusal.
- Journal v14 is additive and immutable. v13→v14 rollback/backup, SIGKILL recovery, restart durability and legacy `RiskIntent` canonical-byte compatibility are green.
- `/strategy-runtime` is authenticated GET/HEAD-only and contains no form, input, textarea, select, button, contenteditable, arbitrary symbol/reason or combined enable action.
- Full `go test ./...`, focused race (strategy/market/official/console and exact journal v14), `go vet ./...`, Windows amd64 build, strict OpenSpec, PM, logic-map and `make validate` passed.

### Dormant handoff

Production can construct only an empty activation repository. A later separately reviewed change must provide the authentic sorted StockOS source manifest, signed activation provisioning and exact a045/a046/a048/a050 bindings before any runtime loop is considered. This change provides no LIVE approval, paper/shadow/canary order path, operational toggle or command wiring; existing exit/reconcile/protection supervision is untouched.

## Verdict

구조, frozen pure lane와 dormant/default-OFF scaffold는 승인한다. a045/a046 evidence 전 exposure-raising runtime implementation/activation 완료를 주장하지 않는다.
