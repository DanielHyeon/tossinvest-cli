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
- Every initial activation/lane/kill/protection/reconcile/scheduler/autostart/gate/LIVE blocker and simultaneous-failure precedence is table-tested with issuer and gateway calls fixed at zero. The positive fixture is explicitly limited to the package-private post-validation core with spies; production-positive `Dispatch` remains intentionally impossible without authentic source/activation authority.
- StockOS `tests/test_parker_vwap_pv2.py::TestEvaluateHappyPath::test_enter_long_full_pipeline` is translated as algorithm-parity evidence: entry `100.5`, stop `99.7965`, target `102.6105`, RR `2.8571428571428571428571428571`. The separate synthetic fixture is no longer described as source-golden or activation evidence.
- Frozen session cutoff is server-derived as `session close - 45m` (regular close `15:30 → 14:45`, early close included), participates in the constants digest, and is shown as a read-only runtime card. No caller/UI cutoff input exists.
- Journal v14 is additive and transition-guarded: receipt/authority columns are immutable, only sanctioned state+revision CAS is accepted, the complete canonical 60-field decision payload/identity is bound to denormalized lineage, and delayed execution-link replay preserves the first timestamp. v13→v14 rollback/backup, SIGKILL recovery, restart durability and legacy `RiskIntent` canonical-byte compatibility are green.
- `/strategy-runtime` is authenticated GET/HEAD-only and contains no form, input, textarea, select, button, contenteditable, arbitrary symbol/reason or combined enable action. The StockOS-style cards expose fixed server provenance/defaults without asking the operator to type values.
- Verification passed: full `go test ./...`; uncached full journal; focused race for strategyengine/strategydispatch/console and strategy/migration journal paths; `go vet ./...`; Windows amd64 CGO-free build; strict OpenSpec; semantic Function/Branch maps; `git diff --check`.

## Independent review findings resolved

The first exact-SHA review found six integration blockers: arbitrary cutoff laundering, missing dispatch blocker evidence, semantically inaccurate branch maps, mutable attempt authority columns, partial decision-payload binding, and timestamp-sensitive execution-link replay. Each now has a structural fix plus RED→GREEN coverage as recorded above. Test-only authority minters were not added.

The dormant build still intentionally leaves these activation-only blockers for a separately reviewed wiring change: semantic provisioning of the manifest, exact final-call manifest/attestation expiry recheck, a sealed single-journal/account Guardian+gateway bundle, structural restriction to the concrete official gateway, removal/sealing of direct Guardian strategy issuance, stronger indirect-import/composite-literal guards, and an authoritative calendar/config adapter. None is reachable in a047 because source proof, manifest installation, protection wiring, runtime entry loop, and production dispatch construction remain absent.

### Dormant handoff

Production can construct only an empty activation repository. A later separately reviewed change must provide the authentic sorted StockOS source manifest, signed activation provisioning and exact a045/a046/a048/a050 bindings before any runtime loop is considered. This change provides no LIVE approval, paper/shadow/canary order path, operational toggle or command wiring; existing exit/reconcile/protection supervision is untouched.

## Verdict

첫 리뷰 blocker 수정 후 exact-SHA 재검토와 gate를 요구한다. a045/a046 evidence 전 exposure-raising runtime implementation/activation 완료를 주장하지 않는다.
