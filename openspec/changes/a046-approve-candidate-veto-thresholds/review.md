# Review: a046-approve-candidate-veto-thresholds

- Date: 2026-07-31
- Voices: Security, Test Architecture, Maintainability

## Security decision

**CLEAN FOR DORMANT INTEGRATION.** activation/evidence binding, veto-order mutation,
UI subtree, markout isolation과 pure approved-candidate boundary를 독립 재리뷰했다.
수치 threshold의 최초 승인은 계속 별도 사람 evidence activation record다.

## Findings and decisions

1. a046이 5/15/30분 target 이후 첫 기존 관측을 +60초 tolerance로 선택하는 `internal/markout` 순수 계약을 소유하고 a049가 재사용한다.
2. 추가 quote poll은 0건이며 누락은 0수익이 아니라 `not_measured`다.
3. 기존 `near_high=2.0`은 `legacy-unapproved` provenance로 남기고 자동 승인하지 않는다.
4. threshold option은 immutable evidence digest/version을 사용하고 UI에서 숫자를 입력하지 않는다.
5. follow-up loader는 opaque evidence bytes의 SHA-256을 재계산하고 strict 별도
   `ActivationRecord`의 version/scope/canonical set digest/evidence digest/approval time을 결합한다.
6. 같은 version의 다른 canonical set digest는 동시성 안전 registry가 거부한다.
7. D3 순서는 private array와 copy accessor로 고정하며 외부 source/compile guard가 mutation 권한 부재를 검증한다.
8. candidate-filters 검사는 DOM 전체 subtree를 순회하고 KR/US 각각 정확히 세 metric을 검증한다.
9. markout isolation 검사는 production import allowlist와 clock/polling positive control을 가진다.

## Verification evidence

- OpenSpec strict validation: pass.
- Numeric evidence approval: absent by design.
- Shadow observation report: `evidence/shadow-observation-report.md` — both KR/US
  regular-session evidence are explicitly `not_measured`; the a049 golden fixture
  is contract evidence only and cannot activate a threshold.

## Verdict

follow-up 구현 후에도 독립 재리뷰와 최종 gate 전에는 승인하지 않는다. `unapproved/passed=0`과
input-free UI는 유지하며, synthetic activation fixture는 numeric human approval이 아니다.

## Independent re-review · 2026-08-01

- Reviewed scope: `178f583..5da9a61`
- Verdict: **CLEAN FOR DORMANT INTEGRATION**
- Previous findings closed: Critical **2/2**, Warning **2/2**
- New findings: **0**

The re-review confirmed that opaque evidence bytes are hashed again at load time
and bound to a strict separate activation record through version, market/session
scope, canonical set digest, evidence digest, and bounded approval time. The
same-version/different-digest registry conflict is fail-closed.

The D3 veto order is now a private array exposed only through a copy accessor.
The `candidate-filters` DOM subtree has no form, textarea, input, or
`contenteditable` surface and contains exactly the KR/US × three-metric matrix.
The markout package remains restricted by a production import allowlist and
positive-control clock/polling detection.

Numeric activation remains absent. Runtime and UI stay
`unapproved / passed=0 / verdict inactive`, and the reviewed dependency guards
show no order or RiskIntent authority. Focused tests, affected-package race
tests, and vet passed for the reviewed scope.

## Independent boundary re-review · 2026-08-01

- Reviewed scope: `30a61ce..fc13b6634866cdff7bc40bc49553792895cc96d5`
- Verdict: **CLEAN**
- New findings: **0**

The final adversarial review verified the `go/types` cycle-safe allowlist over
aliases, named underlying types, embedded fields, fixed arrays and generic type
arguments. Only exact `candidate.ApprovedCandidate` plus immutable scalar/value
composites cross the pure boundary. Pointer, map, slice, channel, signature,
interface (`any`/`error`), unsafe, type-parameter and `candidate.Source`
capabilities fail closed.

Method values/expressions, type assertions, free or injected calls, package
variables, mutation, `go`/`defer`, and external imports are rejected. Reverse
import-graph taint reaches every registered authority root. Focused adversarial
tests, affected-package race tests, full tests, vet, strict OpenSpec, SDD check,
format and whitespace checks all passed on the exact reviewed commit.

## Dependency-integrated rebase verification · 2026-08-01

- Integration base: `70aabdc9936de08df458da13203437ba7d2dd572` (a041/a042 complete)
- Reapplied source range: `178f583..2494cf0`; commits were replayed without conflicts.
- Patch scope remains the independently reviewed a046 implementation; the base marker now
  excludes already-integrated a041/a042 logic from this change's Function Logic Map audit.
- Focused candidate/markout/console/CLI tests, focused vet, strict OpenSpec validation,
  whitespace validation and index synchronization passed on the dependency-integrated tree.
