# Review: a046-approve-candidate-veto-thresholds

- Date: 2026-07-31
- Voices: Security, Test Architecture, Maintainability

## Security decision

**ACCEPTED WITH DORMANT SCOPE.** schema, pure markout, evidence report, fail-closed loader와 read-only `unapproved/passed=0` 상태를 승인한다. 숫자 threshold의 최초 승인은 별도 사람 evidence activation record다.

## Findings and decisions

1. a046이 5/15/30분 target 이후 첫 기존 관측을 +60초 tolerance로 선택하는 `internal/markout` 순수 계약을 소유하고 a049가 재사용한다.
2. 추가 quote poll은 0건이며 누락은 0수익이 아니라 `not_measured`다.
3. 기존 `near_high=2.0`은 `legacy-unapproved` provenance로 남기고 자동 승인하지 않는다.
4. threshold option은 immutable evidence digest/version을 사용하고 UI에서 숫자를 입력하지 않는다.

## Verification evidence

- OpenSpec strict validation: pass.
- Numeric evidence approval: absent by design.

## Verdict

`unapproved/passed=0` 구현과 gate를 승인한다. 숫자 activation은 구현 완료와 분리한다.
