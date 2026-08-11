## 1. Contract and pre-edit logic maps

- [x] 1.1 Capture the implementation base, refresh CodeGraph evidence, and map the current `ProfileProtection=UNWIRED` assembly, attestation loader, protection controller, official gateway and engine interlock call paths without changing runtime state
- [x] 1.2 Create pre-edit Function Logic Maps and Branch Test Maps for `protection.AssessReadiness`, `Controller.Register`, `Controller.Replace`, `Controller.Reconcile`, `Controller.Recover`, `execgw.Gateway.checkProtection`, engine assembly, and every other existing function that implementation will change
- [x] 1.3 Freeze a branch-to-scenario matrix for KR/US independent readiness, signed scope validation, expiry/build drift, restart recovery, duplicate events, replace coverage, oversell prevention and OFF-state safety-loop continuity
- [x] 1.4 Record the new isolated-core Function Logic and Branch Test Map in `analysis/function-logic/isolated-core.md`; no existing runtime function is edited in this wave

## 2. RED attestation and protection lifecycle tests

- [x] 2.1 Add RED strict-attestation tests for pinned trust root, key ID, signature algorithm allowlist, revocation, bounded rotation overlap, monotonic serial anti-rollback, maximum lifetime, trusted-time unavailability/rollback, schema version, account/profile, market, order type, session, quantity range, trigger source, replace semantics, tool/build/evidence digests, issued/expiry times, unknown fields and file ownership/permission failures
- [x] 2.2 Add RED market-isolation tests proving KR-only evidence cannot authorize US, US-only evidence remains valid when KR expires, readiness alone never enables a lane/autostart/automation/LIVE approval, and missing evidence keeps both markets `UNWIRED`
- [x] 2.3 Add RED capability tests requiring exact broker client-key echo, lookup fields, identity uniqueness scope, pending/terminal and cancel-result query semantics, and idempotency/dedup behavior; prove any absent capability yields exactly `UNWIRED` plus a typed refusal
- [x] 2.4 Add RED controller tests for submit/cancel unknown outcomes, register-response crash, duplicate fill, stable generation/revision operation identity, exact broker-ID recovery, orphan discovery, atomic/continuous replacement refusal, non-retreat trigger and sell-claim oversubscription; prove unattested idempotency never permits resubmission and unowned orphans are never guessed/canceled
- [x] 2.5 Add RED engine/Gateway tests proving attestation drift at dispatch blocks only exposure raising, existing broker protection and reduce-only exit continue, and reconciliation/recovery never infer identity from symbol/time
- [x] 2.6 Add RED transport guards for the isolated readiness core proving it has no live-host transport dependency and cannot change a toggle, lane, activation or approval

## 3. Protection readiness implementation

- [x] 3.1 Implement the strict signed protection-attestation schema and canonical verifier with pinned trust roots, key/algorithm/revocation/rotation policy, durable monotonic serial and trusted-time floor, maximum lifetime, typed per-market refusal, current build/evidence binding and fail-closed file validation
- [x] 3.2 Replace the production-global readiness constant at the decision boundary with immutable KR/US readiness snapshots derived from exact attestation scope plus actual supervisor wiring, while preserving the shipped `UNWIRED` default
- [x] 3.3 Implement exact broker identity/query/dedup capability parsing and durable generation/revision operation identity plus desired/observed/unknown broker state needed for registration, safer replacement, cancellation, restart recovery and reconciliation
- [x] 3.4 Enforce attested replace semantics, continuous coverage, broker-reserved/local sell-claim bounds and entry-latch closure without weakening current ACTIVE protection
- [ ] 3.5 **SUPERSEDED by `a100-wire-fill-to-broker-protection` (2026-08-11).** Not delivered here and not counted against this change. See `## 6. Supersession` below. Original text: Wire the official protection gateway and supervisor into production engine assembly without exposing a second mutation path or changing lane, autostart, automation gate or LIVE approval settings
- [x] 3.6 Implement submit/cancel unknown and orphan reconciliation so resubmission occurs only under attested broker idempotency, otherwise remains no-resubmit/entry-latched until exact reconciliation or human resolution

## 4. Isolated integration and failure recovery

- [x] 4.1 Build an isolated KR/US official-broker integration fixture that exercises signed readiness, partial fill, registration, quantity convergence, safer replacement, cancellation and exact broker-state reconciliation with live hostname mutation structurally blocked
- [x] 4.2 Prove process loss after broker acceptance but before local commit recovers the same protection once when exact query/dedup is attested, and otherwise enters no-resubmit reconciliation without duplicating a conditional order while new exposure stays blocked
- [x] 4.3 Prove KR attestation/recovery failure does not lower valid US readiness and vice versa, while protection, exit, reconciliation and fill handling continue in both markets
- [x] 4.4 Run restart/idempotence/crash-point tests and the official-only/WTS-isolation matrix; confirm no fixture flips a toggle, creates approval or sends a live order

## 5. VERIFY and review gates

- [ ] 5.1 Refresh post-edit AST, Function Logic Maps and Branch Test Maps for every changed existing function and pass the repository analysis checker
- [ ] 5.2 Run targeted protection/attestation/execgw/engine tests, race tests for affected packages, journal crash/restart suites, full tests and vet, and strict OpenSpec/PM validation
- [ ] 5.3 Run `make sdd-sync`, `make sdd-check`, and `make gate CHANGE=a071-wire-kr-us-protection-readiness`, then complete adversarial independent review before marking the high-risk change complete
- [x] 5.4 Verify the built default remains lane/autostart/automation/LIVE OFF or unapproved, missing attestation remains `UNWIRED`, and protection/exit/reconciliation/fill paths remain available without any live broker mutation
- [x] 5.5 Run isolated-core unit, race, vet, fuzz, coverage, static dependency and strict OpenSpec validation; preserve the production integration gates above as pending

## 6. Supersession — task 3.5 → a100

2026-08-11. Task 3.5는 `a100-wire-fill-to-broker-protection`이 승계한다. 이 change에서는
전달하지 않으며 완료 조건에서 제외한다.

**왜 여기서 끝내지 못했나.** 이 change의 review.md가 스스로 사유를 적어 뒀다 — "the repository
has no production caller from a journal-committed fill into an exact journal-derived stop/expiry
and durable protection Plan/Register lifecycle". 즉 3.5의 선행 조건은 이 change가 만든 것이
아니라 **이 change 범위 밖에 있는 별도 작업**이다. status.md의 Remaining도 같은 말을 한다.

**왜 이 change를 재개하지 않고 새 change로 가나.**

1. 3.5의 실제 작업은 "assembly에 연결"이 아니다. 아래 §측정 근거대로 배선 대상 함수 13개 분기
   중 9개가 true 결과를 한 번도 실행한 적이 없다. 선행 작업은 배선이 아니라 **거부 경로
   RED 테스트**이며, 이는 a071의 나머지 24개 task와 성격이 다른 별개의 수직 작업이다.
2. a071은 attestation 서명·키 수명·trusted-time floor·monotonic serial까지 포함하는 큰
   change다. 3.5 하나를 위해 재개하면 독립 리뷰가 이미 승인된 21개 task 전체를 다시 범위로
   잡는다. 잘라내는 편이 리뷰 단위를 정확하게 만든다.
3. 이 change가 만든 계약(market-scoped verdict, paired snapshot, sealed supervisor binding,
   Gateway decision boundary)은 **그대로 유효하고 a100의 입력이다.** a100은 그 계약을 다시
   설계하지 않고 소비한다.

**a100이 반드시 지켜야 하는 이 change의 계약** (재설계 금지):

- readiness는 market-scoped `WIRED|UNWIRED` + typed refusal이며 결합 상태를 만들지 않는다
- `WIRED`는 signed attestation과 sealed supervisor binding이 **둘 다** 검증될 때만 생긴다
- reduce-only SELL/CANCEL/축소 AMEND와 stop·긴급청산·대사·체결 경로는 readiness provider를
  읽지 않는다 (정적 격리 테스트가 강제)
- 공개 scalar override나 exported readiness 필드로 entry를 승인할 수 없다
- `internal/protection`은 `net/http`·`internal/official`·`internal/trading`을 import하지 않고,
  app 코드 중 `internal/app/engine/gateway.go`만 이를 import한다

**남은 게이트 task 5.1~5.3의 처리.** 3.5가 빠졌으므로 이 change의 변경 집합은 3.4까지로
확정된다. 5.1~5.3은 그 확정된 집합에 대해 실행한다. a100의 변경분은 a100의 게이트가 본다.

**되돌리는 조건.** a100이 취소되거나 범위에서 프로덕션 배선이 빠지면 이 supersession은
무효이며 3.5는 이 change로 돌아온다. 그때는 이 절과 a100 proposal의 대응 절을 함께 지운다.

### 측정 근거 (a100 `analysis/function-logic/`, 2026-08-11)

| 대상 | 분기 | true 결과 실행됨 | 미실행 |
| --- | ---: | ---: | ---: |
| `execgw.Gateway.checkProtection` | 5 | 5 | 0 |
| `engine.runInterlock` | 3 | 2 | 1 |
| `engine.productionProtectionAssemblies` | 0 | — | — |
| `protectionlifecycle.applyFill` | 7 | 2 | 5 |
| `protectionlifecycle.prepareRegister` | 6 | 2 | 4 |
